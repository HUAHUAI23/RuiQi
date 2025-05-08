// server/dto/rule.go
package dto

import (
	"encoding/json"

	"github.com/HUAHUAI23/simple-waf/pkg/model"
)

// MicroRuleCreateRequest 创建微规则请求
// @Description 创建微规则的请求参数
type MicroRuleCreateRequest struct {
	Name      string          `json:"name" binding:"required" example:"SQL注入防护规则"`                           // 规则名称
	Type      string          `json:"type" binding:"required,oneof=whitelist blacklist" example:"blacklist"` // 规则类型
	Status    string          `json:"status" binding:"required,oneof=enabled disabled" example:"enabled"`    // 规则状态
	Priority  int             `json:"priority" binding:"required" example:"100"`                             // 优先级字段，数字越大优先级越高
	Condition json.RawMessage `json:"condition" binding:"required" swaggertype:"object"`                     // 规则条件
}

// MicroRuleUpdateRequest 更新微规则请求
// @Description 更新微规则的请求参数
type MicroRuleUpdateRequest struct {
	Name      string          `json:"name,omitempty" example:"SQL注入防护规则"`                                               // 规则名称
	Type      string          `json:"type,omitempty" binding:"omitempty,oneof=whitelist blacklist" example:"blacklist"` // 规则类型
	Status    string          `json:"status,omitempty" binding:"omitempty,oneof=enabled disabled" example:"enabled"`    // 规则状态
	Priority  *int            `json:"priority,omitempty" example:"100"`                                                 // 优先级字段，数字越大优先级越高
	Condition json.RawMessage `json:"condition,omitempty" swaggertype:"object"`                                         // 规则条件
}

// MicroRuleListResponse 微规则列表响应
// @Description 微规则列表响应
type MicroRuleListResponse struct {
	Total int64             `json:"total"` // 总数
	Items []model.MicroRule `json:"items"` // 微规则列表
}
