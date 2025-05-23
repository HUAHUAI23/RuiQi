// server/controller/config.go
package controller

import (
	"errors"

	"github.com/HUAHUAI23/simple-waf/pkg/model"
	"github.com/HUAHUAI23/simple-waf/server/config"
	"github.com/HUAHUAI23/simple-waf/server/dto"
	"github.com/HUAHUAI23/simple-waf/server/service"
	"github.com/HUAHUAI23/simple-waf/server/utils/response"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// ConfigController 配置控制器接口
type ConfigController interface {
	GetConfig(ctx *gin.Context)
	PatchConfig(ctx *gin.Context)
}

// ConfigControllerImpl 配置控制器实现
type ConfigControllerImpl struct {
	configService service.ConfigService
	logger        zerolog.Logger
}

// NewConfigController 创建配置控制器
func NewConfigController(configService service.ConfigService) ConfigController {
	logger := config.GetControllerLogger("config")
	return &ConfigControllerImpl{
		configService: configService,
		logger:        logger,
	}
}

// GetConfig 获取配置
//
//	@Summary		获取系统配置
//	@Description	获取当前系统配置信息
//	@Tags			配置管理
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse{data=dto.ConfigResponse}	"获取配置成功"
//	@Failure		401	{object}	model.ErrResponseDontShowError					"未授权访问"
//	@Failure		403	{object}	model.ErrResponseDontShowError					"禁止访问"
//	@Failure		404	{object}	model.ErrResponseDontShowError					"配置不存在"
//	@Failure		500	{object}	model.ErrResponseDontShowError					"服务器内部错误"
//	@Router			/api/v1/config [get]
func (c *ConfigControllerImpl) GetConfig(ctx *gin.Context) {
	cfg, err := c.configService.GetConfig(ctx)
	if err != nil {
		if errors.Is(err, service.ErrConfigNotFound) {
			response.NotFound(ctx, err)
			return
		}
		c.logger.Error().Err(err).Msg("获取配置失败")
		response.InternalServerError(ctx, err, false)
		return
	}

	// 转换为DTO响应
	configResponse := mapConfigToDTO(cfg)

	response.Success(ctx, "获取配置成功", configResponse)
}

// PatchConfig 补丁更新配置
//
//	@Summary		更新系统配置
//	@Description	使用补丁方式更新系统配置
//	@Tags			配置管理
//	@Accept			json
//	@Produce		json
//	@Param			config	body	dto.ConfigPatchRequest	true	"配置更新信息"
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse{data=dto.ConfigResponse}	"配置更新成功"
//	@Failure		400	{object}	model.ErrResponse								"请求参数错误"
//	@Failure		401	{object}	model.ErrResponseDontShowError					"未授权访问"
//	@Failure		403	{object}	model.ErrResponseDontShowError					"禁止访问"
//	@Failure		404	{object}	model.ErrResponseDontShowError					"配置不存在"
//	@Failure		500	{object}	model.ErrResponseDontShowError					"服务器内部错误"
//	@Router			/api/v1/config [patch]
func (c *ConfigControllerImpl) PatchConfig(ctx *gin.Context) {
	var req dto.ConfigPatchRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Warn().Err(err).Msg("请求参数绑定失败")
		response.BadRequest(ctx, err, true)
		return
	}

	cfg, err := c.configService.PatchConfig(ctx, &req)
	if err != nil {
		if errors.Is(err, service.ErrConfigNotFound) {
			response.NotFound(ctx, err)
			return
		}
		c.logger.Error().Err(err).Msg("更新配置失败")
		response.InternalServerError(ctx, err, false)
		return
	}

	// 转换为DTO响应
	configResponse := mapConfigToDTO(cfg)

	response.Success(ctx, "配置更新成功", configResponse)
}

// mapConfigToDTO 将模型转换为DTO
func mapConfigToDTO(cfg *model.Config) dto.ConfigResponse {
	// 将配置模型转换为响应DTO
	engineDTO := dto.EngineDTO{
		Bind:            cfg.Engine.Bind,
		UseBuiltinRules: cfg.Engine.UseBuiltinRules,
		ASNDBPath:       cfg.Engine.ASNDBPath,
		CityDBPath:      cfg.Engine.CityDBPath,
		AppConfig:       make([]dto.AppConfigDTO, len(cfg.Engine.AppConfig)),
		FlowController: dto.FlowControllerDTO{
			VisitLimit: dto.LimitConfigDTO{
				Enabled:        cfg.Engine.FlowController.VisitLimit.Enabled,
				Threshold:      cfg.Engine.FlowController.VisitLimit.Threshold,
				StatDuration:   cfg.Engine.FlowController.VisitLimit.StatDuration,
				BlockDuration:  cfg.Engine.FlowController.VisitLimit.BlockDuration,
				BurstCount:     cfg.Engine.FlowController.VisitLimit.BurstCount,
				ParamsCapacity: cfg.Engine.FlowController.VisitLimit.ParamsCapacity,
			},
			AttackLimit: dto.LimitConfigDTO{
				Enabled:        cfg.Engine.FlowController.AttackLimit.Enabled,
				Threshold:      cfg.Engine.FlowController.AttackLimit.Threshold,
				StatDuration:   cfg.Engine.FlowController.AttackLimit.StatDuration,
				BlockDuration:  cfg.Engine.FlowController.AttackLimit.BlockDuration,
				BurstCount:     cfg.Engine.FlowController.AttackLimit.BurstCount,
				ParamsCapacity: cfg.Engine.FlowController.AttackLimit.ParamsCapacity,
			},
			ErrorLimit: dto.LimitConfigDTO{
				Enabled:        cfg.Engine.FlowController.ErrorLimit.Enabled,
				Threshold:      cfg.Engine.FlowController.ErrorLimit.Threshold,
				StatDuration:   cfg.Engine.FlowController.ErrorLimit.StatDuration,
				BlockDuration:  cfg.Engine.FlowController.ErrorLimit.BlockDuration,
				BurstCount:     cfg.Engine.FlowController.ErrorLimit.BurstCount,
				ParamsCapacity: cfg.Engine.FlowController.ErrorLimit.ParamsCapacity,
			},
		},
	}

	// 转换应用配置
	for i, app := range cfg.Engine.AppConfig {
		engineDTO.AppConfig[i] = dto.AppConfigDTO{
			Name:           app.Name,
			Directives:     app.Directives,
			TransactionTTL: dto.DurationToMillis(app.TransactionTTL),
			LogLevel:       app.LogLevel,
			LogFile:        app.LogFile,
			LogFormat:      app.LogFormat,
		}
	}

	// 转换Haproxy配置
	haproxyDTO := dto.HaproxyDTO{
		ConfigBaseDir: cfg.Haproxy.ConfigBaseDir,
		HaproxyBin:    cfg.Haproxy.HaproxyBin,
		BackupsNumber: cfg.Haproxy.BackupsNumber,
		SpoeAgentAddr: cfg.Haproxy.SpoeAgentAddr,
		SpoeAgentPort: cfg.Haproxy.SpoeAgentPort,
		Thread:        cfg.Haproxy.Thread,
	}

	return dto.ConfigResponse{
		Name:            cfg.Name,
		Engine:          engineDTO,
		Haproxy:         haproxyDTO,
		CreatedAt:       cfg.CreatedAt,
		UpdatedAt:       cfg.UpdatedAt,
		IsResponseCheck: cfg.IsResponseCheck,
		IsDebug:         cfg.IsDebug,
		IsK8s:           cfg.IsK8s,
	}
}
