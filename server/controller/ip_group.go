// server/controller/ip_group.go
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

// IPGroupController IP组控制器接口
type IPGroupController interface {
	CreateIPGroup(ctx *gin.Context)
	GetIPGroups(ctx *gin.Context)
	GetIPGroupByID(ctx *gin.Context)
	UpdateIPGroup(ctx *gin.Context)
	DeleteIPGroup(ctx *gin.Context)
	AddIPToBlacklist(ctx *gin.Context)
}

// IPGroupControllerImpl IP组控制器实现
type IPGroupControllerImpl struct {
	ipGroupService service.IPGroupService
	logger         zerolog.Logger
}

// NewIPGroupController 创建IP组控制器
func NewIPGroupController(ipGroupService service.IPGroupService) IPGroupController {
	logger := config.GetControllerLogger("ipgroup")
	return &IPGroupControllerImpl{
		ipGroupService: ipGroupService,
		logger:         logger,
	}
}

// CreateIPGroup 创建IP组
//
//	@Summary		创建IP组
//	@Description	创建一个新的IP地址组，用于后续IP规则匹配
//	@Tags			IP组管理
//	@Accept			json
//	@Produce		json
//	@Param			ipGroup	body	dto.IPGroupCreateRequest	true	"IP组信息"
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse{data=model.IPGroup}	"IP组创建成功"
//	@Failure		400	{object}	model.ErrResponse							"请求参数错误"
//	@Failure		401	{object}	model.ErrResponseDontShowError				"未授权访问"
//	@Failure		403	{object}	model.ErrResponseDontShowError				"禁止访问"
//	@Failure		409	{object}	model.ErrResponseDontShowError				"IP组名称已存在"
//	@Failure		500	{object}	model.ErrResponseDontShowError				"服务器内部错误"
//	@Router			/api/v1/ip-groups [post]
func (c *IPGroupControllerImpl) CreateIPGroup(ctx *gin.Context) {
	var req dto.IPGroupCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Warn().Err(err).Msg("请求参数绑定失败")
		response.BadRequest(ctx, err, true)
		return
	}

	c.logger.Info().Str("name", req.Name).Msg("创建IP组请求")
	ipGroup, err := c.ipGroupService.CreateIPGroup(ctx, &req)
	if err != nil {
		if errors.Is(err, service.ErrIPGroupNameExists) {
			response.Error(ctx, model.NewAPIError(http.StatusConflict, "IP组名称已存在", err), false)
			return
		}
		c.logger.Error().Err(err).Msg("创建IP组失败")
		response.InternalServerError(ctx, err, false)
		return
	}

	c.logger.Info().Str("id", ipGroup.ID.Hex()).Str("name", ipGroup.Name).Msg("IP组创建成功")
	response.Success(ctx, "IP组创建成功", ipGroup)
}

// GetIPGroups 获取IP组列表
//
//	@Summary		获取IP组列表
//	@Description	获取所有IP组列表，支持分页
//	@Tags			IP组管理
//	@Produce		json
//	@Param			page	query	int	false	"页码"	default(1)
//	@Param			size	query	int	false	"每页数量"	default(10)
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse{data=dto.IPGroupListResponse}	"获取IP组列表成功"
//	@Failure		401	{object}	model.ErrResponseDontShowError						"未授权访问"
//	@Failure		500	{object}	model.ErrResponseDontShowError						"服务器内部错误"
//	@Router			/api/v1/ip-groups [get]
func (c *IPGroupControllerImpl) GetIPGroups(ctx *gin.Context) {
	page := ctx.DefaultQuery("page", "1")
	size := ctx.DefaultQuery("size", "10")

	c.logger.Info().Str("page", page).Str("size", size).Msg("获取IP组列表请求")
	ipGroups, total, err := c.ipGroupService.GetIPGroups(ctx, page, size)
	if err != nil {
		c.logger.Error().Err(err).Msg("获取IP组列表失败")
		response.InternalServerError(ctx, err, false)
		return
	}

	c.logger.Info().Int64("total", total).Msg("获取IP组列表成功")
	response.Success(ctx, "获取IP组列表成功", gin.H{
		"total": total,
		"items": ipGroups,
	})
}

// GetIPGroupByID 获取单个IP组
//
//	@Summary		获取单个IP组
//	@Description	根据ID获取IP组详情
//	@Tags			IP组管理
//	@Produce		json
//	@Param			id	path	string	true	"IP组ID"
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse{data=model.IPGroup}	"获取IP组详情成功"
//	@Failure		400	{object}	model.ErrResponse							"无效的ID格式"
//	@Failure		401	{object}	model.ErrResponseDontShowError				"未授权访问"
//	@Failure		404	{object}	model.ErrResponseDontShowError				"IP组不存在"
//	@Failure		500	{object}	model.ErrResponseDontShowError				"服务器内部错误"
//	@Router			/api/v1/ip-groups/{id} [get]
func (c *IPGroupControllerImpl) GetIPGroupByID(ctx *gin.Context) {
	id := ctx.Param("id")

	c.logger.Info().Str("id", id).Msg("获取IP组详情请求")

	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		c.logger.Error().Err(err).Str("id", id).Msg("无效的ID格式")
		response.BadRequest(ctx, err, true)
		return
	}
	ipGroup, err := c.ipGroupService.GetIPGroupByID(ctx, objectID)
	if err != nil {
		if errors.Is(err, service.ErrIPGroupNotFound) {
			response.NotFound(ctx, err)
			return
		}
		c.logger.Error().Err(err).Str("id", id).Msg("获取IP组详情失败")
		response.InternalServerError(ctx, err, false)
		return
	}

	c.logger.Info().Str("id", id).Str("name", ipGroup.Name).Msg("获取IP组详情成功")
	response.Success(ctx, "获取IP组详情成功", ipGroup)
}

// UpdateIPGroup 更新IP组
//
//	@Summary		更新IP组
//	@Description	更新指定IP组的信息
//	@Tags			IP组管理
//	@Accept			json
//	@Produce		json
//	@Param			id		path	string						true	"IP组ID"
//	@Param			ipGroup	body	dto.IPGroupUpdateRequest	true	"IP组更新信息"
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse{data=model.IPGroup}	"IP组更新成功"
//	@Failure		400	{object}	model.ErrResponse							"请求参数错误"
//	@Failure		401	{object}	model.ErrResponseDontShowError				"未授权访问"
//	@Failure		403	{object}	model.ErrResponseDontShowError				"禁止操作系统默认IP组"
//	@Failure		404	{object}	model.ErrResponseDontShowError				"IP组不存在"
//	@Failure		409	{object}	model.ErrResponseDontShowError				"IP组名称已存在"
//	@Failure		500	{object}	model.ErrResponseDontShowError				"服务器内部错误"
//	@Router			/api/v1/ip-groups/{id} [put]
func (c *IPGroupControllerImpl) UpdateIPGroup(ctx *gin.Context) {
	id := ctx.Param("id")
	var req dto.IPGroupUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Warn().Err(err).Str("id", id).Msg("请求参数绑定失败")
		response.BadRequest(ctx, err, true)
		return
	}

	c.logger.Info().Str("id", id).Msg("更新IP组请求")
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		c.logger.Error().Err(err).Str("id", id).Msg("无效的ID格式")
		response.BadRequest(ctx, err, true)
		return
	}
	ipGroup, err := c.ipGroupService.UpdateIPGroup(ctx, objectID, &req)
	if err != nil {
		if errors.Is(err, service.ErrIPGroupNotFound) {
			response.NotFound(ctx, err)
			return
		} else if errors.Is(err, service.ErrIPGroupNameExists) {
			response.Error(ctx, model.NewAPIError(http.StatusConflict, "IP组名称已存在", err), false)
			return
		} else if errors.Is(err, service.ErrSystemIPGroupNoMod) {
			response.Error(ctx, model.NewAPIError(http.StatusForbidden, "系统默认IP组不允许修改名称", err), false)
			return
		}
		c.logger.Error().Err(err).Str("id", id).Msg("更新IP组失败")
		response.InternalServerError(ctx, err, false)
		return
	}

	c.logger.Info().Str("id", id).Str("name", ipGroup.Name).Msg("IP组更新成功")
	response.Success(ctx, "IP组更新成功", ipGroup)
}

// DeleteIPGroup 删除IP组
//
//	@Summary		删除IP组
//	@Description	删除指定的IP组，系统默认IP组不允许删除
//	@Tags			IP组管理
//	@Produce		json
//	@Param			id	path	string	true	"IP组ID"
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponseNoData		"IP组删除成功"
//	@Failure		400	{object}	model.ErrResponse				"无效的ID格式"
//	@Failure		401	{object}	model.ErrResponseDontShowError	"未授权访问"
//	@Failure		403	{object}	model.ErrResponseDontShowError	"禁止删除系统默认IP组"
//	@Failure		404	{object}	model.ErrResponseDontShowError	"IP组不存在"
//	@Failure		500	{object}	model.ErrResponseDontShowError	"服务器内部错误"
//	@Router			/api/v1/ip-groups/{id} [delete]
func (c *IPGroupControllerImpl) DeleteIPGroup(ctx *gin.Context) {
	id := ctx.Param("id")

	c.logger.Info().Str("id", id).Msg("删除IP组请求")
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		c.logger.Error().Err(err).Str("id", id).Msg("无效的ID格式")
		response.BadRequest(ctx, err, true)
		return
	}
	err = c.ipGroupService.DeleteIPGroup(ctx, objectID)
	if err != nil {
		if errors.Is(err, service.ErrIPGroupNotFound) {
			response.NotFound(ctx, err)
			return
		} else if errors.Is(err, service.ErrSystemIPGroupNoMod) {
			response.Error(ctx, model.NewAPIError(http.StatusForbidden, "系统默认IP组不允许删除", err), false)
			return
		}
		c.logger.Error().Err(err).Str("id", id).Msg("删除IP组失败")
		response.InternalServerError(ctx, err, false)
		return
	}

	c.logger.Info().Str("id", id).Msg("IP组删除成功")
	response.Success(ctx, "IP组删除成功", nil)
}

// AddIPToBlacklist 添加IP到系统默认黑名单
//
//	@Summary		添加IP到黑名单
//	@Description	将指定的IP地址或CIDR添加到系统默认黑名单组中
//	@Tags			IP组管理
//	@Accept			json
//	@Produce		json
//	@Param			request	body	dto.AddIPToBlacklistRequest	true	"IP地址或CIDR"
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponseNoData		"IP添加成功"
//	@Failure		400	{object}	model.ErrResponse				"请求参数错误"
//	@Failure		401	{object}	model.ErrResponseDontShowError	"未授权访问"
//	@Failure		404	{object}	model.ErrResponseDontShowError	"系统默认黑名单不存在"
//	@Failure		500	{object}	model.ErrResponseDontShowError	"服务器内部错误"
//	@Router			/api/v1/ip-groups/blacklist/add [post]
func (c *IPGroupControllerImpl) AddIPToBlacklist(ctx *gin.Context) {
	var req dto.AddIPToBlacklistRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Warn().Err(err).Msg("请求参数绑定失败")
		response.BadRequest(ctx, err, true)
		return
	}

	c.logger.Info().Str("ip", req.IP).Msg("添加IP到黑名单请求")
	err := c.ipGroupService.AddIPToBlacklist(ctx, req.IP)
	if err != nil {
		if errors.Is(err, service.ErrIPGroupNotFound) {
			response.NotFound(ctx, err)
			return
		}
		c.logger.Error().Err(err).Str("ip", req.IP).Msg("添加IP到黑名单失败")
		response.InternalServerError(ctx, err, false)
		return
	}

	c.logger.Info().Str("ip", req.IP).Msg("IP添加到黑名单成功")
	response.Success(ctx, "IP添加到黑名单成功", nil)
}
