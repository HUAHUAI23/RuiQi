package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// WAFLog 表示安全事件日志
// @Description Web应用防火墙安全事件完整记录，包含详细的攻击检测和防护信息
type WAFLog struct {
	ID           bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`                                                                                                     // 日志唯一标识符
	RequestID    string        `json:"requestId" bson:"requestId" example:"a1b2c3d4e5f6"`                                                                                     // 请求唯一标识
	RuleID       int           `json:"ruleId" bson:"ruleId" example:"10086"`                                                                                                  // 触发的规则ID
	SecLangRaw   string        `json:"secLangRaw" bson:"secLangRaw" example:"SecRule REQUEST_HEADERS:User-Agent \"@rx (?:scanner)\" \"id:1008,phase:1,severity:'CRITICAL'\""` // 安全规则原始定义
	Severity     int           `json:"severity" bson:"severity" example:"2"`                                                                                                  // 事件严重级别(0-5)
	Phase        int           `json:"phase" bson:"phase" example:"1"`                                                                                                        // 请求处理阶段
	SecMark      string        `json:"secMark" bson:"secMark" example:"web_scanner"`                                                                                          // 安全标记
	Accuracy     int           `json:"accuracy" bson:"accuracy" example:"9"`                                                                                                  // 规则匹配准确度(0-10)
	Payload      string        `json:"payload" bson:"payload" example:"Scanner/1.0"`                                                                                          // 攻击载荷
	URI          string        `json:"uri" bson:"uri" example:"/api/v1/users"`                                                                                                // 请求URI路径
	SrcIP        string        `json:"srcIp" bson:"srcIp" example:"192.168.1.1"`                                                                                              // 来源IP地址
	SrcIPInfo    *IPInfo       `json:"srcIpInfo,omitempty" bson:"srcIpInfo,omitempty"`                                                                                        // 来源IP地理位置信息
	DstIP        string        `json:"dstIp" bson:"dstIp" example:"10.0.0.1"`                                                                                                 // 目标IP地址
	ClientIP     string        `json:"clientIp" bson:"clientIp" example:"192.168.1.1"`                                                                                        // 来源IP地址
	ServerIP     string        `json:"serverIp" bson:"serverIp" example:"10.0.0.1"`                                                                                           // 目标IP地址
	SrcPort      int           `json:"srcPort" bson:"srcPort" example:"52134"`                                                                                                // 来源端口
	DstPort      int           `json:"dstPort" bson:"dstPort" example:"443"`                                                                                                  // 目标端口
	Domain       string        `json:"domain" bson:"domain" example:"api.example.com"`                                                                                        // 目标域名
	Logs         []Log         `json:"logs" bson:"logs"`                                                                                                                      // 关联的日志条目
	Message      string        `json:"message" bson:"message" example:"恶意扫描器检测"`                                                                                              // 事件描述消息
	Request      string        `json:"request" bson:"request" example:"GET /api/v1/users HTTP/1.1\nHost: api.example.com\nUser-Agent: Scanner/1.0"`                           // 原始HTTP请求
	Response     string        `json:"response" bson:"response" example:"HTTP/1.1 403 Forbidden\nContent-Type: text/html\nContent-Length: 146"`                               // 原始HTTP响应
	Date         string        `json:"date" bson:"date"`
	Hour         int           `json:"hour" bson:"hour"`
	HourGroupSix int           `json:"hourGroupSix" bson:"hourGroupSix" example:"0"`
	Minute       int           `json:"minute" bson:"minute"`
	CreatedAt    time.Time     `json:"createdAt" bson:"createdAt" example:"2024-03-18T08:12:33Z"` // 事件发生时间戳
}

// Log 表示单个日志条目
// @Description 详细的WAF规则匹配记录，包含规则触发的详细信息和原始日志
type Log struct {
	Message    string `json:"message" bson:"message" example:"恶意扫描器检测"`                                                                                                                                                                                                                                       // 日志消息
	Payload    string `json:"payload" bson:"payload" example:"Scanner/1.0"`                                                                                                                                                                                                                                   // 攻击载荷
	RuleID     int    `json:"ruleId" bson:"ruleId" example:"10086"`                                                                                                                                                                                                                                           // 规则ID
	Severity   int    `json:"severity" bson:"severity" example:"2"`                                                                                                                                                                                                                                           // 严重级别(0-5)
	Phase      int    `json:"phase" bson:"phase" example:"1"`                                                                                                                                                                                                                                                 // 请求处理阶段
	SecMark    string `json:"secMark" bson:"secMark" example:"web_scanner"`                                                                                                                                                                                                                                   // 安全标记
	Accuracy   int    `json:"accuracy" bson:"accuracy" example:"9"`                                                                                                                                                                                                                                           // 规则匹配准确度(0-10)
	SecLangRaw string `json:"secLangRaw" bson:"secLangRaw" example:"SecRule REQUEST_HEADERS:User-Agent \"@rx (?:scanner)\" \"id:1008,phase:1,severity:'CRITICAL'\""`                                                                                                                                          // 安全规则原始定义
	LogRaw     string `json:"logRaw" bson:"logRaw" example:"[2024-03-18 08:12:33] [error] ModSecurity: Access denied with code 403 (phase 1). Matched \"Operator 'Rx' with parameter '(?:scanner)' against variable 'REQUEST_HEADERS:User-Agent'\" [id \"10086\"] [msg \"恶意扫描器检测\"] [severity \"CRITICAL\"]"` // 原始日志数据
}

// GetCollectionName 返回WAFLog对应的MongoDB集合名称
// @Description 获取用于存储WAF日志的MongoDB集合名
func (wafLog *WAFLog) GetCollectionName() string {
	return "waf_log"
}
