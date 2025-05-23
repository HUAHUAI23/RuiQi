package repository

import (
	"context"
	"errors"
	"time"

	"github.com/HUAHUAI23/simple-waf/pkg/model"
	"github.com/HUAHUAI23/simple-waf/server/config"
	"github.com/HUAHUAI23/simple-waf/server/dto"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	ErrBlockedIPNotFound = errors.New("封禁IP记录不存在")
)

// BlockedIPRepository 封禁IP仓库接口
type BlockedIPRepository interface {
	GetBlockedIPs(ctx context.Context, req *dto.BlockedIPListRequest) ([]model.BlockedIPRecord, int64, error)
	GetBlockedIPStats(ctx context.Context) (*dto.BlockedIPStatsResponse, error)
	CreateBlockedIP(ctx context.Context, record *model.BlockedIPRecord) error
	DeleteExpiredBlockedIPs(ctx context.Context) (int64, error)
}

// MongoBlockedIPRepository MongoDB实现的封禁IP仓库
type MongoBlockedIPRepository struct {
	collection *mongo.Collection
	logger     zerolog.Logger
}

// NewBlockedIPRepository 创建封禁IP仓库
func NewBlockedIPRepository(db *mongo.Database) BlockedIPRepository {
	var blockedIP model.BlockedIPRecord
	collection := db.Collection(blockedIP.GetCollectionName())
	logger := config.GetRepositoryLogger("blocked_ip")

	// 创建索引
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// IP地址索引
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "ip", Value: 1}},
	})
	if err != nil {
		logger.Error().Err(err).Msg("创建IP地址索引失败")
	}

	// 封禁时间索引（用于过期清理）
	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "blocked_until", Value: 1}},
	})
	if err != nil {
		logger.Error().Err(err).Msg("创建封禁时间索引失败")
	}

	// 复合索引：封禁原因和时间
	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "reason", Value: 1},
			{Key: "blocked_at", Value: -1},
		},
	})
	if err != nil {
		logger.Error().Err(err).Msg("创建复合索引失败")
	}

	return &MongoBlockedIPRepository{
		collection: collection,
		logger:     logger,
	}
}

// GetBlockedIPs 获取封禁IP列表
func (r *MongoBlockedIPRepository) GetBlockedIPs(ctx context.Context, req *dto.BlockedIPListRequest) ([]model.BlockedIPRecord, int64, error) {
	// 构建查询过滤器
	filter := r.buildFilter(req)

	// 计算分页
	page := req.Page
	if page < 1 {
		page = 1
	}
	size := req.Size
	if size < 1 {
		size = 10
	} else if size > 100 {
		size = 100
	}
	skip := int64((page - 1) * size)

	// 构建排序
	sort := r.buildSort(req)

	// 设置查询选项
	findOptions := options.Find().
		SetSkip(skip).
		SetLimit(int64(size)).
		SetSort(sort)

	// 执行查询
	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		r.logger.Error().Err(err).Msg("查询封禁IP列表时出错")
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	// 解析结果
	var records []model.BlockedIPRecord
	if err = cursor.All(ctx, &records); err != nil {
		r.logger.Error().Err(err).Msg("解析封禁IP列表时出错")
		return nil, 0, err
	}

	// 获取总数
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		r.logger.Error().Err(err).Msg("获取封禁IP总数时出错")
		return nil, 0, err
	}

	return records, total, nil
}

// GetBlockedIPStats 获取封禁IP统计信息
func (r *MongoBlockedIPRepository) GetBlockedIPStats(ctx context.Context) (*dto.BlockedIPStatsResponse, error) {
	now := time.Now()

	stats := &dto.BlockedIPStatsResponse{
		ReasonStats: make(map[string]int64),
	}

	// 获取总封禁数量
	total, err := r.collection.CountDocuments(ctx, bson.D{})
	if err != nil {
		r.logger.Error().Err(err).Msg("获取总封禁数量失败")
		return nil, err
	}
	stats.TotalBlocked = total

	// 获取当前生效的封禁数量
	activeFilter := bson.D{{Key: "blocked_until", Value: bson.D{{Key: "$gt", Value: now}}}}
	active, err := r.collection.CountDocuments(ctx, activeFilter)
	if err != nil {
		r.logger.Error().Err(err).Msg("获取生效封禁数量失败")
		return nil, err
	}
	stats.ActiveBlocked = active
	stats.ExpiredBlocked = total - active

	// 按原因统计
	reasonPipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$reason"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
	}

	cursor, err := r.collection.Aggregate(ctx, reasonPipeline)
	if err != nil {
		r.logger.Error().Err(err).Msg("按原因统计失败")
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var result struct {
			ID    string `bson:"_id"`
			Count int64  `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			r.logger.Error().Err(err).Msg("解析原因统计结果失败")
			continue
		}
		stats.ReasonStats[result.ID] = result.Count
	}

	// 最近24小时按小时统计
	last24Hours := now.Add(-24 * time.Hour)
	hourlyStats, err := r.getHourlyStats(ctx, last24Hours, now)
	if err != nil {
		r.logger.Error().Err(err).Msg("获取小时统计失败")
		// 不返回错误，只记录日志
	} else {
		stats.Last24HourStats = hourlyStats
	}

	return stats, nil
}

// CreateBlockedIP 创建封禁IP记录
func (r *MongoBlockedIPRepository) CreateBlockedIP(ctx context.Context, record *model.BlockedIPRecord) error {
	_, err := r.collection.InsertOne(ctx, record)
	if err != nil {
		r.logger.Error().Err(err).Str("ip", record.IP).Msg("插入封禁IP记录时出错")
		return err
	}
	return nil
}

// DeleteExpiredBlockedIPs 删除过期的封禁IP记录
func (r *MongoBlockedIPRepository) DeleteExpiredBlockedIPs(ctx context.Context) (int64, error) {
	now := time.Now()
	filter := bson.D{{Key: "blocked_until", Value: bson.D{{Key: "$lt", Value: now}}}}

	result, err := r.collection.DeleteMany(ctx, filter)
	if err != nil {
		r.logger.Error().Err(err).Msg("删除过期封禁IP记录时出错")
		return 0, err
	}

	r.logger.Info().Int64("count", result.DeletedCount).Msg("已删除过期封禁IP记录")
	return result.DeletedCount, nil
}

// buildFilter 构建查询过滤器
func (r *MongoBlockedIPRepository) buildFilter(req *dto.BlockedIPListRequest) bson.D {
	filter := bson.D{}

	// IP地址过滤
	if req.IP != "" {
		filter = append(filter, bson.E{Key: "ip", Value: bson.D{{Key: "$regex", Value: req.IP}, {Key: "$options", Value: "i"}}})
	}

	// 封禁原因过滤
	if req.Reason != "" {
		filter = append(filter, bson.E{Key: "reason", Value: req.Reason})
	}

	// 状态过滤
	now := time.Now()
	switch req.Status {
	case "active":
		filter = append(filter, bson.E{Key: "blocked_until", Value: bson.D{{Key: "$gt", Value: now}}})
	case "expired":
		filter = append(filter, bson.E{Key: "blocked_until", Value: bson.D{{Key: "$lte", Value: now}}})
	case "all", "":
		// 不添加时间过滤
	}

	return filter
}

// buildSort 构建排序
func (r *MongoBlockedIPRepository) buildSort(req *dto.BlockedIPListRequest) bson.D {
	sortBy := req.SortBy
	if sortBy == "" {
		sortBy = "blocked_at"
	}

	sortDir := 1 // 升序
	if req.SortDir == "desc" {
		sortDir = -1 // 降序
	}

	return bson.D{{Key: sortBy, Value: sortDir}}
}

// getHourlyStats 获取按小时的统计数据
func (r *MongoBlockedIPRepository) getHourlyStats(ctx context.Context, start, end time.Time) ([]dto.BlockedIPHourlyStats, error) {
	pipeline := mongo.Pipeline{
		// 过滤时间范围
		{{Key: "$match", Value: bson.D{
			{Key: "blocked_at", Value: bson.D{
				{Key: "$gte", Value: start},
				{Key: "$lt", Value: end},
			}},
		}}},
		// 按小时分组
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{
				{Key: "$dateToString", Value: bson.D{
					{Key: "format", Value: "%Y-%m-%dT%H:00:00Z"},
					{Key: "date", Value: "$blocked_at"},
				}},
			}},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
		// 排序
		{{Key: "$sort", Value: bson.D{{Key: "_id", Value: 1}}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var stats []dto.BlockedIPHourlyStats
	for cursor.Next(ctx) {
		var result struct {
			ID    string `bson:"_id"`
			Count int64  `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			r.logger.Error().Err(err).Msg("解析小时统计结果失败")
			continue
		}

		stats = append(stats, dto.BlockedIPHourlyStats{
			Hour:  result.ID,
			Count: result.Count,
		})
	}

	return stats, nil
}
