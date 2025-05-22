package model

import (
	"time"
)

// Config WAF系统配置
//	@Description	WAF系统全局配置信息
type Config struct {
	Name            string        `bson:"name" json:"name" example:"AppConfig" description:"配置名称"`
	Engine          EngineConfig  `bson:"engine" json:"engine" description:"引擎配置"`
	Haproxy         HaproxyConfig `bson:"haproxy" json:"haproxy" description:"HAProxy配置"`
	CreatedAt       time.Time     `bson:"createdAt" json:"createdAt" description:"创建时间"`
	UpdatedAt       time.Time     `bson:"updatedAt" json:"updatedAt" description:"更新时间"`
	IsResponseCheck bool          `bson:"isResponseCheck" json:"isResponseCheck" description:"是否检查响应"`
	IsDebug         bool          `bson:"isDebug" json:"isDebug" description:"是否开启调试模式"`
	IsK8s           bool          `bson:"isK8s" json:"isK8s" description:"是否在K8s环境"`
}

// EngineConfig 引擎配置
//	@Description	WAF引擎配置信息
type EngineConfig struct {
	Bind            string            `bson:"bind" json:"bind" example:"0.0.0.0:9000" description:"绑定地址"`
	UseBuiltinRules bool              `bson:"useBuiltinRules" json:"useBuiltinRules" description:"是否使用内置规则"`
	ASNDBPath       string            `bson:"asnDBPath" json:"asnDBPath" example:"/opt/geoip/GeoLite2-ASN.mmdb" description:"ASN数据库路径"`
	CityDBPath      string            `bson:"cityDBPath" json:"cityDBPath" example:"/opt/geoip/GeoLite2-City.mmdb" description:"城市数据库路径"`
	AppConfig       []AppConfig       `bson:"appConfig" json:"appConfig" description:"应用配置列表"`
	FlowController  FlowControlConfig `bson:"flowController" json:"flowController" description:"流量控制配置"`
}

// AppConfig 应用配置
//	@Description	WAF应用配置
type AppConfig struct {
	Name           string        `bson:"name" json:"name" example:"default" description:"应用名称"`
	Directives     string        `bson:"directives" json:"directives" description:"Coraza指令"`
	TransactionTTL time.Duration `bson:"transactionTTL" json:"transactionTTL" example:"10s" description:"事务超时时间"`
	LogLevel       string        `bson:"logLevel" json:"logLevel" example:"info" description:"日志级别"`
	LogFile        string        `bson:"logFile" json:"logFile" example:"/var/log/waf.log" description:"日志文件路径"`
	LogFormat      string        `bson:"logFormat" json:"logFormat" example:"json" description:"日志格式"`
}

// HaproxyConfig HAProxy配置
//	@Description	HAProxy相关配置
type HaproxyConfig struct {
	ConfigBaseDir string `bson:"configBaseDir" json:"configBaseDir" example:"/etc/haproxy" description:"配置基础目录"`
	HaproxyBin    string `bson:"haproxyBin" json:"haproxyBin" example:"/usr/sbin/haproxy" description:"HAProxy可执行文件路径"`
	BackupsNumber int    `bson:"backupsNumber" json:"backupsNumber" example:"5" description:"备份数量"`
	SpoeAgentAddr string `bson:"spoeAgentAddr" json:"spoeAgentAddr" example:"127.0.0.1" description:"SPOE代理地址"`
	SpoeAgentPort int    `bson:"spoeAgentPort" json:"spoeAgentPort" example:"9000" description:"SPOE代理端口"`
	Thread        int    `bson:"thread" json:"thread" example:"4" description:"线程数"`
}

// FlowControlConfig 定义流控配置，用于存储在数据库中
//	@Description	WAF流量控制配置
type FlowControlConfig struct {
	// 高频访问限制配置
	VisitLimit struct {
		Enabled        bool  `bson:"enabled" json:"enabled" example:"true" description:"是否启用访问限制"`
		Threshold      int64 `bson:"threshold" json:"threshold" example:"100" description:"访问阈值，每分钟最大请求数"`
		StatDuration   int64 `bson:"statDuration" json:"statDuration" example:"60" description:"统计时间窗口（秒）"`
		BlockDuration  int64 `bson:"blockDuration" json:"blockDuration" example:"600" description:"封禁时长（秒）"`
		BurstCount     int64 `bson:"burstCount" json:"burstCount" example:"10" description:"允许的突发请求数"`
		ParamsCapacity int64 `bson:"paramsCapacity" json:"paramsCapacity" example:"10000" description:"缓存容量，最多缓存IP数"`
	} `bson:"visitLimit" json:"visitLimit" description:"访问频率限制配置"`

	// 高频攻击限制配置
	AttackLimit struct {
		Enabled        bool  `bson:"enabled" json:"enabled" example:"true" description:"是否启用攻击限制"`
		Threshold      int64 `bson:"threshold" json:"threshold" example:"5" description:"攻击阈值，每分钟最大攻击次数"`
		StatDuration   int64 `bson:"statDuration" json:"statDuration" example:"60" description:"统计时间窗口（秒）"`
		BlockDuration  int64 `bson:"blockDuration" json:"blockDuration" example:"3600" description:"封禁时长（秒）"`
		BurstCount     int64 `bson:"burstCount" json:"burstCount" example:"2" description:"允许的突发攻击次数"`
		ParamsCapacity int64 `bson:"paramsCapacity" json:"paramsCapacity" example:"10000" description:"缓存容量，最多缓存IP数"`
	} `bson:"attackLimit" json:"attackLimit" description:"攻击频率限制配置"`

	// 高频错误限制配置
	ErrorLimit struct {
		Enabled        bool  `bson:"enabled" json:"enabled" example:"true" description:"是否启用错误限制"`
		Threshold      int64 `bson:"threshold" json:"threshold" example:"20" description:"错误阈值，每分钟最大错误次数"`
		StatDuration   int64 `bson:"statDuration" json:"statDuration" example:"60" description:"统计时间窗口（秒）"`
		BlockDuration  int64 `bson:"blockDuration" json:"blockDuration" example:"1800" description:"封禁时长（秒）"`
		BurstCount     int64 `bson:"burstCount" json:"burstCount" example:"5" description:"允许的突发错误次数"`
		ParamsCapacity int64 `bson:"paramsCapacity" json:"paramsCapacity" example:"10000" description:"缓存容量，最多缓存IP数"`
	} `bson:"errorLimit" json:"errorLimit" description:"错误频率限制配置"`
}

// GetDefaultFlowControlConfig 返回默认的流控配置
//	@Summary		获取默认流控配置
//	@Description	获取系统预设的默认流量控制配置
//	@Return			FlowControlConfig 默认流控配置
func GetDefaultFlowControlConfig() FlowControlConfig {
	return FlowControlConfig{
		VisitLimit: struct {
			Enabled        bool  `bson:"enabled" json:"enabled" example:"true" description:"是否启用访问限制"`
			Threshold      int64 `bson:"threshold" json:"threshold" example:"100" description:"访问阈值，每分钟最大请求数"`
			StatDuration   int64 `bson:"statDuration" json:"statDuration" example:"60" description:"统计时间窗口（秒）"`
			BlockDuration  int64 `bson:"blockDuration" json:"blockDuration" example:"600" description:"封禁时长（秒）"`
			BurstCount     int64 `bson:"burstCount" json:"burstCount" example:"10" description:"允许的突发请求数"`
			ParamsCapacity int64 `bson:"paramsCapacity" json:"paramsCapacity" example:"10000" description:"缓存容量，最多缓存IP数"`
		}{
			Enabled:        false,
			Threshold:      100,   // 每分钟100次请求
			StatDuration:   60,    // 统计时间窗口1分钟
			BlockDuration:  600,   // 封禁10分钟
			BurstCount:     10,    // 允许突发10次
			ParamsCapacity: 10000, // 缓存1万个IP
		},
		AttackLimit: struct {
			Enabled        bool  `bson:"enabled" json:"enabled" example:"true" description:"是否启用攻击限制"`
			Threshold      int64 `bson:"threshold" json:"threshold" example:"5" description:"攻击阈值，每分钟最大攻击次数"`
			StatDuration   int64 `bson:"statDuration" json:"statDuration" example:"60" description:"统计时间窗口（秒）"`
			BlockDuration  int64 `bson:"blockDuration" json:"blockDuration" example:"3600" description:"封禁时长（秒）"`
			BurstCount     int64 `bson:"burstCount" json:"burstCount" example:"2" description:"允许的突发攻击次数"`
			ParamsCapacity int64 `bson:"paramsCapacity" json:"paramsCapacity" example:"10000" description:"缓存容量，最多缓存IP数"`
		}{
			Enabled:        false,
			Threshold:      5,     // 每分钟5次攻击
			StatDuration:   60,    // 统计时间窗口1分钟
			BlockDuration:  3600,  // 封禁1小时
			BurstCount:     2,     // 允许突发2次
			ParamsCapacity: 10000, // 缓存1万个IP
		},
		ErrorLimit: struct {
			Enabled        bool  `bson:"enabled" json:"enabled" example:"true" description:"是否启用错误限制"`
			Threshold      int64 `bson:"threshold" json:"threshold" example:"20" description:"错误阈值，每分钟最大错误次数"`
			StatDuration   int64 `bson:"statDuration" json:"statDuration" example:"60" description:"统计时间窗口（秒）"`
			BlockDuration  int64 `bson:"blockDuration" json:"blockDuration" example:"1800" description:"封禁时长（秒）"`
			BurstCount     int64 `bson:"burstCount" json:"burstCount" example:"5" description:"允许的突发错误次数"`
			ParamsCapacity int64 `bson:"paramsCapacity" json:"paramsCapacity" example:"10000" description:"缓存容量，最多缓存IP数"`
		}{
			Enabled:        false,
			Threshold:      20,    // 每分钟20次错误
			StatDuration:   60,    // 统计时间窗口1分钟
			BlockDuration:  1800,  // 封禁30分钟
			BurstCount:     5,     // 允许突发5次
			ParamsCapacity: 10000, // 缓存1万个IP
		},
	}
}

// GetCollectionName 获取集合名称
//	@Summary		获取MongoDB集合名称
//	@Description	获取配置在MongoDB中的集合名称
//	@Return			string 集合名称
func (c *Config) GetCollectionName() string {
	return "config"
}
