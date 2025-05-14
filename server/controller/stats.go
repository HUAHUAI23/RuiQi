package controller

import (
	"strconv"

	"github.com/HUAHUAI23/simple-waf/server/config"
	"github.com/HUAHUAI23/simple-waf/server/dto"
	"github.com/HUAHUAI23/simple-waf/server/service"
	"github.com/HUAHUAI23/simple-waf/server/utils/response"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type StatsController interface {
	GetStats(ctx *gin.Context)
	GetOverviewStats(ctx *gin.Context)
	GetRealtimeQPS(ctx *gin.Context)
	GetTimeSeriesData(ctx *gin.Context)
}

type StatsControllerImpl struct {
	runnerService service.RunnerService
	statsService  service.StatsService
	logger        zerolog.Logger
}

func NewStatsController(runnerService service.RunnerService, statsService service.StatsService) StatsController {
	logger := config.GetControllerLogger("stats")
	return &StatsControllerImpl{
		runnerService: runnerService,
		statsService:  statsService,
		logger:        logger,
	}
}

// GetStats 获取原始统计数据
// @Summary 获取HAProxy原始统计数据
// @Description 获取HAProxy原始的统计信息
// @Tags 统计信息
// @Produce json
// @Security BearerAuth
// @Success 200 {object} model.SuccessResponse "获取统计数据成功"
// @Failure 401 {object} model.ErrResponseDontShowError "未授权访问"
// @Failure 500 {object} model.ErrResponseDontShowError "服务器内部错误"
// @Router /stats [get]
func (c *StatsControllerImpl) GetStats(ctx *gin.Context) {
	stats, err := c.runnerService.GetStats()
	if err != nil {
		c.logger.Error().Err(err).Msg("获取HAProxy统计数据失败")
		response.InternalServerError(ctx, err, false)
		return
	}
	response.Success(ctx, "获取统计数据成功", stats)
}

// GetOverviewStats 获取统计概览数据
// @Summary 获取统计概览数据
// @Description 获取指定时间范围内的统计概览数据，包括请求数、流量、错误率等
// @Tags 统计信息
// @Produce json
// @Param timeRange query string true "时间范围：24h(24小时)、7d(7天)、30d(30天)" Enums(24h, 7d, 30d) default(24h)
// @Security BearerAuth
// @Success 200 {object} model.SuccessResponse{data=dto.OverviewStats} "获取统计概览成功"
// @Failure 400 {object} model.ErrResponse "请求参数错误"
// @Failure 401 {object} model.ErrResponseDontShowError "未授权访问"
// @Failure 500 {object} model.ErrResponseDontShowError "服务器内部错误"
// @Router /api/v1/stats/overview [get]
func (c *StatsControllerImpl) GetOverviewStats(ctx *gin.Context) {
	// 解析请求参数
	var req dto.StatsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		c.logger.Warn().Err(err).Msg("绑定统计概览请求参数失败")
		response.BadRequest(ctx, err, true)
		return
	}

	// 调用服务
	stats, err := c.statsService.GetOverviewStats(ctx, req.TimeRange)
	if err != nil {
		c.logger.Error().Err(err).Str("timeRange", req.TimeRange).Msg("获取统计概览数据失败")
		response.InternalServerError(ctx, err, false)
		return
	}

	response.Success(ctx, "获取统计概览成功", stats)
}

// GetRealtimeQPS 获取实时QPS数据
// @Summary 获取实时QPS数据
// @Description 获取最近的实时QPS数据点
// @Tags 统计信息
// @Produce json
// @Param limit query int false "返回的数据点数量，默认30个点，最大60个点" default(30) minimum(1) maximum(60)
// @Security BearerAuth
// @Success 200 {object} model.SuccessResponse{data=dto.RealtimeQPSResponse} "获取实时QPS数据成功"
// @Failure 400 {object} model.ErrResponse "请求参数错误"
// @Failure 401 {object} model.ErrResponseDontShowError "未授权访问"
// @Failure 500 {object} model.ErrResponseDontShowError "服务器内部错误"
// @Router /api/v1/stats/realtime-qps [get]
func (c *StatsControllerImpl) GetRealtimeQPS(ctx *gin.Context) {
	// 解析请求参数
	limitStr := ctx.DefaultQuery("limit", "30")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 30 // 默认30个点
	}
	if limit > 60 {
		limit = 60 // 最多60个点
	}

	// 调用服务
	data, err := c.statsService.GetRealtimeQPS(ctx, limit)
	if err != nil {
		c.logger.Error().Err(err).Int("limit", limit).Msg("获取实时QPS数据失败")
		response.InternalServerError(ctx, err, false)
		return
	}

	response.Success(ctx, "获取实时QPS数据成功", data)
}

// GetTimeSeriesData 获取时间序列数据
// @Summary 获取时间序列数据
// @Description 获取指定时间范围和指标类型的时间序列数据，用于图表展示
// @Tags 统计信息
// @Produce json
// @Param timeRange query string true "时间范围：24h(24小时)、7d(7天)、30d(30天)" Enums(24h, 7d, 30d) default(24h)
// @Param metric query string true "指标类型：requests(请求数)、blocks(拦截数)" Enums(requests, blocks) default(requests)
// @Security BearerAuth
// @Success 200 {object} model.SuccessResponse{data=dto.TimeSeriesResponse} "获取时间序列数据成功"
// @Failure 400 {object} model.ErrResponse "请求参数错误"
// @Failure 401 {object} model.ErrResponseDontShowError "未授权访问"
// @Failure 500 {object} model.ErrResponseDontShowError "服务器内部错误"
// @Router /api/v1/stats/time-series [get]
func (c *StatsControllerImpl) GetTimeSeriesData(ctx *gin.Context) {
	// 解析请求参数
	var req dto.TimeSeriesDataRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		c.logger.Warn().Err(err).Msg("绑定时间序列请求参数失败")
		response.BadRequest(ctx, err, true)
		return
	}

	// 检查请求参数
	if req.TimeRange == "" {
		req.TimeRange = dto.TimeRange24Hours // 默认24小时
	}

	if req.Metric == "" {
		req.Metric = "requests" // 默认请求数
	}

	// 调用服务
	data, err := c.statsService.GetTimeSeriesData(ctx, req.TimeRange, req.Metric)
	if err != nil {
		c.logger.Error().Err(err).
			Str("timeRange", req.TimeRange).
			Str("metric", req.Metric).
			Msg("获取时间序列数据失败")
		response.InternalServerError(ctx, err, false)
		return
	}

	response.Success(ctx, "获取时间序列数据成功", data)
}
