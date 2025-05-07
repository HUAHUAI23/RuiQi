package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/HUAHUAI23/simple-waf/pkg/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// 匹配方式
type MatchType string

const (
	// IP匹配方式
	MatchEqual        MatchType = "equal"
	MatchNotEqual     MatchType = "not_equal"
	MatchFuzzy        MatchType = "fuzzy"
	MatchInCIDR       MatchType = "in_cidr"
	MatchNotInCIDR    MatchType = "not_in_cidr"
	MatchInIPGroup    MatchType = "in_ipgroup"
	MatchNotInIPGroup MatchType = "not_in_ipgroup"

	// URL和Path匹配方式
	MatchInclude       MatchType = "include"
	MatchContains      MatchType = "contains"
	MatchNotContains   MatchType = "not_contains"
	MatchPrefixKeyword MatchType = "prefix_keyword"
	MatchRegex         MatchType = "regex"
)

// 匹配目标类型
type TargetType string

const (
	SourceIP   TargetType = "source_ip"
	TargetURL  TargetType = "url"
	TargetPath TargetType = "path"
)

// 逻辑操作符
type LogicalOperator string

const (
	LogicalAND LogicalOperator = "AND"
	LogicalOR  LogicalOperator = "OR"
)

// Matcher接口定义了条件匹配的方法
type Matcher interface {
	Match(eng *RuleEngine, ip, url, path string) (bool, error)
}

// 条件类型
type ConditionType string

const (
	SimpleConditionType    ConditionType = "simple"
	CompositeConditionType ConditionType = "composite"
)

// SimpleCondition 简单条件
type SimpleCondition struct {
	Type       ConditionType `json:"type" bson:"type"`
	Target     TargetType    `json:"target" bson:"target"`
	MatchType  MatchType     `json:"match_type" bson:"match_type"`
	MatchValue string        `json:"match_value" bson:"match_value"`
}

// Match 实现Matcher接口
func (c *SimpleCondition) Match(eng *RuleEngine, ip, url, path string) (bool, error) {
	switch c.Target {
	case SourceIP:
		return eng.matchIP(c, ip)
	case TargetURL:
		return eng.matchURL(c, url)
	case TargetPath:
		return eng.matchPath(c, path)
	default:
		return false, fmt.Errorf("不支持的目标类型: %s", c.Target)
	}
}

// CompositeCondition 复合条件
type CompositeCondition struct {
	Type       ConditionType     `json:"type" bson:"type"`
	Operator   LogicalOperator   `json:"operator" bson:"operator"`
	Conditions []json.RawMessage `json:"conditions" bson:"conditions"`

	// 运行时字段，不用于JSON/BSON
	parsedConditions []Matcher
}

// Match 实现Matcher接口
func (c *CompositeCondition) Match(eng *RuleEngine, ip, url, path string) (bool, error) {
	if len(c.parsedConditions) == 0 {
		return false, fmt.Errorf("复合条件未初始化")
	}

	var result bool
	if c.Operator == LogicalAND {
		result = true
	} else {
		result = false
	}

	for _, condition := range c.parsedConditions {
		match, err := condition.Match(eng, ip, url, path)
		if err != nil {
			return false, err
		}

		if c.Operator == LogicalAND {
			result = result && match
			if !result {
				return false, nil // 短路
			}
		} else {
			result = result || match
			if result {
				return true, nil // 短路
			}
		}
	}

	return result, nil
}

// ConditionFactory 条件工厂
type ConditionFactory struct{}

// ParseCondition 解析条件
func (f *ConditionFactory) ParseCondition(data json.RawMessage) (Matcher, error) {
	var baseCondition struct {
		Type ConditionType `json:"type" bson:"type"`
	}

	if err := json.Unmarshal(data, &baseCondition); err != nil {
		return nil, fmt.Errorf("解析条件类型失败: %v", err)
	}

	switch baseCondition.Type {
	case SimpleConditionType:
		var condition SimpleCondition
		if err := json.Unmarshal(data, &condition); err != nil {
			return nil, fmt.Errorf("解析简单条件失败: %v", err)
		}
		return &condition, nil

	case CompositeConditionType:
		var condition CompositeCondition
		if err := json.Unmarshal(data, &condition); err != nil {
			return nil, fmt.Errorf("解析复合条件失败: %v", err)
		}

		condition.parsedConditions = make([]Matcher, 0, len(condition.Conditions))
		for _, rawCondition := range condition.Conditions {
			parsedCondition, err := f.ParseCondition(rawCondition)
			if err != nil {
				return nil, err
			}
			condition.parsedConditions = append(condition.parsedConditions, parsedCondition)
		}

		return &condition, nil

	default:
		return nil, fmt.Errorf("不支持的条件类型: %s", baseCondition.Type)
	}
}

// Rule 规则 - 添加优先级和序列号字段
type Rule struct {
	// 嵌入MicroRule
	model.MicroRule `bson:",inline" json:",inline"`

	// 运行时字段，不用于JSON/BSON
	parsedCondition Matcher `bson:"-" json:"-"`
	sequence        int     `bson:"-" json:"-"`
}

// MongoDB配置
type MongoDBConfig struct {
	mongoClient       *mongo.Client
	Database          string // 数据库名称
	RuleCollection    string // 规则集合名称
	IPGroupCollection string // IP组集合名称
}

// RuleEngine 规则引擎
type RuleEngine struct {
	Rules       []Rule                    `json:"rules"`     // 所有规则列表
	IPGroups    map[string]*model.IPGroup `json:"ip_groups"` // IP组映射表
	regexCache  map[string]*regexp.Regexp // 正则表达式缓存
	factory     ConditionFactory          // 条件工厂
	mongoConfig MongoDBConfig             // MongoDB配置
}

// NewRuleEngine 创建规则引擎
func NewRuleEngine() *RuleEngine {
	return &RuleEngine{
		Rules:      make([]Rule, 0),
		IPGroups:   make(map[string]*model.IPGroup),
		regexCache: make(map[string]*regexp.Regexp),
		factory:    ConditionFactory{},
	}
}

func (e *RuleEngine) InitMongoConfig(config MongoDBConfig) error {
	e.mongoConfig = config
	return nil
}

// LoadRulesFromMongoDB 从MongoDB加载规则

func (e *RuleEngine) LoadRulesFromMongoDB() error {
	if e.mongoConfig.mongoClient == nil {
		return fmt.Errorf("MongoDB客户端未初始化")
	}

	// 获取集合
	collection := e.mongoConfig.mongoClient.
		Database(e.mongoConfig.Database).
		Collection(e.mongoConfig.RuleCollection)

	// 查询所有规则
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.D{})
	if err != nil {
		return fmt.Errorf("查询规则失败: %v", err)
	}
	defer cursor.Close(ctx)

	// 解码规则
	var rules []Rule
	if err = cursor.All(ctx, &rules); err != nil {
		return fmt.Errorf("解码规则失败: %v", err)
	}

	// 设置序列号 - 记录规则在原始配置中的顺序
	for i := range rules {
		rules[i].sequence = i
	}

	// 解析每个规则的条件
	for i := range rules {
		rule := &rules[i]
		parsedCondition, err := e.factory.ParseCondition(rule.Condition)
		if err != nil {
			return fmt.Errorf("解析规则 %s 的条件失败: %v", rule.ID, err)
		}
		rule.parsedCondition = parsedCondition
	}

	// 按照优先级排序，优先级相同时按照原始顺序排序
	sort.Slice(rules, func(i, j int) bool {
		if rules[i].Priority != rules[j].Priority {
			return rules[i].Priority > rules[j].Priority // 优先级高的排在前面
		}
		return rules[i].sequence < rules[j].sequence // 优先级相同时按原始顺序
	})

	e.Rules = rules
	return nil
}

// LoadIPGroupsFromMongoDB 从MongoDB加载IP组
func (e *RuleEngine) LoadIPGroupsFromMongoDB() error {
	if e.mongoConfig.mongoClient == nil {
		return fmt.Errorf("MongoDB客户端未初始化")
	}

	// 获取集合
	collection := e.mongoConfig.mongoClient.
		Database(e.mongoConfig.Database).
		Collection(e.mongoConfig.IPGroupCollection)

	// 查询所有IP组
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 使用bson.D{}代替bson.M{}
	cursor, err := collection.Find(ctx, bson.D{})
	if err != nil {
		return fmt.Errorf("查询IP组失败: %v", err)
	}
	defer cursor.Close(ctx)

	// 解码IP组
	var ipGroups []model.IPGroup
	if err = cursor.All(ctx, &ipGroups); err != nil {
		return fmt.Errorf("解码IP组失败: %v", err)
	}

	// 验证IP组并添加到映射表
	for _, group := range ipGroups {
		for _, item := range group.Items {
			if !isValidIPOrCIDR(item) {
				return fmt.Errorf("IP组 %s 中包含无效的IP或CIDR: %s", group.Name, item)
			}
		}
		e.IPGroups[group.Name] = &group
	}
	return nil
}

// LoadAllFromMongoDB 从MongoDB加载所有规则和IP组
func (e *RuleEngine) LoadAllFromMongoDB() error {
	if err := e.LoadIPGroupsFromMongoDB(); err != nil {
		return err
	}

	return e.LoadRulesFromMongoDB()
}

// AddIPGroup 添加IP组
func (e *RuleEngine) AddIPGroup(group model.IPGroup) error {
	if _, exists := e.IPGroups[group.Name]; exists {
		return fmt.Errorf("IP组 %s 已存在", group.Name)
	}

	for _, item := range group.Items {
		if !isValidIPOrCIDR(item) {
			return fmt.Errorf("IP组 %s 中包含无效的IP或CIDR: %s", group.Name, item)
		}
	}

	e.IPGroups[group.Name] = &group
	return nil
}

// LoadRulesFromJSON 从JSON加载规则 - 修改加载逻辑，增加序列号处理
func (e *RuleEngine) LoadRulesFromJSON(data []byte) error {
	var rules []Rule
	if err := json.Unmarshal(data, &rules); err != nil {
		return err
	}

	// 设置序列号 - 记录规则在原始配置中的顺序
	for i := range rules {
		rules[i].sequence = i
	}

	// 解析每个规则的条件
	for i := range rules {
		rule := &rules[i]
		parsedCondition, err := e.factory.ParseCondition(rule.Condition)
		if err != nil {
			return fmt.Errorf("解析规则 %s 的条件失败: %v", rule.ID, err)
		}
		rule.parsedCondition = parsedCondition
	}

	// 按照优先级排序，优先级相同时按照原始顺序排序
	sort.Slice(rules, func(i, j int) bool {
		if rules[i].Priority != rules[j].Priority {
			return rules[i].Priority > rules[j].Priority // 优先级高的排在前面
		}
		return rules[i].sequence < rules[j].sequence // 优先级相同时按原始顺序
	})

	e.Rules = rules
	return nil
}

// AddRule 添加单个规则
func (e *RuleEngine) AddRule(rule Rule) error {
	// 解析规则条件
	parsedCondition, err := e.factory.ParseCondition(rule.Condition)
	if err != nil {
		return fmt.Errorf("解析规则 %s 的条件失败: %v", rule.ID, err)
	}
	rule.parsedCondition = parsedCondition

	// 设置规则序列号为当前规则列表长度
	rule.sequence = len(e.Rules)

	// 添加规则到列表
	e.Rules = append(e.Rules, rule)

	// 重新排序规则
	sort.Slice(e.Rules, func(i, j int) bool {
		if e.Rules[i].Priority != e.Rules[j].Priority {
			return e.Rules[i].Priority > e.Rules[j].Priority
		}
		return e.Rules[i].sequence < e.Rules[j].sequence
	})

	return nil
}

// MatchRequest 匹配请求
// 参数：
// - ip: 源IP地址
// - url: 请求URL
// - path: 请求路径
// 返回值：
// - shouldBlock: 是否应该拦截请求 (true表示拦截，false表示放行)
// - ruleType: 匹配的规则类型
// - rule: 匹配的规则
// - error: 错误信息
func (e *RuleEngine) MatchRequest(ip string, url string, path string) (shouldBlock bool, ruleType model.RuleType, rule *Rule, err error) {
	// 验证IP地址格式
	if !isValidIP(ip) {
		return false, "", nil, fmt.Errorf("无效的IP地址: %s", ip)
	}

	// 遍历所有规则（已按优先级和序列号排序）
	for _, r := range e.Rules {
		// 跳过禁用的规则
		if r.Status == model.RuleDisabled {
			continue
		}

		// 匹配规则条件
		match, err := r.parsedCondition.Match(e, ip, url, path)
		if err != nil {
			return false, "", nil, err
		}

		// 如果规则条件匹配
		if match {
			// 根据规则类型确定是否需要拦截
			switch r.Type {
			case model.BlacklistRule:
				// 黑名单规则匹配成功 -> 返回true(拦截)
				return true, r.Type, &r, nil
			case model.WhitelistRule:
				// 白名单规则匹配成功 -> 返回false(放行)
				return false, r.Type, &r, nil
			default:
				return false, "", nil, fmt.Errorf("未知的规则类型: %s", r.Type)
			}
		}
	}

	// 如果没有匹配任何规则，默认放行
	return false, "", nil, nil
}

// GetRules 获取当前规则列表
func (e *RuleEngine) GetRules() []Rule {
	return e.Rules
}

// matchIP 匹配IP条件
func (e *RuleEngine) matchIP(cond *SimpleCondition, ip string) (bool, error) {
	switch cond.MatchType {
	case MatchEqual:
		return ip == cond.MatchValue, nil
	case MatchNotEqual:
		return ip != cond.MatchValue, nil
	case MatchFuzzy:
		return matchIPFuzzy(ip, cond.MatchValue)
	case MatchInCIDR:
		return isIPInCIDR(ip, cond.MatchValue)
	case MatchNotInCIDR:
		inCIDR, err := isIPInCIDR(ip, cond.MatchValue)
		return !inCIDR, err
	case MatchInIPGroup:
		return e.isIPInGroup(ip, cond.MatchValue)
	case MatchNotInIPGroup:
		inGroup, err := e.isIPInGroup(ip, cond.MatchValue)
		return !inGroup, err
	default:
		return false, fmt.Errorf("IP不支持匹配方式: %s", cond.MatchType)
	}
}

// matchURL 匹配URL条件
func (e *RuleEngine) matchURL(cond *SimpleCondition, url string) (bool, error) {
	switch cond.MatchType {
	case MatchEqual:
		return url == cond.MatchValue, nil
	case MatchNotEqual:
		return url != cond.MatchValue, nil
	case MatchInclude, MatchContains:
		return strings.Contains(url, cond.MatchValue), nil
	case MatchNotContains:
		return !strings.Contains(url, cond.MatchValue), nil
	case MatchPrefixKeyword:
		return strings.HasPrefix(url, cond.MatchValue), nil
	case MatchRegex:
		return e.matchRegex(url, cond.MatchValue)
	default:
		return false, fmt.Errorf("URL不支持匹配方式: %s", cond.MatchType)
	}
}

// matchPath 匹配Path条件
func (e *RuleEngine) matchPath(cond *SimpleCondition, path string) (bool, error) {
	return e.matchURL(cond, path)
}

// 以下是辅助函数

// isValidIP 检查IP是否有效
func isValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// isValidCIDR 检查CIDR是否有效
func isValidCIDR(cidr string) bool {
	_, _, err := net.ParseCIDR(cidr)
	return err == nil
}

// isValidIPOrCIDR 检查是否为有效的IP或CIDR
func isValidIPOrCIDR(s string) bool {
	return isValidIP(s) || isValidCIDR(s)
}

// isValidIPPattern 检查IP模式是否有效
func isValidIPPattern(pattern string) bool {
	parts := strings.Split(pattern, ".")
	if len(parts) != 4 {
		return false
	}

	for _, part := range parts {
		if part != "*" {
			if part == "" {
				return false
			}

			var num int
			if _, err := fmt.Sscanf(part, "%d", &num); err != nil {
				return false
			}

			if num < 0 || num > 255 {
				return false
			}
		}
	}

	return true
}

// isIPInCIDR 检查IP是否在CIDR范围内
func isIPInCIDR(ipStr, cidrStr string) (bool, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false, fmt.Errorf("无效的IP地址: %s", ipStr)
	}

	_, ipNet, err := net.ParseCIDR(cidrStr)
	if err != nil {
		return false, fmt.Errorf("无效的CIDR: %s", cidrStr)
	}

	return ipNet.Contains(ip), nil
}

// matchIPFuzzy 模糊匹配IP
func matchIPFuzzy(ip, pattern string) (bool, error) {
	ipParts := strings.Split(ip, ".")
	patternParts := strings.Split(pattern, ".")

	if len(ipParts) != 4 || len(patternParts) != 4 {
		return false, fmt.Errorf("IP格式错误")
	}

	for i := 0; i < 4; i++ {
		if patternParts[i] != "*" && ipParts[i] != patternParts[i] {
			return false, nil
		}
	}

	return true, nil
}

// isIPInGroup 检查IP是否在IP组中
func (e *RuleEngine) isIPInGroup(ip, groupName string) (bool, error) {
	group, exists := e.IPGroups[groupName]
	if !exists {
		return false, fmt.Errorf("IP组不存在: %s", groupName)
	}

	for _, item := range group.Items {
		if isValidIP(item) {
			if ip == item {
				return true, nil
			}
		} else if isValidCIDR(item) {
			inCIDR, err := isIPInCIDR(ip, item)
			if err != nil {
				return false, err
			}
			if inCIDR {
				return true, nil
			}
		} else {
			return false, fmt.Errorf("IP组 %s 包含无效项: %s", groupName, item)
		}
	}

	return false, nil
}

// matchRegex 正则表达式匹配
func (e *RuleEngine) matchRegex(s, pattern string) (bool, error) {
	re, exists := e.regexCache[pattern]
	if !exists {
		var err error
		re, err = regexp.Compile(pattern)
		if err != nil {
			return false, fmt.Errorf("无效的正则表达式: %s", pattern)
		}
		e.regexCache[pattern] = re
	}

	return re.MatchString(s), nil
}
