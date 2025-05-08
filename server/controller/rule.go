// server/controller/rule.go
package controller

import (
	"errors"
	"net/http"

	"github.com/HUAHUAI23/simple-waf/server/config"
	"github.com/HUAHUAI23/simple-waf/server/dto"
	"github.com/HUAHUAI23/simple-waf/server/model"
	"github.com/HUAHUAI23/simple-waf/server/service"
	"github.com/HUAHUAI23/simple-waf/server/utils/response"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// MicroRuleController 微规则控制器接口
type MicroRuleController interface {
	CreateMicroRule(ctx *gin.Context)
	GetMicroRules(ctx *gin.Context)
	GetMicroRuleByID(ctx *gin.Context)
	UpdateMicroRule(ctx *gin.Context)
	DeleteMicroRule(ctx *gin.Context)
}

// MicroRuleControllerImpl 微规则控制器实现
type MicroRuleControllerImpl struct {
	ruleService service.MicroRuleService
	logger      zerolog.Logger
}

// NewMicroRuleController 创建微规则控制器
func NewMicroRuleController(ruleService service.MicroRuleService) MicroRuleController {
	logger := config.GetControllerLogger("microrule")
	return &MicroRuleControllerImpl{
		ruleService: ruleService,
		logger:      logger,
	}
}

// CreateMicroRule 创建微规则
//
//	@Summary		创建微规则
//	@Description	创建一个新的WAF微规则，用于匹配和过滤请求
//	@Tags			规则管理
//	@Accept			json
//	@Produce		json
//	@Param			rule	body	dto.MicroRuleCreateRequest	true	"微规则信息"
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse{data=model.MicroRule}	"微规则创建成功"
//	@Failure		400	{object}	model.ErrResponse							"请求参数错误"
//	@Failure		401	{object}	model.ErrResponseDontShowError				"未授权访问"
//	@Failure		403	{object}	model.ErrResponseDontShowError				"禁止访问"
//	@Failure		409	{object}	model.ErrResponseDontShowError				"微规则名称已存在"
//	@Failure		500	{object}	model.ErrResponseDontShowError				"服务器内部错误"
//	@Router			/api/v1/micro-rules [post]
func (c *MicroRuleControllerImpl) CreateMicroRule(ctx *gin.Context) {
	var req dto.MicroRuleCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Warn().Err(err).Msg("请求参数绑定失败")
		response.BadRequest(ctx, err, true)
		return
	}

	c.logger.Info().Str("name", req.Name).Msg("创建微规则请求")
	rule, err := c.ruleService.CreateMicroRule(ctx, &req)
	if err != nil {
		if errors.Is(err, service.ErrMicroRuleNameExists) {
			response.Error(ctx, model.NewAPIError(http.StatusConflict, "微规则名称已存在", err), false)
			return
		}
		c.logger.Error().Err(err).Msg("创建微规则失败")
		response.InternalServerError(ctx, err, false)
		return
	}

	c.logger.Info().Str("id", rule.ID.Hex()).Str("name", rule.Name).Msg("微规则创建成功")
	response.Success(ctx, "微规则创建成功", rule)
}

// GetMicroRules 获取微规则列表
//
//	@Summary		获取微规则列表
//	@Description	获取所有WAF微规则列表，支持分页
//	@Tags			规则管理
//	@Produce		json
//	@Param			page	query	int	false	"页码"	default(1)
//	@Param			size	query	int	false	"每页数量"	default(10)
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse{data=model.MicroRule}	"获取微规则列表成功"
//	@Failure		401	{object}	model.ErrResponseDontShowError				"未授权访问"
//	@Failure		500	{object}	model.ErrResponseDontShowError				"服务器内部错误"
//	@Router			/api/v1/micro-rules [get]
func (c *MicroRuleControllerImpl) GetMicroRules(ctx *gin.Context) {
	page := ctx.DefaultQuery("page", "1")
	size := ctx.DefaultQuery("size", "10")

	c.logger.Info().Str("page", page).Str("size", size).Msg("获取微规则列表请求")
	rules, total, err := c.ruleService.GetMicroRules(ctx, page, size)
	if err != nil {
		c.logger.Error().Err(err).Msg("获取微规则列表失败")
		response.InternalServerError(ctx, err, false)
		return
	}

	c.logger.Info().Int64("total", total).Msg("获取微规则列表成功")
	response.Success(ctx, "获取微规则列表成功", gin.H{
		"total": total,
		"items": rules,
	})
}

// GetMicroRuleByID 获取单个微规则
//
//	@Summary		获取单个微规则
//	@Description	根据ID获取微规则详情
//	@Tags			规则管理
//	@Produce		json
//	@Param			id	path	string	true	"微规则ID"
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse{data=model.MicroRule}	"获取微规则详情成功"
//	@Failure		400	{object}	model.ErrResponse							"无效的ID格式"
//	@Failure		401	{object}	model.ErrResponseDontShowError				"未授权访问"
//	@Failure		404	{object}	model.ErrResponseDontShowError				"微规则不存在"
//	@Failure		500	{object}	model.ErrResponseDontShowError				"服务器内部错误"
//	@Router			/api/v1/micro-rules/{id} [get]
func (c *MicroRuleControllerImpl) GetMicroRuleByID(ctx *gin.Context) {
	id := ctx.Param("id")

	c.logger.Info().Str("id", id).Msg("获取微规则详情请求")

	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		c.logger.Error().Err(err).Str("id", id).Msg("无效的ID格式")
		response.BadRequest(ctx, err, true)
		return
	}
	rule, err := c.ruleService.GetMicroRuleByID(ctx, objectID)
	if err != nil {
		if errors.Is(err, service.ErrMicroRuleNotFound) {
			response.NotFound(ctx, err)
			return
		}
		c.logger.Error().Err(err).Str("id", id).Msg("获取微规则详情失败")
		response.InternalServerError(ctx, err, false)
		return
	}

	c.logger.Info().Str("id", id).Str("name", rule.Name).Msg("获取微规则详情成功")
	response.Success(ctx, "获取微规则详情成功", rule)
}

// UpdateMicroRule 更新微规则
//
//	@Summary		更新微规则
//	@Description	更新指定微规则的信息，系统默认规则不允许修改
//	@Tags			规则管理
//	@Accept			json
//	@Produce		json
//	@Param			id		path	string						true	"微规则ID"
//	@Param			rule	body	dto.MicroRuleUpdateRequest	true	"微规则更新信息"
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse{data=model.MicroRule}	"微规则更新成功"
//	@Failure		400	{object}	model.ErrResponse							"请求参数错误"
//	@Failure		401	{object}	model.ErrResponseDontShowError				"未授权访问"
//	@Failure		403	{object}	model.ErrResponseDontShowError				"禁止修改系统默认规则"
//	@Failure		404	{object}	model.ErrResponseDontShowError				"微规则不存在"
//	@Failure		409	{object}	model.ErrResponseDontShowError				"微规则名称已存在"
//	@Failure		500	{object}	model.ErrResponseDontShowError				"服务器内部错误"
//	@Router			/api/v1/micro-rules/{id} [put]
func (c *MicroRuleControllerImpl) UpdateMicroRule(ctx *gin.Context) {
	id := ctx.Param("id")
	var req dto.MicroRuleUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Warn().Err(err).Str("id", id).Msg("请求参数绑定失败")
		response.BadRequest(ctx, err, true)
		return
	}

	c.logger.Info().Str("id", id).Msg("更新微规则请求")
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		c.logger.Error().Err(err).Str("id", id).Msg("无效的ID格式")
		response.BadRequest(ctx, err, true)
		return
	}
	rule, err := c.ruleService.UpdateMicroRule(ctx, objectID, &req)
	if err != nil {
		if errors.Is(err, service.ErrMicroRuleNotFound) {
			response.NotFound(ctx, err)
			return
		} else if errors.Is(err, service.ErrMicroRuleNameExists) {
			response.Error(ctx, model.NewAPIError(http.StatusConflict, "微规则名称已存在", err), false)
			return
		} else if errors.Is(err, service.ErrSystemRuleNoMod) {
			response.Error(ctx, model.NewAPIError(http.StatusForbidden, "系统默认规则不允许修改", err), false)
			return
		}
		c.logger.Error().Err(err).Str("id", id).Msg("更新微规则失败")
		response.InternalServerError(ctx, err, false)
		return
	}

	c.logger.Info().Str("id", id).Str("name", rule.Name).Msg("微规则更新成功")
	response.Success(ctx, "微规则更新成功", rule)
}

// DeleteMicroRule 删除微规则
//
//	@Summary		删除微规则
//	@Description	删除指定的微规则，系统默认规则不允许删除
//	@Tags			规则管理
//	@Produce		json
//	@Param			id	path	string	true	"微规则ID"
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponseNoData		"微规则删除成功"
//	@Failure		400	{object}	model.ErrResponse				"无效的ID格式"
//	@Failure		401	{object}	model.ErrResponseDontShowError	"未授权访问"
//	@Failure		403	{object}	model.ErrResponseDontShowError	"禁止删除系统默认规则"
//	@Failure		404	{object}	model.ErrResponseDontShowError	"微规则不存在"
//	@Failure		500	{object}	model.ErrResponseDontShowError	"服务器内部错误"
//	@Router			/api/v1/micro-rules/{id} [delete]
func (c *MicroRuleControllerImpl) DeleteMicroRule(ctx *gin.Context) {
	id := ctx.Param("id")

	c.logger.Info().Str("id", id).Msg("删除微规则请求")
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		c.logger.Error().Err(err).Str("id", id).Msg("无效的ID格式")
		response.BadRequest(ctx, err, true)
		return
	}
	err = c.ruleService.DeleteMicroRule(ctx, objectID)
	if err != nil {
		if errors.Is(err, service.ErrMicroRuleNotFound) {
			response.NotFound(ctx, err)
			return
		} else if errors.Is(err, service.ErrSystemRuleNoDelete) {
			response.Error(ctx, model.NewAPIError(http.StatusForbidden, "系统默认规则不允许删除", err), false)
			return
		}
		c.logger.Error().Err(err).Str("id", id).Msg("删除微规则失败")
		response.InternalServerError(ctx, err, false)
		return
	}

	c.logger.Info().Str("id", id).Msg("微规则删除成功")
	response.Success(ctx, "微规则删除成功", nil)
}
