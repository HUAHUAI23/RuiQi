package controller

import (
	"errors"

	"github.com/HUAHUAI23/simple-waf/server/config"
	"github.com/HUAHUAI23/simple-waf/server/dto"
	"github.com/HUAHUAI23/simple-waf/server/service"
	"github.com/HUAHUAI23/simple-waf/server/utils/response"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// BlockedIPController 封禁IP控制器接口
type BlockedIPController interface {
	GetBlockedIPs(ctx *gin.Context)
	GetBlockedIPStats(ctx *gin.Context)
	CleanupExpiredBlockedIPs(ctx *gin.Context)
}

// BlockedIPControllerImpl 封禁IP控制器实现
type BlockedIPControllerImpl struct {
	blockedIPService service.BlockedIPService
	logger           zerolog.Logger
}

// NewBlockedIPController 创建封禁IP控制器
func NewBlockedIPController(blockedIPService service.BlockedIPService) BlockedIPController {
	logger := config.GetControllerLogger("blocked_ip")
	return &BlockedIPControllerImpl{
		blockedIPService: blockedIPService,
		logger:           logger,
	}
}

// GetBlockedIPs 获取封禁IP列表
//
//	@Summary		获取封禁IP列表
//	@Description	获取被封禁的IP地址列表，支持分页、过滤和排序
//	@Tags			封禁IP管理
//	@Produce		json
//	@Param			page	query	int		false	"页码，从1开始"								default(1)	minimum(1)
//	@Param			size	query	int		false	"每页数量，最大100"							default(10)	minimum(1)	maximum(100)
//	@Param			ip		query	string	false	"IP地址过滤，支持模糊匹配"							example(192.168.1.1)
//	@Param			reason	query	string	false	"封禁原因过滤"								example(high_frequency_attack)
//	@Param			status	query	string	false	"状态过滤：active-生效中，expired-已过期，all-全部"	default(all)		Enums(active, expired, all)
//	@Param			sortBy	query	string	false	"排序字段"									default(blocked_at)	Enums(blocked_at, blocked_until, ip)
//	@Param			sortDir	query	string	false	"排序方向：asc-升序，desc-降序"					default(desc)		Enums(asc, desc)
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse{data=dto.BlockedIPListResponse}	"获取封禁IP列表成功"
//	@Failure		400	{object}	model.ErrResponse										"请求参数错误"
//	@Failure		401	{object}	model.ErrResponseDontShowError							"未授权访问"
//	@Failure		403	{object}	model.ErrResponseDontShowError							"禁止访问"
//	@Failure		500	{object}	model.ErrResponseDontShowError							"服务器内部错误"
//	@Router			/api/v1/blocked-ips [get]
func (c *BlockedIPControllerImpl) GetBlockedIPs(ctx *gin.Context) {
	var req dto.BlockedIPListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		c.logger.Warn().Err(err).Msg("请求参数绑定失败")
		response.BadRequest(ctx, err, true)
		return
	}

	c.logger.Info().
		Int("page", req.Page).
		Int("size", req.Size).
		Str("ip", req.IP).
		Str("reason", req.Reason).
		Str("status", req.Status).
		Str("sortBy", req.SortBy).
		Str("sortDir", req.SortDir).
		Msg("获取封禁IP列表请求")

	result, err := c.blockedIPService.GetBlockedIPs(ctx, &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidPageSize) {
			response.BadRequest(ctx, err, true)
			return
		}
		c.logger.Error().Err(err).Msg("获取封禁IP列表失败")
		response.InternalServerError(ctx, err, false)
		return
	}

	c.logger.Info().
		Int64("total", result.Total).
		Int("items_count", len(result.Items)).
		Int("page", result.Page).
		Int("pages", result.Pages).
		Msg("获取封禁IP列表成功")

	response.Success(ctx, "获取封禁IP列表成功", result)
}

// GetBlockedIPStats 获取封禁IP统计信息
//
//	@Summary		获取封禁IP统计信息
//	@Description	获取封禁IP的统计信息，包括总数、生效数、过期数、按原因统计和按小时统计
//	@Tags			封禁IP管理
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse{data=dto.BlockedIPStatsResponse}	"获取统计信息成功"
//	@Failure		401	{object}	model.ErrResponseDontShowError							"未授权访问"
//	@Failure		403	{object}	model.ErrResponseDontShowError							"禁止访问"
//	@Failure		500	{object}	model.ErrResponseDontShowError							"服务器内部错误"
//	@Router			/api/v1/blocked-ips/stats [get]
func (c *BlockedIPControllerImpl) GetBlockedIPStats(ctx *gin.Context) {
	c.logger.Info().Msg("获取封禁IP统计信息请求")

	stats, err := c.blockedIPService.GetBlockedIPStats(ctx)
	if err != nil {
		c.logger.Error().Err(err).Msg("获取封禁IP统计信息失败")
		response.InternalServerError(ctx, err, false)
		return
	}

	c.logger.Info().
		Int64("total_blocked", stats.TotalBlocked).
		Int64("active_blocked", stats.ActiveBlocked).
		Int64("expired_blocked", stats.ExpiredBlocked).
		Int("reason_types", len(stats.ReasonStats)).
		Int("hourly_stats", len(stats.Last24HourStats)).
		Msg("获取封禁IP统计信息成功")

	response.Success(ctx, "获取统计信息成功", stats)
}

// CleanupExpiredBlockedIPs 清理过期的封禁IP记录
//
//	@Summary		清理过期的封禁IP记录
//	@Description	删除已过期的封禁IP记录，释放存储空间
//	@Tags			封禁IP管理
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse{data=dto.BlockedIPCleanupResponse}	"清理完成"
//	@Failure		401	{object}	model.ErrResponseDontShowError								"未授权访问"
//	@Failure		403	{object}	model.ErrResponseDontShowError								"禁止访问"
//	@Failure		500	{object}	model.ErrResponseDontShowError								"服务器内部错误"
//	@Router			/api/v1/blocked-ips/cleanup [delete]
func (c *BlockedIPControllerImpl) CleanupExpiredBlockedIPs(ctx *gin.Context) {
	c.logger.Info().Msg("清理过期封禁IP记录请求")

	deletedCount, err := c.blockedIPService.CleanupExpiredBlockedIPs(ctx)
	if err != nil {
		c.logger.Error().Err(err).Msg("清理过期封禁IP记录失败")
		response.InternalServerError(ctx, err, false)
		return
	}

	c.logger.Info().Int64("deleted_count", deletedCount).Msg("清理过期封禁IP记录成功")

	cleanupResponse := dto.BlockedIPCleanupResponse{
		DeletedCount: deletedCount,
		Message:      "已成功清理过期的封禁IP记录",
	}

	response.Success(ctx, "清理完成", cleanupResponse)
}
