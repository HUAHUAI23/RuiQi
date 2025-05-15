package model

import (
	"time"

	"github.com/haproxytech/client-native/v6/models"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// HAProxyStats 统计数据的具体字段定义
// 这个结构仅用于转换和差值计算，不再用于存储
// @Description HAProxy 统计数据模型，用于统计指标的计算和转换
type HAProxyStats struct {
	// 流量相关统计
	// @Description 入站流量字节数
	Bin int64 `bson:"bin" json:"bin"`
	// @Description 出站流量字节数
	Bout int64 `bson:"bout" json:"bout"`

	// HTTP响应状态码统计
	// @Description 1xx 状态码响应数量
	Hrsp1xx int64 `bson:"hrsp_1xx" json:"hrsp_1xx"`
	// @Description 2xx 状态码响应数量
	Hrsp2xx int64 `bson:"hrsp_2xx" json:"hrsp_2xx"`
	// @Description 3xx 状态码响应数量
	Hrsp3xx int64 `bson:"hrsp_3xx" json:"hrsp_3xx"`
	// @Description 4xx 状态码响应数量
	Hrsp4xx int64 `bson:"hrsp_4xx" json:"hrsp_4xx"`
	// @Description 5xx 状态码响应数量
	Hrsp5xx int64 `bson:"hrsp_5xx" json:"hrsp_5xx"`
	// @Description 其他状态码响应数量
	HrspOther int64 `bson:"hrsp_other" json:"hrsp_other"`

	// 错误相关统计
	// @Description 请求被拒绝的数量
	Dreq int64 `bson:"dreq" json:"dreq"`
	// @Description 响应被拒绝的数量
	Dresp int64 `bson:"dresp" json:"dresp"`
	// @Description 请求错误数量
	Ereq int64 `bson:"ereq" json:"ereq"`
	// @Description 连接被拒绝的数量
	Dcon int64 `bson:"dcon" json:"dcon"`
	// @Description 会话被拒绝的数量
	Dses int64 `bson:"dses" json:"dses"`
	// @Description 连接错误数量
	Econ int64 `bson:"econ" json:"econ"`
	// @Description 响应错误数量
	Eresp int64 `bson:"eresp" json:"eresp"`

	// 速率最大值
	// @Description 最大请求速率
	ReqRateMax int64 `bson:"req_rate_max" json:"req_rate_max"`
	// @Description 最大连接速率
	ConnRateMax int64 `bson:"conn_rate_max" json:"conn_rate_max"`
	// @Description 最大传输速率(字节/秒)
	RateMax int64 `bson:"rate_max" json:"rate_max"`
	// @Description 最大并发会话数
	Smax int64 `bson:"smax" json:"smax"`

	// 总计值
	// @Description 总连接数
	ConnTot int64 `bson:"conn_tot" json:"conn_tot"`
	// @Description 总会话数
	Stot int64 `bson:"stot" json:"stot"`
	// @Description 总请求数
	ReqTot int64 `bson:"req_tot" json:"req_tot"`
}

// HAProxyStatsBaseline 存储HAProxy统计数据基准线
// @Description HAProxy 统计数据基准线，用于记录和恢复统计基准
type HAProxyStatsBaseline struct {
	// @Description 记录的唯一标识符
	ID bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	// @Description 目标前端名称
	TargetName string `bson:"target_name" json:"target_name"`
	// @Description 记录时间戳
	Timestamp time.Time `bson:"timestamp" json:"timestamp"`
	// @Description HAProxy 重启计数
	ResetCount int32 `bson:"reset_count" json:"reset_count"`

	// 直接将统计字段扁平化在结构中
	// @Description 入站流量字节数
	Bin int64 `bson:"bin" json:"bin"`
	// @Description 出站流量字节数
	Bout int64 `bson:"bout" json:"bout"`
	// @Description 1xx 状态码响应数量
	Hrsp1xx int64 `bson:"hrsp_1xx" json:"hrsp_1xx"`
	// @Description 2xx 状态码响应数量
	Hrsp2xx int64 `bson:"hrsp_2xx" json:"hrsp_2xx"`
	// @Description 3xx 状态码响应数量
	Hrsp3xx int64 `bson:"hrsp_3xx" json:"hrsp_3xx"`
	// @Description 4xx 状态码响应数量
	Hrsp4xx int64 `bson:"hrsp_4xx" json:"hrsp_4xx"`
	// @Description 5xx 状态码响应数量
	Hrsp5xx int64 `bson:"hrsp_5xx" json:"hrsp_5xx"`
	// @Description 其他状态码响应数量
	HrspOther int64 `bson:"hrsp_other" json:"hrsp_other"`
	// @Description 请求被拒绝的数量
	Dreq int64 `bson:"dreq" json:"dreq"`
	// @Description 响应被拒绝的数量
	Dresp int64 `bson:"dresp" json:"dresp"`
	// @Description 请求错误数量
	Ereq int64 `bson:"ereq" json:"ereq"`
	// @Description 连接被拒绝的数量
	Dcon int64 `bson:"dcon" json:"dcon"`
	// @Description 会话被拒绝的数量
	Dses int64 `bson:"dses" json:"dses"`
	// @Description 连接错误数量
	Econ int64 `bson:"econ" json:"econ"`
	// @Description 响应错误数量
	Eresp int64 `bson:"eresp" json:"eresp"`
	// @Description 最大请求速率
	ReqRateMax int64 `bson:"req_rate_max" json:"req_rate_max"`
	// @Description 最大连接速率
	ConnRateMax int64 `bson:"conn_rate_max" json:"conn_rate_max"`
	// @Description 最大传输速率(字节/秒)
	RateMax int64 `bson:"rate_max" json:"rate_max"`
	// @Description 最大并发会话数
	Smax int64 `bson:"smax" json:"smax"`
	// @Description 总连接数
	ConnTot int64 `bson:"conn_tot" json:"conn_tot"`
	// @Description 总会话数
	Stot int64 `bson:"stot" json:"stot"`
	// @Description 总请求数
	ReqTot int64 `bson:"req_tot" json:"req_tot"`
}

// GetCollectionName 返回集合名称
// @Description 获取数据库集合名称
func (h *HAProxyStatsBaseline) GetCollectionName() string {
	return "haproxy_baseline"
}

// GetStats 从扁平结构获取HAProxyStats
// @Description 从基准线数据获取 HAProxyStats 结构
func (h *HAProxyStatsBaseline) GetStats() HAProxyStats {
	return HAProxyStats{
		Bin:         h.Bin,
		Bout:        h.Bout,
		Hrsp1xx:     h.Hrsp1xx,
		Hrsp2xx:     h.Hrsp2xx,
		Hrsp3xx:     h.Hrsp3xx,
		Hrsp4xx:     h.Hrsp4xx,
		Hrsp5xx:     h.Hrsp5xx,
		HrspOther:   h.HrspOther,
		Dreq:        h.Dreq,
		Dresp:       h.Dresp,
		Ereq:        h.Ereq,
		Dcon:        h.Dcon,
		Dses:        h.Dses,
		Econ:        h.Econ,
		Eresp:       h.Eresp,
		ReqRateMax:  h.ReqRateMax,
		ConnRateMax: h.ConnRateMax,
		RateMax:     h.RateMax,
		Smax:        h.Smax,
		ConnTot:     h.ConnTot,
		Stot:        h.Stot,
		ReqTot:      h.ReqTot,
	}
}

// SetStats 将HAProxyStats应用到扁平结构
// @Description 将统计数据应用到基准线结构
func (h *HAProxyStatsBaseline) SetStats(stats HAProxyStats) {
	h.Bin = stats.Bin
	h.Bout = stats.Bout
	h.Hrsp1xx = stats.Hrsp1xx
	h.Hrsp2xx = stats.Hrsp2xx
	h.Hrsp3xx = stats.Hrsp3xx
	h.Hrsp4xx = stats.Hrsp4xx
	h.Hrsp5xx = stats.Hrsp5xx
	h.HrspOther = stats.HrspOther
	h.Dreq = stats.Dreq
	h.Dresp = stats.Dresp
	h.Ereq = stats.Ereq
	h.Dcon = stats.Dcon
	h.Dses = stats.Dses
	h.Econ = stats.Econ
	h.Eresp = stats.Eresp
	h.ReqRateMax = stats.ReqRateMax
	h.ConnRateMax = stats.ConnRateMax
	h.RateMax = stats.RateMax
	h.Smax = stats.Smax
	h.ConnTot = stats.ConnTot
	h.Stot = stats.Stot
	h.ReqTot = stats.ReqTot
}

// HAProxyMinuteStats 存储HAProxy分钟统计数据 - 完全扁平化结构
// @Description HAProxy 分钟级统计数据，用于存储每分钟的统计指标
type HAProxyMinuteStats struct {
	// @Description 记录的唯一标识符
	ID bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	// @Description 目标前端名称
	TargetName string `bson:"target_name" json:"target_name"`
	// @Description 日期，格式为 YYYY-MM-DD
	Date string `bson:"date" json:"date"`
	// @Description 小时 (0-23)
	Hour int `bson:"hour" json:"hour"`
	// @Description 分钟 (0-59)
	Minute int `bson:"minute" json:"minute"`
	// @Description 小时组 (0-4) 0-5点为0组，6-11点为1组，12-17点为2组，18-23点为3组
	HourGroupSix int `bson:"hourGroupSix" json:"hourGroupSix"`
	// @Description 完整时间戳
	Timestamp time.Time `bson:"timestamp" json:"timestamp"`

	// 直接将所有指标字段放在第一级别
	// @Description 入站流量字节数
	Bin int64 `bson:"bin" json:"bin"`
	// @Description 出站流量字节数
	Bout int64 `bson:"bout" json:"bout"`
	// @Description 1xx 状态码响应数量
	Hrsp1xx int64 `bson:"hrsp_1xx" json:"hrsp_1xx"`
	// @Description 2xx 状态码响应数量
	Hrsp2xx int64 `bson:"hrsp_2xx" json:"hrsp_2xx"`
	// @Description 3xx 状态码响应数量
	Hrsp3xx int64 `bson:"hrsp_3xx" json:"hrsp_3xx"`
	// @Description 4xx 状态码响应数量
	Hrsp4xx int64 `bson:"hrsp_4xx" json:"hrsp_4xx"`
	// @Description 5xx 状态码响应数量
	Hrsp5xx int64 `bson:"hrsp_5xx" json:"hrsp_5xx"`
	// @Description 其他状态码响应数量
	HrspOther int64 `bson:"hrsp_other" json:"hrsp_other"`
	// @Description 请求被拒绝的数量
	Dreq int64 `bson:"dreq" json:"dreq"`
	// @Description 响应被拒绝的数量
	Dresp int64 `bson:"dresp" json:"dresp"`
	// @Description 请求错误数量
	Ereq int64 `bson:"ereq" json:"ereq"`
	// @Description 连接被拒绝的数量
	Dcon int64 `bson:"dcon" json:"dcon"`
	// @Description 会话被拒绝的数量
	Dses int64 `bson:"dses" json:"dses"`
	// @Description 连接错误数量
	Econ int64 `bson:"econ" json:"econ"`
	// @Description 响应错误数量
	Eresp int64 `bson:"eresp" json:"eresp"`
	// @Description 最大请求速率
	ReqRateMax int64 `bson:"req_rate_max" json:"req_rate_max"`
	// @Description 最大连接速率
	ConnRateMax int64 `bson:"conn_rate_max" json:"conn_rate_max"`
	// @Description 最大传输速率(字节/秒)
	RateMax int64 `bson:"rate_max" json:"rate_max"`
	// @Description 最大并发会话数
	Smax int64 `bson:"smax" json:"smax"`
	// @Description 总连接数
	ConnTot int64 `bson:"conn_tot" json:"conn_tot"`
	// @Description 总会话数
	Stot int64 `bson:"stot" json:"stot"`
	// @Description 总请求数
	ReqTot int64 `bson:"req_tot" json:"req_tot"`
}

// GetCollectionName 返回集合名称
// @Description 获取分钟统计数据的数据库集合名称
func (h *HAProxyMinuteStats) GetCollectionName() string {
	return "haproxy_minute_stats"
}

// GetStats 从扁平结构获取HAProxyStats
// @Description 从分钟统计数据获取 HAProxyStats 结构
func (h *HAProxyMinuteStats) GetStats() HAProxyStats {
	return HAProxyStats{
		Bin:         h.Bin,
		Bout:        h.Bout,
		Hrsp1xx:     h.Hrsp1xx,
		Hrsp2xx:     h.Hrsp2xx,
		Hrsp3xx:     h.Hrsp3xx,
		Hrsp4xx:     h.Hrsp4xx,
		Hrsp5xx:     h.Hrsp5xx,
		HrspOther:   h.HrspOther,
		Dreq:        h.Dreq,
		Dresp:       h.Dresp,
		Ereq:        h.Ereq,
		Dcon:        h.Dcon,
		Dses:        h.Dses,
		Econ:        h.Econ,
		Eresp:       h.Eresp,
		ReqRateMax:  h.ReqRateMax,
		ConnRateMax: h.ConnRateMax,
		RateMax:     h.RateMax,
		Smax:        h.Smax,
		ConnTot:     h.ConnTot,
		Stot:        h.Stot,
		ReqTot:      h.ReqTot,
	}
}

// SetStats 将HAProxyStats应用到扁平结构
// @Description 将统计数据应用到分钟统计结构
func (h *HAProxyMinuteStats) SetStats(stats HAProxyStats) {
	h.Bin = stats.Bin
	h.Bout = stats.Bout
	h.Hrsp1xx = stats.Hrsp1xx
	h.Hrsp2xx = stats.Hrsp2xx
	h.Hrsp3xx = stats.Hrsp3xx
	h.Hrsp4xx = stats.Hrsp4xx
	h.Hrsp5xx = stats.Hrsp5xx
	h.HrspOther = stats.HrspOther
	h.Dreq = stats.Dreq
	h.Dresp = stats.Dresp
	h.Ereq = stats.Ereq
	h.Dcon = stats.Dcon
	h.Dses = stats.Dses
	h.Econ = stats.Econ
	h.Eresp = stats.Eresp
	h.ReqRateMax = stats.ReqRateMax
	h.ConnRateMax = stats.ConnRateMax
	h.RateMax = stats.RateMax
	h.Smax = stats.Smax
	h.ConnTot = stats.ConnTot
	h.Stot = stats.Stot
	h.ReqTot = stats.ReqTot
}

// HAProxyRealTimeStats 存储HAProxy实时统计数据
// @Description HAProxy 实时统计数据，用于存储和查询实时指标
type HAProxyRealTimeStats struct {
	// @Description 目标前端名称
	TargetName string `bson:"target_name" json:"target_name"`
	// @Description 度量指标名称
	MetricName string `bson:"metric_name" json:"metric_name"`
	// @Description 度量指标值
	Value int64 `bson:"value" json:"value"`
	// @Description 记录时间戳
	Timestamp time.Time `bson:"timestamp" json:"timestamp"`
}

// GetCollectionName 返回集合名称
// @Description 获取实时统计数据的数据库集合名称
func (h *HAProxyRealTimeStats) GetCollectionName() string {
	return "haproxy_real_time_stats"
}

// TimeSeriesMetric 时间序列指标
// @Description 时间序列度量指标结构，用于存储和查询时间序列数据
type TimeSeriesMetric struct {
	// @Description 记录时间戳
	Timestamp time.Time `bson:"timestamp"`
	// @Description 度量指标值
	Value int64 `bson:"value"`
	// @Description 度量指标元数据
	Metadata TimeSeriesMeta `bson:"metadata"`
}

// TimeSeriesMeta 时间序列元数据
// @Description 时间序列度量指标的元数据
type TimeSeriesMeta struct {
	// @Description 目标名称
	Target string `bson:"target"`
}

// NativeStatsToHAProxyStats 将NativeStatStats转换为HAProxyStats
// @Description 将 HAProxy 原生统计数据转换为内部使用的统计结构
func NativeStatsToHAProxyStats(native *models.NativeStatStats) HAProxyStats {
	stats := HAProxyStats{}

	// 设置流量相关字段
	if native.Bin != nil {
		stats.Bin = *native.Bin
	}
	if native.Bout != nil {
		stats.Bout = *native.Bout
	}

	// 设置HTTP响应状态码
	if native.Hrsp1xx != nil {
		stats.Hrsp1xx = *native.Hrsp1xx
	}
	if native.Hrsp2xx != nil {
		stats.Hrsp2xx = *native.Hrsp2xx
	}
	if native.Hrsp3xx != nil {
		stats.Hrsp3xx = *native.Hrsp3xx
	}
	if native.Hrsp4xx != nil {
		stats.Hrsp4xx = *native.Hrsp4xx
	}
	if native.Hrsp5xx != nil {
		stats.Hrsp5xx = *native.Hrsp5xx
	}
	if native.HrspOther != nil {
		stats.HrspOther = *native.HrspOther
	}

	// 设置错误相关字段
	if native.Dreq != nil {
		stats.Dreq = *native.Dreq
	}
	if native.Dresp != nil {
		stats.Dresp = *native.Dresp
	}
	if native.Ereq != nil {
		stats.Ereq = *native.Ereq
	}
	if native.Dcon != nil {
		stats.Dcon = *native.Dcon
	}
	if native.Dses != nil {
		stats.Dses = *native.Dses
	}
	if native.Econ != nil {
		stats.Econ = *native.Econ
	}
	if native.Eresp != nil {
		stats.Eresp = *native.Eresp
	}

	// 设置速率最大值
	if native.ReqRateMax != nil {
		stats.ReqRateMax = *native.ReqRateMax
	}
	if native.ConnRateMax != nil {
		stats.ConnRateMax = *native.ConnRateMax
	}
	if native.RateMax != nil {
		stats.RateMax = *native.RateMax
	}
	if native.Smax != nil {
		stats.Smax = *native.Smax
	}

	// 设置总计值
	if native.ConnTot != nil {
		stats.ConnTot = *native.ConnTot
	}
	if native.Stot != nil {
		stats.Stot = *native.Stot
	}
	if native.ReqTot != nil {
		stats.ReqTot = *native.ReqTot
	}

	return stats
}

// HAProxyStatsToNative 将HAProxyStats转换为NativeStatStats
// @Description 将内部使用的统计结构转换为 HAProxy 原生统计数据格式
func HAProxyStatsToNative(stats HAProxyStats) *models.NativeStatStats {
	native := &models.NativeStatStats{}

	// 设置流量相关字段
	bin := stats.Bin
	native.Bin = &bin

	bout := stats.Bout
	native.Bout = &bout

	// 设置HTTP响应状态码
	hrsp1xx := stats.Hrsp1xx
	native.Hrsp1xx = &hrsp1xx

	hrsp2xx := stats.Hrsp2xx
	native.Hrsp2xx = &hrsp2xx

	hrsp3xx := stats.Hrsp3xx
	native.Hrsp3xx = &hrsp3xx

	hrsp4xx := stats.Hrsp4xx
	native.Hrsp4xx = &hrsp4xx

	hrsp5xx := stats.Hrsp5xx
	native.Hrsp5xx = &hrsp5xx

	hrspOther := stats.HrspOther
	native.HrspOther = &hrspOther

	// 设置错误相关字段
	dreq := stats.Dreq
	native.Dreq = &dreq

	dresp := stats.Dresp
	native.Dresp = &dresp

	ereq := stats.Ereq
	native.Ereq = &ereq

	dcon := stats.Dcon
	native.Dcon = &dcon

	dses := stats.Dses
	native.Dses = &dses

	econ := stats.Econ
	native.Econ = &econ

	eresp := stats.Eresp
	native.Eresp = &eresp

	// 设置速率最大值
	reqRateMax := stats.ReqRateMax
	native.ReqRateMax = &reqRateMax

	connRateMax := stats.ConnRateMax
	native.ConnRateMax = &connRateMax

	rateMax := stats.RateMax
	native.RateMax = &rateMax

	smax := stats.Smax
	native.Smax = &smax

	// 设置总计值
	connTot := stats.ConnTot
	native.ConnTot = &connTot

	stot := stats.Stot
	native.Stot = &stot

	reqTot := stats.ReqTot
	native.ReqTot = &reqTot

	return native
}

// CalculateStatsDelta 计算两个HAProxyStats之间的差值
// @Description 计算两个时间点之间的统计数据差值
func CalculateStatsDelta(last, current HAProxyStats) HAProxyStats {
	delta := HAProxyStats{}

	// 计算流量字段差值
	delta.Bin = safeSubtract(current.Bin, last.Bin)
	delta.Bout = safeSubtract(current.Bout, last.Bout)

	// 计算HTTP响应状态码差值
	delta.Hrsp1xx = safeSubtract(current.Hrsp1xx, last.Hrsp1xx)
	delta.Hrsp2xx = safeSubtract(current.Hrsp2xx, last.Hrsp2xx)
	delta.Hrsp3xx = safeSubtract(current.Hrsp3xx, last.Hrsp3xx)
	delta.Hrsp4xx = safeSubtract(current.Hrsp4xx, last.Hrsp4xx)
	delta.Hrsp5xx = safeSubtract(current.Hrsp5xx, last.Hrsp5xx)
	delta.HrspOther = safeSubtract(current.HrspOther, last.HrspOther)

	// 计算错误相关字段差值
	delta.Dreq = safeSubtract(current.Dreq, last.Dreq)
	delta.Dresp = safeSubtract(current.Dresp, last.Dresp)
	delta.Ereq = safeSubtract(current.Ereq, last.Ereq)
	delta.Dcon = safeSubtract(current.Dcon, last.Dcon)
	delta.Dses = safeSubtract(current.Dses, last.Dses)
	delta.Econ = safeSubtract(current.Econ, last.Econ)
	delta.Eresp = safeSubtract(current.Eresp, last.Eresp)

	// 速率最大值直接使用当前值
	delta.ReqRateMax = current.ReqRateMax
	delta.ConnRateMax = current.ConnRateMax
	delta.RateMax = current.RateMax
	delta.Smax = current.Smax

	// 请求数 连接数 会话数 直接使用差值
	delta.ConnTot = safeSubtract(current.ConnTot, last.ConnTot)
	delta.Stot = safeSubtract(current.Stot, last.Stot)
	delta.ReqTot = safeSubtract(current.ReqTot, last.ReqTot)

	return delta
}

// safeSubtract 安全的减法操作，确保结果不为负
// @Description 安全减法操作，处理 HAProxy 重启等导致的计数器重置情况
func safeSubtract(current, last int64) int64 {
	if current < last {
		// 可能是由于HAProxy重启导致的，返回当前值作为增量
		return current
	}
	return current - last
}

// CreateZeroStats 创建所有指标为0的HAProxyStats，用于重启后记录
// @Description 创建全零统计数据，用于 HAProxy 重启后的记录
func CreateZeroStats() HAProxyStats {
	return HAProxyStats{} // Go结构体的零值，所有字段都为0
}

// DetectReset 检测HAProxy是否已重启
// @Description 通过比较两个时间点的统计数据，检测 HAProxy 是否已重启
func DetectReset(lastStats, currentStats HAProxyStats) bool {
	// 检查bin（入站字节数）
	if currentStats.Bin < lastStats.Bin {
		return true
	}

	// 检查bout（出站字节数）
	if currentStats.Bout < lastStats.Bout {
		return true
	}

	// 检查总连接数
	if currentStats.ConnTot < lastStats.ConnTot {
		return true
	}

	// 检查总会话数
	if currentStats.Stot < lastStats.Stot {
		return true
	}

	// 检查总请求数
	if currentStats.ReqTot < lastStats.ReqTot {
		return true
	}

	return false
}
