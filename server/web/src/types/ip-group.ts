/**
 * IP组数据结构
 */
export interface IPGroup {
    id: string             // IP组唯一标识符
    name: string           // IP组名称
    items: string[]        // IP地址或CIDR列表
}

/**
 * IP组创建请求
 */
export interface IPGroupCreateRequest {
    name: string           // IP组名称
    items: string[]        // IP地址或CIDR列表
}

/**
 * IP组更新请求
 */
export interface IPGroupUpdateRequest {
    name?: string          // IP组名称
    items?: string[]       // IP地址或CIDR列表
}

/**
 * IP组列表响应
 */
export interface IPGroupListResponse {
    total: number          // 总数
    items: IPGroup[]       // IP组列表
}

/**
 * 用于表单的IP组数据
 */
export interface IPGroupFormValues {
    name: string
    items: string[]
}