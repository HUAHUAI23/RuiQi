// 时间范围类型
export type TimeRange = '24h' | '7d' | '30d';

// 概览统计数据
export interface OverviewStats {
  timeRange: TimeRange;
  totalRequests: number;
  inboundTraffic: number;
  outboundTraffic: number;
  maxQPS: number;
  error4xx: number;
  error4xxRate: number;
  error5xx: number;
  error5xxRate: number;
  blockCount: number;
  attackIPCount: number;
}

// 实时QPS数据点
export interface RealtimeQPSData {
  timestamp: string;
  value: number;
}

// 实时QPS响应
export interface RealtimeQPSResponse {
  data: RealtimeQPSData[];
}

// 时间序列数据点
export interface TimeSeriesDataPoint {
  timestamp: string;
  value: number;
}

// 请求和拦截数据的时间序列响应
export interface TimeSeriesResponse {
  metric: 'requests' | 'blocks';
  timeRange: TimeRange;
  data: TimeSeriesDataPoint[];
}

// 组合时间序列响应(请求和拦截)
export interface CombinedTimeSeriesResponse {
  timeRange: TimeRange;
  requests: TimeSeriesResponse;
  blocks: TimeSeriesResponse;
}

// 流量数据点
export interface TrafficDataPoint {
  timestamp: string;
  inboundTraffic: number;
  outboundTraffic: number;
}

// 流量时间序列响应
export interface TrafficTimeSeriesResponse {
  timeRange: TimeRange;
  data: TrafficDataPoint[];
}