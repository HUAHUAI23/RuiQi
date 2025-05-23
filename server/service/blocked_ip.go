package service

import (
	"context"
	"errors"
	"math"

	"github.com/HUAHUAI23/simple-waf/pkg/model"
	"github.com/HUAHUAI23/simple-waf/server/config"
	"github.com/HUAHUAI23/simple-waf/server/dto"
	"github.com/HUAHUAI23/simple-waf/server/repository"
	"github.com/rs/zerolog"
)

var (
	ErrBlockedIPNotFound = errors.New("封禁IP记录不存在")
	ErrInvalidPageSize   = errors.New("无效的分页参数")
)

// BlockedIPService 封禁IP服务接口
type BlockedIPService interface {
	GetBlockedIPs(ctx context.Context, req *dto.BlockedIPListRequest) (*dto.BlockedIPListResponse, error)
	GetBlockedIPStats(ctx context.Context) (*dto.BlockedIPStatsResponse, error)
	CreateBlockedIP(ctx context.Context, record *model.BlockedIPRecord) error
	CleanupExpiredBlockedIPs(ctx context.Context) (int64, error)
}

// BlockedIPServiceImpl 封禁IP服务实现
type BlockedIPServiceImpl struct {
	blockedIPRepo repository.BlockedIPRepository
	logger        zerolog.Logger
}

// NewBlockedIPService 创建封禁IP服务
func NewBlockedIPService(blockedIPRepo repository.BlockedIPRepository) BlockedIPService {
	logger := config.GetServiceLogger("blocked_ip")
	return &BlockedIPServiceImpl{
		blockedIPRepo: blockedIPRepo,
		logger:        logger,
	}
}

// GetBlockedIPs 获取封禁IP列表
func (s *BlockedIPServiceImpl) GetBlockedIPs(ctx context.Context, req *dto.BlockedIPListRequest) (*dto.BlockedIPListResponse, error) {
	// 验证分页参数
	if err := s.validatePaginationParams(req); err != nil {
		return nil, err
	}

	// 设置默认值
	s.setDefaultParams(req)

	s.logger.Info().
		Int("page", req.Page).
		Int("size", req.Size).
		Str("ip", req.IP).
		Str("reason", req.Reason).
		Str("status", req.Status).
		Msg("获取封禁IP列表请求")

	// 调用仓库层
	records, total, err := s.blockedIPRepo.GetBlockedIPs(ctx, req)
	if err != nil {
		s.logger.Error().Err(err).Msg("获取封禁IP列表失败")
		return nil, err
	}

	// 转换为响应DTO
	items := dto.MapToResponseList(records)

	// 计算总页数
	pages := int(math.Ceil(float64(total) / float64(req.Size)))
	if pages == 0 {
		pages = 1
	}

	response := &dto.BlockedIPListResponse{
		Total: total,
		Items: items,
		Page:  req.Page,
		Size:  req.Size,
		Pages: pages,
	}

	s.logger.Info().
		Int64("total", total).
		Int("pages", pages).
		Int("items_count", len(items)).
		Msg("获取封禁IP列表成功")

	return response, nil
}

// GetBlockedIPStats 获取封禁IP统计信息
func (s *BlockedIPServiceImpl) GetBlockedIPStats(ctx context.Context) (*dto.BlockedIPStatsResponse, error) {
	s.logger.Info().Msg("获取封禁IP统计信息请求")

	stats, err := s.blockedIPRepo.GetBlockedIPStats(ctx)
	if err != nil {
		s.logger.Error().Err(err).Msg("获取封禁IP统计信息失败")
		return nil, err
	}

	s.logger.Info().
		Int64("total_blocked", stats.TotalBlocked).
		Int64("active_blocked", stats.ActiveBlocked).
		Int64("expired_blocked", stats.ExpiredBlocked).
		Int("reason_types", len(stats.ReasonStats)).
		Int("hourly_stats", len(stats.Last24HourStats)).
		Msg("获取封禁IP统计信息成功")

	return stats, nil
}

// CreateBlockedIP 创建封禁IP记录
func (s *BlockedIPServiceImpl) CreateBlockedIP(ctx context.Context, record *model.BlockedIPRecord) error {
	if record.IP == "" {
		return errors.New("IP地址不能为空")
	}

	if record.BlockedUntil.Before(record.BlockedAt) {
		return errors.New("封禁结束时间不能早于开始时间")
	}

	s.logger.Info().
		Str("ip", record.IP).
		Str("reason", record.Reason).
		Str("request_uri", record.RequestUri).
		Time("blocked_at", record.BlockedAt).
		Time("blocked_until", record.BlockedUntil).
		Msg("创建封禁IP记录请求")

	err := s.blockedIPRepo.CreateBlockedIP(ctx, record)
	if err != nil {
		s.logger.Error().Err(err).Str("ip", record.IP).Msg("创建封禁IP记录失败")
		return err
	}

	s.logger.Info().Str("ip", record.IP).Str("reason", record.Reason).Msg("封禁IP记录创建成功")
	return nil
}

// CleanupExpiredBlockedIPs 清理过期的封禁IP记录
func (s *BlockedIPServiceImpl) CleanupExpiredBlockedIPs(ctx context.Context) (int64, error) {
	s.logger.Info().Msg("开始清理过期封禁IP记录")

	deletedCount, err := s.blockedIPRepo.DeleteExpiredBlockedIPs(ctx)
	if err != nil {
		s.logger.Error().Err(err).Msg("清理过期封禁IP记录失败")
		return 0, err
	}

	s.logger.Info().Int64("deleted_count", deletedCount).Msg("清理过期封禁IP记录完成")
	return deletedCount, nil
}

// validatePaginationParams 验证分页参数
func (s *BlockedIPServiceImpl) validatePaginationParams(req *dto.BlockedIPListRequest) error {
	if req.Page < 0 {
		return ErrInvalidPageSize
	}
	if req.Size < 0 || req.Size > 100 {
		return ErrInvalidPageSize
	}
	return nil
}

// setDefaultParams 设置默认参数
func (s *BlockedIPServiceImpl) setDefaultParams(req *dto.BlockedIPListRequest) {
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Size == 0 {
		req.Size = 10
	}
	if req.Status == "" {
		req.Status = "all"
	}
	if req.SortBy == "" {
		req.SortBy = "blocked_at"
	}
	if req.SortDir == "" {
		req.SortDir = "desc"
	}
}
