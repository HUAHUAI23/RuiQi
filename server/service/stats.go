package service

import (
	"context"
	"fmt"
	"math"
	"time"

	mongodb "github.com/HUAHUAI23/simple-waf/pkg/database/mongo"
	"github.com/HUAHUAI23/simple-waf/server/config"
	"github.com/HUAHUAI23/simple-waf/server/dto"
	"github.com/HUAHUAI23/simple-waf/server/repository"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type StatsService interface {
	GetOverviewStats(ctx context.Context, timeRange string) (*dto.OverviewStats, error)
	GetRealtimeQPS(ctx context.Context, limit int) (*dto.RealtimeQPSResponse, error)
	GetTimeSeriesData(ctx context.Context, timeRange string, metric string) (*dto.TimeSeriesResponse, error)
	GetCombinedTimeSeriesData(ctx context.Context, timeRange string) (*dto.CombinedTimeSeriesResponse, error)
	GetTrafficTimeSeriesData(ctx context.Context, timeRange string) (*dto.TrafficTimeSeriesResponse, error)
}

type StatsServiceImpl struct {
	wafLogRepository repository.WAFLogRepository
	dbName           string
	logger           zerolog.Logger
}

func NewStatsService(wafLogRepository repository.WAFLogRepository) StatsService {
	dbName := config.Global.DBConfig.Database
	logger := config.GetServiceLogger("stats")
	return &StatsServiceImpl{
		wafLogRepository: wafLogRepository,
		dbName:           dbName,
		logger:           logger,
	}
}

// GetOverviewStats 获取概览统计数据
func (s *StatsServiceImpl) GetOverviewStats(ctx context.Context, timeRange string) (*dto.OverviewStats, error) {
	// 确定时间范围
	startTime, err := s.getTimeRangeStart(timeRange)
	if err != nil {
		return nil, err
	}

	// 获取MongoDB数据库连接
	db, err := mongodb.GetDatabase(s.dbName)
	if err != nil {
		return nil, fmt.Errorf("获取数据库连接失败: %w", err)
	}

	// 1. 获取 HAProxyMinuteStats 的统计数据
	minuteStatsResult, err := s.getMinuteStatsAggregate(ctx, db, startTime)
	if err != nil {
		return nil, fmt.Errorf("获取分钟统计数据失败: %w", err)
	}

	// 2. 获取 WAF 拦截统计数据
	blockCount, attackIPCount, err := s.getWAFBlockStats(ctx, startTime)
	if err != nil {
		return nil, fmt.Errorf("获取WAF拦截统计失败: %w", err)
	}

	// 3. 计算错误率
	error4xxRate := 0.0
	error5xxRate := 0.0

	if minuteStatsResult.TotalRequests > 0 {
		error4xxRate = float64(minuteStatsResult.Error4xx) / float64(minuteStatsResult.TotalRequests) * 100
		error5xxRate = float64(minuteStatsResult.Error5xx) / float64(minuteStatsResult.TotalRequests) * 100
	}

	// 构建结果
	result := &dto.OverviewStats{
		TimeRange:       timeRange,
		TotalRequests:   minuteStatsResult.TotalRequests,
		InboundTraffic:  minuteStatsResult.InboundTraffic,
		OutboundTraffic: minuteStatsResult.OutboundTraffic,
		MaxQPS:          minuteStatsResult.MaxQPS,
		Error4xx:        minuteStatsResult.Error4xx,
		Error4xxRate:    math.Round(error4xxRate*100) / 100, // 保留两位小数
		Error5xx:        minuteStatsResult.Error5xx,
		Error5xxRate:    math.Round(error5xxRate*100) / 100, // 保留两位小数
		BlockCount:      blockCount,
		AttackIPCount:   attackIPCount,
	}

	return result, nil
}

// GetRealtimeQPS 获取实时QPS数据
func (s *StatsServiceImpl) GetRealtimeQPS(ctx context.Context, limit int) (*dto.RealtimeQPSResponse, error) {
	if limit <= 0 {
		limit = 30 // 默认获取30个点
	}
	if limit > 60 {
		limit = 60 // 最多不超过60个点
	}

	// 获取MongoDB数据库连接
	db, err := mongodb.GetDatabase(s.dbName)
	if err != nil {
		return nil, fmt.Errorf("获取数据库连接失败: %w", err)
	}

	// 从时序数据库中查询数据
	matchStage := bson.D{{Key: "$match", Value: bson.D{
		{Key: "metadata.target", Value: "all"},
	}}}

	sortStage := bson.D{{Key: "$sort", Value: bson.D{
		{Key: "timestamp", Value: -1}, // 按时间降序排序
	}}}

	limitStage := bson.D{{Key: "$limit", Value: limit}}

	pipeline := mongo.Pipeline{matchStage, sortStage, limitStage}

	// 执行聚合查询
	reqRateColl := db.Collection("req_rate")
	cursor, err := reqRateColl.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("查询实时QPS数据失败: %w", err)
	}
	defer cursor.Close(ctx)

	// 解析结果
	var results []struct {
		Timestamp time.Time `bson:"timestamp"`
		Value     int64     `bson:"value"`
		Metadata  bson.D    `bson:"metadata"`
	}
	if err = cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("解析实时QPS数据失败: %w", err)
	}

	// 构建结果
	dataPoints := make([]dto.RealtimeQPSData, 0, len(results))
	// 反转顺序，让时间从旧到新
	for i := len(results) - 1; i >= 0; i-- {
		dataPoints = append(dataPoints, dto.RealtimeQPSData{
			Timestamp: results[i].Timestamp,
			Value:     results[i].Value,
		})
	}

	return &dto.RealtimeQPSResponse{
		Data: dataPoints,
	}, nil
}

// GetTimeSeriesData 获取时间序列数据
func (s *StatsServiceImpl) GetTimeSeriesData(ctx context.Context, timeRange string, metric string) (*dto.TimeSeriesResponse, error) {
	// 确定时间范围
	startTime, err := s.getTimeRangeStart(timeRange)
	if err != nil {
		return nil, err
	}

	// 根据时间范围决定数据聚合粒度
	var interval string
	var groupByFields []string

	switch timeRange {
	case dto.TimeRange24Hours:
		// 24小时内按小时聚合
		interval = "hour"
		groupByFields = []string{"date", "hour"}
	case dto.TimeRange7Days:
		// 7天内按6小时聚合，提供更多数据点
		interval = "6hour"
		groupByFields = []string{"date", "hourGroupSix"}
	case dto.TimeRange30Days:
		// 30天内按天聚合
		interval = "day"
		groupByFields = []string{"date"}
	default:
		return nil, fmt.Errorf("无效的时间范围: %s", timeRange)
	}

	// 根据指标类型获取数据
	var dataPoints []dto.TimeSeriesDataPoint

	switch metric {
	case "requests":
		// 请求数时间序列
		dataPoints, err = s.getRequestTimeSeries(ctx, startTime, interval, groupByFields)
	case "blocks":
		// 拦截数时间序列
		dataPoints, err = s.getBlockTimeSeries(ctx, startTime, interval, groupByFields)
	default:
		return nil, fmt.Errorf("无效的指标类型: %s", metric)
	}

	if err != nil {
		return nil, err
	}

	return &dto.TimeSeriesResponse{
		Metric:    metric,
		TimeRange: timeRange,
		Data:      dataPoints,
	}, nil
}

// GetCombinedTimeSeriesData 同时获取请求数和拦截数的时间序列数据
func (s *StatsServiceImpl) GetCombinedTimeSeriesData(ctx context.Context, timeRange string) (*dto.CombinedTimeSeriesResponse, error) {
	// 确定时间范围
	startTime, err := s.getTimeRangeStart(timeRange)
	if err != nil {
		return nil, err
	}

	// 根据时间范围决定数据聚合粒度
	var interval string
	var groupByFields []string

	switch timeRange {
	case dto.TimeRange24Hours:
		// 24小时内按小时聚合
		interval = "hour"
		groupByFields = []string{"date", "hour"}
	case dto.TimeRange7Days:
		// 7天内按6小时聚合，提供更多数据点
		interval = "6hour"
		groupByFields = []string{"date", "hourGroupSix"}
	case dto.TimeRange30Days:
		// 30天内按天聚合
		interval = "day"
		groupByFields = []string{"date"}
	default:
		return nil, fmt.Errorf("无效的时间范围: %s", timeRange)
	}

	// 并行获取两种数据
	requestCh := make(chan struct {
		data []dto.TimeSeriesDataPoint
		err  error
	})
	blockCh := make(chan struct {
		data []dto.TimeSeriesDataPoint
		err  error
	})

	// 获取请求数据
	go func() {
		data, err := s.getRequestTimeSeries(ctx, startTime, interval, groupByFields)
		requestCh <- struct {
			data []dto.TimeSeriesDataPoint
			err  error
		}{data, err}
	}()

	// 获取拦截数据
	go func() {
		data, err := s.getBlockTimeSeries(ctx, startTime, interval, groupByFields)
		blockCh <- struct {
			data []dto.TimeSeriesDataPoint
			err  error
		}{data, err}
	}()

	// 接收结果
	requestResult := <-requestCh
	blockResult := <-blockCh

	// 检查错误
	if requestResult.err != nil {
		return nil, fmt.Errorf("获取请求数据失败: %w", requestResult.err)
	}
	if blockResult.err != nil {
		return nil, fmt.Errorf("获取拦截数据失败: %w", blockResult.err)
	}

	// 返回组合结果
	return &dto.CombinedTimeSeriesResponse{
		TimeRange: timeRange,
		Requests: dto.TimeSeriesResponse{
			Metric:    "requests",
			TimeRange: timeRange,
			Data:      requestResult.data,
		},
		Blocks: dto.TimeSeriesResponse{
			Metric:    "blocks",
			TimeRange: timeRange,
			Data:      blockResult.data,
		},
	}, nil
}

// GetTrafficTimeSeriesData 获取流量时间序列数据
func (s *StatsServiceImpl) GetTrafficTimeSeriesData(ctx context.Context, timeRange string) (*dto.TrafficTimeSeriesResponse, error) {
	// 确定时间范围
	startTime, err := s.getTimeRangeStart(timeRange)
	if err != nil {
		return nil, err
	}

	// 根据时间范围决定数据聚合粒度
	var interval string
	var groupByFields []string

	switch timeRange {
	case dto.TimeRange24Hours:
		// 24小时内按小时聚合
		interval = "hour"
		groupByFields = []string{"date", "hour"}
	case dto.TimeRange7Days:
		// 7天内按6小时聚合，提供更多数据点
		interval = "6hour"
		groupByFields = []string{"date", "hourGroupSix"}
	case dto.TimeRange30Days:
		// 30天内按天聚合
		interval = "day"
		groupByFields = []string{"date"}
	default:
		return nil, fmt.Errorf("无效的时间范围: %s", timeRange)
	}

	// 获取MongoDB数据库连接
	db, err := mongodb.GetDatabase(s.dbName)
	if err != nil {
		return nil, fmt.Errorf("获取数据库连接失败: %w", err)
	}

	collection := db.Collection("haproxy_minute_stats")

	// 构建时间过滤条件
	matchStage := bson.D{{Key: "$match", Value: bson.D{
		{Key: "target_name", Value: "all"},
		{Key: "timestamp", Value: bson.D{{Key: "$gte", Value: startTime}}},
	}}}

	// 构建分组条件
	var groupStage bson.D

	if interval == "6hour" {
		// 使用预计算的HourGroupSix字段进行分组
		groupStage = bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{
				{Key: "date", Value: "$date"},
				{Key: "hourGroup", Value: "$hourGroupSix"},
			}},
			{Key: "inboundTraffic", Value: bson.D{{Key: "$sum", Value: "$bin"}}},
			{Key: "outboundTraffic", Value: bson.D{{Key: "$sum", Value: "$bout"}}},
			{Key: "timestamp", Value: bson.D{{Key: "$min", Value: "$timestamp"}}},
		}}}
	} else {
		// 构建常规分组ID
		groupID := bson.D{}
		for _, field := range groupByFields {
			groupID = append(groupID, bson.E{Key: field, Value: fmt.Sprintf("$%s", field)})
		}

		groupStage = bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: groupID},
			{Key: "inboundTraffic", Value: bson.D{{Key: "$sum", Value: "$bin"}}},
			{Key: "outboundTraffic", Value: bson.D{{Key: "$sum", Value: "$bout"}}},
			{Key: "timestamp", Value: bson.D{{Key: "$min", Value: "$timestamp"}}},
		}}}
	}

	sortStage := bson.D{{Key: "$sort", Value: bson.D{
		{Key: "timestamp", Value: 1}, // 按时间升序排序
	}}}

	pipeline := mongo.Pipeline{matchStage, groupStage, sortStage}

	// 执行聚合查询
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("执行流量时间序列聚合查询失败: %w", err)
	}
	defer cursor.Close(ctx)

	// 解析结果
	var results []struct {
		ID              bson.D    `bson:"_id"`
		InboundTraffic  int64     `bson:"inboundTraffic"`
		OutboundTraffic int64     `bson:"outboundTraffic"`
		Timestamp       time.Time `bson:"timestamp"`
	}
	if err = cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("解析流量时间序列结果失败: %w", err)
	}

	// 构建数据点
	dataPoints := make([]dto.TrafficDataPoint, 0, len(results))
	for _, result := range results {
		dataPoints = append(dataPoints, dto.TrafficDataPoint{
			Timestamp:       result.Timestamp,
			InboundTraffic:  result.InboundTraffic,
			OutboundTraffic: result.OutboundTraffic,
		})
	}

	return &dto.TrafficTimeSeriesResponse{
		TimeRange: timeRange,
		Data:      dataPoints,
	}, nil
}

// 辅助方法 - 获取时间范围的开始时间
func (s *StatsServiceImpl) getTimeRangeStart(timeRange string) (time.Time, error) {
	now := time.Now()

	switch timeRange {
	case dto.TimeRange24Hours:
		return now.Add(-24 * time.Hour), nil
	case dto.TimeRange7Days:
		return now.AddDate(0, 0, -7), nil
	case dto.TimeRange30Days:
		return now.AddDate(0, 0, -30), nil
	default:
		return time.Time{}, fmt.Errorf("无效的时间范围: %s", timeRange)
	}
}

// 辅助方法 - 分钟统计聚合结果
type minuteStatsAggregateResult struct {
	TotalRequests   int64 // 总请求数
	InboundTraffic  int64 // 入站流量
	OutboundTraffic int64 // 出站流量
	MaxQPS          int64 // 最大QPS
	Error4xx        int64 // 4xx错误数
	Error5xx        int64 // 5xx错误数
}

// 辅助方法 - 获取分钟统计数据聚合结果
func (s *StatsServiceImpl) getMinuteStatsAggregate(ctx context.Context, db *mongo.Database, startTime time.Time) (*minuteStatsAggregateResult, error) {
	// 创建聚合管道
	matchStage := bson.D{{Key: "$match", Value: bson.D{
		{Key: "target_name", Value: "all"},
		{Key: "timestamp", Value: bson.D{{Key: "$gte", Value: startTime}}},
	}}}

	groupStage := bson.D{{Key: "$group", Value: bson.D{
		{Key: "_id", Value: nil},
		{Key: "totalRequests", Value: bson.D{{Key: "$sum", Value: "$req_tot"}}},
		{Key: "inboundTraffic", Value: bson.D{{Key: "$sum", Value: "$bin"}}},
		{Key: "outboundTraffic", Value: bson.D{{Key: "$sum", Value: "$bout"}}},
		{Key: "maxQPS", Value: bson.D{{Key: "$max", Value: "$req_rate_max"}}},
		{Key: "error4xx", Value: bson.D{{Key: "$sum", Value: "$hrsp_4xx"}}},
		{Key: "error5xx", Value: bson.D{{Key: "$sum", Value: "$hrsp_5xx"}}},
	}}}

	pipeline := mongo.Pipeline{matchStage, groupStage}

	// 执行聚合查询
	collection := db.Collection("haproxy_minute_stats")
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("执行统计聚合查询失败: %w", err)
	}
	defer cursor.Close(ctx)

	// 解析结果
	var results []struct {
		TotalRequests   int64 `bson:"totalRequests"`
		InboundTraffic  int64 `bson:"inboundTraffic"`
		OutboundTraffic int64 `bson:"outboundTraffic"`
		MaxQPS          int64 `bson:"maxQPS"`
		Error4xx        int64 `bson:"error4xx"`
		Error5xx        int64 `bson:"error5xx"`
	}
	if err = cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("解析统计聚合结果失败: %w", err)
	}

	// 构建结果
	result := &minuteStatsAggregateResult{}
	if len(results) > 0 {
		result.TotalRequests = results[0].TotalRequests
		result.InboundTraffic = results[0].InboundTraffic
		result.OutboundTraffic = results[0].OutboundTraffic
		result.MaxQPS = results[0].MaxQPS
		result.Error4xx = results[0].Error4xx
		result.Error5xx = results[0].Error5xx
	}

	return result, nil
}

// 辅助方法 - 获取WAF拦截统计
func (s *StatsServiceImpl) getWAFBlockStats(ctx context.Context, startTime time.Time) (int64, int64, error) {
	// 构建时间过滤条件
	timeFilter := bson.D{
		{Key: "createdAt", Value: bson.D{{Key: "$gte", Value: startTime}}},
	}

	// 获取拦截总数
	blockCount, err := s.wafLogRepository.CountAttackLogs(ctx, timeFilter)
	if err != nil {
		return 0, 0, fmt.Errorf("计算拦截总数失败: %w", err)
	}

	// 聚合不同攻击IP数
	ipCountPipeline := mongo.Pipeline{
		{
			{Key: "$match", Value: bson.D{
				{Key: "createdAt", Value: bson.D{{Key: "$gte", Value: startTime}}},
			}},
		},
		{
			{Key: "$group", Value: bson.D{
				{Key: "_id", Value: "$srcIp"},
			}},
		},
		{
			{Key: "$count", Value: "count"},
		},
	}

	attackIPCount, err := s.wafLogRepository.CountAggregateAttackEvents(ctx, ipCountPipeline)
	if err != nil {
		s.logger.Error().Err(err).Msg("计算攻击IP数量失败")
		// 忽略错误，降级返回0
		attackIPCount = 0
	}

	return blockCount, attackIPCount, nil
}

// 辅助方法 - 获取请求数时间序列
func (s *StatsServiceImpl) getRequestTimeSeries(ctx context.Context, startTime time.Time, interval string, groupByFields []string) ([]dto.TimeSeriesDataPoint, error) {
	// 获取MongoDB数据库连接
	db, err := mongodb.GetDatabase(s.dbName)
	if err != nil {
		return nil, fmt.Errorf("获取数据库连接失败: %w", err)
	}

	collection := db.Collection("haproxy_minute_stats")

	// 构建时间过滤条件
	matchStage := bson.D{{Key: "$match", Value: bson.D{
		{Key: "target_name", Value: "all"},
		{Key: "timestamp", Value: bson.D{{Key: "$gte", Value: startTime}}},
	}}}

	// 构建分组条件
	var groupStage bson.D

	if interval == "6hour" {
		// 使用预计算的HourGroupSix字段进行分组
		groupStage = bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{
				{Key: "date", Value: "$date"},
				{Key: "hourGroup", Value: "$hourGroupSix"},
			}},
			{Key: "requests", Value: bson.D{{Key: "$sum", Value: "$req_tot"}}},
			{Key: "timestamp", Value: bson.D{{Key: "$min", Value: "$timestamp"}}},
		}}}
	} else {
		// 构建常规分组ID
		groupID := bson.D{}
		for _, field := range groupByFields {
			groupID = append(groupID, bson.E{Key: field, Value: fmt.Sprintf("$%s", field)})
		}

		groupStage = bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: groupID},
			{Key: "requests", Value: bson.D{{Key: "$sum", Value: "$req_tot"}}},
			{Key: "timestamp", Value: bson.D{{Key: "$min", Value: "$timestamp"}}},
		}}}
	}

	sortStage := bson.D{{Key: "$sort", Value: bson.D{
		{Key: "timestamp", Value: 1}, // 按时间升序排序
	}}}

	pipeline := mongo.Pipeline{matchStage, groupStage, sortStage}

	// 执行聚合查询
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("执行请求时间序列聚合查询失败: %w", err)
	}
	defer cursor.Close(ctx)

	// 解析结果
	var results []struct {
		ID        bson.D    `bson:"_id"`
		Requests  int64     `bson:"requests"`
		Timestamp time.Time `bson:"timestamp"`
	}
	if err = cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("解析请求时间序列结果失败: %w", err)
	}

	// 构建数据点
	dataPoints := make([]dto.TimeSeriesDataPoint, 0, len(results))
	for _, result := range results {
		dataPoints = append(dataPoints, dto.TimeSeriesDataPoint{
			Timestamp: result.Timestamp,
			Value:     result.Requests,
		})
	}

	return dataPoints, nil
}

// 辅助方法 - 获取拦截数时间序列
func (s *StatsServiceImpl) getBlockTimeSeries(ctx context.Context, startTime time.Time, interval string, groupByFields []string) ([]dto.TimeSeriesDataPoint, error) {
	// 获取MongoDB数据库连接
	db, err := mongodb.GetDatabase(s.dbName)
	if err != nil {
		return nil, fmt.Errorf("获取数据库连接失败: %w", err)
	}

	collection := db.Collection("waf_log")

	// 构建时间过滤条件
	matchStage := bson.D{{Key: "$match", Value: bson.D{
		{Key: "createdAt", Value: bson.D{{Key: "$gte", Value: startTime}}},
	}}}

	// 根据interval决定如何分组
	var groupStage bson.D

	if interval == "hour" {
		// 按小时分组
		groupStage = bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{
				{Key: "date", Value: "$date"},
				{Key: "hour", Value: "$hour"},
			}},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
			{Key: "timestamp", Value: bson.D{{Key: "$min", Value: "$createdAt"}}},
		}}}
	} else if interval == "6hour" {
		// 使用预计算的HourGroupSix字段分组
		groupStage = bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{
				{Key: "date", Value: "$date"},
				{Key: "hourGroup", Value: "$hourGroupSix"},
			}},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
			{Key: "timestamp", Value: bson.D{{Key: "$min", Value: "$createdAt"}}},
		}}}
	} else {
		// 按日期分组
		groupStage = bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{{Key: "date", Value: "$date"}}},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
			{Key: "timestamp", Value: bson.D{{Key: "$min", Value: "$createdAt"}}},
		}}}
	}

	sortStage := bson.D{{Key: "$sort", Value: bson.D{
		{Key: "timestamp", Value: 1}, // 按时间升序排序
	}}}

	// 限制返回项目 - 优化性能
	projectStage := bson.D{{Key: "$project", Value: bson.D{
		{Key: "_id", Value: 0},
		{Key: "count", Value: 1},
		{Key: "timestamp", Value: 1},
	}}}

	pipeline := mongo.Pipeline{matchStage, groupStage, sortStage, projectStage}

	// 执行聚合查询
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("执行WAF拦截时间序列聚合查询失败: %w", err)
	}
	defer cursor.Close(ctx)

	// 解析结果
	var results []struct {
		Count     int64     `bson:"count"`
		Timestamp time.Time `bson:"timestamp"`
	}

	if err = cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("解析WAF拦截时间序列结果失败: %w", err)
	}

	// 构建数据点
	dataPoints := make([]dto.TimeSeriesDataPoint, 0, len(results))
	for _, result := range results {
		dataPoints = append(dataPoints, dto.TimeSeriesDataPoint{
			Timestamp: result.Timestamp,
			Value:     result.Count,
		})
	}

	return dataPoints, nil
}
