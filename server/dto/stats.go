package dto

import "time"

// 时间范围常量
const (
	TimeRange24Hours = "24h"
	TimeRange7Days   = "7d"
	TimeRange30Days  = "30d"
)

// StatsRequest 统计数据请求
// @Description 统计数据请求参数
type StatsRequest struct {
	TimeRange string `json:"timeRange" form:"timeRange" binding:"required,oneof=24h 7d 30d" example:"24h"` // 时间范围: 24h, 7d, 30d
}

// RealtimeQPSRequest 实时QPS请求
// @Description 实时QPS数据请求参数
type RealtimeQPSRequest struct {
	Limit int `json:"limit" form:"limit" binding:"omitempty,min=1,max=60" default:"30" example:"30"` // 数据点数量限制，默认30个点，最大60个点
}

// TimeSeriesDataRequest 时间序列数据请求
// @Description 时间序列数据请求参数
type TimeSeriesDataRequest struct {
	TimeRange string `json:"timeRange" form:"timeRange" binding:"required,oneof=24h 7d 30d" example:"24h"`     // 时间范围: 24h, 7d, 30d
	Metric    string `json:"metric" form:"metric" binding:"required,oneof=requests blocks" example:"requests"` // 指标类型: requests(请求数), blocks(拦截数)
}

// TrafficTimeSeriesRequest 流量时间序列数据请求
// @Description 流量时间序列数据请求参数
type TrafficTimeSeriesRequest struct {
	TimeRange string `json:"timeRange" form:"timeRange" binding:"required,oneof=24h 7d 30d" example:"24h"` // 时间范围: 24h, 7d, 30d
}

// OverviewStats 概览统计数据
// @Description 概览统计数据，包含各项关键指标
type OverviewStats struct {
	// 时间范围
	TimeRange string `json:"timeRange" example:"24h"` // 统计数据的时间范围

	// 流量统计
	TotalRequests   int64 `json:"totalRequests" example:"123456"`     // 总请求数
	InboundTraffic  int64 `json:"inboundTraffic" example:"67890123"`  // 入站流量(字节)
	OutboundTraffic int64 `json:"outboundTraffic" example:"12345678"` // 出站流量(字节)
	MaxQPS          int64 `json:"maxQPS" example:"150"`               // 最大QPS

	// 错误统计
	Error4xx     int64   `json:"error4xx" example:"234"`      // 4xx错误数量
	Error4xxRate float64 `json:"error4xxRate" example:"0.23"` // 4xx错误率(百分比)
	Error5xx     int64   `json:"error5xx" example:"45"`       // 5xx错误数量
	Error5xxRate float64 `json:"error5xxRate" example:"0.05"` // 5xx错误率(百分比)

	// 安全统计
	BlockCount    int64 `json:"blockCount" example:"123"`   // 拦截数量
	AttackIPCount int64 `json:"attackIPCount" example:"45"` // 攻击IP数量
}

// RealtimeQPSData 实时QPS数据
// @Description 实时QPS数据点，包含时间戳和值
type RealtimeQPSData struct {
	Timestamp time.Time `json:"timestamp" example:"2024-01-01T12:30:45Z"` // 时间戳
	Value     int64     `json:"value" example:"120"`                      // QPS值
}

// RealtimeQPSResponse 实时QPS响应
// @Description 实时QPS数据响应
type RealtimeQPSResponse struct {
	Data []RealtimeQPSData `json:"data"` // QPS数据点列表
}

// TimeSeriesDataPoint 时间序列数据点
// @Description 时间序列图表数据点
type TimeSeriesDataPoint struct {
	Timestamp time.Time `json:"timestamp" example:"2024-01-01T12:00:00Z"` // 时间戳，表示数据点的时间
	Value     int64     `json:"value" example:"128"`                      // 数值，表示该时间点的指标值
}

// TrafficDataPoint 流量数据点
// @Description 流量时间序列图表数据点
type TrafficDataPoint struct {
	Timestamp       time.Time `json:"timestamp" example:"2024-01-01T12:00:00Z"` // 时间戳，表示数据点的时间
	InboundTraffic  int64     `json:"inboundTraffic" example:"1024000"`         // 入站流量(字节)
	OutboundTraffic int64     `json:"outboundTraffic" example:"2048000"`        // 出站流量(字节)
}

// TimeSeriesResponse 时间序列响应
// @Description 时间序列图表数据响应
type TimeSeriesResponse struct {
	Metric    string                `json:"metric" example:"requests"` // 指标名称
	TimeRange string                `json:"timeRange" example:"24h"`   // 时间范围
	Data      []TimeSeriesDataPoint `json:"data"`                      // 数据点列表
}

// TrafficTimeSeriesResponse 流量时间序列响应
// @Description 流量时间序列图表数据响应
type TrafficTimeSeriesResponse struct {
	TimeRange string             `json:"timeRange" example:"24h"` // 时间范围
	Data      []TrafficDataPoint `json:"data"`                    // 流量数据点列表
}

// CombinedTimeSeriesResponse 组合时间序列响应
// @Description 同时包含请求数和拦截数的时间序列数据
type CombinedTimeSeriesResponse struct {
	TimeRange string             `json:"timeRange" example:"24h"` // 时间范围
	Requests  TimeSeriesResponse `json:"requests"`                // 请求数时间序列
	Blocks    TimeSeriesResponse `json:"blocks"`                  // 拦截数时间序列
}
