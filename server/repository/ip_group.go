// server/repository/ip_group.go
package repository

import (
	"context"
	"errors"
	"time"

	"github.com/HUAHUAI23/simple-waf/pkg/model"
	"github.com/HUAHUAI23/simple-waf/server/config"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	ErrIPGroupNotFound = errors.New("IP组不存在")
)

// IPGroupRepository IP组仓库接口
type IPGroupRepository interface {
	CreateIPGroup(ctx context.Context, ipGroup *model.IPGroup) error
	GetIPGroups(ctx context.Context, page, size int64) ([]model.IPGroup, int64, error)
	GetIPGroupByID(ctx context.Context, id bson.ObjectID) (*model.IPGroup, error)
	GetIPGroupByName(ctx context.Context, name string) (*model.IPGroup, error)
	UpdateIPGroup(ctx context.Context, ipGroup *model.IPGroup) error
	DeleteIPGroup(ctx context.Context, id bson.ObjectID) error
	CheckIPGroupNameExists(ctx context.Context, name string, excludeID bson.ObjectID) (bool, error)
}

// MongoIPGroupRepository MongoDB实现的IP组仓库
type MongoIPGroupRepository struct {
	collection *mongo.Collection
	logger     zerolog.Logger
}

// NewIPGroupRepository 创建IP组仓库
func NewIPGroupRepository(db *mongo.Database) IPGroupRepository {
	var ipGroup model.IPGroup
	collection := db.Collection(ipGroup.GetCollectionName())
	logger := config.GetRepositoryLogger("ipgroup")

	// 创建索引
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// IP组名称唯一索引
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		logger.Error().Err(err).Msg("创建IP组名称索引失败")
	}

	return &MongoIPGroupRepository{
		collection: collection,
		logger:     logger,
	}
}

// CreateIPGroup 创建IP组
func (r *MongoIPGroupRepository) CreateIPGroup(ctx context.Context, ipGroup *model.IPGroup) error {
	// 插入新IP组
	result, err := r.collection.InsertOne(ctx, ipGroup)
	if err != nil {
		r.logger.Error().Err(err).Str("name", ipGroup.Name).Msg("插入IP组时出错")
		return err
	}

	ipGroup.ID = result.InsertedID.(bson.ObjectID)
	return nil
}

// GetIPGroups 获取IP组列表
func (r *MongoIPGroupRepository) GetIPGroups(ctx context.Context, page, size int64) ([]model.IPGroup, int64, error) {
	// 计算分页
	skip := (page - 1) * size

	// 设置查询选项
	findOptions := options.Find().
		SetSkip(skip).
		SetLimit(size).
		SetSort(bson.D{{Key: "name", Value: 1}}) // 按名称升序排序

	// 执行查询
	cursor, err := r.collection.Find(ctx, bson.D{}, findOptions)
	if err != nil {
		r.logger.Error().Err(err).Msg("查询IP组列表时出错")
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	// 解析结果
	var ipGroups []model.IPGroup
	if err = cursor.All(ctx, &ipGroups); err != nil {
		r.logger.Error().Err(err).Msg("解析IP组列表时出错")
		return nil, 0, err
	}

	// 获取总数
	total, err := r.collection.CountDocuments(ctx, bson.D{})
	if err != nil {
		r.logger.Error().Err(err).Msg("获取IP组总数时出错")
		return nil, 0, err
	}

	return ipGroups, total, nil
}

// GetIPGroupByID 根据ID获取IP组
func (r *MongoIPGroupRepository) GetIPGroupByID(ctx context.Context, id bson.ObjectID) (*model.IPGroup, error) {
	var ipGroup model.IPGroup
	err := r.collection.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&ipGroup)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrIPGroupNotFound
		}
		r.logger.Error().Err(err).Str("id", id.Hex()).Msg("查询IP组时出错")
		return nil, err
	}

	return &ipGroup, nil
}

// GetIPGroupByName 根据名称获取IP组
func (r *MongoIPGroupRepository) GetIPGroupByName(ctx context.Context, name string) (*model.IPGroup, error) {
	var ipGroup model.IPGroup
	err := r.collection.FindOne(ctx, bson.D{{Key: "name", Value: name}}).Decode(&ipGroup)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrIPGroupNotFound
		}
		r.logger.Error().Err(err).Str("name", name).Msg("按名称查询IP组时出错")
		return nil, err
	}

	return &ipGroup, nil
}

// UpdateIPGroup 更新IP组
func (r *MongoIPGroupRepository) UpdateIPGroup(ctx context.Context, ipGroup *model.IPGroup) error {
	// 更新IP组
	_, err := r.collection.ReplaceOne(
		ctx,
		bson.D{{Key: "_id", Value: ipGroup.ID}},
		ipGroup,
	)

	if err != nil {
		r.logger.Error().Err(err).Str("id", ipGroup.ID.Hex()).Msg("更新IP组时出错")
		return err
	}

	return nil
}

// DeleteIPGroup 删除IP组
func (r *MongoIPGroupRepository) DeleteIPGroup(ctx context.Context, id bson.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.D{{Key: "_id", Value: id}})
	if err != nil {
		r.logger.Error().Err(err).Str("id", id.Hex()).Msg("删除IP组时出错")
		return err
	}

	if result.DeletedCount == 0 {
		return ErrIPGroupNotFound
	}

	return nil
}

// CheckIPGroupNameExists 检查IP组名称是否已存在
func (r *MongoIPGroupRepository) CheckIPGroupNameExists(ctx context.Context, name string, excludeID bson.ObjectID) (bool, error) {
	filter := bson.D{{Key: "name", Value: name}}

	// 如果是更新操作，需要排除当前IP组ID
	if excludeID != bson.NilObjectID {
		filter = append(filter, bson.E{Key: "_id", Value: bson.D{{Key: "$ne", Value: excludeID}}})
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		r.logger.Error().Err(err).Str("name", name).Msg("检查IP组名称是否存在时出错")
		return false, err
	}

	return count > 0, nil
}
