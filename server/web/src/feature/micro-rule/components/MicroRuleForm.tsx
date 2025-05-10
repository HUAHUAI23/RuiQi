import { useCallback } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import {
    Form,
    FormControl,
    FormField,
    FormItem,
    FormLabel,
    FormMessage,
} from '@/components/ui/form'
import { AlertCircle } from 'lucide-react'
import { ruleCreateSchema } from '@/validation/rule'
import { MicroRuleCreateRequest, Condition, MicroRule } from '@/types/rule'
import { useCreateMicroRule, useUpdateMicroRule } from '../hooks/useMicroRule'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { AnimatedContainer } from '@/components/ui/animation/components/animated-container'
import { useTranslation } from 'react-i18next'
import { AnimatedButton } from '@/components/ui/animation/components/animated-button'
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select'
import { ConditionBuilder } from './ConditionBuilder'

// 默认复合条件
const DEFAULT_CONDITION: Condition = {
    type: 'composite',
    operator: 'AND',
    conditions: [
        {
            type: 'simple',
            target: 'source_ip',
            match_type: 'equal',
            match_value: '',
        }
    ]
}

interface MicroRuleFormProps {
    mode?: 'create' | 'update'
    ruleId?: string
    onSuccess?: () => void
    defaultValues?: MicroRule
}

export function MicroRuleForm({
    mode = 'create',
    ruleId,
    onSuccess,
    defaultValues,
}: MicroRuleFormProps) {
    const { t } = useTranslation()

    // 准备默认值
    const initialValues: MicroRuleCreateRequest = defaultValues || {
        name: '',
        type: 'blacklist',
        status: 'enabled',
        priority: 100,
        condition: DEFAULT_CONDITION,
    }

    // API钩子
    const {
        createMicroRule,
        isLoading: isCreating,
        error: createError,
        clearError: clearCreateError
    } = useCreateMicroRule()

    const {
        updateMicroRule,
        isLoading: isUpdating,
        error: updateError,
        clearError: clearUpdateError
    } = useUpdateMicroRule()

    // 动态状态
    const isLoading = mode === 'create' ? isCreating : isUpdating
    const error = mode === 'create' ? createError : updateError
    const clearError = mode === 'create' ? clearCreateError : clearUpdateError

    // 表单设置
    const form = useForm<MicroRuleCreateRequest>({
        resolver: zodResolver(ruleCreateSchema),
        defaultValues: initialValues,
    })


    // 表单提交处理
    const handleFormSubmit = useCallback((data: MicroRuleCreateRequest) => {
        // 清除之前的错误
        if (clearError) clearError()


        // 根据模式执行创建或更新操作
        if (mode === 'create') {
            createMicroRule(data, {
                onSuccess: () => {
                    // 重置表单
                    form.reset({
                        name: '',
                        type: 'blacklist',
                        status: 'enabled',
                        priority: 100,
                        condition: DEFAULT_CONDITION,
                    })
                    // 通知父组件成功
                    if (onSuccess) onSuccess()
                }
            })
        } else if (mode === 'update' && ruleId) {
            updateMicroRule({ id: ruleId, data }, {
                onSuccess: () => {
                    // 通知父组件成功
                    if (onSuccess) onSuccess()
                }
            })
        }
    }, [mode, ruleId, clearError, createMicroRule, updateMicroRule, form, onSuccess])

    return (
        <AnimatedContainer>
            <Form {...form}>
                <form onSubmit={form.handleSubmit(handleFormSubmit)} className="space-y-6">
                    {/* API错误提示 */}
                    {error && (
                        <Alert variant="destructive">
                            <AlertCircle className="h-4 w-4" />
                            <AlertDescription>{error}</AlertDescription>
                        </Alert>
                    )}

                    {/* 基本信息字段 */}
                    <div className="grid grid-cols-2 gap-4">
                        <FormField
                            control={form.control}
                            name="name"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel className="dark:text-shadow-glow-white">{t("microRule.form.name")}</FormLabel>
                                    <FormControl>
                                        <Input className="dark:text-shadow-glow-white" placeholder={t("microRule.form.namePlaceholder")} {...field} />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        <FormField
                            control={form.control}
                            name="priority"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel className="dark:text-shadow-glow-white">{t("microRule.form.priority")}</FormLabel>
                                    <FormControl>
                                        <Input
                                            className="dark:text-shadow-glow-white"
                                            placeholder={t("microRule.form.priorityPlaceholder")}
                                            {...field}
                                            onChange={(e) => field.onChange(Number(e.target.value))}
                                        />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        {/* 规则类型 */}
                        <FormField
                            control={form.control}
                            name="type"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel className="dark:text-shadow-glow-white">{t("microRule.form.type")}</FormLabel>
                                    <Select
                                        onValueChange={field.onChange}
                                        defaultValue={field.value}
                                        value={field.value}
                                    >
                                        <FormControl>
                                            <SelectTrigger className="dark:text-shadow-glow-white">
                                                <SelectValue placeholder={t("microRule.form.selectType")} />
                                            </SelectTrigger>
                                        </FormControl>
                                        <SelectContent>
                                            <SelectItem value="whitelist">{t("microRule.form.whitelist")}</SelectItem>
                                            <SelectItem value="blacklist">{t("microRule.form.blacklist")}</SelectItem>
                                        </SelectContent>
                                    </Select>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        {/* 规则状态 */}
                        <FormField
                            control={form.control}
                            name="status"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel className="dark:text-shadow-glow-white">{t("microRule.form.status")}</FormLabel>
                                    <Select
                                        onValueChange={field.onChange}
                                        defaultValue={field.value}
                                        value={field.value}
                                    >
                                        <FormControl>
                                            <SelectTrigger className="dark:text-shadow-glow-white">
                                                <SelectValue placeholder={t("microRule.form.selectStatus")} />
                                            </SelectTrigger>
                                        </FormControl>
                                        <SelectContent>
                                            <SelectItem value="enabled">{t("microRule.form.enabled")}</SelectItem>
                                            <SelectItem value="disabled">{t("microRule.form.disabled")}</SelectItem>
                                        </SelectContent>
                                    </Select>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />
                    </div>

                    {/* 条件构建器 */}

                    <div className="space-y-4 mt-6">
                        <FormLabel className="dark:text-shadow-glow-white text-base">{t("microRule.form.condition")}</FormLabel>
                        <div className="border p-6 rounded-md dark:border-gray-700">
                            <ConditionBuilder
                                form={form}
                                path="condition"
                                isRoot={true}
                            />
                        </div>
                    </div>


                    {/* 提交按钮 */}
                    <div className="flex justify-end mt-6">
                        <AnimatedButton>
                            <Button type="submit" disabled={isLoading}>
                                {isLoading ? t("common.submitting") : mode === 'create' ? t("common.create") : t("common.save")}
                            </Button>
                        </AnimatedButton>
                    </div>
                </form>
            </Form>
        </AnimatedContainer>
    )
}