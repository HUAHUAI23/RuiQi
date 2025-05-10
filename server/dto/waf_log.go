package dto

import (
	"time"

	"github.com/HUAHUAI23/simple-waf/pkg/model"
)

// AttackEventRequset 攻击事件查询请求
// @Description 用于攻击事件聚合查询的参数结构体，支持多维度筛选，包括来源/目标IP地址、域名、端口和时间范围，并提供分页功能
type AttackEventRequset struct {
	SrcIP     string    `json:"srcIp" form:"srcIp" binding:"omitempty" example:"192.168.1.100"`                                                   // 来源IP地址，用于追踪攻击源
	DstIP     string    `json:"dstIp" form:"dstIp" binding:"omitempty" example:"10.0.0.5"`                                                        // 目标IP地址，被攻击的服务器地址
	Domain    string    `json:"domain" form:"domain" binding:"omitempty" example:"example.com"`                                                   // 域名，被攻击的站点域名
	SrcPort   int       `json:"srcPort" form:"srcPort" binding:"omitempty,min=1,max=65535" example:"443"`                                         // 来源端口号，发起攻击的端口
	DstPort   int       `json:"dstPort" form:"dstPort" binding:"omitempty,min=1,max=65535" example:"443"`                                         // 目标端口号，被攻击的服务端口
	StartTime time.Time `json:"startTime" form:"startTime" binding:"omitempty" time_format:"2006-01-02T15:04:05Z" example:"2024-03-17T00:00:00Z"` // 查询起始时间，ISO8601格式
	EndTime   time.Time `json:"endTime" form:"endTime" binding:"omitempty" time_format:"2006-01-02T15:04:05Z" example:"2024-03-18T23:59:59Z"`     // 查询结束时间，ISO8601格式
	Page      int       `json:"page" form:"page" binding:"omitempty,min=1" default:"1" example:"1"`                                               // 当前页码，从1开始
	PageSize  int       `json:"pageSize" form:"pageSize" binding:"omitempty,min=1,max=100" default:"10" example:"10"`                             // 每页记录数，最大100条
}

// AttackLogRequest 攻击日志查询请求
// @Description 用于详细攻击日志查询的参数结构体，提供精细化的条件筛选，支持通过规则ID、来源/目标IP、域名、端口、请求ID和时间范围进行过滤
type AttackLogRequest struct {
	RuleID    int       `json:"ruleId" form:"ruleId" binding:"omitempty" example:"100012"`                                                        // 规则ID，触发攻击检测的WAF规则标识
	SrcPort   int       `json:"srcPort" form:"srcPort" binding:"omitempty,min=1,max=65535" example:"443"`                                         // 来源端口号，发起攻击的端口
	DstPort   int       `json:"dstPort" form:"dstPort" binding:"omitempty,min=1,max=65535" example:"443"`                                         // 目标端口号，被攻击的服务端口
	Domain    string    `json:"domain" form:"domain" binding:"omitempty" example:"example.com"`                                                   // 域名，被攻击的站点域名
	SrcIP     string    `json:"srcIp" form:"srcIp" binding:"omitempty" example:"192.168.1.100"`                                                   // 来源IP地址，用于追踪攻击源
	DstIP     string    `json:"dstIp" form:"dstIp" binding:"omitempty" example:"10.0.0.5"`                                                        // 目标IP地址，被攻击的服务器地址
	RequestID string    `json:"requestId" form:"requestId" binding:"omitempty" example:"1234567890"`                                              // 请求ID，唯一标识HTTP请求的ID
	StartTime time.Time `json:"startTime" form:"startTime" binding:"omitempty" time_format:"2006-01-02T15:04:05Z" example:"2024-03-17T00:00:00Z"` // 查询起始时间，ISO8601格式
	EndTime   time.Time `json:"endTime" form:"endTime" binding:"omitempty" time_format:"2006-01-02T15:04:05Z" example:"2024-03-18T23:59:59Z"`     // 查询结束时间，ISO8601格式
	Page      int       `json:"page" form:"page" binding:"omitempty,min=1" default:"1" example:"1"`                                               // 当前页码，从1开始
	PageSize  int       `json:"pageSize" form:"pageSize" binding:"omitempty,min=1,max=100" default:"10" example:"10"`                             // 每页记录数，最大100条
}

// AttackEventAggregateResult 攻击事件聚合结果
// @Description 攻击事件的聚合统计结果，提供IP、域名、端口维度的攻击信息汇总，包含攻击次数、首次和最近攻击时间、持续时间等关键指标
type AttackEventAggregateResult struct {
	SrcIP             string        `bson:"srcIp" json:"srcIp" example:"192.168.1.100"`                                    // 来源IP地址，攻击者地址
	SrcIPInfo         *model.IPInfo `bson:"srcIpInfo" json:"srcIpInfo"`                                                    // 来源IP地址，攻击者地址
	DstPort           int           `bson:"dstPort" json:"dstPort" example:"443"`                                          // 目标端口号，被攻击的服务端口
	Domain            string        `bson:"domain" json:"domain" example:"example.com"`                                    // 域名，被攻击的站点
	Count             int           `bson:"count" json:"count" example:"15"`                                               // 攻击总次数，同一来源的攻击计数
	FirstAttackTime   time.Time     `bson:"firstAttackTime" json:"firstAttackTime" example:"2024-03-18T08:12:33Z"`         // 首次攻击时间，该IP首次发起攻击的时间点
	LastAttackTime    time.Time     `bson:"lastAttackTime" json:"lastAttackTime" example:"2024-03-18T08:30:45Z"`           // 最近攻击时间，该IP最后一次攻击的时间点
	DurationInMinutes float64       `bson:"durationInMinutes,omitempty" json:"durationInMinutes,omitempty" example:"18.2"` // 攻击持续时间(分钟)，从首次到最近攻击的时间跨度
	IsOngoing         bool          `bson:"isOngoing" json:"isOngoing" example:"true"`                                     // 是否正在进行中，标识攻击是否仍在持续
}

// AttackEventResponse 攻击事件响应
// @Description 攻击事件查询的分页响应结构体，包含聚合结果列表及分页元数据，用于前端展示和翻页控制
type AttackEventResponse struct {
	Results     []AttackEventAggregateResult `json:"results"`                 // 聚合结果列表，当前页的攻击事件记录
	TotalCount  int64                        `json:"totalCount" example:"35"` // 总记录数，符合条件的攻击事件总数
	PageSize    int                          `json:"pageSize" example:"10"`   // 每页大小，当前设置的每页记录数
	CurrentPage int                          `json:"currentPage" example:"1"` // 当前页码，从1开始计数
	TotalPages  int                          `json:"totalPages" example:"4"`  // 总页数，根据总记录数和每页大小计算
}

// AttackLogResponse 攻击日志响应
// @Description 攻击日志查询的分页响应结构体，返回详细的WAF日志记录及分页信息，便于安全分析和事件追踪
type AttackLogResponse struct {
	Results     []model.WAFLog `json:"results"`                  // 日志记录列表，当前页的WAF攻击日志详情
	TotalCount  int64          `json:"totalCount" example:"128"` // 总记录数，符合查询条件的日志总数
	PageSize    int            `json:"pageSize" example:"10"`    // 每页大小，当前设置的每页记录数
	CurrentPage int            `json:"currentPage" example:"1"`  // 当前页码，从1开始计数
	TotalPages  int            `json:"totalPages" example:"13"`  // 总页数，根据总记录数和每页大小计算
}
