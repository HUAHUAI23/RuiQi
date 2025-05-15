import { get } from './index'
import {
  OverviewStats,
  RealtimeQPSResponse,
  TimeSeriesResponse,
  CombinedTimeSeriesResponse,
  TrafficTimeSeriesResponse,
  TimeRange
} from '@/types/stats'

// 统计API接口基础路径
const BASE_URL = '/stats'

/**
 * 统计数据相关API服务
 */
export const statsApi = {
  /**
   * 获取统计概览数据
   * @param timeRange 时间范围
   * @returns 统计概览数据
   */
  getOverviewStats: (timeRange: TimeRange = '24h'): Promise<OverviewStats> => {
    return get<OverviewStats>(`${BASE_URL}/overview`, {
      params: { timeRange }
    })
  },

  /**
   * 获取实时QPS数据
   * @param limit 数据点数量限制，默认30
   * @returns 实时QPS数据
   */
  getRealtimeQPS: (limit: number = 30): Promise<RealtimeQPSResponse> => {
    return get<RealtimeQPSResponse>(`${BASE_URL}/realtime-qps`, {
      params: { limit }
    })
  },

  /**
   * 获取时间序列数据
   * @param timeRange 时间范围
   * @param metric 指标类型
   * @returns 时间序列数据
   */
  getTimeSeriesData: (
    timeRange: TimeRange = '24h',
    metric: 'requests' | 'blocks' = 'requests'
  ): Promise<TimeSeriesResponse> => {
    return get<TimeSeriesResponse>(`${BASE_URL}/time-series`, {
      params: { timeRange, metric }
    })
  },

  /**
   * 获取组合时间序列数据(请求数和拦截数)
   * @param timeRange 时间范围
   * @returns 组合时间序列数据
   */
  getCombinedTimeSeriesData: (timeRange: TimeRange = '24h'): Promise<CombinedTimeSeriesResponse> => {
    return get<CombinedTimeSeriesResponse>(`${BASE_URL}/combined-time-series`, {
      params: { timeRange }
    })
  },

  /**
   * 获取流量时间序列数据
   * @param timeRange 时间范围
   * @returns 流量时间序列数据
   */
  getTrafficTimeSeriesData: (timeRange: TimeRange = '24h'): Promise<TrafficTimeSeriesResponse> => {
    return get<TrafficTimeSeriesResponse>(`${BASE_URL}/traffic-time-series`, {
      params: { timeRange }
    })
  }
}