package cornjob

import (
	"context"
	"fmt"
	"sync"
	"time"

	mongodb "github.com/HUAHUAI23/simple-waf/pkg/database/mongo"
	"github.com/HUAHUAI23/simple-waf/server/model"
	"github.com/HUAHUAI23/simple-waf/server/service/daemon"
	"github.com/haproxytech/client-native/v6/models"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// HAProxyStatsData 结构用于记录最后一次的统计数据，用于计算差值
type HAProxyStatsData struct {
	TargetName string
	LastStats  *models.NativeStatStats
	LastTime   time.Time
	ResetCount int // 用于跟踪重启次数
}

// StatsAggregator HAProxy统计数据聚合器
type StatsAggregator struct {
	runner          daemon.ServiceRunner
	dbName          string
	lastStats       map[string]*HAProxyStatsData // 用backend名称做key
	lastStatsLoaded bool                         // 标记是否已从数据库加载基准数据
	TargetFilter    map[string]bool              // 过滤的backend列表
	mu              sync.RWMutex
	log             zerolog.Logger
	stopCh          chan struct{}
}

// NewStatsAggregator 创建新的数据聚合器
func NewStatsAggregator(runner daemon.ServiceRunner, dbName string, TargetList []string, logger zerolog.Logger) (*StatsAggregator, error) {
	// 初始化TargetFilter
	TargetFilter := make(map[string]bool)
	for _, target := range TargetList {
		TargetFilter[target] = true
	}

	agg := &StatsAggregator{
		runner:          runner,
		dbName:          dbName,
		lastStats:       make(map[string]*HAProxyStatsData),
		lastStatsLoaded: false,
		TargetFilter:    TargetFilter,
		log:             logger.With().Str("component", "haproxy_stats_aggregator").Logger(),
		stopCh:          make(chan struct{}),
	}

	// 确保必要的集合和索引存在
	err := agg.ensureCollections()
	if err != nil {
		return nil, err
	}

	return agg, nil
}

// UpdateTargetList 更新监控的后端列表
func (a *StatsAggregator) UpdateTargetList(targetList []string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// 创建新的过滤器
	newFilter := make(map[string]bool)
	for _, target := range targetList {
		newFilter[target] = true
	}

	// 查找不再监控的目标
	for target := range a.TargetFilter {
		if !newFilter[target] {
			a.log.Info().Str("target", target).Msg("Removing target from monitoring")
			// 从内存中移除，但保留数据库中的基准线(以防后续重新添加)
			delete(a.lastStats, target)
		}
	}

	// 记录新增的目标
	for target := range newFilter {
		if !a.TargetFilter[target] {
			a.log.Info().Str("target", target).Msg("Adding new target to monitoring")
			// 新增的目标会在下次数据采集时自动初始化
		}
	}

	// 更新过滤器
	a.TargetFilter = newFilter
}

// ensureCollections 确保必要的集合和索引存在
func (a *StatsAggregator) ensureCollections() error {
	db, err := mongodb.GetDatabase(a.dbName)
	if err != nil {
		return err
	}

	ctx := context.Background()

	var haproxyStatsBaseline model.HAProxyStatsBaseline
	// 确保基准数据集合索引
	baselineColl := db.Collection(haproxyStatsBaseline.GetCollectionName())
	_, err = baselineColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "target_name", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return fmt.Errorf("failed to create index for baseline collection: %v", err)
	}

	var haproxyMinuteStats model.HAProxyMinuteStats
	// 确保分钟统计数据集合索引
	minuteStatsColl := db.Collection(haproxyMinuteStats.GetCollectionName())
	_, err = minuteStatsColl.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "target_name", Value: 1},
				{Key: "date", Value: 1},
				{Key: "hour", Value: 1},
				{Key: "minute", Value: 1},
			},
		},
		{
			Keys: bson.D{{Key: "timestamp", Value: 1}},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create indexes for minute stats collection: %v", err)
	}

	// 确保实时统计数据时间序列集合
	// 确保实时统计数据时间序列集合
	realtimeMetrics := []string{"conn_rate", "scur", "rate", "req_rate"}
	for _, metric := range realtimeMetrics {
		// 检查集合是否存在
		names, err := db.ListCollectionNames(ctx, bson.M{"name": metric})
		if err != nil {
			return err
		}

		if len(names) == 0 {
			// 创建时间序列集合
			timeSeriesOpts := options.TimeSeries().
				SetTimeField("timestamp").
				SetMetaField("metadata").
				SetGranularity("seconds")

			// 在MongoDB v2驱动中，过期时间设置在CreateCollection选项中，而不是TimeSeriesOptions中
			createOpts := options.CreateCollection().
				SetTimeSeriesOptions(timeSeriesOpts).
				SetExpireAfterSeconds(3600) // 设置1小时过期

			if err := db.CreateCollection(ctx, metric, createOpts); err != nil {
				return fmt.Errorf("failed to create timeseries collection %s: %v", metric, err)
			}
		}
	}

	return nil
}

// Start 启动聚合器
func (a *StatsAggregator) Start(ctx context.Context) error {
	// 尝试从数据库加载基准数据
	if err := a.loadLastStats(ctx); err != nil {
		a.log.Warn().Err(err).Msg("Failed to load lastStats, will initialize new baseline")
		// 如果加载失败，初始化新基准
		a.initializeStats(ctx)
	} else {
		a.log.Info().Msg("Successfully loaded lastStats data from database")
		a.lastStatsLoaded = true
		// 获取最新数据并检查HAProxy是否重启
		a.checkHAProxyResetOnStartup(ctx)
	}

	return nil
}

// Stop 停止聚合器
func (a *StatsAggregator) Stop(ctx context.Context) error {
	// 保存最后的状态
	if err := a.saveLastStats(ctx); err != nil {
		a.log.Warn().Err(err).Msg("Failed to save lastStats during shutdown")
	}

	close(a.stopCh)
	return nil
}

// loadLastStats 从MongoDB加载lastStats基准数据
func (a *StatsAggregator) loadLastStats(ctx context.Context) error {
	db, err := mongodb.GetDatabase(a.dbName)
	if err != nil {
		return err
	}

	var haproxyStatsBaseline model.HAProxyStatsBaseline
	collection := db.Collection(haproxyStatsBaseline.GetCollectionName())
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("failed to query lastStats: %v", err)
	}
	defer cursor.Close(ctx)

	var baselines []bson.M
	if err = cursor.All(ctx, &baselines); err != nil {
		return fmt.Errorf("failed to decode lastStats documents: %v", err)
	}

	if len(baselines) == 0 {
		return fmt.Errorf("no baseline data found")
	}

	loadCount := 0
	for _, doc := range baselines {
		targetName, ok := doc["target_name"].(string)
		if !ok {
			a.log.Warn().Interface("doc", doc).Msg("Invalid target_name in baseline document")
			continue
		}

		// 加载所有目标的数据，即使当前不在监控列表中
		// 这样如果目标列表变化，我们已经有数据了

		// 重建NativeStatStats结构
		stats := &models.NativeStatStats{}
		statsMap, ok := doc["stats"].(map[string]interface{})
		if !ok {
			a.log.Warn().Str("target", targetName).Msg("Invalid stats map for target")
			continue
		}

		// 设置关键计数器字段
		if binVal, ok := statsMap["bin"].(int64); ok {
			stats.Bin = &binVal
		}

		if boutVal, ok := statsMap["bout"].(int64); ok {
			stats.Bout = &boutVal
		}

		if connTotVal, ok := statsMap["conn_tot"].(int64); ok {
			stats.ConnTot = &connTotVal
		}

		if stotVal, ok := statsMap["stot"].(int64); ok {
			stats.Stot = &stotVal
		}

		if reqTotVal, ok := statsMap["req_tot"].(int64); ok {
			stats.ReqTot = &reqTotVal
		}

		// HTTP响应状态码
		if hrsp1xxVal, ok := statsMap["hrsp_1xx"].(int64); ok {
			stats.Hrsp1xx = &hrsp1xxVal
		}

		if hrsp2xxVal, ok := statsMap["hrsp_2xx"].(int64); ok {
			stats.Hrsp2xx = &hrsp2xxVal
		}

		if hrsp3xxVal, ok := statsMap["hrsp_3xx"].(int64); ok {
			stats.Hrsp3xx = &hrsp3xxVal
		}

		if hrsp4xxVal, ok := statsMap["hrsp_4xx"].(int64); ok {
			stats.Hrsp4xx = &hrsp4xxVal
		}

		if hrsp5xxVal, ok := statsMap["hrsp_5xx"].(int64); ok {
			stats.Hrsp5xx = &hrsp5xxVal
		}

		if hrspOtherVal, ok := statsMap["hrsp_other"].(int64); ok {
			stats.HrspOther = &hrspOtherVal
		}

		// 错误相关
		if dreqVal, ok := statsMap["dreq"].(int64); ok {
			stats.Dreq = &dreqVal
		}

		if drespVal, ok := statsMap["dresp"].(int64); ok {
			stats.Dresp = &drespVal
		}

		if ereqVal, ok := statsMap["ereq"].(int64); ok {
			stats.Ereq = &ereqVal
		}

		if dconVal, ok := statsMap["dcon"].(int64); ok {
			stats.Dcon = &dconVal
		}

		if dsesVal, ok := statsMap["dses"].(int64); ok {
			stats.Dses = &dsesVal
		}

		if econVal, ok := statsMap["econ"].(int64); ok {
			stats.Econ = &econVal
		}

		if erespVal, ok := statsMap["eresp"].(int64); ok {
			stats.Eresp = &erespVal
		}

		timestamp, ok := doc["timestamp"].(time.Time)
		if !ok {
			timestamp = time.Now()
		}

		resetCount := 0
		if resetCountVal, ok := doc["reset_count"].(int32); ok {
			resetCount = int(resetCountVal)
		}

		// 加载到内存中
		a.mu.Lock()
		a.lastStats[targetName] = &HAProxyStatsData{
			TargetName: targetName,
			LastStats:  stats,
			LastTime:   timestamp,
			ResetCount: resetCount,
		}
		a.mu.Unlock()

		loadCount++
	}

	a.log.Info().Int("count", loadCount).Msg("Loaded target baselines from database")
	return nil
}

// saveLastStats 将lastStats基准数据保存到MongoDB
func (a *StatsAggregator) saveLastStats(ctx context.Context) error {
	db, err := mongodb.GetDatabase(a.dbName)
	if err != nil {
		return err
	}

	var haproxyStatsBaseline model.HAProxyStatsBaseline
	collection := db.Collection(haproxyStatsBaseline.GetCollectionName())

	a.mu.RLock()
	defer a.mu.RUnlock()

	for targetName, stats := range a.lastStats {
		if stats.LastStats == nil {
			continue // 跳过无效数据
		}

		// 提取统计数据
		statsMap := make(map[string]int64)

		// 流量相关
		if stats.LastStats.Bin != nil {
			statsMap["bin"] = *stats.LastStats.Bin
		}

		if stats.LastStats.Bout != nil {
			statsMap["bout"] = *stats.LastStats.Bout
		}

		// HTTP响应状态码
		if stats.LastStats.Hrsp1xx != nil {
			statsMap["hrsp_1xx"] = *stats.LastStats.Hrsp1xx
		}

		if stats.LastStats.Hrsp2xx != nil {
			statsMap["hrsp_2xx"] = *stats.LastStats.Hrsp2xx
		}

		if stats.LastStats.Hrsp3xx != nil {
			statsMap["hrsp_3xx"] = *stats.LastStats.Hrsp3xx
		}

		if stats.LastStats.Hrsp4xx != nil {
			statsMap["hrsp_4xx"] = *stats.LastStats.Hrsp4xx
		}

		if stats.LastStats.Hrsp5xx != nil {
			statsMap["hrsp_5xx"] = *stats.LastStats.Hrsp5xx
		}

		if stats.LastStats.HrspOther != nil {
			statsMap["hrsp_other"] = *stats.LastStats.HrspOther
		}

		// 错误相关
		if stats.LastStats.Dreq != nil {
			statsMap["dreq"] = *stats.LastStats.Dreq
		}

		if stats.LastStats.Dresp != nil {
			statsMap["dresp"] = *stats.LastStats.Dresp
		}

		if stats.LastStats.Ereq != nil {
			statsMap["ereq"] = *stats.LastStats.Ereq
		}

		if stats.LastStats.Dcon != nil {
			statsMap["dcon"] = *stats.LastStats.Dcon
		}

		if stats.LastStats.Dses != nil {
			statsMap["dses"] = *stats.LastStats.Dses
		}

		if stats.LastStats.Econ != nil {
			statsMap["econ"] = *stats.LastStats.Econ
		}

		if stats.LastStats.Eresp != nil {
			statsMap["eresp"] = *stats.LastStats.Eresp
		}

		// 总计值
		if stats.LastStats.ConnTot != nil {
			statsMap["conn_tot"] = *stats.LastStats.ConnTot
		}

		if stats.LastStats.Stot != nil {
			statsMap["stot"] = *stats.LastStats.Stot
		}

		if stats.LastStats.ReqTot != nil {
			statsMap["req_tot"] = *stats.LastStats.ReqTot
		}

		// 创建更新文档
		doc := bson.M{
			"target_name": targetName,
			"stats":       statsMap,
			"timestamp":   stats.LastTime,
			"reset_count": stats.ResetCount,
		}

		// 使用upsert操作保存 - 采用v2驱动的builder模式
		filter := bson.M{"target_name": targetName}
		update := bson.M{"$set": doc}

		// 在v2版本中，使用options.UpdateOne()方法构建选项
		opts := options.UpdateOne().SetUpsert(true)
		_, err := collection.UpdateOne(ctx, filter, update, opts)

		if err != nil {
			return fmt.Errorf("failed to save lastStats for %s: %v", targetName, err)
		}
	}

	return nil
}

// checkHAProxyResetOnStartup 启动时检查HAProxy是否重启
func (a *StatsAggregator) checkHAProxyResetOnStartup(ctx context.Context) {
	stats, err := a.runner.GetStats()
	if err != nil {
		a.log.Error().Err(err).Msg("Failed to get stats during startup check")
		return
	}

	resetDetected := false
	newTargets := make(map[string]bool)

	a.mu.Lock()
	defer a.mu.Unlock()

	// 找出所有当前活跃的后端
	for _, stat := range stats.Stats {
		if stat.Type == "frontend" {
			newTargets[stat.Name] = true
		}
	}

	// 检查当前监控的后端
	for _, stat := range stats.Stats {
		if stat.Type != "frontend" || !a.TargetFilter[stat.Name] || stat.Stats == nil {
			continue
		}

		lastStat, exists := a.lastStats[stat.Name]
		if !exists {
			// 新的后端，初始化
			a.lastStats[stat.Name] = &HAProxyStatsData{
				TargetName: stat.Name,
				LastStats:  stat.Stats,
				LastTime:   time.Now(),
				ResetCount: 0,
			}
			a.log.Info().Str("target", stat.Name).Msg("New target detected during startup")
			continue
		}

		// 检测是否重启
		if a.detectReset(lastStat.LastStats, stat.Stats) {
			a.log.Warn().
				Str("target", stat.Name).
				Int("reset_count", lastStat.ResetCount+1).
				Msg("Detected HAProxy reset for target during startup")

			resetDetected = true

			// 保存零增量记录
			a.mu.Unlock() // 临时释放锁以防死锁
			err := a.saveMinuteMetrics(ctx, stat.Name, a.createZeroMetrics(), time.Now())
			a.mu.Lock() // 重新获取锁

			if err != nil {
				a.log.Error().
					Err(err).
					Str("target", stat.Name).
					Msg("Failed to save zero metrics for target after reset")
			}

			// 更新重启计数和基准
			lastStat.ResetCount++
			lastStat.LastStats = stat.Stats
			lastStat.LastTime = time.Now()
		}
	}

	// 检查不再存在的后端
	for targetName := range a.lastStats {
		if a.TargetFilter[targetName] && !newTargets[targetName] {
			a.log.Warn().
				Str("target", targetName).
				Msg("Target is in monitoring list but not found in HAProxy stats")
		}
	}

	if resetDetected {
		a.log.Warn().Msg("HAProxy reset detected during startup, metrics will be adjusted")
		// 保存重置后的状态
		a.mu.Unlock() // 临时释放锁以防死锁
		if err := a.saveLastStats(ctx); err != nil {
			a.log.Error().Err(err).Msg("Failed to save lastStats after reset detection")
		}
		a.mu.Lock() // 重新获取锁
	} else {
		a.log.Info().Msg("No HAProxy reset detected during startup")
	}
}

// initializeStats 初始化统计数据
func (a *StatsAggregator) initializeStats(ctx context.Context) {
	stats, err := a.runner.GetStats()
	if err != nil {
		a.log.Error().Err(err).Msg("Failed to initialize stats")
		return
	}

	now := time.Now()
	a.mu.Lock()
	defer a.mu.Unlock()

	a.lastStats = make(map[string]*HAProxyStatsData) // 清除现有数据

	for _, stat := range stats.Stats {
		if stat.Type == "frontend" && a.TargetFilter[stat.Name] && stat.Stats != nil {
			a.lastStats[stat.Name] = &HAProxyStatsData{
				TargetName: stat.Name,
				LastStats:  stat.Stats,
				LastTime:   now,
				ResetCount: 0,
			}
		}
	}

	// 保存初始基准到数据库
	a.mu.Unlock() // 临时释放锁以防死锁
	if err := a.saveLastStats(ctx); err != nil {
		a.log.Error().Err(err).Msg("Failed to save initial lastStats")
	}
	a.mu.Lock() // 重新获取锁

	a.log.Info().Int("count", len(a.lastStats)).Msg("Statistics initialized for backends")
}

// CollectRealtimeMetrics 收集实时指标
func (a *StatsAggregator) CollectRealtimeMetrics(ctx context.Context) error {
	stats, err := a.runner.GetStats()
	if err != nil {
		return fmt.Errorf("failed to collect realtime metrics: %v", err)
	}

	return a.processRealtimeMetrics(ctx, stats, time.Now())
}

// processRealtimeMetrics 处理实时指标
func (a *StatsAggregator) processRealtimeMetrics(ctx context.Context, stats models.NativeStats, t time.Time) error {
	db, err := mongodb.GetDatabase(a.dbName)
	if err != nil {
		return err
	}

	// 获取当前过滤器的副本以避免并发访问
	a.mu.RLock()
	targetFilter := make(map[string]bool)
	for k, v := range a.TargetFilter {
		targetFilter[k] = v
	}
	a.mu.RUnlock()

	// 用于存储聚合数据
	totalConnRate := int64(0)
	totalScur := int64(0)
	totalRate := int64(0)
	totalReqRate := int64(0)

	// 准备保存到时间序列的文档
	documents := make(map[string][]interface{})
	for _, metric := range []string{"conn_rate", "scur", "rate", "req_rate"} {
		documents[metric] = make([]interface{}, 0)
	}

	for _, stat := range stats.Stats {
		if stat.Type != "frontend" || !targetFilter[stat.Name] || stat.Stats == nil {
			continue
		}

		// 收集各backend的实时指标
		// 添加单个backend的指标
		if stat.Stats.ConnRate != nil {
			totalConnRate += *stat.Stats.ConnRate
			documents["conn_rate"] = append(documents["conn_rate"], bson.M{
				"timestamp": t,
				"value":     *stat.Stats.ConnRate,
				"metadata":  bson.M{"target": stat.Name},
			})
		}

		if stat.Stats.Scur != nil {
			totalScur += *stat.Stats.Scur
			documents["scur"] = append(documents["scur"], bson.M{
				"timestamp": t,
				"value":     *stat.Stats.Scur,
				"metadata":  bson.M{"target": stat.Name},
			})
		}

		if stat.Stats.Rate != nil {
			totalRate += *stat.Stats.Rate
			documents["rate"] = append(documents["rate"], bson.M{
				"timestamp": t,
				"value":     *stat.Stats.Rate,
				"metadata":  bson.M{"target": stat.Name},
			})
		}

		if stat.Stats.ReqRate != nil {
			totalReqRate += *stat.Stats.ReqRate
			documents["req_rate"] = append(documents["req_rate"], bson.M{
				"timestamp": t,
				"value":     *stat.Stats.ReqRate,
				"metadata":  bson.M{"target": stat.Name},
			})
		}
	}

	// 添加所有backend的总计指标
	documents["conn_rate"] = append(documents["conn_rate"], bson.M{
		"timestamp": t,
		"value":     totalConnRate,
		"metadata":  bson.M{"target": "all"},
	})

	documents["scur"] = append(documents["scur"], bson.M{
		"timestamp": t,
		"value":     totalScur,
		"metadata":  bson.M{"target": "all"},
	})

	documents["rate"] = append(documents["rate"], bson.M{
		"timestamp": t,
		"value":     totalRate,
		"metadata":  bson.M{"target": "all"},
	})

	documents["req_rate"] = append(documents["req_rate"], bson.M{
		"timestamp": t,
		"value":     totalReqRate,
		"metadata":  bson.M{"target": "all"},
	})

	// 保存到MongoDB时间序列集合
	for metric, docs := range documents {
		if len(docs) > 0 {
			collection := db.Collection(metric)
			_, err := collection.InsertMany(ctx, docs)
			if err != nil {
				return fmt.Errorf("failed to insert %s metrics: %v", metric, err)
			}
		}
	}

	return nil
}

// CollectMinuteMetrics 收集分钟级指标
func (a *StatsAggregator) CollectMinuteMetrics(ctx context.Context) error {
	stats, err := a.runner.GetStats()
	if err != nil {
		return fmt.Errorf("failed to collect minute metrics: %v", err)
	}

	t := time.Now()
	err = a.processMinuteMetrics(ctx, stats, t)
	if err != nil {
		return fmt.Errorf("failed to process minute metrics: %v", err)
	}

	// 每处理完一个周期就保存基准数据，确保崩溃时最多丢失一个周期
	if err := a.saveLastStats(ctx); err != nil {
		a.log.Error().Err(err).Msg("Failed to save lastStats after minute metrics")
	}

	return nil
}

// processMinuteMetrics 处理分钟级指标
func (a *StatsAggregator) processMinuteMetrics(ctx context.Context, stats models.NativeStats, t time.Time) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// 用于聚合所有backend的总指标
	aggregateStats := make(map[string]int64)

	// 创建当前活跃后端Map
	activeTargets := make(map[string]bool)
	for _, stat := range stats.Stats {
		if stat.Type == "frontend" {
			activeTargets[stat.Name] = true
		}
	}

	// 检测并记录监控列表中但不再活跃的后端
	for target := range a.TargetFilter {
		if !activeTargets[target] {
			a.log.Warn().Str("target", target).Msg("Target is in monitoring list but not active")
		}
	}

	// 按每个backend处理
	for _, stat := range stats.Stats {
		if stat.Type != "frontend" || !a.TargetFilter[stat.Name] || stat.Stats == nil {
			continue
		}

		lastStat, exists := a.lastStats[stat.Name]
		if !exists {
			// 新的backend，初始化
			a.lastStats[stat.Name] = &HAProxyStatsData{
				TargetName: stat.Name,
				LastStats:  stat.Stats,
				LastTime:   t,
				ResetCount: 0,
			}
			a.log.Info().Str("target", stat.Name).Msg("New target detected during metrics collection")
			continue // 跳过计算差值，因为这是首次见到该backend
		}

		// 检测重启
		if a.detectReset(lastStat.LastStats, stat.Stats) {
			a.log.Warn().
				Str("target", stat.Name).
				Int("reset_count", lastStat.ResetCount+1).
				Msg("Detected HAProxy reset for target")

			// 记录零增量而非跳过
			zeroMetrics := a.createZeroMetrics()

			// 临时释放锁以防死锁
			a.mu.Unlock()
			err := a.saveMinuteMetrics(ctx, stat.Name, zeroMetrics, t)
			a.mu.Lock()

			if err != nil {
				a.log.Error().
					Err(err).
					Str("target", stat.Name).
					Msg("Failed to save zero metrics for target after reset")
			}

			// 更新重启计数和基准
			lastStat.ResetCount++
			lastStat.LastStats = stat.Stats
			lastStat.LastTime = t

			// 也将零值添加到聚合统计中
			for k, v := range zeroMetrics {
				aggregateStats[k] += v // 零值对聚合无影响
			}

			continue
		}

		// 计算差值并保存
		deltas := a.calculateDeltas(lastStat.LastStats, stat.Stats)

		// 保存单个backend的差值
		a.mu.Unlock() // 临时释放锁以防死锁
		err := a.saveMinuteMetrics(ctx, stat.Name, deltas, t)
		a.mu.Lock() // 重新获取锁

		if err != nil {
			a.log.Error().
				Err(err).
				Str("target", stat.Name).
				Msg("Failed to save minute metrics for target")
		}

		// 累加到聚合统计中
		for k, v := range deltas {
			aggregateStats[k] += v
		}

		// 更新上一次的统计数据
		lastStat.LastStats = stat.Stats
		lastStat.LastTime = t
	}

	// 保存所有backend的总计指标
	if len(aggregateStats) > 0 {
		a.mu.Unlock() // 临时释放锁以防死锁
		err := a.saveMinuteMetrics(ctx, "all", aggregateStats, t)
		a.mu.Lock() // 重新获取锁

		if err != nil {
			a.log.Error().Err(err).Msg("Failed to save aggregate minute metrics")
		}
	}

	return nil
}

// detectReset 检测HAProxy是否已重启
func (a *StatsAggregator) detectReset(lastStats, currentStats *models.NativeStatStats) bool {
	// 检查几个关键指标，如果当前值比上次值小，说明可能发生了重置
	if lastStats == nil || currentStats == nil {
		return false
	}

	// 检查bin（入站字节数）
	if lastStats.Bin != nil && currentStats.Bin != nil && *currentStats.Bin < *lastStats.Bin {
		return true
	}

	// 检查bout（出站字节数）
	if lastStats.Bout != nil && currentStats.Bout != nil && *currentStats.Bout < *lastStats.Bout {
		return true
	}

	// 检查总连接数
	if lastStats.ConnTot != nil && currentStats.ConnTot != nil && *currentStats.ConnTot < *lastStats.ConnTot {
		return true
	}

	// 检查总会话数
	if lastStats.Stot != nil && currentStats.Stot != nil && *currentStats.Stot < *lastStats.Stot {
		return true
	}

	// 检查总请求数
	if lastStats.ReqTot != nil && currentStats.ReqTot != nil && *currentStats.ReqTot < *lastStats.ReqTot {
		return true
	}

	return false
}

// createZeroMetrics 创建所有指标为0的映射，用于重启后记录
func (a *StatsAggregator) createZeroMetrics() map[string]int64 {
	metrics := make(map[string]int64)
	// 添加所有需要跟踪的指标，值为0
	fields := []string{
		"bin", "bout",
		"hrsp_1xx", "hrsp_2xx", "hrsp_3xx", "hrsp_4xx", "hrsp_5xx", "hrsp_other",
		"dreq", "dresp", "ereq", "dcon", "dses", "econ", "eresp",
		"req_rate_max", "conn_rate_max", "rate_max", "smax",
		"conn_tot", "stot", "req_tot",
	}

	for _, field := range fields {
		metrics[field] = 0
	}

	return metrics
}

// calculateDeltas 计算差值
func (a *StatsAggregator) calculateDeltas(lastStats, currentStats *models.NativeStatStats) map[string]int64 {
	deltas := make(map[string]int64)

	// 只计算我们关心的字段
	// 流量相关统计
	if lastStats.Bin != nil && currentStats.Bin != nil {
		deltas["bin"] = a.safeSubtract(*currentStats.Bin, *lastStats.Bin)
	}

	if lastStats.Bout != nil && currentStats.Bout != nil {
		deltas["bout"] = a.safeSubtract(*currentStats.Bout, *lastStats.Bout)
	}

	// HTTP响应状态码统计
	if lastStats.Hrsp1xx != nil && currentStats.Hrsp1xx != nil {
		deltas["hrsp_1xx"] = a.safeSubtract(*currentStats.Hrsp1xx, *lastStats.Hrsp1xx)
	}

	if lastStats.Hrsp2xx != nil && currentStats.Hrsp2xx != nil {
		deltas["hrsp_2xx"] = a.safeSubtract(*currentStats.Hrsp2xx, *lastStats.Hrsp2xx)
	}

	if lastStats.Hrsp3xx != nil && currentStats.Hrsp3xx != nil {
		deltas["hrsp_3xx"] = a.safeSubtract(*currentStats.Hrsp3xx, *lastStats.Hrsp3xx)
	}

	if lastStats.Hrsp4xx != nil && currentStats.Hrsp4xx != nil {
		deltas["hrsp_4xx"] = a.safeSubtract(*currentStats.Hrsp4xx, *lastStats.Hrsp4xx)
	}

	if lastStats.Hrsp5xx != nil && currentStats.Hrsp5xx != nil {
		deltas["hrsp_5xx"] = a.safeSubtract(*currentStats.Hrsp5xx, *lastStats.Hrsp5xx)
	}

	if lastStats.HrspOther != nil && currentStats.HrspOther != nil {
		deltas["hrsp_other"] = a.safeSubtract(*currentStats.HrspOther, *lastStats.HrspOther)
	}

	// 错误相关统计
	if lastStats.Dreq != nil && currentStats.Dreq != nil {
		deltas["dreq"] = a.safeSubtract(*currentStats.Dreq, *lastStats.Dreq)
	}

	if lastStats.Dresp != nil && currentStats.Dresp != nil {
		deltas["dresp"] = a.safeSubtract(*currentStats.Dresp, *lastStats.Dresp)
	}

	if lastStats.Ereq != nil && currentStats.Ereq != nil {
		deltas["ereq"] = a.safeSubtract(*currentStats.Ereq, *lastStats.Ereq)
	}

	if lastStats.Dcon != nil && currentStats.Dcon != nil {
		deltas["dcon"] = a.safeSubtract(*currentStats.Dcon, *lastStats.Dcon)
	}

	if lastStats.Dses != nil && currentStats.Dses != nil {
		deltas["dses"] = a.safeSubtract(*currentStats.Dses, *lastStats.Dses)
	}

	if lastStats.Econ != nil && currentStats.Econ != nil {
		deltas["econ"] = a.safeSubtract(*currentStats.Econ, *lastStats.Econ)
	}

	if lastStats.Eresp != nil && currentStats.Eresp != nil {
		deltas["eresp"] = a.safeSubtract(*currentStats.Eresp, *lastStats.Eresp)
	}

	// 保存最大值
	if currentStats.ReqRateMax != nil {
		deltas["req_rate_max"] = *currentStats.ReqRateMax
	}

	if currentStats.ConnRateMax != nil {
		deltas["conn_rate_max"] = *currentStats.ConnRateMax
	}

	if currentStats.RateMax != nil {
		deltas["rate_max"] = *currentStats.RateMax
	}

	if currentStats.Smax != nil {
		deltas["smax"] = *currentStats.Smax
	}

	// 总计值
	if currentStats.ConnTot != nil {
		deltas["conn_tot"] = *currentStats.ConnTot
	}

	if currentStats.Stot != nil {
		deltas["stot"] = *currentStats.Stot
	}

	if currentStats.ReqTot != nil {
		deltas["req_tot"] = *currentStats.ReqTot
	}

	return deltas
}

// safeSubtract 安全的减法操作，确保结果不为负
func (a *StatsAggregator) safeSubtract(current, last int64) int64 {
	if current < last {
		// 可能是由于HAProxy重启导致的，返回当前值作为增量
		return current
	}
	return current - last
}

// saveMinuteMetrics 保存分钟级指标到MongoDB
func (a *StatsAggregator) saveMinuteMetrics(ctx context.Context, targetName string, metrics map[string]int64, timestamp time.Time) error {
	db, err := mongodb.GetDatabase(a.dbName)
	if err != nil {
		return err
	}

	// 创建文档
	doc := bson.M{
		"target_name": targetName,
		"timestamp":   timestamp,
		"date":        timestamp.Format("2006-01-02"),
		"hour":        timestamp.Hour(),
		"minute":      timestamp.Minute(),
		"stats":       metrics,
	}

	var haproxyMinuteStats model.HAProxyMinuteStats
	// 插入到数据库集合
	collection := db.Collection(haproxyMinuteStats.GetCollectionName())
	_, err = collection.InsertOne(ctx, doc)
	return err
}
