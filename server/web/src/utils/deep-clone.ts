// 确保添加完整的解析和格式化函数
import { Condition } from "@/types/rule"

/**
 * 深度克隆对象或数组
 * @param obj 要克隆的对象或数组
 * @returns 克隆后的对象或数组
 */
export function deepClone<T>(obj: T): T {
    if (obj === null || typeof obj !== 'object') {
        return obj
    }

    // 处理日期对象
    if (obj instanceof Date) {
        return new Date(obj.getTime()) as unknown as T
    }

    // 处理数组
    if (Array.isArray(obj)) {
        return obj.map(item => deepClone(item)) as unknown as T
    }

    // 处理对象
    const clonedObj = {} as Record<string, any>
    for (const key in obj) {
        if (Object.prototype.hasOwnProperty.call(obj, key)) {
            clonedObj[key] = deepClone((obj as Record<string, any>)[key])
        }
    }

    return clonedObj as T
}

/**
 * 将表单数据中的条件格式化为后端接受的格式
 * 确保复杂嵌套条件可以正确序列化给后端
 */
export function formatConditionForApi(condition: Condition): Condition {
    // 深度克隆条件以避免修改原始对象
    const formattedCondition = deepClone(condition)

    // 如果是复合条件，递归格式化其子条件
    if (formattedCondition.type === 'composite') {
        formattedCondition.conditions = formattedCondition.conditions.map(
            child => formatConditionForApi(child)
        )
    }

    return formattedCondition
}

/**
 * 将后端返回的条件解析为表单可用的格式
 * 确保复杂嵌套条件可以正确展示在UI中
 */
export function parseConditionFromApi(condition: any): Condition {
    // 如果是null或undefined，返回默认条件
    if (!condition) {
        return {
            type: 'simple',
            target: 'source_ip',
            match_type: 'equal',
            match_value: '',
        }
    }

    // 如果是简单条件，确保类型正确
    if (condition.type === 'simple') {
        return {
            type: 'simple',
            target: condition.target || 'source_ip',
            match_type: condition.match_type || 'equal',
            match_value: condition.match_value || '',
        }
    }

    // 如果是复合条件，递归解析子条件
    if (condition.type === 'composite') {
        return {
            type: 'composite',
            operator: condition.operator || 'AND',
            conditions: Array.isArray(condition.conditions)
                ? condition.conditions.map(parseConditionFromApi)
                : [],
        }
    }

    // 如果无法确定类型，返回默认简单条件
    return {
        type: 'simple',
        target: 'source_ip',
        match_type: 'equal',
        match_value: '',
    }
}