import { get, post, put, del } from './index'
import {
    MicroRule,
    MicroRuleListResponse,
    MicroRuleCreateRequest,
    MicroRuleUpdateRequest,
} from '@/types/rule'

// 规则API接口基础路径
const BASE_URL = '/micro-rules'

/**
 * 规则相关API服务
 */
export const ruleApi = {
    /**
     * 获取规则列表
     * @param page 页码
     * @param size 每页数量
     * @returns 规则列表响应数据
     */
    getMicroRules: (page: number = 1, size: number = 10): Promise<MicroRuleListResponse> => {
        return get<MicroRuleListResponse>(BASE_URL, {
            params: { page, size }
        })
    },

    /**
     * 创建新规则
     * @param rule 规则创建请求数据
     * @returns 创建后的规则详情
     */
    createMicroRule: (rule: MicroRuleCreateRequest): Promise<MicroRule> => {
        return post<MicroRule>(BASE_URL, rule)
    },

    /**
     * 获取单个规则详情
     * @param id 规则ID
     * @returns 规则详情
     */
    getMicroRule: (id: string): Promise<MicroRule> => {
        return get<MicroRule>(`${BASE_URL}/${id}`)
    },

    /**
     * 更新规则
     * @param id 规则ID
     * @param rule 规则更新请求数据
     * @returns 更新后的规则详情
     */
    updateMicroRule: (id: string, rule: MicroRuleUpdateRequest): Promise<MicroRule> => {
        return put<MicroRule>(`${BASE_URL}/${id}`, rule)
    },

    /**
     * 删除规则
     * @param id 规则ID
     * @returns void
     */
    deleteMicroRule: (id: string): Promise<void> => {
        return del<void>(`${BASE_URL}/${id}`)
    }
}