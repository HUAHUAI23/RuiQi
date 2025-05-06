import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { attackLogQuerySchema, AttackLogQueryFormValues } from "@/validation/log"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Form, FormControl, FormField, FormItem, FormLabel } from "@/components/ui/form"
import { Card } from "@/components/ui/card"
import { Search, RefreshCw, ChevronDown, ChevronUp, RotateCcw } from "lucide-react"
import { useTranslation } from "react-i18next"
import { useState } from "react"
import { DateTimePicker24h } from "@/components/common/date"
import { Collapse } from "@/components/ui/animation/components/collapse"
import { AnimatedButton } from "@/components/ui/animation/components/animated-button"
import { AnimatedIcon } from "@/components/ui/animation/components/animated-icon"

interface AttackLogFilterProps {
    onFilter: (values: AttackLogQueryFormValues) => void
    onRefresh?: () => void
    defaultValues?: Partial<AttackLogQueryFormValues>
}

export function AttackLogFilter({ onFilter, onRefresh, defaultValues = {} }: AttackLogFilterProps) {
    const { t } = useTranslation()
    const [expanded, setExpanded] = useState(false)

    const [isRefreshAnimating, setIsRefreshAnimating] = useState(false)
    const [isResetAnimating, setIsResetAnimating] = useState(false)

    const form = useForm<AttackLogQueryFormValues>({
        resolver: zodResolver(attackLogQuerySchema),
        defaultValues: {
            ruleId: defaultValues.ruleId || undefined,
            srcIp: defaultValues.srcIp || "",
            dstIp: defaultValues.dstIp || "",
            domain: defaultValues.domain || "",
            srcPort: defaultValues.srcPort || undefined,
            dstPort: defaultValues.dstPort || undefined,
            requestId: defaultValues.requestId || "",
            startTime: defaultValues.startTime || "",
            endTime: defaultValues.endTime || "",
            page: 1,
            pageSize: 10
        }
    })

    const handleSubmit = (values: AttackLogQueryFormValues) => {
        onFilter(values)
    }

    const handleReset = () => {
        setIsResetAnimating(true)
        form.reset({
            ruleId: undefined,
            srcIp: "",
            dstIp: "",
            domain: "",
            srcPort: undefined,
            dstPort: undefined,
            requestId: "",
            startTime: "",
            endTime: "",
            page: 1,
            pageSize: 10
        })
        onFilter(form.getValues())
        setTimeout(() => {
            setIsResetAnimating(false)
        }, 1000)
    }

    const handleRefresh = () => {
        setIsRefreshAnimating(true)
        if (onRefresh) onRefresh()
        setTimeout(() => {
            setIsRefreshAnimating(false)
        }, 1000)
    }


    return (
        <Card className="p-4 bg-zinc-50 dark:bg-muted/30 border-none shadow-none rounded-sm">
            <Form {...form}>
                <form onSubmit={form.handleSubmit(handleSubmit)}>
                    <div className="flex items-center justify-between mb-2">
                        <Button
                            type="button"
                            variant="ghost"
                            size="sm"
                            onClick={() => setExpanded(!expanded)}
                            className="flex items-center gap-1 font-medium"
                        >
                            {t('filter')} {expanded ? <ChevronUp className="h-4 w-4" /> : <ChevronDown className="h-4 w-4" />}
                        </Button>

                        <div className="flex gap-2">
                            <AnimatedButton>
                                <Button
                                    type="button"
                                    variant="outline"
                                    size="sm"
                                    onClick={handleReset}
                                    className="flex items-center gap-1"
                                >
                                    <AnimatedIcon animationVariant="continuous-spin" isAnimating={isResetAnimating} className="h-4 w-4">
                                        <RotateCcw className="h-4 w-4" />
                                    </AnimatedIcon>
                                    {t('reset')}
                                </Button>
                            </AnimatedButton>
                            <AnimatedButton>
                                <Button
                                    type="button"
                                    variant="outline"
                                    size="sm"
                                    onClick={handleRefresh}
                                    className="flex items-center gap-1"
                                >
                                    <AnimatedIcon animationVariant="continuous-spin" isAnimating={isRefreshAnimating} className="h-4 w-4">
                                        <RefreshCw className="h-4 w-4" />
                                    </AnimatedIcon>
                                    {t('refresh')}
                                </Button>
                            </AnimatedButton>
                            <AnimatedButton>
                                <Button
                                    type="submit"
                                    size="sm"
                                    className="flex items-center gap-1"
                                >
                                    <Search className="h-4 w-4" />
                                    {t('search')}
                                </Button>
                            </AnimatedButton>
                        </div>
                    </div>

                    <Collapse isOpen={expanded} animationType="scale">
                        <div className="flex flex-wrap gap-3 mt-3">
                            <FormField
                                control={form.control}
                                name="ruleId"
                                render={({ field }) => (
                                    <FormItem className="justify-between w-full sm:w-[calc(50%-0.375rem)] md:w-[calc(33.33%-0.5rem)] lg:w-[calc(20%-0.6rem)]">
                                        <FormLabel className="text-xs dark:text-shadow-glow-white">{t('ruleId')}</FormLabel>
                                        <FormControl>
                                            <Input
                                                placeholder={t('ruleIdPlaceholder')}
                                                {...field}
                                                onChange={(e) => field.onChange(e.target.value === "" ? undefined : parseInt(e.target.value))}
                                                className="h-8 text-sm bg-white dark:bg-background dark:border-muted-foreground/30 dark:text-shadow-glow-white"
                                            />
                                        </FormControl>
                                    </FormItem>
                                )}
                            />

                            <FormField
                                control={form.control}
                                name="domain"
                                render={({ field }) => (
                                    <FormItem className="justify-between w-full sm:w-[calc(50%-0.375rem)] md:w-[calc(33.33%-0.5rem)] lg:w-[calc(20%-0.6rem)]">
                                        <FormLabel className="text-xs dark:text-shadow-glow-white">{t('domain')}</FormLabel>
                                        <FormControl>
                                            <Input placeholder={t('domainPlaceholder')} {...field} className="h-8 text-sm bg-white dark:bg-background dark:border-muted-foreground/30 dark:text-shadow-glow-white" />
                                        </FormControl>
                                    </FormItem>
                                )}
                            />

                            <FormField
                                control={form.control}
                                name="srcIp"
                                render={({ field }) => (
                                    <FormItem className="justify-between w-full sm:w-[calc(50%-0.375rem)] md:w-[calc(33.33%-0.5rem)] lg:w-[calc(20%-0.6rem)]">
                                        <FormLabel className="text-xs dark:text-shadow-glow-white">{t('srcIp')}</FormLabel>
                                        <FormControl>
                                            <Input placeholder={t('ipPlaceholder')} {...field} className="h-8 text-sm bg-white dark:bg-background dark:border-muted-foreground/30 dark:text-shadow-glow-white" />
                                        </FormControl>
                                    </FormItem>
                                )}
                            />

                            <FormField
                                control={form.control}
                                name="dstIp"
                                render={({ field }) => (
                                    <FormItem className="justify-between w-full sm:w-[calc(50%-0.375rem)] md:w-[calc(33.33%-0.5rem)] lg:w-[calc(20%-0.6rem)]">
                                        <FormLabel className="text-xs dark:text-shadow-glow-white">{t('dstIp')}</FormLabel>
                                        <FormControl>
                                            <Input placeholder={t('ipPlaceholder')} {...field} className="h-8 text-sm bg-white dark:bg-background dark:border-muted-foreground/30 dark:text-shadow-glow-white" />
                                        </FormControl>
                                    </FormItem>
                                )}
                            />

                            <FormField
                                control={form.control}
                                name="srcPort"
                                render={({ field }) => (
                                    <FormItem className="justify-between w-full sm:w-[calc(50%-0.375rem)] md:w-[calc(33.33%-0.5rem)] lg:w-[calc(20%-0.6rem)]">
                                        <FormLabel className="text-xs dark:text-shadow-glow-white">{t('srcPort')}</FormLabel>
                                        <FormControl>
                                            <Input
                                                placeholder={t('portPlaceholder')}
                                                {...field}
                                                onChange={(e) => field.onChange(e.target.value === "" ? undefined : parseInt(e.target.value))}
                                                className="h-8 text-sm bg-white dark:bg-background dark:border-muted-foreground/30 dark:text-shadow-glow-white"
                                            />
                                        </FormControl>
                                    </FormItem>
                                )}
                            />

                            <FormField
                                control={form.control}
                                name="dstPort"
                                render={({ field }) => (
                                    <FormItem className="justify-between w-full sm:w-[calc(50%-0.375rem)] md:w-[calc(33.33%-0.5rem)] lg:w-[calc(20%-0.6rem)]">
                                        <FormLabel className="text-xs dark:text-shadow-glow-white">{t('dstPort')}</FormLabel>
                                        <FormControl>
                                            <Input
                                                placeholder={t('portPlaceholder')}
                                                {...field}
                                                onChange={(e) => field.onChange(e.target.value === "" ? undefined : parseInt(e.target.value))}
                                                className="h-8 text-sm bg-white dark:bg-background dark:border-muted-foreground/30 dark:text-shadow-glow-white"
                                            />
                                        </FormControl>
                                    </FormItem>
                                )}
                            />

                            <FormField
                                control={form.control}
                                name="requestId"
                                render={({ field }) => (
                                    <FormItem className="justify-between w-full sm:w-[calc(50%-0.375rem)] md:w-[calc(33.33%-0.5rem)] lg:w-[calc(20%-0.6rem)]">
                                        <FormLabel className="text-xs dark:text-shadow-glow-white">{t('requestId')}</FormLabel>
                                        <FormControl>
                                            <Input placeholder={t('requestIdPlaceholder')} {...field} className="h-8 text-sm bg-white dark:bg-background dark:border-muted-foreground/30 dark:text-shadow-glow-white" />
                                        </FormControl>
                                    </FormItem>
                                )}
                            />

                            <FormField
                                control={form.control}
                                name="startTime"
                                render={({ field }) => (
                                    <FormItem className="justify-between w-full sm:w-[calc(50%-0.375rem)] md:w-[calc(33.33%-0.5rem)] lg:w-[calc(20%-0.6rem)]">
                                        <FormLabel className="text-xs dark:text-shadow-glow-white">{t('startTime')}</FormLabel>
                                        <FormControl>
                                            <DateTimePicker24h
                                                type="dateHourMinuteSecond"
                                                value={field.value ? new Date(field.value) : undefined}
                                                onChange={(date) => {
                                                    if (!date) {
                                                        // 用户清除了日期
                                                        field.onChange("")
                                                        return
                                                    }

                                                    try {
                                                        const isoString = date.toISOString()
                                                        const formattedDate = isoString.substring(0, 19) + 'Z'
                                                        field.onChange(formattedDate)
                                                    } catch (error) {
                                                        console.error("Invalid date format:", error)
                                                        field.onChange("")
                                                    }
                                                }}
                                                className="dark:bg-background dark:border-muted-foreground/30 dark:text-shadow-glow-white"
                                            />
                                        </FormControl>
                                    </FormItem>
                                )}
                            />

                            <FormField
                                control={form.control}
                                name="endTime"
                                render={({ field }) => (
                                    <FormItem className="justify-between w-full sm:w-[calc(50%-0.375rem)] md:w-[calc(33.33%-0.5rem)] lg:w-[calc(20%-0.6rem)]">
                                        <FormLabel className="text-xs dark:text-shadow-glow-white">{t('endTime')}</FormLabel>
                                        <FormControl>
                                            <DateTimePicker24h
                                                type="dateHourMinuteSecond"
                                                value={field.value ? new Date(field.value) : undefined}
                                                onChange={(date) => {
                                                    if (!date) {
                                                        // 用户清除了日期
                                                        field.onChange("")
                                                        return
                                                    }

                                                    try {
                                                        const isoString = date.toISOString()
                                                        const formattedDate = isoString.substring(0, 19) + 'Z'
                                                        field.onChange(formattedDate)
                                                    } catch (error) {
                                                        console.error("Invalid date format:", error)
                                                        field.onChange("")
                                                    }
                                                }}
                                                className="dark:bg-background dark:border-muted-foreground/30 dark:text-shadow-glow-white"
                                            />
                                        </FormControl>
                                    </FormItem>
                                )}
                            />
                        </div>
                    </Collapse>

                </form>
            </Form>
        </Card>
    )
} 