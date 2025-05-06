import { useState } from "react"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { HelpCircle, RefreshCw, ChevronDown, ChevronUp } from "lucide-react"
import { Button } from "@/components/ui/button"
import {
    Collapsible,
    CollapsibleContent,
    CollapsibleTrigger
} from "@/components/ui/collapsible"
import { cn } from "@/lib/utils"
import { ApiError } from "@/api/index"
import { getErrorType, ErrorType } from "./errorTypes"
import { errorConfigs } from "./errorConfig"
import { useTranslation } from "react-i18next"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"

interface AdvancedErrorDisplayProps {
    /** 错误对象 */
    error: unknown
    /** 重试回调函数 */
    onRetry?: () => void
    /** 额外的CSS类名 */
    className?: string
    /** 是否使用卡片样式 (更现代化的UI) */
    useCardStyle?: boolean
}

/**
 * 高级错误展示组件
 */
export const AdvancedErrorDisplay = ({
    error,
    onRetry,
    className,
    useCardStyle = true
}: AdvancedErrorDisplayProps) => {
    const { t } = useTranslation()
    const [showDetails, setShowDetails] = useState(false)

    // 处理错误信息
    const isApiError = error instanceof Error && error.name === 'ApiError'
    const errorMessage = error instanceof Error ? error.message : t('error.loading', '加载数据时发生错误')
    const apiError = isApiError ? (error as ApiError) : undefined
    const errorCode = apiError?.code || (error instanceof Error && 'status' in error ? (error as unknown as { status: number }).status : undefined)
    const errorType = getErrorType(error)
    const requestId = apiError?.requestId
    const errorDetail = apiError?.errorDetail

    // 获取错误配置
    const config = errorConfigs[errorType]
    const title = t(config.titleKey, config.titleKey)
    const description = t(config.descriptionKey, config.descriptionKey)
    const finalDescription = errorType === ErrorType.CLIENT || errorType === ErrorType.VALIDATION
        ? errorMessage || description
        : description

    /**
     * 格式化错误详情对象
     * @param detail - 错误详情
     * @returns 格式化后的错误详情字符串
     */
    const formatErrorDetail = (detail: unknown): string | null => {
        if (!detail) return null

        try {
            if (typeof detail === 'string') {
                // 尝试解析JSON字符串
                try {
                    const parsed = JSON.parse(detail)
                    return JSON.stringify(parsed, null, 2)
                } catch {
                    // 如果不是JSON，直接返回字符串
                    return detail
                }
            } else if (typeof detail === 'object') {
                return JSON.stringify(detail, null, 2)
            }

            return String(detail)
        } catch {
            return t('error.cannotParse', '无法解析错误详情')
        }
    }

    const formattedErrorDetail = formatErrorDetail(errorDetail)

    if (useCardStyle) {
        return (
            <Card className={cn("mx-auto my-6 max-w-xl border-red-200 dark:border-red-900/50 shadow-md", className)}>
                <CardHeader className="pb-2">
                    <div className="flex items-center gap-3">
                        <div className="bg-red-50 dark:bg-red-900/30 p-2 rounded-full">
                            {config.icon}
                        </div>
                        <div className="flex-1">
                            <CardTitle className="text-lg text-red-700 dark:text-red-400">{title}</CardTitle>
                            {errorCode && (
                                <Badge variant="outline" className="mt-1 text-xs border-red-200 dark:border-red-900/50 text-red-500 dark:text-red-400">
                                    {t('error.code', '错误代码')}: {errorCode}
                                </Badge>
                            )}
                        </div>
                    </div>
                </CardHeader>
                <CardContent className="pb-2">
                    <CardDescription className="text-sm text-red-600 dark:text-red-400">
                        {finalDescription}
                    </CardDescription>
                    
                    <Collapsible open={showDetails} onOpenChange={setShowDetails} className="mt-4">
                        <div className="flex items-center justify-between">
                            <CollapsibleTrigger asChild>
                                <Button
                                    variant="ghost"
                                    size="sm"
                                    className="px-2 text-xs flex items-center gap-1 text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20"
                                >
                                    {showDetails ? <ChevronUp className="h-3 w-3" /> : <ChevronDown className="h-3 w-3" />}
                                    {showDetails ? t('error.hideDetails', '隐藏详情') : t('error.viewDetails', '查看详情')}
                                </Button>
                            </CollapsibleTrigger>
                        </div>

                        <CollapsibleContent>
                            <div className="mt-3 p-3 bg-red-50 dark:bg-red-900/20 rounded-md text-xs font-mono overflow-auto">
                                <p className="break-words overflow-hidden max-h-[100px] dark:text-gray-300">{String(errorMessage)}</p>
                                {errorCode && <p className="mt-1 dark:text-gray-300">{t('error.code', '错误代码')}: {errorCode}</p>}
                                {requestId && <p className="mt-1 dark:text-gray-300">{t('error.requestId', '请求ID')}: {requestId}</p>}
                                {formattedErrorDetail && (
                                    <div className="mt-1">
                                        <p className="dark:text-gray-300">{t('error.details', '错误详情')}:</p>
                                        <pre className="mt-1 p-2 bg-red-100 dark:bg-red-900/30 rounded-md overflow-x-auto dark:text-gray-300">
                                            {formattedErrorDetail}
                                        </pre>
                                    </div>
                                )}
                            </div>
                        </CollapsibleContent>
                    </Collapsible>
                </CardContent>
                {onRetry && errorType !== ErrorType.UNAUTHORIZED && (
                    <CardFooter className="pt-2">
                        <Button
                            variant="outline"
                            size="sm"
                            onClick={onRetry}
                            className="flex items-center gap-2 text-xs border-red-300 dark:border-red-900/50 text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 mt-2"
                        >
                            <RefreshCw className="h-3 w-3" />
                            {t('error.retry', '重试')}
                        </Button>
                    </CardFooter>
                )}
            </Card>
        )
    }

    // 经典 Alert 样式
    return (
        <Alert variant="destructive" className={cn("mx-auto my-6 max-w-xl", className)}>
            {config.icon}
            <AlertTitle className="font-medium mt-0">{title}</AlertTitle>
            <AlertDescription className="mt-2 flex flex-col gap-4">
                <p className="text-sm">{finalDescription}</p>

                <Collapsible open={showDetails} onOpenChange={setShowDetails}>
                    <div className="flex items-center justify-between">
                        <CollapsibleTrigger asChild>
                            <Button
                                variant="ghost"
                                size="sm"
                                className="px-2 text-xs flex items-center gap-1 hover:bg-red-50 dark:hover:bg-red-900/20"
                            >
                                <HelpCircle className="h-3 w-3" />
                                {showDetails ? t('error.hideDetails', '隐藏详情') : t('error.viewDetails', '查看详情')}
                            </Button>
                        </CollapsibleTrigger>

                        {onRetry && errorType !== ErrorType.UNAUTHORIZED && (
                            <Button
                                variant="outline"
                                size="sm"
                                onClick={onRetry}
                                className="flex items-center gap-2 text-xs border-red-300 dark:border-red-900/50 hover:bg-red-50 dark:hover:bg-red-900/20 hover:text-red-600 dark:hover:text-red-400"
                            >
                                <RefreshCw className="h-3 w-3" />
                                {t('error.retry', '重试')}
                            </Button>
                        )}
                    </div>

                    <CollapsibleContent>
                        <div className="mt-3 p-3 bg-red-50 dark:bg-red-900/20 rounded text-xs font-mono overflow-auto">
                            <p className="break-words overflow-hidden max-h-[100px] dark:text-gray-300">{String(errorMessage)}</p>
                            {errorCode && <p className="mt-1 dark:text-gray-300">{t('error.code', '错误代码')}: {errorCode}</p>}
                            {requestId && <p className="mt-1 dark:text-gray-300">{t('error.requestId', '请求ID')}: {requestId}</p>}
                            {formattedErrorDetail && (
                                <div className="mt-1">
                                    <p className="dark:text-gray-300">{t('error.details', '错误详情')}:</p>
                                    <pre className="mt-1 p-2 bg-red-100 dark:bg-red-900/30 rounded overflow-x-auto dark:text-gray-300">
                                        {formattedErrorDetail}
                                    </pre>
                                </div>
                            )}
                        </div>
                    </CollapsibleContent>
                </Collapsible>
            </AlertDescription>
        </Alert>
    )
}