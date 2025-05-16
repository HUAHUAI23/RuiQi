import { Card, CardContent, CardTitle } from "@/components/ui/card"
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip"
import { ReactNode } from "react"
import { useNavigate } from "react-router"

interface StatsCardProps {
    title: string
    value: string | number
    icon?: ReactNode
    change?: string | number
    trend?: 'up' | 'down' | 'neutral'
    loading?: boolean
    isTraffic?: boolean
    link?: string
}

export function StatsCard({
    title,
    value,
    icon,
    change,
    trend,
    loading = false,
    isTraffic = false,
    link,
}: StatsCardProps) {
    const navigate = useNavigate()

    // 用于格式化流量数据
    const formatTraffic = (bytes: number): string => {
        if (bytes < 1024) return `${bytes} B`
        if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(2)} KB`
        if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(2)} MB`
        return `${(bytes / (1024 * 1024 * 1024)).toFixed(2)} GB`
    }

    // 颜色处理
    const getTrendColor = () => {
        if (!trend) return 'text-muted-foreground'
        return trend === 'up'
            ? 'text-emerald-500 dark:text-emerald-400'
            : trend === 'down'
                ? 'text-red-500 dark:text-red-400'
                : 'text-muted-foreground'
    }

    // Format value for display
    const displayValue = isTraffic && typeof value === 'number'
        ? formatTraffic(value)
        : String(value)

    const handleCardClick = () => {
        if (!link) return
        
        // 检查是否为外部链接
        if (link.startsWith('http://') || link.startsWith('https://')) {
            window.open(link, '_blank')
        } else {
            // 内部路由跳转
            navigate(link)
        }
    }

    return (
        <Card 
            className={`border-none shadow-none p-4 hover:bg-gray-50 dark:hover:bg-gray-900/10 transition-colors ${link ? 'cursor-pointer' : ''}`}
            onClick={link ? handleCardClick : undefined}
        >
            <CardTitle className="text-sm font-medium text-muted-foreground mb-2 flex items-center gap-2 dark:text-shadow-glow-white">
                {icon && <span className="text-primary dark:text-white flex-shrink-0">{icon}</span>}
                {title}
            </CardTitle>

            <CardContent className="p-0">
                {loading ? (
                    <div className="h-9 w-24 animate-pulse bg-gray-200 dark:bg-gray-800 rounded"></div>
                ) : (
                    <div className="flex flex-col">
                        <TooltipProvider delayDuration={300}>
                            <Tooltip>
                                <TooltipTrigger asChild>
                                    <div className="text-2xl font-bold dark:text-shadow-glow-white truncate">
                                        {displayValue}
                                    </div>
                                </TooltipTrigger>
                                <TooltipContent
                                    className="max-w-[350px] break-all bg-white border border-slate-200 shadow-md py-2 px-3 text-sm text-slate-800 dark:bg-slate-800 dark:border-slate-700 dark:!text-slate-200 dark:text-shadow-glow-white"
                                    side="top"
                                >
                                    {displayValue}
                                </TooltipContent>
                            </Tooltip>
                        </TooltipProvider>

                        {change && (
                            <TooltipProvider delayDuration={300}>
                                <Tooltip>
                                    <TooltipTrigger asChild>
                                        <div className={`text-xs ${getTrendColor()} flex items-center mt-1 truncate`}>
                                            {trend === 'up' && '↑ '}
                                            {trend === 'down' && '↓ '}
                                            {change}
                                        </div>
                                    </TooltipTrigger>
                                    <TooltipContent
                                        className="max-w-[350px] break-all bg-white border border-slate-200 shadow-md py-2 px-3 text-sm text-slate-800 dark:bg-slate-800 dark:border-slate-700 dark:!text-slate-200 dark:text-shadow-glow-white"
                                        side="top"
                                    >
                                        {trend === 'up' ? '↑ ' : trend === 'down' ? '↓ ' : ''}{change}
                                    </TooltipContent>
                                </Tooltip>
                            </TooltipProvider>
                        )}
                    </div>
                )}
            </CardContent>
        </Card>
    )
}