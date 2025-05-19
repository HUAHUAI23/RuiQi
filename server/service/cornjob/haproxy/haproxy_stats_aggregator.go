package cornjob

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	mongodb "github.com/HUAHUAI23/simple-waf/pkg/database/mongo"
	"github.com/HUAHUAI23/simple-waf/server/config"
	"github.com/HUAHUAI23/simple-waf/server/model"
	"github.com/HUAHUAI23/simple-waf/server/service/daemon"
	"github.com/haproxytech/client-native/v6/models"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// StatsData 结构用于记录最后一次的统计数据，用于计算差值
type StatsData struct {
	TargetName string
	LastStats  model.HAProxyStats
	LastTime   time.Time
	ResetCount int // 用于跟踪重启次数
}

// StatsAggregator HAProxy统计数据聚合器
type StatsAggregator struct {
	runner          daemon.ServiceRunner
	dbName          string
	lastStats       map[string]*StatsData // 用backend名称做key
	lastStatsLoaded bool                  // 标记是否已从数据库加载基准数据
	targetFilter    map[string]bool       // 过滤的backend列表
	log             zerolog.Logger
	isRunning       bool // 表示聚合器是否在运行
}

// NewStatsAggregator 创建新的数据聚合器
func NewStatsAggregator(runner daemon.ServiceRunner, dbName string, targetList []string) (*StatsAggregator, error) {
	logger := config.GetLogger().With().Str("component", "cronjob-haproxy-stats-aggregator").Logger()
	// 初始化targetFilter
	targetFilter := make(map[string]bool)
	for _, target := range targetList {
		targetFilter[target] = true
	}

	agg := &StatsAggregator{
		runner:          runner,
		dbName:          dbName,
		lastStats:       make(map[string]*StatsData),
		lastStatsLoaded: false,
		targetFilter:    targetFilter,
		log:             logger,
		isRunning:       false,
	}

	// 确保必要的集合和索引存在
	err := agg.ensureCollections()
	if err != nil {
		return nil, fmt.Errorf("failed to ensure collections: %w", err)
	}

	return agg, nil
}

// UpdateTargetList 更新监控的后端列表
func (a *StatsAggregator) UpdateTargetList(targetList []string) {
	// 创建新的过滤器
	newFilter := make(map[string]bool)
	for _, target := range targetList {
		newFilter[target] = true
	}

	// 检查状态变化
	hadNoTargets := len(a.targetFilter) == 0
	willHaveNoTargets := len(newFilter) == 0
	hasNewTargets := len(newFilter) > 0

	// 查找不再监控的目标
	for target := range a.targetFilter {
		if !newFilter[target] {
			a.log.Info().Str("target", target).Msg("Removing target from monitoring")
			// 从内存中移除，但保留数据库中的基准线(以防后续重新添加)
			delete(a.lastStats, target)
		}
	}

	// 更新过滤器
	prevFilter := a.targetFilter
	a.targetFilter = newFilter

	// 处理从有目标到无目标的转变
	if !hadNoTargets && willHaveNoTargets {
		a.log.Warn().Int("previous_targets", len(prevFilter)).Msg("All targets removed, aggregator entering standby mode")

		// 清空统计数据，减少内存占用
		a.lastStats = make(map[string]*StatsData)

		// 重置lastStatsLoaded状态，以便下次添加目标时能正确初始化
		a.lastStatsLoaded = false

		// 保存最终状态到数据库，标记完成
		if a.isRunning {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if err := a.saveLastStats(ctx); err != nil {
				a.log.Error().Err(err).Msg("Failed to save final stats when entering standby mode")
			}
		}
	} else if hadNoTargets && hasNewTargets && a.isRunning {
		// 如果从无目标状态变为有目标状态，需要初始化状态
		a.log.Info().Msg("Targets added to previously empty target list, initializing stats collection")

		// 创建上下文进行初始化
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// 尝试初始化统计数据
		if err := a.initializeStats(ctx); err != nil {
			a.log.Error().Err(err).Msg("Failed to initialize stats after targets were added")
		}
	} else {
		// 记录新增的目标
		for target := range newFilter {
			if !prevFilter[target] {
				a.log.Info().Str("target", target).Msg("Adding new target to monitoring")
				// 新增的目标会在下次数据采集时自动初始化
			}
		}
	}
}

// ensureCollections 确保必要的集合和索引存在
func (a *StatsAggregator) ensureCollections() error {
	db, err := mongodb.GetDatabase(a.dbName)
	if err != nil {
		return fmt.Errorf("failed to get database: %w", err)
	}

	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 获取所有集合名称
	collections, err := db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return fmt.Errorf("failed to list collections: %w", err)
	}

	// 定义需要检查的集合
	var haproxyStatsBaseline model.HAProxyStatsBaseline
	var haproxyMinuteStats model.HAProxyMinuteStats
	baselineCollName := haproxyStatsBaseline.GetCollectionName()
	minuteStatsCollName := haproxyMinuteStats.GetCollectionName()

	// 检查时间序列集合是否存在
	requiredCollections := []string{
		baselineCollName,
		minuteStatsCollName,
		"conn_rate",
		"scur",
		"rate",
		"req_rate",
	}

	// 使用 slices.Contains 检查所有必需的集合是否存在
	allCollectionsExist := true
	for _, collName := range requiredCollections {
		if !slices.Contains(collections, collName) {
			allCollectionsExist = false
			break
		}
	}

	// 如果所有必需的集合都存在，则认为数据库已初始化
	if allCollectionsExist {
		a.log.Info().
			Strs("requiredCollections", requiredCollections).
			Msg("All required collections exist, skipping initialization")
		return nil
	}

	// 如果不是所有集合都存在，执行完整的初始化
	a.log.Info().Msg("Some required collections are missing, initializing database")

	// 创建基准数据集合索引
	baselineColl := db.Collection(baselineCollName)
	indexOpts := options.Index().SetUnique(true)
	_, err = baselineColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "target_name", Value: 1}},
		Options: indexOpts,
	})
	if err != nil {
		return fmt.Errorf("failed to create index for baseline collection: %w", err)
	}
	a.log.Info().Msg("Created baseline collection index")

	// 创建分钟统计数据集合索引
	minuteStatsColl := db.Collection(minuteStatsCollName)
	_, err = minuteStatsColl.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			// 时间戳索引 - 支持时间范围查询
			Keys: bson.D{{Key: "timestamp", Value: 1}},
		},
		{
			// 主查询索引 - 按目标和时间查询的主要索引
			Keys: bson.D{
				{Key: "target_name", Value: 1},
				{Key: "timestamp", Value: 1},
			},
		},
		{
			// 按小时聚合索引
			Keys: bson.D{
				{Key: "target_name", Value: 1},
				{Key: "date", Value: 1},
				{Key: "hour", Value: 1},
			},
		},
		{
			// 按六小时组聚合索引
			Keys: bson.D{
				{Key: "target_name", Value: 1},
				{Key: "date", Value: 1},
				{Key: "hourGroupSix", Value: 1},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create indexes for minute stats collection: %w", err)
	}
	a.log.Info().Msg("Created minute stats collection indexes")

	// 创建实时统计数据时间序列集合
	realtimeMetrics := []string{"conn_rate", "scur", "rate", "req_rate"}
	for _, metric := range realtimeMetrics {
		// 使用 slices.Contains 检查集合是否已存在
		if !slices.Contains(collections, metric) {
			// 创建时间序列集合
			timeSeriesOpts := options.TimeSeries().
				SetTimeField("timestamp").
				SetMetaField("metadata").
				SetGranularity("seconds")
			createOpts := options.CreateCollection().
				SetTimeSeriesOptions(timeSeriesOpts).
				SetExpireAfterSeconds(3600) // 设置1小时过期
			if err := db.CreateCollection(ctx, metric, createOpts); err != nil {
				return fmt.Errorf("failed to create timeseries collection %s: %w", metric, err)
			}
			a.log.Info().Msgf("Created timeseries collection %s", metric)
		}
	}

	a.log.Info().Msg("Database initialization completed successfully")
	return nil
}

// Start 启动聚合器
func (a *StatsAggregator) Start(ctx context.Context) error {
	if a.isRunning {
		return errors.New("aggregator is already running")
	}
	a.isRunning = true

	// 使用带超时的上下文进行初始化操作
	initCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// 尝试从数据库加载基准数据
	if err := a.loadLastStats(initCtx); err != nil {
		a.log.Warn().Err(err).Msg("Failed to load lastStats, will initialize new baseline")
		// 如果加载失败，初始化新基准
		if err := a.initializeStats(initCtx); err != nil {
			a.isRunning = false
			return fmt.Errorf("failed to initialize stats: %w", err)
		}
	} else {
		a.log.Info().Msg("Successfully loaded lastStats data from database")
		a.lastStatsLoaded = true
		// 获取最新数据并检查HAProxy是否重启
		if err := a.checkHAProxyResetOnStartup(initCtx); err != nil {
			a.log.Warn().Err(err).Msg("Error checking HAProxy reset on startup, continuing anyway")
		}
	}

	return nil
}

// Stop 停止聚合器
func (a *StatsAggregator) Stop(ctx context.Context) error {
	if !a.isRunning {
		return nil // 已经停止，不需要再做任何事
	}
	a.isRunning = false

	// 创建一个带超时的上下文
	saveCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// 保存最后的状态
	if err := a.saveLastStats(saveCtx); err != nil {
		a.log.Warn().Err(err).Msg("Failed to save lastStats during shutdown")
		// 不返回错误，继续关闭流程
	}

	a.log.Info().Msg("Aggregator successfully stopped")
	return nil
}

// loadLastStats 从MongoDB加载lastStats基准数据
func (a *StatsAggregator) loadLastStats(ctx context.Context) error {
	db, err := mongodb.GetDatabase(a.dbName)
	if err != nil {
		return fmt.Errorf("failed to get database: %w", err)
	}

	var haproxyStatsBaseline model.HAProxyStatsBaseline
	collection := db.Collection(haproxyStatsBaseline.GetCollectionName())

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("failed to query lastStats: %w", err)
	}
	defer cursor.Close(ctx)

	var baselines []model.HAProxyStatsBaseline
	if err = cursor.All(ctx, &baselines); err != nil {
		return fmt.Errorf("failed to decode lastStats documents: %w", err)
	}

	if len(baselines) == 0 {
		return errors.New("no baseline data found")
	}

	loadCount := 0
	for _, doc := range baselines {
		targetName := doc.TargetName
		if targetName == "" {
			a.log.Warn().Interface("doc", doc).Msg("Invalid target_name in baseline document")
			continue
		}

		// 加载到内存中
		a.lastStats[targetName] = &StatsData{
			TargetName: targetName,
			LastStats:  doc.GetStats(), // 使用新方法获取统计数据
			LastTime:   doc.Timestamp,
			ResetCount: int(doc.ResetCount),
		}

		loadCount++
	}

	a.log.Info().Int("count", loadCount).Msg("Loaded target baselines from database")
	if loadCount == 0 {
		return errors.New("no valid baseline data loaded")
	}
	return nil
}

// saveLastStats 将lastStats基准数据保存到MongoDB
func (a *StatsAggregator) saveLastStats(ctx context.Context) error {
	db, err := mongodb.GetDatabase(a.dbName)
	if err != nil {
		return fmt.Errorf("failed to get database: %w", err)
	}

	var haproxyStatsBaseline model.HAProxyStatsBaseline
	collection := db.Collection(haproxyStatsBaseline.GetCollectionName())

	var errs []error
	for targetName, stats := range a.lastStats {
		// 创建基准文档
		doc := model.HAProxyStatsBaseline{
			TargetName: targetName,
			Timestamp:  stats.LastTime,
			ResetCount: int32(stats.ResetCount),
		}
		// 设置统计数据
		doc.SetStats(stats.LastStats)

		// 使用upsert操作保存
		filter := bson.M{"target_name": targetName}
		update := bson.M{"$set": doc}

		opts := options.UpdateOne().SetUpsert(true)
		_, err := collection.UpdateOne(ctx, filter, update, opts)

		if err != nil {
			errs = append(errs, fmt.Errorf("failed to save lastStats for %s: %w", targetName, err))
			a.log.Error().Err(err).Str("target", targetName).Msg("Failed to save baseline for target")
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors occurred while saving baselines: %v", errs)
	}
	return nil
}

// checkHAProxyResetOnStartup 启动时检查HAProxy是否重启
func (a *StatsAggregator) checkHAProxyResetOnStartup(ctx context.Context) error {
	stats, err := a.runner.GetStats()
	if err != nil {
		a.log.Error().Err(err).Msg("checkHAProxyResetOnStartup: failed to get stats")
	}

	resetDetected := false
	newTargets := make(map[string]bool)

	// 找出所有当前活跃的后端
	for _, stat := range stats.Stats {
		if stat.Type == "frontend" {
			newTargets[stat.Name] = true
		}
	}

	// 检查当前监控的后端
	for _, stat := range stats.Stats {
		if stat.Type != "frontend" || !a.targetFilter[stat.Name] || stat.Stats == nil {
			continue
		}

		currentStats := model.NativeStatsToHAProxyStats(stat.Stats)
		lastStat, exists := a.lastStats[stat.Name]

		if !exists {
			// 新的后端，初始化
			a.lastStats[stat.Name] = &StatsData{
				TargetName: stat.Name,
				LastStats:  currentStats,
				LastTime:   time.Now(),
				ResetCount: 0,
			}
			a.log.Info().Str("target", stat.Name).Msg("New target detected during startup")
			continue
		}

		// 检测是否重启
		if model.DetectReset(lastStat.LastStats, currentStats) {
			a.log.Warn().
				Str("target", stat.Name).
				Int("reset_count", lastStat.ResetCount+1).
				Msg("Detected HAProxy reset for target during startup")

			resetDetected = true

			// 保存零增量记录
			err := a.saveMinuteMetrics(ctx, stat.Name, model.CreateZeroStats(), time.Now())

			if err != nil {
				a.log.Error().
					Err(err).
					Str("target", stat.Name).
					Msg("Failed to save zero metrics for target after reset")
			}

			// 更新重启计数和基准
			lastStat.ResetCount++
			lastStat.LastStats = currentStats
			lastStat.LastTime = time.Now()
		}
	}

	// 检查不再存在的后端
	for targetName := range a.lastStats {
		if a.targetFilter[targetName] && !newTargets[targetName] {
			a.log.Warn().
				Str("target", targetName).
				Msg("Target is in monitoring list but not found in HAProxy stats")
		}
	}

	if resetDetected {
		a.log.Warn().Msg("HAProxy reset detected during startup, metrics will be adjusted")
		// 保存重置后的状态
		if err := a.saveLastStats(ctx); err != nil {
			a.log.Error().Err(err).Msg("Failed to save lastStats after reset detection")
			return fmt.Errorf("failed to save state after reset: %w", err)
		}
	} else {
		a.log.Info().Msg("No HAProxy reset detected during startup")
	}

	return nil
}

// initializeStats 初始化统计数据
func (a *StatsAggregator) initializeStats(ctx context.Context) error {
	stats, err := a.runner.GetStats()
	if err != nil {
		a.log.Error().Err(err).Msg("initializeStats: failed to get stats")
	}

	now := time.Now()
	a.lastStats = make(map[string]*StatsData) // 清除现有数据

	for _, stat := range stats.Stats {
		if stat.Type == "frontend" && a.targetFilter[stat.Name] && stat.Stats != nil {
			currentStats := model.NativeStatsToHAProxyStats(stat.Stats)

			a.lastStats[stat.Name] = &StatsData{
				TargetName: stat.Name,
				LastStats:  currentStats,
				LastTime:   now,
				ResetCount: 0,
			}
		}
	}

	// 检查是否至少有一个目标
	if len(a.lastStats) == 0 {
		a.log.Warn().Msg("No valid targets found to initialize, aggregator will wait for targets to be added")
		return nil // 返回nil而不是错误，允许服务启动但处于空闲状态
	}

	// 保存初始基准到数据库
	if err := a.saveLastStats(ctx); err != nil {
		a.log.Error().Err(err).Msg("Failed to save initial lastStats")
		return fmt.Errorf("failed to save initial state: %w", err)
	}

	a.log.Info().Int("count", len(a.lastStats)).Msg("Statistics initialized for backends")
	return nil
}

// CollectRealtimeMetrics 收集实时指标
func (a *StatsAggregator) CollectRealtimeMetrics(ctx context.Context) error {
	if !a.isRunning {
		return errors.New("aggregator is not running")
	}

	// 如果没有目标，则跳过收集
	if len(a.targetFilter) == 0 {
		a.log.Debug().Msg("Skipping realtime metrics collection: no targets configured")
		return nil
	}

	stats, err := a.runner.GetStats()
	if err != nil {
		a.log.Error().Err(err).Msg("CollectRealtimeMetrics: failed to get stats")
	}

	return a.processRealtimeMetrics(ctx, stats, time.Now())
}

// processRealtimeMetrics 处理实时指标
func (a *StatsAggregator) processRealtimeMetrics(ctx context.Context, stats models.NativeStats, t time.Time) error {
	db, err := mongodb.GetDatabase(a.dbName)
	if err != nil {
		return fmt.Errorf("failed to get database: %w", err)
	}

	// 用于存储聚合数据
	totalConnRate := int64(0)
	totalScur := int64(0)
	totalRate := int64(0)
	totalReqRate := int64(0)

	// 准备保存到时间序列的文档
	documents := map[string][]model.TimeSeriesMetric{
		"conn_rate": {},
		"scur":      {},
		"rate":      {},
		"req_rate":  {},
	}

	for _, stat := range stats.Stats {
		if stat.Type != "frontend" || !a.targetFilter[stat.Name] || stat.Stats == nil {
			continue
		}

		// 收集各backend的实时指标
		if stat.Stats.ConnRate != nil {
			totalConnRate += *stat.Stats.ConnRate
			documents["conn_rate"] = append(documents["conn_rate"], model.TimeSeriesMetric{
				Timestamp: t,
				Value:     *stat.Stats.ConnRate,
				Metadata:  model.TimeSeriesMeta{Target: stat.Name},
			})
		}

		if stat.Stats.Scur != nil {
			totalScur += *stat.Stats.Scur
			documents["scur"] = append(documents["scur"], model.TimeSeriesMetric{
				Timestamp: t,
				Value:     *stat.Stats.Scur,
				Metadata:  model.TimeSeriesMeta{Target: stat.Name},
			})
		}

		if stat.Stats.Rate != nil {
			totalRate += *stat.Stats.Rate
			documents["rate"] = append(documents["rate"], model.TimeSeriesMetric{
				Timestamp: t,
				Value:     *stat.Stats.Rate,
				Metadata:  model.TimeSeriesMeta{Target: stat.Name},
			})
		}

		if stat.Stats.ReqRate != nil {
			totalReqRate += *stat.Stats.ReqRate
			documents["req_rate"] = append(documents["req_rate"], model.TimeSeriesMetric{
				Timestamp: t,
				Value:     *stat.Stats.ReqRate,
				Metadata:  model.TimeSeriesMeta{Target: stat.Name},
			})
		}
	}

	// 添加所有backend的总计指标
	documents["conn_rate"] = append(documents["conn_rate"], model.TimeSeriesMetric{
		Timestamp: t,
		Value:     totalConnRate,
		Metadata:  model.TimeSeriesMeta{Target: "all"},
	})

	documents["scur"] = append(documents["scur"], model.TimeSeriesMetric{
		Timestamp: t,
		Value:     totalScur,
		Metadata:  model.TimeSeriesMeta{Target: "all"},
	})

	documents["rate"] = append(documents["rate"], model.TimeSeriesMetric{
		Timestamp: t,
		Value:     totalRate,
		Metadata:  model.TimeSeriesMeta{Target: "all"},
	})

	documents["req_rate"] = append(documents["req_rate"], model.TimeSeriesMetric{
		Timestamp: t,
		Value:     totalReqRate,
		Metadata:  model.TimeSeriesMeta{Target: "all"},
	})

	// 保存到MongoDB时间序列集合
	var errs []error
	for metric, docs := range documents {
		if len(docs) > 0 {
			collection := db.Collection(metric)

			// 转换为适合InsertMany的格式
			insertDocs := make([]interface{}, len(docs))
			for i, doc := range docs {
				insertDocs[i] = doc
			}

			_, err := collection.InsertMany(ctx, insertDocs)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to insert %s metrics: %w", metric, err))
				a.log.Error().Err(err).Str("metric", metric).Msg("Failed to insert metrics")
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors occurred while inserting metrics: %v", errs)
	}
	return nil
}

// CollectMinuteMetrics 收集分钟级指标
func (a *StatsAggregator) CollectMinuteMetrics(ctx context.Context) error {
	if !a.isRunning {
		return errors.New("aggregator is not running")
	}

	// 如果没有目标，则跳过收集
	if len(a.targetFilter) == 0 {
		a.log.Debug().Msg("Skipping minute metrics collection: no targets configured")
		return nil
	}

	stats, err := a.runner.GetStats()
	if err != nil {
		a.log.Error().Err(err).Msg("CollectMinuteMetrics: failed to get stats")
	}

	t := time.Now()
	err = a.processMinuteMetrics(ctx, stats, t)
	if err != nil {
		return fmt.Errorf("failed to process minute metrics: %w", err)
	}

	// 每处理完一个周期就保存基准数据，确保崩溃时最多丢失一个周期
	saveCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := a.saveLastStats(saveCtx); err != nil {
		a.log.Error().Err(err).Msg("Failed to save lastStats after minute metrics")
		// 不返回错误，继续执行
	}

	return nil
}

// processMinuteMetrics 处理分钟级指标
func (a *StatsAggregator) processMinuteMetrics(ctx context.Context, stats models.NativeStats, t time.Time) error {
	// 用于聚合所有backend的总指标
	aggregateStats := model.HAProxyStats{}

	// 创建当前活跃后端Map
	activeTargets := make(map[string]bool)
	for _, stat := range stats.Stats {
		if stat.Type == "frontend" {
			activeTargets[stat.Name] = true
		}
	}

	// 检测并记录监控列表中但不再活跃的后端
	for target := range a.targetFilter {
		if !activeTargets[target] {
			a.log.Warn().Str("target", target).Msg("Target is in monitoring list but not active")
		}
	}

	// 按每个backend处理
	var errs []error
	for _, stat := range stats.Stats {
		if stat.Type != "frontend" || !a.targetFilter[stat.Name] || stat.Stats == nil {
			continue
		}

		currentStats := model.NativeStatsToHAProxyStats(stat.Stats)
		lastStat, exists := a.lastStats[stat.Name]

		if !exists {
			// 新的backend，初始化
			a.lastStats[stat.Name] = &StatsData{
				TargetName: stat.Name,
				LastStats:  currentStats,
				LastTime:   t,
				ResetCount: 0,
			}
			a.log.Info().Str("target", stat.Name).Msg("New target detected during metrics collection")
			continue // 跳过计算差值，因为这是首次见到该backend
		}

		// 检测重启
		if model.DetectReset(lastStat.LastStats, currentStats) {
			a.log.Warn().
				Str("target", stat.Name).
				Int("reset_count", lastStat.ResetCount+1).
				Msg("Detected HAProxy reset for target")

			// 记录零增量而非跳过
			zeroMetrics := model.CreateZeroStats()

			err := a.saveMinuteMetrics(ctx, stat.Name, zeroMetrics, t)

			if err != nil {
				a.log.Error().
					Err(err).
					Str("target", stat.Name).
					Msg("Failed to save zero metrics for target after reset")
				errs = append(errs, fmt.Errorf("failed to save zero metrics for %s: %w", stat.Name, err))
			}

			// 更新重启计数和基准
			lastStat.ResetCount++
			lastStat.LastStats = currentStats
			lastStat.LastTime = t

			// 也将零值添加到聚合统计中 - 零值对聚合无影响
			continue
		}

		// 计算差值并保存
		deltas := model.CalculateStatsDelta(lastStat.LastStats, currentStats)

		// 保存单个backend的差值
		err := a.saveMinuteMetrics(ctx, stat.Name, deltas, t)

		if err != nil {
			a.log.Error().
				Err(err).
				Str("target", stat.Name).
				Msg("Failed to save minute metrics for target")
			errs = append(errs, fmt.Errorf("failed to save metrics for %s: %w", stat.Name, err))
		}

		// 累加到聚合统计中
		aggregateStats.Bin += deltas.Bin
		aggregateStats.Bout += deltas.Bout

		aggregateStats.Hrsp1xx += deltas.Hrsp1xx
		aggregateStats.Hrsp2xx += deltas.Hrsp2xx
		aggregateStats.Hrsp3xx += deltas.Hrsp3xx
		aggregateStats.Hrsp4xx += deltas.Hrsp4xx
		aggregateStats.Hrsp5xx += deltas.Hrsp5xx
		aggregateStats.HrspOther += deltas.HrspOther

		aggregateStats.Dreq += deltas.Dreq
		aggregateStats.Dresp += deltas.Dresp
		aggregateStats.Ereq += deltas.Ereq
		aggregateStats.Dcon += deltas.Dcon
		aggregateStats.Dses += deltas.Dses
		aggregateStats.Econ += deltas.Econ
		aggregateStats.Eresp += deltas.Eresp

		// 对于最大值，我们改为使用所有后端中的总和
		aggregateStats.ReqRateMax += deltas.ReqRateMax
		aggregateStats.ConnRateMax += deltas.ConnRateMax
		aggregateStats.RateMax += deltas.RateMax
		aggregateStats.Smax += deltas.Smax

		// 对于总计值，使用所有后端的总和
		aggregateStats.ConnTot += deltas.ConnTot
		aggregateStats.Stot += deltas.Stot
		aggregateStats.ReqTot += deltas.ReqTot

		// 更新上一次的统计数据
		lastStat.LastStats = currentStats
		lastStat.LastTime = t
	}

	// 保存所有backend的总计指标
	err := a.saveMinuteMetrics(ctx, "all", aggregateStats, t)

	if err != nil {
		a.log.Error().Err(err).Msg("Failed to save aggregate minute metrics")
		errs = append(errs, fmt.Errorf("failed to save aggregate metrics: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors occurred while processing minute metrics: %v", errs)
	}
	return nil
}

// saveMinuteMetrics 保存分钟级指标到MongoDB - 使用扁平化结构
func (a *StatsAggregator) saveMinuteMetrics(ctx context.Context, targetName string, metrics model.HAProxyStats, timestamp time.Time) error {
	db, err := mongodb.GetDatabase(a.dbName)
	if err != nil {
		return fmt.Errorf("failed to get database: %w", err)
	}

	// 创建扁平化的文档结构
	minuteStats := model.HAProxyMinuteStats{
		TargetName:   targetName,
		Date:         timestamp.Format("2006-01-02"),
		Hour:         timestamp.Hour(),
		HourGroupSix: timestamp.Hour() / 6,
		Minute:       timestamp.Minute(),
		Timestamp:    timestamp,
	}
	// 设置统计数据
	minuteStats.SetStats(metrics)

	// 插入到数据库集合
	collection := db.Collection(minuteStats.GetCollectionName())
	_, err = collection.InsertOne(ctx, minuteStats)
	if err != nil {
		return fmt.Errorf("failed to insert minute stats: %w", err)
	}

	return nil
}
