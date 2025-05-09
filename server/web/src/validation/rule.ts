import { z } from 'zod'
import { MatchType, TARGET_MATCH_TYPES } from '@/types/rule'

// 简单条件验证
const simpleConditionSchema = z.object({
    type: z.literal('simple'),
    target: z.enum(['source_ip', 'url', 'path']),
    match_type: z.string().refine(
        (value) => {
            // 确保匹配方式正确
            const allMatchTypes = Object.values(TARGET_MATCH_TYPES).flat()
            return allMatchTypes.includes(value as MatchType)
        },
        { message: 'Invalid match type for the selected target' }
    ),
    match_value: z.string().min(1, { message: 'Match value is required' })
})

// 递归定义复合条件验证
const conditionSchema: z.ZodType<any> = z.lazy(() =>
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
    priority: z.number().int().min(1).max(1000),
    condition: conditionSchema,
})

// 更新规则请求验证
export const ruleUpdateSchema = ruleCreateSchema.partial()