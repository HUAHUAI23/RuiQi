package dto

import (
	"time"

	"github.com/HUAHUAI23/simple-waf/pkg/model"
)

// BlockedIPListRequest 封禁IP列表请求参数
// @Description 获取封禁IP列表的请求参数
type BlockedIPListRequest struct {
	Page    int    `form:"page" binding:"omitempty,min=1" example:"1"`                                        // 页码
	Size    int    `form:"size" binding:"omitempty,min=1,max=100" example:"10"`                               // 每页数量
	IP      string `form:"ip" binding:"omitempty" example:"192.168.1.1"`                                      // IP地址过滤
	Reason  string `form:"reason" binding:"omitempty" example:"high_frequency_attack"`                        // 封禁原因过滤
	Status  string `form:"status" binding:"omitempty,oneof=active expired all" example:"active"`              // 状态过滤：active-生效中，expired-已过期，all-全部
	SortBy  string `form:"sortBy" binding:"omitempty,oneof=blocked_at blocked_until ip" example:"blocked_at"` // 排序字段
	SortDir string `form:"sortDir" binding:"omitempty,oneof=asc desc" example:"desc"`                         // 排序方向
}

// BlockedIPResponse 封禁IP响应
// @Description 封禁IP详细信息
type BlockedIPResponse struct {
	IP           string    `json:"ip" example:"192.168.1.1"`                    // 被封禁的IP地址
	Reason       string    `json:"reason" example:"high_frequency_attack"`      // 封禁原因
	RequestUri   string    `json:"requestUri" example:"/api/v1/login"`          // 请求URI
	BlockedAt    time.Time `json:"blockedAt" example:"2023-12-01T10:00:00Z"`    // 封禁开始时间
	BlockedUntil time.Time `json:"blockedUntil" example:"2023-12-01T11:00:00Z"` // 封禁结束时间
	IsActive     bool      `json:"isActive" example:"true"`                     // 是否仍在封禁中
	RemainingTTL int64     `json:"remainingTTL" example:"3600"`                 // 剩余封禁时间（秒）
}

// BlockedIPListResponse 封禁IP列表响应
// @Description 封禁IP分页列表响应
type BlockedIPListResponse struct {
	Total int64               `json:"total" example:"100"` // 总数量
	Items []BlockedIPResponse `json:"items"`               // IP列表
	Page  int                 `json:"page" example:"1"`    // 当前页码
	Size  int                 `json:"size" example:"10"`   // 每页数量
	Pages int                 `json:"pages" example:"10"`  // 总页数
}

// BlockedIPStatsResponse 封禁IP统计响应
// @Description 封禁IP统计信息
type BlockedIPStatsResponse struct {
	TotalBlocked    int64                  `json:"totalBlocked" example:"1000"`  // 总封禁数量
	ActiveBlocked   int64                  `json:"activeBlocked" example:"50"`   // 当前生效的封禁数量
	ExpiredBlocked  int64                  `json:"expiredBlocked" example:"950"` // 已过期的封禁数量
	ReasonStats     map[string]int64       `json:"reasonStats"`                  // 按原因统计
	Last24HourStats []BlockedIPHourlyStats `json:"last24HourStats"`              // 最近24小时统计
}

// BlockedIPHourlyStats 按小时统计
// @Description 按小时的封禁统计
type BlockedIPHourlyStats struct {
	Hour  string `json:"hour" example:"2023-12-01T10:00:00Z"` // 小时时间点
	Count int64  `json:"count" example:"10"`                  // 该小时的封禁数量
}

// MapToResponse 将模型转换为响应DTO
func (r *BlockedIPResponse) MapFromModel(record *model.BlockedIPRecord) {
	r.IP = record.IP
	r.Reason = record.Reason
	r.RequestUri = record.RequestUri
	r.BlockedAt = record.BlockedAt
	r.BlockedUntil = record.BlockedUntil

	// 计算是否仍在封禁中
	now := time.Now()
	r.IsActive = now.Before(record.BlockedUntil)

	// 计算剩余封禁时间
	if r.IsActive {
		r.RemainingTTL = int64(record.BlockedUntil.Sub(now).Seconds())
	} else {
		r.RemainingTTL = 0
	}
}

// MapToResponseList 将模型列表转换为响应DTO列表
func MapToResponseList(records []model.BlockedIPRecord) []BlockedIPResponse {
	responses := make([]BlockedIPResponse, len(records))
	for i, record := range records {
		responses[i].MapFromModel(&record)
	}
	return responses
}

// BlockedIPCleanupResponse 清理过期封禁IP记录响应
// @Description 清理过期封禁IP记录的响应数据
type BlockedIPCleanupResponse struct {
	DeletedCount int64  `json:"deletedCount" example:"25"`        // 删除的记录数量
	Message      string `json:"message" example:"已成功清理过期的封禁IP记录"` // 操作结果消息
}
