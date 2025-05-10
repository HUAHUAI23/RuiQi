import { z } from 'zod'
import { Condition, MatchType, TARGET_MATCH_TYPES, TargetType } from '@/types/rule'

const allMatchTypesArray = Object.values(TARGET_MATCH_TYPES).flat() as MatchType[]
const matchTypeEnum = z.enum(allMatchTypesArray as [MatchType, ...MatchType[]])


// 简单条件验证 - 使用superRefine来处理依赖验证
// TODO: 不同的 match_value 类型需要不同的验证方式
const simpleConditionSchema = z.object({
    type: z.literal('simple'),
    target: z.enum(['source_ip', 'url', 'path']),
    match_type: matchTypeEnum,
    match_value: z.string().min(1, { message: 'Match value is required' })
}).superRefine((data, ctx) => {
    // 验证目标类型与匹配方式的兼容性
    const validMatchTypes = TARGET_MATCH_TYPES[data.target as TargetType]
    if (!validMatchTypes.includes(data.match_type as MatchType)) {
        ctx.addIssue({
            code: z.ZodIssueCode.custom,
            message: `Match type '${data.match_type}' is not valid for target '${data.target}'`,
            path: ['match_type']
        })
    }
})

// 递归定义复合条件验证

const conditionSchema: z.ZodType<Condition> = z.lazy(() =>
    z.union([
        simpleConditionSchema,
        z.object({
            type: z.literal('composite'),
            operator: z.enum(['AND', 'OR']),
            conditions: z.array(conditionSchema).min(1)
        })
    ])
)

// 创建规则请求验证
export const ruleCreateSchema = z.object({
    name: z.string().min(1, { message: 'Name is required' }),
    type: z.enum(['whitelist', 'blacklist']),
    status: z.enum(['enabled', 'disabled']),
    priority: z.number().int().min(1).max(10000),
    condition: conditionSchema,
})

// 更新规则请求验证
export const ruleUpdateSchema = ruleCreateSchema.partial()