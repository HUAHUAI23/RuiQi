import { EChartWrapper } from './EChartWrapper'
import { useTranslation } from 'react-i18next'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { useCombinedTimeSeriesData } from '../../hooks/useStats'
import { TimeRange } from '@/types/stats'
import * as echarts from 'echarts'
import { useTheme } from '@/provider/theme-context'

interface RequestsBlocksChartProps {
    timeRange: TimeRange
}

export function RequestsBlocksChart({ timeRange }: RequestsBlocksChartProps) {
    const { t } = useTranslation()
    const { data, isLoading } = useCombinedTimeSeriesData(timeRange)
    const { theme } = useTheme()

    // 判断是否为暗色模式
    const isDarkMode = theme === 'dark'

    // 格式化时间
    const formatTime = (timestamp: string) => {
        const date = new Date(timestamp)

        switch (timeRange) {
            case '24h':
                return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
            case '7d':
                return `${date.getMonth() + 1}/${date.getDate()} ${date.getHours()}:00`
            case '30d':
                return `${date.getMonth() + 1}/${date.getDate()}`
            default:
                return date.toLocaleString()
        }
    }

    // 使用项目的紫色主题
    const THEME_PRIMARY = '#9d76db'         // 主色紫 (157, 118, 219)
    // const THEME_SECONDARY = '#a346ff'       // 更鲜艳的紫色
    const THEME_ERROR = '#f43f5e'           // 错误红色

    // 暗色模式下加强色彩亮度和辉光效果
    const THEME_PRIMARY_DARK = '#b394e9'    // 更亮的紫色
    const THEME_ERROR_DARK = '#fb6d88'      // 更亮的红色

    // 获取当前主题下的颜色
    const primaryColor = isDarkMode ? THEME_PRIMARY_DARK : THEME_PRIMARY
    const errorColor = isDarkMode ? THEME_ERROR_DARK : THEME_ERROR

    // 图表配置
    const chartOptions: echarts.EChartsOption = {
        tooltip: {
            trigger: 'axis',
            backgroundColor: isDarkMode ? 'rgba(36, 37, 46, 0.95)' : 'rgba(255, 255, 255, 0.95)',
            borderColor: isDarkMode ? 'rgba(179, 148, 233, 0.2)' : 'rgba(157, 118, 219, 0.2)',
            borderWidth: 1,
            padding: [12, 16],
            textStyle: {
                color: isDarkMode ? '#e0e0e0' : '#333',
                fontSize: 12
            },
            formatter: function (params) {
                // Ensure params is an array
                const paramArray = Array.isArray(params) ? params : [params]
                let result = `<div style="font-weight: bold; margin-bottom: 8px; color: ${isDarkMode ? '#fff' : '#333'};">${paramArray[0].name}</div>`

                paramArray.forEach((param) => {
                    const seriesName = param.seriesName as string
                    const value = param.value as number
                    const isTotalRequests = seriesName === t('stats.requests')
                    const color = isTotalRequests ? primaryColor : errorColor

                    result += `<div style="display: flex; align-items: center; margin: 6px 0;">
                                <span style="display: inline-block; margin-right: 8px; width: 8px; height: 8px; border-radius: 50%; background: ${color}; box-shadow: 0 0 ${isDarkMode ? '6px' : '4px'} ${color}"></span>
                                <span style="flex: 1; color: ${isDarkMode ? '#ccc' : '#555'};">${seriesName}</span>
                                <span style="font-weight: bold; margin-left: 15px; color: ${color};">${value.toLocaleString()}</span>
                              </div>`
                })

                return result
            },
            axisPointer: {
                type: 'line',
                lineStyle: {
                    color: isDarkMode ? 'rgba(179, 148, 233, 0.4)' : 'rgba(157, 118, 219, 0.3)',
                    width: 1,
                    type: 'dashed'
                },
                shadowStyle: {
                    color: isDarkMode ? 'rgba(179, 148, 233, 0.15)' : 'rgba(157, 118, 219, 0.05)'
                }
            }
        },
        legend: {
            data: [t('stats.requests'), t('stats.blocks')],
            right: 10,
            top: 0,
            textStyle: {
                fontSize: 12,
                color: isDarkMode ? '#ddd' : '#666'
            },
            icon: 'circle',
            itemWidth: 8,
            itemHeight: 8,
            itemGap: 20
        },
        grid: {
            left: '3%',
            right: '4%',
            bottom: '3%',
            top: '40px',
            containLabel: true
        },
        xAxis: {
            type: 'category',
            boundaryGap: false,
            data: data?.requests.data.map(item => formatTime(item.timestamp)) || [],
            axisLabel: {
                fontSize: 10,
                color: isDarkMode ? '#aaa' : '#999',
                interval: timeRange === '24h' ? 3 : timeRange === '7d' ? 6 : 2
            },
            axisLine: {
                lineStyle: {
                    color: isDarkMode ? 'rgba(255, 255, 255, 0.09)' : 'rgba(0, 0, 0, 0.09)'
                }
            },
            axisTick: {
                show: false
            }
        },
        yAxis: {
            type: 'value',
            position: 'left',
            axisLabel: {
                fontSize: 10,
                color: isDarkMode ? '#aaa' : '#999',
                formatter: (value) => value.toLocaleString()
            },
            axisLine: {
                show: false
            },
            axisTick: {
                show: false
            },
            splitLine: {
                lineStyle: {
                    color: isDarkMode ? 'rgba(255, 255, 255, 0.04)' : 'rgba(0, 0, 0, 0.04)',
                    type: 'dashed'
                }
            }
        },
        series: [
            {
                name: t('stats.requests'),
                type: 'line',
                smooth: true,
                symbol: 'emptyCircle',
                symbolSize: 6,
                showSymbol: false,
                emphasis: {
                    focus: 'series',
                    scale: true,
                    itemStyle: {
                        borderWidth: 2,
                        shadowBlur: isDarkMode ? 15 : 10,
                        shadowColor: isDarkMode
                            ? 'rgba(179, 148, 233, 0.7)'
                            : 'rgba(157, 118, 219, 0.5)'
                    }
                },
                lineStyle: {
                    width: 3,
                    shadowColor: isDarkMode
                        ? 'rgba(179, 148, 233, 0.4)'
                        : 'rgba(157, 118, 219, 0.3)',
                    shadowBlur: isDarkMode ? 15 : 10
                },
                itemStyle: {
                    color: primaryColor,
                    borderWidth: 2,
                    borderColor: isDarkMode ? '#2d2d3a' : '#fff'
                },
                areaStyle: {
                    color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                        {
                            offset: 0,
                            color: isDarkMode
                                ? 'rgba(179, 148, 233, 0.5)'
                                : 'rgba(157, 118, 219, 0.4)'
                        },
                        {
                            offset: 0.5,
                            color: isDarkMode
                                ? 'rgba(179, 148, 233, 0.2)'
                                : 'rgba(157, 118, 219, 0.15)'
                        },
                        {
                            offset: 1,
                            color: isDarkMode
                                ? 'rgba(179, 148, 233, 0.05)'
                                : 'rgba(157, 118, 219, 0.02)'
                        }
                    ]),
                    opacity: 0.8
                },
                data: data?.requests.data.map(item => item.value) || [],
                z: 10
            },
            {
                name: t('stats.blocks'),
                type: 'line',
                smooth: true,
                symbol: 'emptyCircle',
                symbolSize: 6,
                showSymbol: false,
                emphasis: {
                    focus: 'series',
                    scale: true,
                    itemStyle: {
                        borderWidth: 2,
                        shadowBlur: isDarkMode ? 15 : 10,
                        shadowColor: isDarkMode
                            ? 'rgba(251, 109, 136, 0.7)'
                            : 'rgba(244, 63, 94, 0.5)'
                    }
                },
                lineStyle: {
                    width: 3,
                    shadowColor: isDarkMode
                        ? 'rgba(251, 109, 136, 0.4)'
                        : 'rgba(244, 63, 94, 0.3)',
                    shadowBlur: isDarkMode ? 15 : 10
                },
                itemStyle: {
                    color: errorColor,
                    borderWidth: 2,
                    borderColor: isDarkMode ? '#2d2d3a' : '#fff'
                },
                areaStyle: {
                    color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                        {
                            offset: 0,
                            color: isDarkMode
                                ? 'rgba(251, 109, 136, 0.5)'
                                : 'rgba(244, 63, 94, 0.4)'
                        },
                        {
                            offset: 0.5,
                            color: isDarkMode
                                ? 'rgba(251, 109, 136, 0.2)'
                                : 'rgba(244, 63, 94, 0.15)'
                        },
                        {
                            offset: 1,
                            color: isDarkMode
                                ? 'rgba(251, 109, 136, 0.05)'
                                : 'rgba(244, 63, 94, 0.02)'
                        }
                    ]),
                    opacity: 0.8
                },
                data: data?.blocks.data.map(item => item.value) || [],
                z: 9
            }
        ],
        animation: true,
        animationDuration: 1000,
        animationEasing: 'cubicOut',
        animationDelay: function (idx) {
            return idx * 50
        }
    }

    return (
        <Card className="border-none shadow-none">
            <CardHeader className="p-4 pb-0">
                <CardTitle className="text-lg font-medium dark:text-shadow-glow-white">{t('stats.requestsAndBlocks')}</CardTitle>
            </CardHeader>
            <CardContent className="p-4">
                <div className="h-[300px]">
                    <EChartWrapper options={chartOptions} loading={isLoading} height={300} />
                </div>
            </CardContent>
        </Card>
    )
}