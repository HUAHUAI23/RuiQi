// server/dto/ip_group.go
package dto

import (
	"github.com/HUAHUAI23/simple-waf/pkg/model"
)

// IPGroupCreateRequest IP组创建请求
// @Description 创建IP组的请求参数
type IPGroupCreateRequest struct {
	Name  string   `json:"name" binding:"required" example:"内部服务器"`              // IP组名称
	Items []string `json:"items" binding:"required" example:"[\"192.168.1.1\"]"` // IP地址或CIDR列表
}

// IPGroupUpdateRequest IP组更新请求
// @Description 更新IP组的请求参数
type IPGroupUpdateRequest struct {
	Name  string   `json:"name,omitempty" example:"内部服务器"`              // IP组名称
	Items []string `json:"items,omitempty" example:"[\"192.168.1.1\"]"` // IP地址或CIDR列表
}

// IPGroupListResponse IP组列表响应
// @Description IP组列表响应
type IPGroupListResponse struct {
	Total int64           `json:"total"` // 总数
	Items []model.IPGroup `json:"items"` // IP组列表
}
