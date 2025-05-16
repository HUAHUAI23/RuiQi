import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useRealtimeQPS } from '../../hooks/useStats'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { History } from 'lucide-react'
import { AnimatedIcon } from '@/components/ui/animation/components/animated-icon'
import * as echarts from 'echarts'
import { useTheme } from '@/provider/theme-context'
import WaveIcon from '@/components/icon/wave-icon'

export function RealtimeQPSChart() {
    const { t } = useTranslation()
    const chartRef = useRef<HTMLDivElement>(null)
    const chartInstanceRef = useRef<echarts.ECharts | null>(null)
    const { theme } = useTheme()
    const [isResetAnimating, setIsResetAnimating] = useState(false)
    // 判断是否为暗色模式
    const isDarkMode = theme === 'dark'

    // 使用修改后的hook获取数据，现在包含refreshInterval
    const { localData, isLoading, isInitialized, refreshInterval } = useRealtimeQPS(30)

    // 当前QPS值 - 取最新的数据点值
    const currentQPS = localData.length > 0 ? localData[localData.length - 1].value : 0

    // 格式化时间戳 - 显示时:分:秒
    const formatTimeLabel = (timestamp: string) => {
        const date = new Date(timestamp)
        return date.toLocaleTimeString([], {
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit'
        })
    }

    // when data initialized and refreshInterval is set, start the animation
    useEffect(() => {
        if (!refreshInterval || !isInitialized) return

        // 创建定时器，与数据刷新周期同步
        const animationTimer = setInterval(() => {
            // 触发动画
            setIsResetAnimating(true)

            // 1秒后结束动画
            setTimeout(() => {
                setIsResetAnimating(false)
            }, 1000)
        }, refreshInterval)

        // 清理函数
        return () => {
            clearInterval(animationTimer)
        }
    }, [refreshInterval, isInitialized])


    // 初始化和更新图表
    useEffect(() => {
        if (!chartRef.current || localData.length === 0) return

        // 如果已经有图表实例，则不需要重新创建
        if (!chartInstanceRef.current) {
            chartInstanceRef.current = echarts.init(chartRef.current)
        }

        // 根据主题使用对应颜色
        const CHART_PRIMARY_COLOR = isDarkMode ? '#b394e9' : '#9d76db' // 紫色主题
        const CHART_SECONDARY_COLOR = isDarkMode ? 'rgba(179, 148, 233, 0.5)' : 'rgba(157, 118, 219, 0.5)' // 半透明紫色

        // 处理数据，确保0值也显示小柱子
        const processedData = localData.map(item => {
            // 如果值为0，给一个最小值0.05以便显示
            return item.value === 0 ? 0.05 : item.value
        })

        // 设置图表选项
        const option: echarts.EChartsOption = {
            tooltip: {
                trigger: 'axis',
                axisPointer: {
                    type: 'shadow',
                    shadowStyle: {
                        color: isDarkMode ? 'rgba(179, 148, 233, 0.15)' : 'rgba(157, 118, 219, 0.05)'
                    }
                },
                formatter: function (params) {
                    if (!params || !Array.isArray(params) || params.length === 0) {
                        return ''
                    }

                    const dataIndex = params[0].dataIndex
                    if (typeof dataIndex === 'number' && dataIndex >= 0 && dataIndex < localData.length) {
                        const item = localData[dataIndex]
                        // 显示实际值，而不是处理后的值
                        return `${formatTimeLabel(item.timestamp)}: <span style="font-weight: bold; color: ${CHART_PRIMARY_COLOR}">${item.value}</span> QPS`
                    }
                    return ''
                },
                backgroundColor: isDarkMode ? 'rgba(36, 37, 46, 0.95)' : 'rgba(255, 255, 255, 0.9)',
                borderColor: isDarkMode ? 'rgba(179, 148, 233, 0.2)' : 'rgba(157, 118, 219, 0.2)',
                textStyle: {
                    color: isDarkMode ? '#e0e0e0' : '#333'
                }
            },
            grid: {
                left: 0,  // 完全移除左侧margin
                right: 0, // 完全移除右侧margin
                bottom: 0, // 移除底部margin
                top: '5%',
                containLabel: false // 不为标签预留空间
            },
            xAxis: {
                type: 'category',
                boundaryGap: true,
                data: localData.map(item => formatTimeLabel(item.timestamp)),
                axisLabel: {
                    show: false // 隐藏X轴标签
                },
                axisLine: {
                    show: false // 隐藏X轴线
                },
                axisTick: {
                    show: false // 隐藏刻度线
                },
                splitLine: {
                    show: false // 隐藏分割线
                }
            },
            yAxis: {
                type: 'value',
                show: false, // 完全隐藏Y轴
                max: function (value) {
                    // 让图表顶部有一些留白，最大值上浮20%
                    return Math.max(10, value.max * 1.2)
                },
                min: 0 // 确保Y轴从0开始，这样小柱子能显示
            },
            series: [
                {
                    name: 'QPS',
                    type: 'bar',
                    barWidth: '65%',  // 调整柱子宽度为65%
                    barCategoryGap: '10%',  // 适当增加类目间距
                    barGap: '0%',  // 消除柱子间隙
                    itemStyle: {
                        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                            { offset: 0, color: CHART_PRIMARY_COLOR },
                            { offset: 1, color: CHART_SECONDARY_COLOR }
                        ]),
                        borderRadius: [4, 4, 0, 0], // 柱状图顶部圆角
                        shadowColor: isDarkMode ? 'rgba(179, 148, 233, 0.6)' : 'rgba(157, 118, 219, 0.4)',
                        shadowBlur: isDarkMode ? 15 : 10
                    },
                    data: processedData, // 使用处理后的数据
                    animationDuration: 300,
                    animationEasing: 'cubicOut',
                    animationDelay: function (idx) {
                        return idx * 5 // 快速动画
                    }
                }
            ],
            // 添加底部的虚线
            markLine: {
                silent: true,
                symbol: 'none',
                lineStyle: {
                    color: isDarkMode ? 'rgba(179, 148, 233, 0.25)' : 'rgba(157, 118, 219, 0.15)',
                    type: 'dashed'
                },
                data: [
                    {
                        yAxis: 0 // Y轴0点处的线
                    }
                ]
            },
            animation: true
        }

        // 应用选项
        chartInstanceRef.current.setOption(option)

        // 调整窗口大小时重绘图表
        const handleResize = () => {
            if (chartInstanceRef.current) {
                chartInstanceRef.current.resize()
            }
        }

        window.addEventListener('resize', handleResize)

        // 清理函数
        return () => {
            window.removeEventListener('resize', handleResize)
        }
    }, [localData, isInitialized, isDarkMode])

    return (
        <Card className="border-none shadow-none">
            <CardHeader className="flex flex-row items-center justify-between p-4 pb-2">
                <div className="flex items-center gap-3">
                    <CardTitle className="text-lg font-medium dark:text-shadow-glow-white">
                        {t('stats.realtimeQPS')}
                    </CardTitle>
                    <div className="flex items-center bg-gray-50 dark:bg-gray-800 rounded-md py-1 px-3">
                        <span className="text-primary dark:text-primary mr-2 mb-2">
                            {isDarkMode ? <WaveIcon width={20} height={20} /> : <WaveIcon width={20} height={20} color="#B68AFE" />}
                        </span>
                        <span className="font-medium text-gray-800 dark:text-white">{currentQPS}</span>
                    </div>
                </div>
                <div className="h-8 w-8 flex items-center justify-center text-gray-500 dark:text-gray-400 qps-refresh-icon">
                    <AnimatedIcon animationVariant="continuous-spin" isAnimating={isResetAnimating}>
                        <History className="h-5 w-5" />
                    </AnimatedIcon>
                </div>
            </CardHeader>
            <CardContent className="px-4 pb-4 pt-0">
                <div className="h-[180px]">
                    <div
                        ref={chartRef}
                        style={{ width: '100%', height: '100%' }}
                        className={isLoading && localData.length === 0 ? "opacity-50" : ""}
                    />
                </div>
            </CardContent>
        </Card>
    )
}