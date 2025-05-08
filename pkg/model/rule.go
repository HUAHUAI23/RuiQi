package model

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

// RuleType 规则类型
//
//	@Description	规则类型，表示规则是白名单还是黑名单
type RuleType string

const (
	WhitelistRule RuleType = "whitelist" // 白名单规则
	BlacklistRule RuleType = "blacklist" // 黑名单规则
)

// RuleStatus 规则状态
//
//	@Description	规则状态，表示规则是启用还是禁用
type RuleStatus string

const (
	RuleEnabled  RuleStatus = "enabled"  // 规则已启用
	RuleDisabled RuleStatus = "disabled" // 规则已禁用
)

// MicroRule 表示WAF微规则信息
// @Description WAF微规则信息，包含规则名称、类型、状态、优先级和条件
type MicroRule struct {
	ID       bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty" example:"60d21b4367d0d8992e89e964"` // 规则唯一标识符
	Name     string        `json:"name" bson:"name" example:"SQL注入防护规则"`                                 // 规则名称
	Type     RuleType      `json:"type" bson:"type" example:"blacklist"`                                 // 规则类型
	Status   RuleStatus    `json:"status" bson:"status" example:"enabled"`                               // 规则状态
	Priority int           `json:"priority" bson:"priority" example:"100"`                               // 优先级字段，数字越大优先级越高
	// @Schema(type=object, example={"type":"composite","operator":"AND","conditions":[{"type":"simple","target":"source_ip","match_type":"in_ipgroup","match_value":"blocked_ips"},{"type":"simple","target":"path","match_type":"regex","match_value":"^/admin/.*$"}]})
	Condition bson.Raw `json:"condition" bson:"condition" swaggertype:"object"`
}

func (r *MicroRule) GetCollectionName() string {
	return "micro_rule"
}
