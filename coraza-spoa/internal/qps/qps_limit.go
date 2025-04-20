package qps

import (
	"context"
	"fmt"
	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/alibaba/sentinel-golang/core/hotspot"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"sync"
	"time"
)

type QPS struct {
	Db             *mongo.Database
	Logger         *zerolog.Logger
	DstIpLimiter   bool
	DstPathLimiter bool
	SrcIpLimiter   bool
	Lock           sync.RWMutex // 读写锁
}

// Site 原mongo的site结构，部分字段
type Site struct {
	ID           []primitive.ObjectID `bson:"_id"`
	Name         string               `bson:"name"`
	Domain       string               `bson:"domain"`
	Port         int                  `bson:"listenPort"`
	HasDstLimit  bool                 `bson:"hasDstLimit"`  // 是否启用本机限流
	DstLimit     int64                `bson:"dstLimit"`     // 本机限流速率
	HasSrcLimit  bool                 `bson:"hasSrcLimit"`  // 是否启用源ip限流
	IpGroupIds   []primitive.ObjectID `bson:"ipGroupIds"`   // 源ip限流组id列表
	IpGroups     []LimitRateGroup     `bson:"ipGroups"`     // 源ip限流组
	HasPathLimit bool                 `bson:"hasPathLimit"` // 是否启用路径限流
	PathGroups   []PathLimitRule      `bson:"pathGroups"`   // 路径限流组
}

// 路径限流Group
type PathLimitRule struct {
	Target    string `bson:"target"`    // 路径
	Threshold int64  `bson:"threshold"` // 阈值，即QPS限制
}

// LimitRateGroup 限流配置结构
type LimitRateGroup struct {
	ID         []primitive.ObjectID `bson:"_id"`
	Name       string               `bson:"name"`        // 限流分组名称
	Type       string               `bson:"type"`        // 限流类型：scr_ip、dst_ip、path等
	Rules      []IpLimitRule        `bson:"rules"`       // 限流规则配置
	CreateTime time.Time            `bson:"create_time"` // 创建时间
	UpdateTime time.Time            `bson:"update_time"` // 更新时间
}

// LimitRule 限流规则结构
type IpLimitRule struct {
	Target    string `bson:"target"`    // 限流目标，如IP地址
	Threshold int64  `bson:"threshold"` // 阈值，即QPS限制
}

// 初始化
func InitQpsLimit(dbClient *mongo.Client, dbName string, log *zerolog.Logger) (*QPS, error) {
	err := sentinel.InitDefault()
	if err != nil {
		return nil, err
	}
	db := dbClient.Database(dbName)
	qps := &QPS{
		Db:     db,
		Logger: log,
	}
	// TODO : 读取配置文件，设置流量控制参数
	qps.DstIpLimiter = false
	qps.DstPathLimiter = false
	qps.SrcIpLimiter = false
	return qps, nil
}

// IPLimiter 处理基于IP的流量控制
func (qps *QPS) IPLimiter(name, ip string) bool {
	// 使用Sentinel进行流量控制
	resource := fmt.Sprintf("ip-limt:%s", name)
	entry, blockError := sentinel.Entry(resource, sentinel.WithTrafficType(base.Inbound), sentinel.WithArgs(ip))
	if blockError != nil {
		// 请求被限流
		return false
	}
	defer entry.Exit()
	return true
}

// PathLimiter 处理基于路径的流量控制
func (qps *QPS) PathLimiter(name, path string) bool {
	resource := fmt.Sprintf("path-limt:%s", name)
	// 使用Sentinel进行流量控制
	entry, blockError := sentinel.Entry(resource, sentinel.WithTrafficType(base.Inbound), sentinel.WithArgs(path))
	if blockError != nil {
		// 请求被限流
		return false
	}
	defer entry.Exit()
	return true
}

// LoadLimitRulesFromDB 从MongoDB加载限流规则
func (q *QPS) LoadLimitRulesFromDB() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 执行聚合查询，相当于联表查询
	pipeline := mongo.Pipeline{
		// 第一次关联：group_ids -> limit_rate_group
		{{"$lookup", bson.D{
			{"from", "limit_rate_group"},
			{"localField", "ipGroupIds"},
			{"foreignField", "_id"},
			{"as", "ipGroups"},
		}}},
	}
	// 获取limit_rate_group集合
	cursor, err := q.Db.Collection("site").Aggregate(ctx, pipeline)
	if err != nil {
		return fmt.Errorf("查询限流配置失败，联表查询失败: %w", err)
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			q.Logger.Error().Err(err).Msg("关闭限流配置查询游标失败")
		}
	}(cursor, ctx)

	// 解析结果 ip限流重组数据 + 本机ip限流
	result := make(map[string]map[interface{}]int64)
	// 路径限流重组数据
	pathResult := make(map[string]map[interface{}]int64)
	// 遍历查询结果
	for cursor.Next(ctx) {
		var site Site
		if err := cursor.Decode(&site); err != nil {
			q.Logger.Error().Err(err).Msg("解析限流配置记录失败")
			continue
		}
		// 源ip限流
		if site.HasSrcLimit {
			q.SetSrcIpLimiter(true)
			// 合并所有限流组中的IP规则
			var allIPRates []IpLimitRule
			// 源ip限流
			for _, group := range site.IpGroups {
				allIPRates = append(allIPRates, group.Rules...)
			}
			// 处理可能的IP重复情况，如果同一IP在不同限流组中有不同速率，可以选择最严格的
			ipRateMap := make(map[string]int64)
			for _, rate := range allIPRates {
				if existingRate, exists := ipRateMap[rate.Target]; !exists || rate.Threshold < existingRate {
					ipRateMap[rate.Target] = rate.Threshold
				}
			}
			// 转换回数组形式
			uniqueIPRates := make([]IpLimitRule, 0, len(ipRateMap))
			for ip, threshold := range ipRateMap {
				uniqueIPRates = append(uniqueIPRates, IpLimitRule{
					Target:    ip,
					Threshold: threshold,
				})
			}
			result[fmt.Sprintf("ip-limt:%s:%s", site.Domain, site.Port)] = convertIpToMap(uniqueIPRates)
		} else {
			q.SetSrcIpLimiter(false)
		}
		// 本机ip限流
		if site.HasDstLimit {
			q.SetDstIpLimiter(true)
			// 本机ip限流
			dstIpRate := map[interface{}]int64{
				site.Domain: site.DstLimit,
			}
			result[fmt.Sprintf("ip-limt:%s:%s", site.Domain, site.Port)] = dstIpRate
		} else {
			q.SetDstIpLimiter(false)
		}

		if site.HasPathLimit {
			q.SetDstPathLimiter(true)
			// 路径限流
			pathResult[fmt.Sprintf("path-limt:%s:%s", site.Domain, site.Port)] = convertPathToMap(site.PathGroups)
		} else {
			q.SetDstPathLimiter(false)
		}
	}

	if err := cursor.Err(); err != nil {
		return fmt.Errorf("遍历限流配置失败: %w", err)
	}
	err1 := q.AddRule(result, pathResult)
	if err1 != nil {
		return fmt.Errorf("添加限流规则失败: %w", err)
	}
	return nil
}

func (q *QPS) AddRule(ipRules map[string]map[interface{}]int64, pathRules map[string]map[interface{}]int64) error {
	// 存储所有限流规则
	var sentinelRules []*hotspot.Rule
	if len(ipRules) > 0 {
		for resource, rules := range ipRules {
			// 创建sentinel规则
			sentinelRule := &hotspot.Rule{
				Resource:      resource,
				MetricType:    hotspot.QPS,
				ParamIndex:    0,    // 第一个参数，即IP
				Threshold:     1000, // 默认阈值
				BurstCount:    10,
				DurationInSec: 1,
				SpecificItems: rules,
			}
			sentinelRules = append(sentinelRules, sentinelRule)
			q.Logger.Info().Msgf("添加ip限流规则资源: %s, 规则: %v", resource, rules)
		}
	}
	if len(pathRules) > 0 {
		for resource, rules := range pathRules {
			// 创建sentinel规则
			sentinelRule := &hotspot.Rule{
				Resource:      resource,
				MetricType:    hotspot.QPS,
				ParamIndex:    0,    // 第一个参数，即IP
				Threshold:     1000, // 默认阈值
				BurstCount:    10,
				DurationInSec: 1,
				SpecificItems: rules,
			}
			sentinelRules = append(sentinelRules, sentinelRule)
			q.Logger.Info().Msgf("添加路径限流规则资源: %s, 规则: %v", resource, rules)
		}
	}
	// 加载所有规则到Sentinel
	if len(sentinelRules) > 0 {
		_, err := hotspot.LoadRules(sentinelRules)
		if err != nil {
			return err
		}
	}
	return nil
}

// RefreshLimitRules 刷新限流规则
func (q *QPS) RefreshLimitRules() error {
	// 清空现有规则
	err := flow.ClearRules()
	if err != nil {
		return fmt.Errorf("清空限流规则失败: %w", err)
	}
	// 重新加载规则
	return q.LoadLimitRulesFromDB()
}

func convertIpToMap(rules []IpLimitRule) map[interface{}]int64 {
	result := make(map[interface{}]int64)
	for _, rule := range rules {
		result[rule.Target] = rule.Threshold
	}
	return result
}

func convertPathToMap(rules []PathLimitRule) map[interface{}]int64 {
	result := make(map[interface{}]int64)
	for _, rule := range rules {
		result[rule.Target] = rule.Threshold
	}
	return result
}

// 设置DstIpLimiter
func (q *QPS) SetDstIpLimiter(dstIpLimiter bool) {
	q.Lock.Lock()
	defer q.Lock.Unlock()
	q.DstIpLimiter = dstIpLimiter
}

// 设置DstPathLimiter
func (q *QPS) SetDstPathLimiter(dstPathLimiter bool) {
	q.Lock.Lock()
	defer q.Lock.Unlock()
	q.DstPathLimiter = dstPathLimiter
}

// 设置SrcIpLimiter
func (q *QPS) SetSrcIpLimiter(srcIpLimiter bool) {
	q.Lock.Lock()
	defer q.Lock.Unlock()
	q.SrcIpLimiter = srcIpLimiter
}
