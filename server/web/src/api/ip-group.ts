import { get, post, put, del } from './index'
import {
    IPGroup,
    IPGroupListResponse,
    IPGroupCreateRequest,
    IPGroupUpdateRequest,
} from '@/types/ip-group'

// IP组API接口基础路径
const BASE_URL = '/ip-groups'

/**
 * IP组相关API服务
 */
export const ipGroupApi = {
    /**
     * 获取IP组列表
     * @param page 页码
     * @param size 每页数量
     * @returns IP组列表响应数据
     */
    getIPGroups: (page: number = 1, size: number = 10): Promise<IPGroupListResponse> => {
        return get<IPGroupListResponse>(BASE_URL, {
            params: { page, size }
        })
    },

    /**
     * 创建新IP组
     * @param ipGroup IP组创建请求数据
     * @returns 创建后的IP组详情
     */
    createIPGroup: (ipGroup: IPGroupCreateRequest): Promise<IPGroup> => {
        return post<IPGroup>(BASE_URL, ipGroup)
    },

    /**
     * 获取单个IP组详情
     * @param id IP组ID
     * @returns IP组详情
     */
    getIPGroup: (id: string): Promise<IPGroup> => {
        return get<IPGroup>(`${BASE_URL}/${id}`)
    },

    /**
     * 更新IP组
     * @param id IP组ID
     * @param ipGroup IP组更新请求数据
     * @returns 更新后的IP组详情
     */
    updateIPGroup: (id: string, ipGroup: IPGroupUpdateRequest): Promise<IPGroup> => {
        return put<IPGroup>(`${BASE_URL}/${id}`, ipGroup)
    },

    /**
     * 删除IP组
     * @param id IP组ID
     * @returns void
     */
    deleteIPGroup: (id: string): Promise<void> => {
        return del<void>(`${BASE_URL}/${id}`)
    },

    /**
     * 将IP添加到系统黑名单
     * @param ip IP地址或CIDR
     * @returns void
     */
    blockIP: (ip: string): Promise<void> => {
        return post<void>(`${BASE_URL}/blacklist/add`, { ip })
    }
}