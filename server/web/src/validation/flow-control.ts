import { z } from 'zod'

export const limitConfigSchema = z.object({
    enabled: z.boolean().default(false),
    threshold: z.number()
        .min(1, "阈值不能小于1")
        .max(10000, "阈值不能超过10000")
        .default(100),
    statDuration: z.number()
        .min(10, "统计时间窗口不能小于10秒")
        .max(3600, "统计时间窗口不能超过3600秒")
        .default(60),
    blockDuration: z.number()
        .min(60, "封禁时长不能小于60秒")
        .max(86400, "封禁时长不能超过86400秒")
        .default(600),
    burstCount: z.number()
        .min(1, "突发请求数不能小于1")
        .max(1000, "突发请求数不能超过1000")
        .default(10),
    paramsCapacity: z.number()
        .min(1000, "缓存容量不能小于1000")
        .max(100000, "缓存容量不能超过100000")
        .default(10000),
})

export const flowControlConfigSchema = z.object({
    visitLimit: limitConfigSchema,
    attackLimit: limitConfigSchema,
    errorLimit: limitConfigSchema,
})

export const blockedIPListRequestSchema = z.object({
    page: z.number().min(1).default(1),
    size: z.number().min(1).max(100).default(10),
    ip: z.string().optional(),
    reason: z.string().optional(),
    status: z.enum(['active', 'expired', 'all']).default('all'),
    sortBy: z.enum(['blocked_at', 'blocked_until', 'ip']).default('blocked_at'),
    sortDir: z.enum(['asc', 'desc']).default('desc'),
})

export type LimitConfigFormValues = z.infer<typeof limitConfigSchema>
export type FlowControlConfigFormValues = z.infer<typeof flowControlConfigSchema>
export type BlockedIPListRequestFormValues = z.infer<typeof blockedIPListRequestSchema> 