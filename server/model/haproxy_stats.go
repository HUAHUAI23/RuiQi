package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// HAProxyStatsBaseline 存储HAProxy统计数据基准线
type HAProxyStatsBaseline struct {
	ID         bson.ObjectID    `bson:"_id,omitempty" json:"id,omitempty"`
	TargetName string           `bson:"target_name" json:"target_name"`
	Stats      map[string]int64 `bson:"stats" json:"stats"`
	Timestamp  time.Time        `bson:"timestamp" json:"timestamp"`
	ResetCount int              `bson:"reset_count" json:"reset_count"`
}

// GetCollectionName 返回集合名称
func (h *HAProxyStatsBaseline) GetCollectionName() string {
	return "haproxy_baseline"
}

// HAProxyMinuteStats 存储HAProxy分钟统计数据
type HAProxyMinuteStats struct {
	ID         bson.ObjectID    `bson:"_id,omitempty" json:"id,omitempty"`
	TargetName string           `bson:"target_name" json:"target_name"`
	Date       string           `bson:"date" json:"date"`
	Hour       int              `bson:"hour" json:"hour"`
	Minute     int              `bson:"minute" json:"minute"`
	Timestamp  time.Time        `bson:"timestamp" json:"timestamp"`
	Stats      map[string]int64 `bson:"stats" json:"stats"`
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
