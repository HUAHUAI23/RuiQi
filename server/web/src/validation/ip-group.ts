import { z } from 'zod'

// IP组创建/编辑表单验证
export const ipGroupFormSchema = z.object({
    name: z
        .string()
        .min(1, { message: 'ipGroup.validation.nameRequired' })
        .max(50, { message: 'ipGroup.validation.nameMaxLength' }),
    items: z
        .array(
            z.string()
                .min(1, { message: 'ipGroup.validation.ipItemRequired' })
                .regex(/^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])(\/([0-9]|[1-2][0-9]|3[0-2]))?$/, {
                    message: 'ipGroup.validation.ipItemInvalid'
                })
        )
        .min(1, { message: 'ipGroup.validation.itemsRequired' })
})

// 查询参数验证
export const ipGroupQuerySchema = z.object({
    page: z.number().optional().default(1),
    pageSize: z.number().optional().default(10)
})

export type IPGroupFormValues = z.infer<typeof ipGroupFormSchema>
export type IPGroupQueryParams = z.infer<typeof ipGroupQuerySchema>