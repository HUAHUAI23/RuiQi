import { useEffect, useRef } from 'react'
import * as echarts from 'echarts'
import { useResizeObserver } from '@/feature/stats/hooks/use-resize-observer'
import { useTheme } from "@/provider/theme-context"

interface EChartWrapperProps {
    options: echarts.EChartsOption
    loading?: boolean
    height?: number | string
    className?: string
}

export function EChartWrapper({
    options,
    loading = false,
    height = 300,
    className = '',
}: EChartWrapperProps) {
    const chartRef = useRef<HTMLDivElement>(null)
    const chartInstanceRef = useRef<echarts.ECharts | null>(null)
    const { theme } = useTheme()

    // 监听容器大小变化
    const { width } = useResizeObserver(chartRef)

    // 初始化图表和处理主题变化 - 合并为一个useEffect
    useEffect(() => {
        if (!chartRef.current) return

        const isDarkMode = theme === 'dark'

        // 如果实例不存在，创建新实例；如果已存在，不重新创建而是应用主题
        if (!chartInstanceRef.current) {
            chartInstanceRef.current = echarts.init(chartRef.current, isDarkMode ? 'dark' : undefined)
        } else {
            // 如果主题改变，应用新主题
            chartInstanceRef.current.dispose()
            chartInstanceRef.current = echarts.init(chartRef.current, isDarkMode ? 'dark' : undefined)
        }

        // 设置加载状态
        if (loading) {
            chartInstanceRef.current.showLoading({
                text: '',
                color: isDarkMode ? '#ffffff' : '#1f2937',
                textColor: isDarkMode ? '#ffffff' : '#1f2937',
                maskColor: isDarkMode ? 'rgba(0, 0, 0, 0.1)' : 'rgba(255, 255, 255, 0.8)',
            })
        } else {
            chartInstanceRef.current.hideLoading()
        }

        // 更新图表
        chartInstanceRef.current.setOption(options)

        // 清理函数
        return () => {
            if (chartInstanceRef.current) {
                chartInstanceRef.current.dispose()
                chartInstanceRef.current = null
            }
        }
    }, [options, theme, loading])

    // 响应容器大小变化
    useEffect(() => {
        if (chartInstanceRef.current && width) {
            chartInstanceRef.current.resize()
        }
    }, [width])

    return (
        <div
            ref={chartRef}
            style={{ height: typeof height === 'number' ? `${height}px` : height }}
            className={`w-full ${className}`}
        />
    )
}