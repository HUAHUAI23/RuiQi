import { Card, CardContent, CardTitle } from "@/components/ui/card"
import { ReactNode } from "react"

interface StatsCardProps {
    title: string
    value: string | number
    icon?: ReactNode
    change?: string | number
    trend?: 'up' | 'down' | 'neutral'
    loading?: boolean
    isTraffic?: boolean
}

export function StatsCard({
    title,
    value,
    icon,
    change,
    trend,
    loading = false,
    isTraffic = false,
}: StatsCardProps) {
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

    return (
        <Card className="border-none shadow-none p-4 hover:bg-gray-50 dark:hover:bg-gray-900/10 transition-colors">
            <CardTitle className="text-sm font-medium text-muted-foreground mb-2 flex items-center gap-2 dark:text-shadow-glow-white">
                {icon && <span className="text-primary dark:text-white">{icon}</span>}
                {title}
            </CardTitle>
            <CardContent className="p-0">
                {loading ? (
                    <div className="h-9 w-24 animate-pulse bg-gray-200 dark:bg-gray-800 rounded"></div>
                ) : (
                    <div className="flex flex-col">
                        <div className="text-2xl font-bold dark:text-shadow-glow-white">
                            {isTraffic && typeof value === 'number' ? formatTraffic(value) : value}
                        </div>
                        {change && (
                            <div className={`text-xs ${getTrendColor()} flex items-center mt-1`}>
                                {trend === 'up' && '↑ '}
                                {trend === 'down' && '↓ '}
                                {change}
                            </div>
                        )}
                    </div>
                )}
            </CardContent>
        </Card>
    )
}