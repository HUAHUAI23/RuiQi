package model

import (
	"time"

	"github.com/haproxytech/client-native/v6/models"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// HAProxyStats 统计数据的具体字段定义
type HAProxyStats struct {
	// 流量相关统计
	Bin  int64 `bson:"bin" json:"bin"`
	Bout int64 `bson:"bout" json:"bout"`

	// HTTP响应状态码统计
	Hrsp1xx   int64 `bson:"hrsp_1xx" json:"hrsp_1xx"`
	Hrsp2xx   int64 `bson:"hrsp_2xx" json:"hrsp_2xx"`
	Hrsp3xx   int64 `bson:"hrsp_3xx" json:"hrsp_3xx"`
	Hrsp4xx   int64 `bson:"hrsp_4xx" json:"hrsp_4xx"`
	Hrsp5xx   int64 `bson:"hrsp_5xx" json:"hrsp_5xx"`
	HrspOther int64 `bson:"hrsp_other" json:"hrsp_other"`

	// 错误相关统计
	Dreq  int64 `bson:"dreq" json:"dreq"`
	Dresp int64 `bson:"dresp" json:"dresp"`
	Ereq  int64 `bson:"ereq" json:"ereq"`
	Dcon  int64 `bson:"dcon" json:"dcon"`
	Dses  int64 `bson:"dses" json:"dses"`
	Econ  int64 `bson:"econ" json:"econ"`
	Eresp int64 `bson:"eresp" json:"eresp"`

	// 速率最大值
	ReqRateMax  int64 `bson:"req_rate_max" json:"req_rate_max"`
	ConnRateMax int64 `bson:"conn_rate_max" json:"conn_rate_max"`
	RateMax     int64 `bson:"rate_max" json:"rate_max"`
	Smax        int64 `bson:"smax" json:"smax"`

	// 总计值
	ConnTot int64 `bson:"conn_tot" json:"conn_tot"`
	Stot    int64 `bson:"stot" json:"stot"`
	ReqTot  int64 `bson:"req_tot" json:"req_tot"`
}

// HAProxyStatsBaseline 存储HAProxy统计数据基准线
type HAProxyStatsBaseline struct {
	ID         bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	TargetName string        `bson:"target_name" json:"target_name"`
	Stats      HAProxyStats  `bson:"stats" json:"stats"`
	Timestamp  time.Time     `bson:"timestamp" json:"timestamp"`
	ResetCount int32         `bson:"reset_count" json:"reset_count"`
}

// GetCollectionName 返回集合名称
func (h *HAProxyStatsBaseline) GetCollectionName() string {
	return "haproxy_baseline"
}

// HAProxyMinuteStats 存储HAProxy分钟统计数据
type HAProxyMinuteStats struct {
	ID         bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	TargetName string        `bson:"target_name" json:"target_name"`
	Date       string        `bson:"date" json:"date"`
	Hour       int           `bson:"hour" json:"hour"`
	Minute     int           `bson:"minute" json:"minute"`
	Timestamp  time.Time     `bson:"timestamp" json:"timestamp"`
	Stats      HAProxyStats  `bson:"stats" json:"stats"`
}

// GetCollectionName 返回集合名称
func (h *HAProxyMinuteStats) GetCollectionName() string {
	return "haproxy_minute_stats"
}

// HAProxyRealTimeStats 存储HAProxy实时统计数据
type HAProxyRealTimeStats struct {
	TargetName string    `bson:"target_name" json:"target_name"`
	MetricName string    `bson:"metric_name" json:"metric_name"`
	Value      int64     `bson:"value" json:"value"`
	Timestamp  time.Time `bson:"timestamp" json:"timestamp"`
}

// GetCollectionName 返回集合名称
func (h *HAProxyRealTimeStats) GetCollectionName() string {
	return "haproxy_real_time_stats"
}

// TimeSeriesMetric 时间序列指标
type TimeSeriesMetric struct {
	Timestamp time.Time      `bson:"timestamp"`
	Value     int64          `bson:"value"`
	Metadata  TimeSeriesMeta `bson:"metadata"`
}

// TimeSeriesMeta 时间序列元数据
type TimeSeriesMeta struct {
	Target string `bson:"target"`
}

// NativeStatsToHAProxyStats 将NativeStatStats转换为HAProxyStats
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
func safeSubtract(current, last int64) int64 {
	if current < last {
		// 可能是由于HAProxy重启导致的，返回当前值作为增量
		return current
	}
	return current - last
}

// CreateZeroStats 创建所有指标为0的HAProxyStats，用于重启后记录
func CreateZeroStats() HAProxyStats {
	return HAProxyStats{} // Go结构体的零值，所有字段都为0
}

// DetectReset 检测HAProxy是否已重启
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
