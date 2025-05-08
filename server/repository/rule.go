// server/repository/rule.go
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
	ErrRuleNotFound = errors.New("规则不存在")
)

// MicroRuleRepository 微规则仓库接口
type MicroRuleRepository interface {
	CreateMicroRule(ctx context.Context, rule *model.MicroRule) error
	GetMicroRules(ctx context.Context, page, size int64) ([]model.MicroRule, int64, error)
	GetMicroRuleByID(ctx context.Context, id bson.ObjectID) (*model.MicroRule, error)
	GetMicroRuleByName(ctx context.Context, name string) (*model.MicroRule, error)
	UpdateMicroRule(ctx context.Context, rule *model.MicroRule) error
	DeleteMicroRule(ctx context.Context, id bson.ObjectID) error
	CheckMicroRuleNameExists(ctx context.Context, name string, excludeID bson.ObjectID) (bool, error)
}

// MongoMicroRuleRepository MongoDB实现的微规则仓库
type MongoMicroRuleRepository struct {
	collection *mongo.Collection
	logger     zerolog.Logger
}

// NewMicroRuleRepository 创建微规则仓库
func NewMicroRuleRepository(db *mongo.Database) MicroRuleRepository {
	var rule model.MicroRule
	collection := db.Collection(rule.GetCollectionName())
	logger := config.GetRepositoryLogger("microrule")

	// 创建索引
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 规则名称唯一索引
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		logger.Error().Err(err).Msg("创建规则名称索引失败")
	}

	return &MongoMicroRuleRepository{
		collection: collection,
		logger:     logger,
	}
}

// CreateMicroRule 创建微规则
func (r *MongoMicroRuleRepository) CreateMicroRule(ctx context.Context, rule *model.MicroRule) error {
	// 插入新微规则
	result, err := r.collection.InsertOne(ctx, rule)
	if err != nil {
		r.logger.Error().Err(err).Str("name", rule.Name).Msg("插入微规则时出错")
		return err
	}

	rule.ID = result.InsertedID.(bson.ObjectID)
	return nil
}

// GetMicroRules 获取微规则列表
func (r *MongoMicroRuleRepository) GetMicroRules(ctx context.Context, page, size int64) ([]model.MicroRule, int64, error) {
	// 计算分页
	skip := (page - 1) * size

	// 设置查询选项，按优先级降序排序
	findOptions := options.Find().
		SetSkip(skip).
		SetLimit(size).
		SetSort(bson.D{{Key: "priority", Value: -1}})

	// 执行查询
	cursor, err := r.collection.Find(ctx, bson.D{}, findOptions)
	if err != nil {
		r.logger.Error().Err(err).Msg("查询微规则列表时出错")
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	// 解析结果
	var rules []model.MicroRule
	if err = cursor.All(ctx, &rules); err != nil {
		r.logger.Error().Err(err).Msg("解析微规则列表时出错")
		return nil, 0, err
	}

	// 获取总数
	total, err := r.collection.CountDocuments(ctx, bson.D{})
	if err != nil {
		r.logger.Error().Err(err).Msg("获取微规则总数时出错")
		return nil, 0, err
	}

	return rules, total, nil
}

// GetMicroRuleByID 根据ID获取微规则
func (r *MongoMicroRuleRepository) GetMicroRuleByID(ctx context.Context, id bson.ObjectID) (*model.MicroRule, error) {
	var rule model.MicroRule
	err := r.collection.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&rule)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrRuleNotFound
		}
		r.logger.Error().Err(err).Str("id", id.Hex()).Msg("查询微规则时出错")
		return nil, err
	}

	return &rule, nil
}

// GetMicroRuleByName 根据名称获取微规则
func (r *MongoMicroRuleRepository) GetMicroRuleByName(ctx context.Context, name string) (*model.MicroRule, error) {
	var rule model.MicroRule
	err := r.collection.FindOne(ctx, bson.D{{Key: "name", Value: name}}).Decode(&rule)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrRuleNotFound
		}
		r.logger.Error().Err(err).Str("name", name).Msg("按名称查询微规则时出错")
		return nil, err
	}

	return &rule, nil
}

// UpdateMicroRule 更新微规则
func (r *MongoMicroRuleRepository) UpdateMicroRule(ctx context.Context, rule *model.MicroRule) error {
	// 更新微规则
	_, err := r.collection.ReplaceOne(
		ctx,
		bson.D{{Key: "_id", Value: rule.ID}},
		rule,
	)

	if err != nil {
		r.logger.Error().Err(err).Str("id", rule.ID.Hex()).Msg("更新微规则时出错")
		return err
	}

	return nil
}

// DeleteMicroRule 删除微规则
func (r *MongoMicroRuleRepository) DeleteMicroRule(ctx context.Context, id bson.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.D{{Key: "_id", Value: id}})
	if err != nil {
		r.logger.Error().Err(err).Str("id", id.Hex()).Msg("删除微规则时出错")
		return err
	}

	if result.DeletedCount == 0 {
		return ErrRuleNotFound
	}

	return nil
}

// CheckMicroRuleNameExists 检查微规则名称是否已存在
func (r *MongoMicroRuleRepository) CheckMicroRuleNameExists(ctx context.Context, name string, excludeID bson.ObjectID) (bool, error) {
	filter := bson.D{{Key: "name", Value: name}}

	// 如果是更新操作，需要排除当前规则ID
	if excludeID != bson.NilObjectID {
		filter = append(filter, bson.E{Key: "_id", Value: bson.D{{Key: "$ne", Value: excludeID}}})
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		r.logger.Error().Err(err).Str("name", name).Msg("检查微规则名称是否存在时出错")
		return false, err
	}

	return count > 0, nil
}
