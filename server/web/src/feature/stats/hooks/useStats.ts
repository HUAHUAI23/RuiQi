import { useQuery } from '@tanstack/react-query'
import { statsApi } from '@/api/services'
import { TimeRange } from '@/types/stats'
import { useState, useRef, useEffect } from 'react'

const REFRESH_INTERVAL = 60000 // 60秒
const QPS_REFRESH_INTERVAL = 5000 // 5秒
const QPS_STALETIME = 4000 // 4秒

// 定义QPS数据点类型
export interface QPSDataPoint {
    timestamp: string
    value: number
}

// 统计概览数据hook
export const useOverviewStats = (timeRange: TimeRange) => {
    return useQuery({
        queryKey: ['stats', 'overview', timeRange],
        queryFn: () => statsApi.getOverviewStats(timeRange),
        refetchInterval: REFRESH_INTERVAL, // 使用常量
    })
}

// 实时QPS数据hook - 增量更新滚动版本
export const useRealtimeQPS = (limit: number = 30) => {
    // 本地状态存储
    const [localQPSData, setLocalQPSData] = useState<QPSDataPoint[]>([])
    const lastTimestampRef = useRef<string | null>(null)
    const isInitializedRef = useRef(false)
    const shouldScrollRef = useRef(true)

    // 查询API数据
    const queryResult = useQuery({
        queryKey: ['stats', 'realtimeQPS', limit],
        queryFn: () => statsApi.getRealtimeQPS(limit),
        refetchInterval: QPS_REFRESH_INTERVAL, // 使用常量
        refetchIntervalInBackground: true, // 页面不活跃时也继续更新
        staleTime: QPS_STALETIME, // 数据保持新鲜状态的时间
        refetchOnWindowFocus: false, // 窗口获得焦点时不自动刷新
    })

    // 数据处理
    useEffect(() => {
        if (!queryResult.data?.data || queryResult.data.data.length === 0) return

        const apiData = queryResult.data.data

        // 如果数据为空或首次加载
        if (localQPSData.length === 0 || !isInitializedRef.current) {
            setLocalQPSData(apiData)
            lastTimestampRef.current = apiData[apiData.length - 1].timestamp
            isInitializedRef.current = true
            return
        }

        // 获取最新的时间戳
        const newestTimestamp = apiData[apiData.length - 1].timestamp

        // 如果有新数据
        if (newestTimestamp !== lastTimestampRef.current && shouldScrollRef.current) {
            // 找到增量数据
            const lastIndex = apiData.findIndex(item => item.timestamp === lastTimestampRef.current)

            if (lastIndex !== -1 && lastIndex < apiData.length - 1) {
                // 有新数据
                const newPoints = apiData.slice(lastIndex + 1)

                setLocalQPSData(prevData => {
                    const newData = [...prevData, ...newPoints]
                    return newData.slice(-limit) // 保持数据量在限制范围内
                })
            } else {
                // 数据结构变化，完全替换
                setLocalQPSData(apiData)
            }

            // 更新最后时间戳
            lastTimestampRef.current = newestTimestamp
        }
    }, [queryResult.data, limit, localQPSData.length])

    // 暂停/恢复滚动更新
    const toggleScroll = (shouldScroll: boolean) => {
        shouldScrollRef.current = shouldScroll
    }

    // 返回扩展的结果对象
    return {
        ...queryResult,
        localData: localQPSData,
        isInitialized: isInitializedRef.current,
        toggleScroll, // 暴露控制滚动的方法
        refreshInterval: QPS_REFRESH_INTERVAL // 暴露刷新间隔
    }
}

// 流量时间序列数据hook
export const useTrafficTimeSeriesData = (timeRange: TimeRange) => {
    return useQuery({
        queryKey: ['stats', 'trafficTimeSeries', timeRange],
        queryFn: () => statsApi.getTrafficTimeSeriesData(timeRange),
        refetchInterval: REFRESH_INTERVAL, // 使用常量
    })
}

// 请求和拦截组合时间序列数据hook
export const useCombinedTimeSeriesData = (timeRange: TimeRange) => {
    return useQuery({
        queryKey: ['stats', 'combinedTimeSeries', timeRange],
        queryFn: () => statsApi.getCombinedTimeSeriesData(timeRange),
        refetchInterval: REFRESH_INTERVAL, // 使用常量
    })
}

// 时间范围选择hook
export const useTimeRangeSelector = (initialRange: TimeRange = '24h') => {
    const [timeRange, setTimeRange] = useState<TimeRange>(initialRange)

    return {
        timeRange,
        setTimeRange,
    }
}