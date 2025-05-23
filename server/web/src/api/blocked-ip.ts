import { get, del } from './index'
import {
    BlockedIPListRequest,
    BlockedIPListResponse,
    BlockedIPStatsResponse,
    BlockedIPCleanupResponse,
} from '@/types/blocked-ip'

// Blocked IP API接口基础路径
const BASE_URL = '/blocked-ips'

/**
 * Blocked IP相关API服务
 */
export const blockedIPApi = {
    /**
     * 获取封禁IP列表
     * @param params 查询参数
     * @returns 封禁IP列表响应数据
     */
    getBlockedIPs: (params: BlockedIPListRequest = {}): Promise<BlockedIPListResponse> => {
        // 映射前端参数到后端期望的参数名，确保默认值
        const queryParams: Record<string, string | number> = {
            page: params.page || 1,
            size: params.size || 20,
            status: params.status || 'all',
            sortBy: params.sortBy || 'blocked_at',
            sortDir: params.sortDir || 'desc'
        }
        
        // 只有当有值时才添加可选参数
        if (params.ip) {
            queryParams.ip = params.ip
        }
        if (params.reason) {
            queryParams.reason = params.reason
        }

        return get<BlockedIPListResponse>(BASE_URL, {
            params: queryParams
        })
    },

    /**
     * 获取封禁IP统计信息
     * @returns 统计信息
     */
    getBlockedIPStats: (): Promise<BlockedIPStatsResponse> => {
        return get<BlockedIPStatsResponse>(`${BASE_URL}/stats`)
    },

    /**
     * 清理过期的封禁IP记录
     * @returns 清理结果
     */
    cleanupExpiredBlockedIPs: (): Promise<BlockedIPCleanupResponse> => {
        return del<BlockedIPCleanupResponse>(`${BASE_URL}/cleanup`)
    }
} 