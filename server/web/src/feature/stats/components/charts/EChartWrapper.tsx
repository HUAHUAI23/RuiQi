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

    // 初始化图表
    useEffect(() => {
        if (!chartRef.current) return

        const isDarkMode = theme === 'dark'

        // 如果实例已存在，先销毁
        if (chartInstanceRef.current) {
            chartInstanceRef.current.dispose()
        }

        // 创建新实例
        const newChart = echarts.init(chartRef.current, isDarkMode ? 'dark' : undefined)
        chartInstanceRef.current = newChart

        // 设置加载状态
        if (loading) {
            newChart.showLoading({
                text: '',
                color: isDarkMode ? '#ffffff' : '#1f2937',
                textColor: isDarkMode ? '#ffffff' : '#1f2937',
                maskColor: isDarkMode ? 'rgba(0, 0, 0, 0.1)' : 'rgba(255, 255, 255, 0.8)',
            })
        } else {
            newChart.hideLoading()
        }

        // 更新图表
        newChart.setOption(options)

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

    // 主题变化时更新图表
    useEffect(() => {
        if (chartInstanceRef.current) {
            const isDarkMode = theme === 'dark'

            // 重新初始化图表以应用主题
            if (chartRef.current) {
                const newChart = echarts.init(chartRef.current, isDarkMode ? 'dark' : undefined)
                newChart.setOption(chartInstanceRef.current.getOption())

                // 清理旧实例
                chartInstanceRef.current.dispose()
                chartInstanceRef.current = newChart
            }
        }
    }, [theme])

    return (
        <div
            ref={chartRef}
            style={{ height: typeof height === 'number' ? `${height}px` : height }}
            className={`w-full ${className}`}
        />
    )
}