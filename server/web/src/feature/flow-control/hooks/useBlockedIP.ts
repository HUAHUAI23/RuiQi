import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { blockedIPApi } from '@/api/blocked-ip'
import { BlockedIPListRequest } from '@/types/blocked-ip'
import { useToast } from '@/hooks/use-toast'
import { useTranslation } from 'react-i18next'

// 获取封禁IP列表查询hook
export const useBlockedIPsQuery = (params: BlockedIPListRequest) => {
    const query = useQuery({
        queryKey: ['blocked-ips', params],
        queryFn: () => blockedIPApi.getBlockedIPs(params)
    })

    return {
        blockedIPs: query.data,
        isLoading: query.isPending,
        error: query.error,
        refetch: query.refetch
    }
}

// 获取封禁IP统计查询hook
export const useBlockedIPStatsQuery = () => {
    const query = useQuery({
        queryKey: ['blocked-ip-stats'],
        queryFn: blockedIPApi.getBlockedIPStats
    })

    return {
        stats: query.data,
        isLoading: query.isPending,
        error: query.error,
        refetch: query.refetch
    }
}

// 清理过期封禁IP mutation hook
export const useCleanupExpiredBlockedIPs = () => {
    const queryClient = useQueryClient()
    const { toast } = useToast()
    const { t } = useTranslation()
    const [error, setError] = useState<string | null>(null)

    const mutation = useMutation({
        mutationFn: blockedIPApi.cleanupExpiredBlockedIPs,
        onSuccess: (data) => {
            toast({
                title: t('flowControl.toast.cleanupSuccess', '清理成功'),
                description: t('flowControl.toast.cleanupMessage', `已清理 ${data.deletedCount} 条过期记录`),
            })
            queryClient.invalidateQueries({ queryKey: ['blocked-ips'] })
            queryClient.invalidateQueries({ queryKey: ['blocked-ip-stats'] })
        },
        onError: (error: ApiError) => {
            console.error('清理过期封禁IP失败:', error)
            setError(error.message || t('flowControl.toast.cleanupError', '清理过期记录时出现错误'))
            toast({
                title: t('flowControl.toast.cleanupFailed', '清理失败'),
                description: error.message || t('flowControl.toast.cleanupError', '清理过期记录时出现错误'),
                variant: "destructive",
            })
        }
    })

    return {
        cleanupExpiredBlockedIPs: mutation.mutate,
        isLoading: mutation.isPending,
        error,
        clearError: () => setError(null),
    }
} 