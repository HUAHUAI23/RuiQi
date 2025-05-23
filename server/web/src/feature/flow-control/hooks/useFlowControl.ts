import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { configApi } from '@/api/config'
import { ConfigPatchRequest } from '@/types/config'
import { useToast } from '@/hooks/use-toast'
import { useTranslation } from 'react-i18next'
import { useConfigQuery } from '@/feature/global-setting/hooks/useConfig'

// 更新流量控制配置mutation hook
export const useUpdateFlowControlConfig = () => {
    const queryClient = useQueryClient()
    const { toast } = useToast()
    const { t } = useTranslation()
    const [error, setError] = useState<string | null>(null)

    const mutation = useMutation({
        mutationFn: (data: ConfigPatchRequest) => configApi.updateConfig(data),
        onSuccess: () => {
            toast({
                title: t('flowControl.toast.updateSuccess', '更新成功'),
                description: t('flowControl.toast.configUpdated', '流量控制配置已成功更新'),
            })
            queryClient.invalidateQueries({ queryKey: ['config'] })
        },
        onError: (error: ApiError) => {
            console.error('更新流量控制配置失败:', error)
            setError(error.message || t('flowControl.toast.updateError', '更新配置时出现错误'))
            toast({
                title: t('flowControl.toast.updateFailed', '更新失败'),
                description: error.message || t('flowControl.toast.updateError', '更新配置时出现错误'),
                variant: "destructive",
            })
        }
    })

    return {
        updateFlowControlConfig: mutation.mutate,
        isLoading: mutation.isPending,
        error,
        clearError: () => setError(null),
    }
}

// 获取流量控制配置
export const useFlowControlConfig = () => {
    const { config, isLoading, error, refetch } = useConfigQuery()
    
    return {
        flowControlConfig: config?.engine?.flowController,
        isLoading,
        error,
        refetch
    }
} 