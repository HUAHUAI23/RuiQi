package model

import "time"

// BlockedIPRecord IP封禁记录
// @Description 被封禁的IP记录信息
type BlockedIPRecord struct {
	IP           string    `bson:"ip" json:"ip" example:"192.168.1.1" description:"被封禁的IP地址"`
	Reason       string    `bson:"reason" json:"reason" example:"high_frequency_attack" description:"封禁原因"`
	RequestUri   string    `bson:"request_uri" json:"requestUri" example:"/api/v1/login" description:"请求URI"`
	BlockedAt    time.Time `bson:"blocked_at" json:"blockedAt" description:"封禁开始时间"`
	BlockedUntil time.Time `bson:"blocked_until" json:"blockedUntil" description:"封禁结束时间"`
}

func (b *BlockedIPRecord) GetCollectionName() string {
	return "blocked_ips"
}
