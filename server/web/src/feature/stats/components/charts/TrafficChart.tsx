import { EChartWrapper } from './EChartWrapper'
import { useTranslation } from 'react-i18next'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { useTrafficTimeSeriesData } from '../../hooks/useStats'
import { TimeRange } from '@/types/stats'
import * as echarts from 'echarts'
import { useTheme } from '@/provider/theme-context'

interface TrafficChartProps {
    timeRange: TimeRange
}

export function TrafficChart({ timeRange }: TrafficChartProps) {
    const { t } = useTranslation()
    const { data, isLoading } = useTrafficTimeSeriesData(timeRange)
    const { theme } = useTheme()
    
    // 判断是否为暗色模式
    const isDarkMode = theme === 'dark'

    // 使用项目的紫色主题
    const THEME_PRIMARY = '#9d76db'         // 主色紫 (157, 118, 219)
    const THEME_SECONDARY = '#10b981'       // 绿色 - 搭配紫色
    
    // 暗色模式下加强色彩亮度和辉光效果
    const THEME_PRIMARY_DARK = '#b394e9'    // 更亮的紫色
    const THEME_SECONDARY_DARK = '#34d399'  // 更亮的绿色

    // 获取当前主题下的颜色
    const primaryColor = isDarkMode ? THEME_PRIMARY_DARK : THEME_PRIMARY
    const secondaryColor = isDarkMode ? THEME_SECONDARY_DARK : THEME_SECONDARY

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

    // 格式化流量
    const formatTraffic = (bytes: number) => {
        if (bytes < 1024) return `${bytes} B`
        if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
        if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
        return `${(bytes / (1024 * 1024 * 1024)).toFixed(1)} GB`
    }

    // 图表配置
    const chartOptions: echarts.EChartsOption = {
        tooltip: {
            trigger: 'axis',
            backgroundColor: isDarkMode ? 'rgba(36, 37, 46, 0.95)' : 'rgba(255, 255, 255, 0.95)',
            borderColor: `rgba(${isDarkMode ? '179, 148, 233' : '157, 118, 219'}, 0.2)`,
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
                    const isInbound = seriesName === t('stats.inboundTraffic')
                    const color = isInbound ? primaryColor : secondaryColor
                    
                    result += `<div style="display: flex; align-items: center; margin: 6px 0;">
                              <span style="display: inline-block; margin-right: 8px; width: 8px; height: 8px; border-radius: 2px; background: ${color}; box-shadow: 0 0 ${isDarkMode ? '6px' : '4px'} ${color}"></span>
                              <span style="flex: 1; color: ${isDarkMode ? '#ccc' : '#555'};">${seriesName}</span>
                              <span style="font-weight: bold; margin-left: 15px; color: ${color};">${formatTraffic(value)}</span>
                            </div>`
                })
                return result
            },
            axisPointer: {
                type: 'shadow',
                shadowStyle: {
                    color: isDarkMode 
                        ? 'rgba(157, 118, 219, 0.15)' 
                        : 'rgba(157, 118, 219, 0.05)'
                }
            }
        },
        legend: {
            data: [t('stats.inboundTraffic'), t('stats.outboundTraffic')],
            right: 10,
            top: 0,
            textStyle: {
                fontSize: 12,
                color: isDarkMode ? '#ddd' : '#666'
            },
            icon: 'rect',
            itemWidth: 10,
            itemHeight: 10,
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
            data: data?.data.map(item => formatTime(item.timestamp)) || [],
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
            axisLabel: {
                formatter: (value: number) => formatTraffic(value),
                fontSize: 10,
                color: isDarkMode ? '#aaa' : '#999'
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
                name: t('stats.inboundTraffic'),
                type: 'bar',
                stack: 'traffic',
                emphasis: {
                    focus: 'series',
                    itemStyle: {
                        shadowBlur: isDarkMode ? 15 : 10,
                        shadowColor: isDarkMode 
                            ? 'rgba(179, 148, 233, 0.7)' 
                            : 'rgba(157, 118, 219, 0.5)'
                    }
                },
                data: data?.data.map(item => item.inboundTraffic) || [],
                itemStyle: {
                    color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                        { offset: 0, color: isDarkMode ? 'rgba(179, 148, 233, 0.95)' : 'rgba(157, 118, 219, 0.9)' },
                        { offset: 1, color: isDarkMode ? 'rgba(179, 148, 233, 0.65)' : 'rgba(157, 118, 219, 0.6)' }
                    ]),
                    borderRadius: [3, 3, 0, 0]
                },
                barWidth: '12px',
                barGap: '10%'
            },
            {
                name: t('stats.outboundTraffic'),
                type: 'bar',
                stack: 'traffic',
                emphasis: {
                    focus: 'series',
                    itemStyle: {
                        shadowBlur: isDarkMode ? 15 : 10,
                        shadowColor: isDarkMode 
                            ? 'rgba(52, 211, 153, 0.7)' 
                            : 'rgba(16, 185, 129, 0.5)'
                    }
                },
                data: data?.data.map(item => item.outboundTraffic) || [],
                itemStyle: {
                    color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                        { offset: 0, color: isDarkMode ? 'rgba(52, 211, 153, 0.95)' : 'rgba(16, 185, 129, 0.9)' },
                        { offset: 1, color: isDarkMode ? 'rgba(52, 211, 153, 0.65)' : 'rgba(16, 185, 129, 0.6)' }
                    ]),
                    borderRadius: [3, 3, 0, 0]
                },
                barWidth: '12px',
                barGap: '10%'
            }
        ],
        animation: true,
        animationDuration: 1000,
        animationEasing: 'cubicOut',
        animationDelay: function(idx) {
            return idx * 30
        }
    }

    return (
        <Card className="border-none shadow-none">
            <CardHeader className="p-4 pb-0">
                <CardTitle className="text-lg font-medium dark:text-shadow-glow-white">{t('stats.trafficTrend')}</CardTitle>
            </CardHeader>
            <CardContent className="p-4">
                <div className="h-[300px]">
                    <EChartWrapper options={chartOptions} loading={isLoading} height={300} />
                </div>
            </CardContent>
        </Card>
    )
}