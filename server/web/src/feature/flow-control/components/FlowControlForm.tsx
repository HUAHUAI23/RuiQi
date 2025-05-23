import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { Skeleton } from '@/components/ui/skeleton'
import {
    Form,
    FormControl,
    FormField,
    FormItem,
    FormLabel,
    FormMessage,
    FormDescription,
} from '@/components/ui/form'
import { flowControlConfigSchema, FlowControlConfigFormValues } from '@/validation/flow-control'
import { useUpdateFlowControlConfig, useFlowControlConfig } from '../hooks/useFlowControl'
import { AnimatedContainer } from '@/components/ui/animation/components/animated-container'
import { AnimatedButton } from '@/components/ui/animation/components/animated-button'
import { useTranslation } from 'react-i18next'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { AlertCircle } from 'lucide-react'
import { useEffect } from 'react'

interface FlowControlFormProps {
    onSuccess?: () => void
}

// 流控配置表单骨架屏组件
const FlowControlFormSkeleton = () => {
    return (
        <AnimatedContainer>
            <div className="space-y-8">
                {/* 访问频率限制配置骨架 */}
                <div className="bg-background/50 rounded-xl p-6 space-y-6">
                    <div className="space-y-2">
                        <Skeleton className="h-7 w-40" />
                        <Skeleton className="h-4 w-96" />
                    </div>
                    
                    {/* 启用开关骨架 */}
                    <div className="flex flex-row items-center justify-between py-4">
                        <div className="space-y-1">
                            <Skeleton className="h-5 w-24" />
                            <Skeleton className="h-4 w-32" />
                        </div>
                        <Skeleton className="h-6 w-11 rounded-full" />
                    </div>

                    {/* 配置项网格骨架 */}
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                        {Array.from({ length: 5 }).map((_, index) => (
                            <div key={index} className="space-y-2">
                                <Skeleton className="h-4 w-20" />
                                <Skeleton className="h-10 w-full" />
                                <Skeleton className="h-3 w-32" />
                            </div>
                        ))}
                    </div>
                </div>

                {/* 攻击频率限制配置骨架 */}
                <div className="bg-background/50 rounded-xl p-6 space-y-6">
                    <div className="space-y-2">
                        <Skeleton className="h-7 w-40" />
                        <Skeleton className="h-4 w-96" />
                    </div>
                    
                    <div className="flex flex-row items-center justify-between py-4">
                        <div className="space-y-1">
                            <Skeleton className="h-5 w-24" />
                            <Skeleton className="h-4 w-32" />
                        </div>
                        <Skeleton className="h-6 w-11 rounded-full" />
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                        {Array.from({ length: 5 }).map((_, index) => (
                            <div key={index} className="space-y-2">
                                <Skeleton className="h-4 w-20" />
                                <Skeleton className="h-10 w-full" />
                                <Skeleton className="h-3 w-32" />
                            </div>
                        ))}
                    </div>
                </div>

                {/* 错误频率限制配置骨架 */}
                <div className="bg-background/50 rounded-xl p-6 space-y-6">
                    <div className="space-y-2">
                        <Skeleton className="h-7 w-40" />
                        <Skeleton className="h-4 w-96" />
                    </div>
                    
                    <div className="flex flex-row items-center justify-between py-4">
                        <div className="space-y-1">
                            <Skeleton className="h-5 w-24" />
                            <Skeleton className="h-4 w-32" />
                        </div>
                        <Skeleton className="h-6 w-11 rounded-full" />
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                        {Array.from({ length: 5 }).map((_, index) => (
                            <div key={index} className="space-y-2">
                                <Skeleton className="h-4 w-20" />
                                <Skeleton className="h-10 w-full" />
                                <Skeleton className="h-3 w-32" />
                            </div>
                        ))}
                    </div>
                </div>

                {/* 提交按钮骨架 */}
                <div className="flex justify-end pt-4">
                    <Skeleton className="h-11 w-24" />
                </div>
            </div>
        </AnimatedContainer>
    )
}

export function FlowControlForm({ onSuccess }: FlowControlFormProps) {
    const { t } = useTranslation()

    // 获取当前配置
    const { flowControlConfig, isLoading: isConfigLoading } = useFlowControlConfig()

    // 更新配置
    const {
        updateFlowControlConfig,
        isLoading: isUpdating,
        error,
        clearError
    } = useUpdateFlowControlConfig()

    // 表单设置
    const form = useForm<FlowControlConfigFormValues>({
        resolver: zodResolver(flowControlConfigSchema),
        defaultValues: flowControlConfig || {
            visitLimit: {
                enabled: false,
                threshold: 100,
                statDuration: 60,
                blockDuration: 600,
                burstCount: 10,
                paramsCapacity: 10000,
            },
            attackLimit: {
                enabled: false,
                threshold: 5,
                statDuration: 60,
                blockDuration: 3600,
                burstCount: 2,
                paramsCapacity: 10000,
            },
            errorLimit: {
                enabled: false,
                threshold: 20,
                statDuration: 60,
                blockDuration: 1800,
                burstCount: 5,
                paramsCapacity: 10000,
            },
        },
    })

    // 当配置加载完成时重置表单
    useEffect(() => {
        if (flowControlConfig) {
            form.reset(flowControlConfig)
        }
    }, [flowControlConfig, form])

    // 表单提交处理
    const handleSubmit = (data: FlowControlConfigFormValues) => {
        if (clearError) clearError()

        updateFlowControlConfig({
            engine: {
                flowController: data
            }
        }, {
            onSuccess: () => {
                if (onSuccess) onSuccess()
            }
        })
    }

    if (isConfigLoading) {
        return <FlowControlFormSkeleton />
    }

    return (
        <AnimatedContainer>
            <Form {...form}>
                <form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-8">
                    {/* API错误提示 */}
                    {error && (
                        <Alert variant="destructive">
                            <AlertCircle className="h-4 w-4" />
                            <AlertDescription>{error}</AlertDescription>
                        </Alert>
                    )}

                    {/* 访问频率限制配置 */}
                    <div className="bg-background/50 rounded-xl p-6 space-y-6">
                        <div className="space-y-2">
                            <h3 className="text-xl font-semibold dark:text-shadow-glow-white">
                                {t('flowControl.visitLimit.title', '访问频率限制')}
                            </h3>
                            <p className="text-muted-foreground dark:text-shadow-glow-white">
                                {t('flowControl.visitLimit.description', '限制单个IP在指定时间窗口内的访问次数，防止高频访问')}
                            </p>
                        </div>

                        <FormField
                            control={form.control}
                            name="visitLimit.enabled"
                            render={({ field }) => (
                                <FormItem className="flex flex-row items-center justify-between py-4">
                                    <div className="space-y-1">
                                        <FormLabel className="text-base font-medium dark:text-shadow-glow-white">
                                            {t('flowControl.enabled', '启用限制')}
                                        </FormLabel>
                                        <FormDescription className="text-sm dark:text-shadow-glow-white">
                                            {t('flowControl.enabledDescription', '开启此限制规则')}
                                        </FormDescription>
                                    </div>
                                    <FormControl>
                                        <Switch
                                            checked={field.value}
                                            onCheckedChange={field.onChange}
                                        />
                                    </FormControl>
                                </FormItem>
                            )}
                        />

                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                            <FormField
                                control={form.control}
                                name="visitLimit.threshold"
                                render={({ field }) => (
                                    <FormItem className="space-y-2">
                                        <FormLabel className="font-medium dark:text-shadow-glow-white">
                                            {t('flowControl.threshold', '触发阈值')}
                                        </FormLabel>
                                        <FormControl>
                                            <Input
                                                type="number"
                                                placeholder="100"
                                                className="dark:text-shadow-glow-white"
                                                {...field}
                                                onChange={(e) => field.onChange(parseInt(e.target.value))}
                                            />
                                        </FormControl>
                                        <FormDescription className="text-xs dark:text-shadow-glow-white">
                                            {t('flowControl.thresholdDescription', '统计时间窗口内的最大请求数')}
                                        </FormDescription>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />

                            <FormField
                                control={form.control}
                                name="visitLimit.statDuration"
                                render={({ field }) => (
                                    <FormItem className="space-y-2">
                                        <FormLabel className="font-medium dark:text-shadow-glow-white">
                                            {t('flowControl.statDuration', '统计时间窗口')}
                                        </FormLabel>
                                        <FormControl>
                                            <Input
                                                type="number"
                                                placeholder="60"
                                                className="dark:text-shadow-glow-white"
                                                {...field}
                                                onChange={(e) => field.onChange(parseInt(e.target.value))}
                                            />
                                        </FormControl>
                                        <FormDescription className="text-xs dark:text-shadow-glow-white">
                                            {t('flowControl.statDurationDescription', '统计时间窗口长度（秒）')}
                                        </FormDescription>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />

                            <FormField
                                control={form.control}
                                name="visitLimit.blockDuration"
                                render={({ field }) => (
                                    <FormItem className="space-y-2">
                                        <FormLabel className="font-medium dark:text-shadow-glow-white">
                                            {t('flowControl.blockDuration', '封禁时长')}
                                        </FormLabel>
                                        <FormControl>
                                            <Input
                                                type="number"
                                                placeholder="600"
                                                className="dark:text-shadow-glow-white"
                                                {...field}
                                                onChange={(e) => field.onChange(parseInt(e.target.value))}
                                            />
                                        </FormControl>
                                        <FormDescription className="text-xs dark:text-shadow-glow-white">
                                            {t('flowControl.blockDurationDescription', '触发限制后的封禁时长（秒）')}
                                        </FormDescription>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />

                            <FormField
                                control={form.control}
                                name="visitLimit.burstCount"
                                render={({ field }) => (
                                    <FormItem className="space-y-2">
                                        <FormLabel className="font-medium dark:text-shadow-glow-white">
                                            {t('flowControl.burstCount', '突发请求数')}
                                        </FormLabel>
                                        <FormControl>
                                            <Input
                                                type="number"
                                                placeholder="10"
                                                className="dark:text-shadow-glow-white"
                                                {...field}
                                                onChange={(e) => field.onChange(parseInt(e.target.value))}
                                            />
                                        </FormControl>
                                        <FormDescription className="text-xs dark:text-shadow-glow-white">
                                            {t('flowControl.burstCountDescription', '允许的突发请求数量')}
                                        </FormDescription>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />

                            <FormField
                                control={form.control}
                                name="visitLimit.paramsCapacity"
                                render={({ field }) => (
                                    <FormItem className="space-y-2">
                                        <FormLabel className="font-medium dark:text-shadow-glow-white">
                                            {t('flowControl.paramsCapacity', '缓存容量')}
                                        </FormLabel>
                                        <FormControl>
                                            <Input
                                                type="number"
                                                placeholder="10000"
                                                className="dark:text-shadow-glow-white"
                                                {...field}
                                                onChange={(e) => field.onChange(parseInt(e.target.value))}
                                            />
                                        </FormControl>
                                        <FormDescription className="text-xs dark:text-shadow-glow-white">
                                            {t('flowControl.paramsCapacityDescription', '最多缓存的IP数量')}
                                        </FormDescription>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />
                        </div>
                    </div>

                    {/* 攻击频率限制配置 */}
                    <div className="bg-background/50 rounded-xl p-6 space-y-6">
                        <div className="space-y-2">
                            <h3 className="text-xl font-semibold dark:text-shadow-glow-white">
                                {t('flowControl.attackLimit.title', '攻击频率限制')}
                            </h3>
                            <p className="text-muted-foreground dark:text-shadow-glow-white">
                                {t('flowControl.attackLimit.description', '限制单个IP在指定时间窗口内的攻击次数，防止恶意攻击')}
                            </p>
                        </div>

                        <FormField
                            control={form.control}
                            name="attackLimit.enabled"
                            render={({ field }) => (
                                <FormItem className="flex flex-row items-center justify-between py-4">
                                    <div className="space-y-1">
                                        <FormLabel className="text-base font-medium dark:text-shadow-glow-white">
                                            {t('flowControl.enabled', '启用限制')}
                                        </FormLabel>
                                        <FormDescription className="text-sm dark:text-shadow-glow-white">
                                            {t('flowControl.enabledDescription', '开启此限制规则')}
                                        </FormDescription>
                                    </div>
                                    <FormControl>
                                        <Switch
                                            checked={field.value}
                                            onCheckedChange={field.onChange}
                                        />
                                    </FormControl>
                                </FormItem>
                            )}
                        />

                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                            <FormField
                                control={form.control}
                                name="attackLimit.threshold"
                                render={({ field }) => (
                                    <FormItem className="space-y-2">
                                        <FormLabel className="font-medium dark:text-shadow-glow-white">
                                            {t('flowControl.threshold', '触发阈值')}
                                        </FormLabel>
                                        <FormControl>
                                            <Input
                                                type="number"
                                                placeholder="5"
                                                className="dark:text-shadow-glow-white"
                                                {...field}
                                                onChange={(e) => field.onChange(parseInt(e.target.value))}
                                            />
                                        </FormControl>
                                        <FormDescription className="text-xs dark:text-shadow-glow-white">
                                            {t('flowControl.attackThresholdDescription', '统计时间窗口内的最大攻击次数')}
                                        </FormDescription>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />

                            <FormField
                                control={form.control}
                                name="attackLimit.statDuration"
                                render={({ field }) => (
                                    <FormItem className="space-y-2">
                                        <FormLabel className="font-medium dark:text-shadow-glow-white">
                                            {t('flowControl.statDuration', '统计时间窗口')}
                                        </FormLabel>
                                        <FormControl>
                                            <Input
                                                type="number"
                                                placeholder="60"
                                                className="dark:text-shadow-glow-white"
                                                {...field}
                                                onChange={(e) => field.onChange(parseInt(e.target.value))}
                                            />
                                        </FormControl>
                                        <FormDescription className="text-xs dark:text-shadow-glow-white">
                                            {t('flowControl.statDurationDescription', '统计时间窗口长度（秒）')}
                                        </FormDescription>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />

                            <FormField
                                control={form.control}
                                name="attackLimit.blockDuration"
                                render={({ field }) => (
                                    <FormItem className="space-y-2">
                                        <FormLabel className="font-medium dark:text-shadow-glow-white">
                                            {t('flowControl.blockDuration', '封禁时长')}
                                        </FormLabel>
                                        <FormControl>
                                            <Input
                                                type="number"
                                                placeholder="3600"
                                                className="dark:text-shadow-glow-white"
                                                {...field}
                                                onChange={(e) => field.onChange(parseInt(e.target.value))}
                                            />
                                        </FormControl>
                                        <FormDescription className="text-xs dark:text-shadow-glow-white">
                                            {t('flowControl.blockDurationDescription', '触发限制后的封禁时长（秒）')}
                                        </FormDescription>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />

                            <FormField
                                control={form.control}
                                name="attackLimit.burstCount"
                                render={({ field }) => (
                                    <FormItem className="space-y-2">
                                        <FormLabel className="font-medium dark:text-shadow-glow-white">
                                            {t('flowControl.burstCount', '突发请求数')}
                                        </FormLabel>
                                        <FormControl>
                                            <Input
                                                type="number"
                                                placeholder="2"
                                                className="dark:text-shadow-glow-white"
                                                {...field}
                                                onChange={(e) => field.onChange(parseInt(e.target.value))}
                                            />
                                        </FormControl>
                                        <FormDescription className="text-xs dark:text-shadow-glow-white">
                                            {t('flowControl.burstCountDescription', '允许的突发请求数量')}
                                        </FormDescription>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />

                            <FormField
                                control={form.control}
                                name="attackLimit.paramsCapacity"
                                render={({ field }) => (
                                    <FormItem className="space-y-2">
                                        <FormLabel className="font-medium dark:text-shadow-glow-white">
                                            {t('flowControl.paramsCapacity', '缓存容量')}
                                        </FormLabel>
                                        <FormControl>
                                            <Input
                                                type="number"
                                                placeholder="10000"
                                                className="dark:text-shadow-glow-white"
                                                {...field}
                                                onChange={(e) => field.onChange(parseInt(e.target.value))}
                                            />
                                        </FormControl>
                                        <FormDescription className="text-xs dark:text-shadow-glow-white">
                                            {t('flowControl.paramsCapacityDescription', '最多缓存的IP数量')}
                                        </FormDescription>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />
                        </div>
                    </div>

                    {/* 错误频率限制配置 */}
                    <div className="bg-background/50 rounded-xl p-6 space-y-6">
                        <div className="space-y-2">
                            <h3 className="text-xl font-semibold dark:text-shadow-glow-white">
                                {t('flowControl.errorLimit.title', '错误频率限制')}
                            </h3>
                            <p className="text-muted-foreground dark:text-shadow-glow-white">
                                {t('flowControl.errorLimit.description', '限制单个IP在指定时间窗口内的错误响应次数，防止恶意探测')}
                            </p>
                        </div>

                        <FormField
                            control={form.control}
                            name="errorLimit.enabled"
                            render={({ field }) => (
                                <FormItem className="flex flex-row items-center justify-between py-4">
                                    <div className="space-y-1">
                                        <FormLabel className="text-base font-medium dark:text-shadow-glow-white">
                                            {t('flowControl.enabled', '启用限制')}
                                        </FormLabel>
                                        <FormDescription className="text-sm dark:text-shadow-glow-white">
                                            {t('flowControl.enabledDescription', '开启此限制规则')}
                                        </FormDescription>
                                    </div>
                                    <FormControl>
                                        <Switch
                                            checked={field.value}
                                            onCheckedChange={field.onChange}
                                        />
                                    </FormControl>
                                </FormItem>
                            )}
                        />

                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                            <FormField
                                control={form.control}
                                name="errorLimit.threshold"
                                render={({ field }) => (
                                    <FormItem className="space-y-2">
                                        <FormLabel className="font-medium dark:text-shadow-glow-white">
                                            {t('flowControl.threshold', '触发阈值')}
                                        </FormLabel>
                                        <FormControl>
                                            <Input
                                                type="number"
                                                placeholder="20"
                                                className="dark:text-shadow-glow-white"
                                                {...field}
                                                onChange={(e) => field.onChange(parseInt(e.target.value))}
                                            />
                                        </FormControl>
                                        <FormDescription className="text-xs dark:text-shadow-glow-white">
                                            {t('flowControl.errorThresholdDescription', '统计时间窗口内的最大错误次数')}
                                        </FormDescription>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />

                            <FormField
                                control={form.control}
                                name="errorLimit.statDuration"
                                render={({ field }) => (
                                    <FormItem className="space-y-2">
                                        <FormLabel className="font-medium dark:text-shadow-glow-white">
                                            {t('flowControl.statDuration', '统计时间窗口')}
                                        </FormLabel>
                                        <FormControl>
                                            <Input
                                                type="number"
                                                placeholder="60"
                                                className="dark:text-shadow-glow-white"
                                                {...field}
                                                onChange={(e) => field.onChange(parseInt(e.target.value))}
                                            />
                                        </FormControl>
                                        <FormDescription className="text-xs dark:text-shadow-glow-white">
                                            {t('flowControl.statDurationDescription', '统计时间窗口长度（秒）')}
                                        </FormDescription>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />

                            <FormField
                                control={form.control}
                                name="errorLimit.blockDuration"
                                render={({ field }) => (
                                    <FormItem className="space-y-2">
                                        <FormLabel className="font-medium dark:text-shadow-glow-white">
                                            {t('flowControl.blockDuration', '封禁时长')}
                                        </FormLabel>
                                        <FormControl>
                                            <Input
                                                type="number"
                                                placeholder="1800"
                                                className="dark:text-shadow-glow-white"
                                                {...field}
                                                onChange={(e) => field.onChange(parseInt(e.target.value))}
                                            />
                                        </FormControl>
                                        <FormDescription className="text-xs dark:text-shadow-glow-white">
                                            {t('flowControl.blockDurationDescription', '触发限制后的封禁时长（秒）')}
                                        </FormDescription>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />

                            <FormField
                                control={form.control}
                                name="errorLimit.burstCount"
                                render={({ field }) => (
                                    <FormItem className="space-y-2">
                                        <FormLabel className="font-medium dark:text-shadow-glow-white">
                                            {t('flowControl.burstCount', '突发请求数')}
                                        </FormLabel>
                                        <FormControl>
                                            <Input
                                                type="number"
                                                placeholder="5"
                                                className="dark:text-shadow-glow-white"
                                                {...field}
                                                onChange={(e) => field.onChange(parseInt(e.target.value))}
                                            />
                                        </FormControl>
                                        <FormDescription className="text-xs dark:text-shadow-glow-white">
                                            {t('flowControl.burstCountDescription', '允许的突发请求数量')}
                                        </FormDescription>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />

                            <FormField
                                control={form.control}
                                name="errorLimit.paramsCapacity"
                                render={({ field }) => (
                                    <FormItem className="space-y-2">
                                        <FormLabel className="font-medium dark:text-shadow-glow-white">
                                            {t('flowControl.paramsCapacity', '缓存容量')}
                                        </FormLabel>
                                        <FormControl>
                                            <Input
                                                type="number"
                                                placeholder="10000"
                                                className="dark:text-shadow-glow-white"
                                                {...field}
                                                onChange={(e) => field.onChange(parseInt(e.target.value))}
                                            />
                                        </FormControl>
                                        <FormDescription className="text-xs dark:text-shadow-glow-white">
                                            {t('flowControl.paramsCapacityDescription', '最多缓存的IP数量')}
                                        </FormDescription>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />
                        </div>
                    </div>

                    {/* 提交按钮 */}
                    <div className="flex justify-end pt-4">
                        <AnimatedButton>
                            <Button type="submit" disabled={isUpdating} size="lg">
                                {isUpdating
                                    ? t('flowControl.updating', '更新中...')
                                    : t('flowControl.updateConfig', '更新配置')
                                }
                            </Button>
                        </AnimatedButton>
                    </div>
                </form>
            </Form>
        </AnimatedContainer>
    )
} 