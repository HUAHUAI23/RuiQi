// 匹配目标类型
export type TargetType = 'source_ip' | 'url' | 'path'

// 匹配方式类型
export type MatchType =
    // IP匹配方式
    | 'equal'
    | 'not_equal'
    | 'fuzzy'
    | 'in_cidr'
    | 'not_in_cidr'
    | 'in_ipgroup'
    | 'not_in_ipgroup'
    // URL和Path匹配方式
    | 'include'
    | 'contains'
    | 'not_contains'
    | 'prefix_keyword'
    | 'regex'

// 逻辑操作符
export type LogicalOperator = 'AND' | 'OR'

// 条件类型
export type ConditionType = 'simple' | 'composite'

// 规则类型
export type RuleType = 'whitelist' | 'blacklist'

// 规则状态
export type RuleStatus = 'enabled' | 'disabled'

// 简单条件
export interface SimpleCondition {
    type: 'simple'
    target: TargetType
    match_type: MatchType
    match_value: string
}

// 复合条件
export interface CompositeCondition {
    type: 'composite'
    operator: LogicalOperator
    conditions: (SimpleCondition | CompositeCondition)[]
}

// 条件类型(联合类型)
export type Condition = SimpleCondition | CompositeCondition

// 微规则模型
export interface MicroRule {
    id: string
    name: string
    type: RuleType
    status: RuleStatus
    priority: number
    condition: Condition
    createdAt?: string
    updatedAt?: string
}

// 创建规则请求
export interface MicroRuleCreateRequest {
    name: string
    type: RuleType
    status: RuleStatus
    priority: number
    condition: Condition
}

// 更新规则请求
export interface MicroRuleUpdateRequest {
    name?: string
    type?: RuleType
    status?: RuleStatus
    priority?: number
    condition?: Condition
}

// 规则列表响应
export interface MicroRuleListResponse {
    total: number
    items: MicroRule[]
}

// 目标类型与匹配方式的映射关系
export const TARGET_MATCH_TYPES: Record<TargetType, MatchType[]> = {
    'source_ip': ['equal', 'not_equal', 'fuzzy', 'in_cidr', 'not_in_cidr', 'in_ipgroup', 'not_in_ipgroup'],
    'url': ['equal', 'not_equal', 'contains', 'not_contains', 'prefix_keyword', 'regex'],
    'path': ['equal', 'not_equal', 'contains', 'not_contains', 'prefix_keyword', 'regex']
}