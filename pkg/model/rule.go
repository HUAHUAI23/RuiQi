package model

import (
	"encoding/json"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// 规则类型
type RuleType string

const (
	WhitelistRule RuleType = "whitelist"
	BlacklistRule RuleType = "blacklist"
)

// 规则状态
type RuleStatus string

const (
	RuleEnabled  RuleStatus = "enabled"
	RuleDisabled RuleStatus = "disabled"
)

type MicroRule struct {
	ID        bson.ObjectID   `bson:"_id,omitempty" json:"id,omitempty"` // 规则唯一标识符
	Name      string          `json:"name" bson:"name"`
	Type      RuleType        `json:"type" bson:"type"`
	Status    RuleStatus      `json:"status" bson:"status"`
	Priority  int             `json:"priority" bson:"priority"` // 优先级字段，数字越大优先级越高
	Condition json.RawMessage `json:"condition" bson:"condition"`
}

func (r *MicroRule) GetCollectionName() string {
	return "rule"
}
