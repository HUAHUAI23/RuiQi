// server/service/rule.go
package service

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/HUAHUAI23/simple-waf/pkg/model"
	"github.com/HUAHUAI23/simple-waf/server/config"
	"github.com/HUAHUAI23/simple-waf/server/dto"
	"github.com/HUAHUAI23/simple-waf/server/repository"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	SystemDefaultIPBlockRule = "system_default_ip_block" // 系统默认IP阻止规则名称
)

var (
	ErrMicroRuleNotFound   = errors.New("微规则不存在")
	ErrMicroRuleNameExists = errors.New("微规则名称已存在")
	ErrSystemRuleNoMod     = errors.New("系统默认规则不允许修改")
	ErrSystemRuleNoDelete  = errors.New("系统默认规则不允许删除")
)

// MicroRuleService 微规则服务接口
type MicroRuleService interface {
	CreateMicroRule(ctx context.Context, req *dto.MicroRuleCreateRequest) (*model.MicroRule, error)
	GetMicroRules(ctx context.Context, pageStr, sizeStr string) ([]model.MicroRule, int64, error)
	GetMicroRuleByID(ctx context.Context, id bson.ObjectID) (*model.MicroRule, error)
	UpdateMicroRule(ctx context.Context, id bson.ObjectID, req *dto.MicroRuleUpdateRequest) (*model.MicroRule, error)
	DeleteMicroRule(ctx context.Context, id bson.ObjectID) error
}

// MicroRuleServiceImpl 微规则服务实现
type MicroRuleServiceImpl struct {
	ruleRepo repository.MicroRuleRepository
	logger   zerolog.Logger
}

// NewMicroRuleService 创建微规则服务
func NewMicroRuleService(ruleRepo repository.MicroRuleRepository) MicroRuleService {
	logger := config.GetServiceLogger("microrule")
	return &MicroRuleServiceImpl{
		ruleRepo: ruleRepo,
		logger:   logger,
	}
}

// CreateMicroRule 创建微规则
func (s *MicroRuleServiceImpl) CreateMicroRule(ctx context.Context, req *dto.MicroRuleCreateRequest) (*model.MicroRule, error) {
	// 检查规则名称是否已存在
	if req.Name != "" {
		exists, err := s.ruleRepo.CheckMicroRuleNameExists(ctx, req.Name, bson.NilObjectID)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrMicroRuleNameExists
		}
	}

	// 将JSON条件转换为BSON
	var condition bson.Raw
	if len(req.Condition) > 0 {
		// 使用JSON解析器将JSON解析为interface{}
		var anyValue interface{}
		if err := json.Unmarshal(req.Condition, &anyValue); err != nil {
			s.logger.Error().Err(err).Msg("解析JSON条件失败")
			return nil, err
		}

		// 将interface{}转换为BSON
		bsonData, err := bson.Marshal(anyValue)
		if err != nil {
			s.logger.Error().Err(err).Msg("转换条件为BSON失败")
			return nil, err
		}

		condition = bsonData
	}

	// 创建新微规则
	rule := &model.MicroRule{
		Name:      req.Name,
		Type:      model.RuleType(req.Type),
		Status:    model.RuleStatus(req.Status),
		Priority:  req.Priority,
		Condition: condition,
	}

	// 保存微规则
	err := s.ruleRepo.CreateMicroRule(ctx, rule)
	if err != nil {
		s.logger.Error().Err(err).Msg("创建微规则失败")
		return nil, err
	}

	s.logger.Info().Str("id", rule.ID.Hex()).Str("name", rule.Name).Msg("微规则创建成功")
	return rule, nil
}

// GetMicroRules 获取微规则列表
func (s *MicroRuleServiceImpl) GetMicroRules(ctx context.Context, pageStr, sizeStr string) ([]model.MicroRule, int64, error) {
	page, err := strconv.ParseInt(pageStr, 10, 64)
	if err != nil || page < 1 {
		page = 1
	}

	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil || size < 1 {
		size = 10
	}

	rules, total, err := s.ruleRepo.GetMicroRules(ctx, page, size)
	if err != nil {
		s.logger.Error().Err(err).Msg("获取微规则列表失败")
		return nil, 0, err
	}

	return rules, total, nil
}

// GetMicroRuleByID 根据ID获取微规则
func (s *MicroRuleServiceImpl) GetMicroRuleByID(ctx context.Context, id bson.ObjectID) (*model.MicroRule, error) {
	rule, err := s.ruleRepo.GetMicroRuleByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrRuleNotFound) {
			return nil, ErrMicroRuleNotFound
		}
		s.logger.Error().Err(err).Str("id", id.Hex()).Msg("获取微规则失败")
		return nil, err
	}

	return rule, nil
}

// UpdateMicroRule 更新微规则
func (s *MicroRuleServiceImpl) UpdateMicroRule(ctx context.Context, id bson.ObjectID, req *dto.MicroRuleUpdateRequest) (*model.MicroRule, error) {
	// 获取现有微规则
	rule, err := s.ruleRepo.GetMicroRuleByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrRuleNotFound) {
			return nil, ErrMicroRuleNotFound
		}
		return nil, err
	}

	// 检查是否是系统默认规则
	if rule.Name == SystemDefaultIPBlockRule {
		s.logger.Warn().Str("id", id.Hex()).Msg("尝试修改系统默认规则")
		return nil, ErrSystemRuleNoMod
	}

	// 检查规则名称是否已存在（如果要更新名称）
	if req.Name != "" && req.Name != rule.Name {
		exists, err := s.ruleRepo.CheckMicroRuleNameExists(ctx, req.Name, id)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrMicroRuleNameExists
		}
		rule.Name = req.Name
	}

	// 更新其他字段（只更新非空字段）
	if req.Type != "" {
		rule.Type = model.RuleType(req.Type)
	}
	if req.Status != "" {
		rule.Status = model.RuleStatus(req.Status)
	}
	if req.Priority != nil {
		rule.Priority = *req.Priority
	}
	if len(req.Condition) > 0 {
		// 使用JSON解析器将JSON解析为interface{}
		var anyValue interface{}
		if err := json.Unmarshal(req.Condition, &anyValue); err != nil {
			s.logger.Error().Err(err).Msg("解析JSON条件失败")
			return nil, err
		}

		// 将interface{}转换为BSON
		bsonData, err := bson.Marshal(anyValue)
		if err != nil {
			s.logger.Error().Err(err).Msg("转换条件为BSON失败")
			return nil, err
		}

		rule.Condition = bsonData
	}

	// 保存更新
	err = s.ruleRepo.UpdateMicroRule(ctx, rule)
	if err != nil {
		s.logger.Error().Err(err).Str("id", id.Hex()).Msg("更新微规则失败")
		return nil, err
	}

	s.logger.Info().Str("id", id.Hex()).Str("name", rule.Name).Msg("微规则更新成功")
	return rule, nil
}

// DeleteMicroRule 删除微规则
func (s *MicroRuleServiceImpl) DeleteMicroRule(ctx context.Context, id bson.ObjectID) error {
	// 检查微规则是否存在
	rule, err := s.ruleRepo.GetMicroRuleByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrRuleNotFound) {
			return ErrMicroRuleNotFound
		}
		return err
	}

	// 检查是否是系统默认规则
	if rule.Name == SystemDefaultIPBlockRule {
		s.logger.Warn().Str("id", id.Hex()).Msg("尝试删除系统默认规则")
		return ErrSystemRuleNoDelete
	}

	// 删除微规则
	err = s.ruleRepo.DeleteMicroRule(ctx, id)
	if err != nil {
		s.logger.Error().Err(err).Str("id", id.Hex()).Msg("删除微规则失败")
		return err
	}

	s.logger.Info().Str("id", id.Hex()).Msg("微规则删除成功")
	return nil
}
