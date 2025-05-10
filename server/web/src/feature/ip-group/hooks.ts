import { useMutation, useQueryClient, useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { ipGroupApi } from '@/api/ip-group'
import { IPGroupCreateRequest, IPGroupUpdateRequest, IPGroupListResponse, IPGroup } from '@/types/ip-group'
import { useToast } from '@/hooks/use-toast'
import { useTranslation } from 'react-i18next'

export const useIPGroups = (page: number = 1, size: number = 10) => {
    return useQuery<IPGroupListResponse>({
        queryKey: ['ipGroups', page, size],
        queryFn: () => ipGroupApi.getIPGroups(page, size),
    })
}

export const useIPGroup = (id: string) => {
    return useQuery<IPGroup>({
        queryKey: ['ipGroup', id],
        queryFn: () => ipGroupApi.getIPGroup(id),
        enabled: !!id,
    })
}

export const useCreateIPGroup = () => {
    const queryClient = useQueryClient()
    const { toast } = useToast()
    const { t } = useTranslation()
    const [error, setError] = useState<string | null>(null)

    const mutation = useMutation({
        mutationFn: (data: IPGroupCreateRequest) => ipGroupApi.createIPGroup(data),
        onSuccess: () => {
            toast({
                title: t('ipGroup.toast.createSuccess'),
                description: t('ipGroup.toast.ipGroupCreated'),
            })
            queryClient.invalidateQueries({ queryKey: ['ipGroups'] })
        },
        onError: (error: ApiError) => {
            console.error(t('ipGroup.toast.createFailed'), error)
            setError(error.message || t('ipGroup.toast.createError'))
            toast({
                title: t('ipGroup.toast.createFailed'),
                description: error.message || t('ipGroup.toast.createError'),
                variant: "destructive",
            })
        }
    })

    return {
        createIPGroup: mutation.mutate,
        isLoading: mutation.isPending,
        error,
        clearError: () => setError(null),
    }
}

export const useDeleteIPGroup = () => {
    const queryClient = useQueryClient()
    const { toast } = useToast()
    const { t } = useTranslation()
    const [error, setError] = useState<string | null>(null)

    const mutation = useMutation({
        mutationFn: (id: string) => ipGroupApi.deleteIPGroup(id),
        onSuccess: () => {
            toast({
                title: t('ipGroup.toast.deleteSuccess'),
                description: t('ipGroup.toast.ipGroupDeleted'),
            })
            queryClient.invalidateQueries({ queryKey: ['ipGroups'] })
        },
        onError: (error: ApiError) => {
            console.error(t('ipGroup.toast.deleteFailed'), error)
            setError(error.message || t('ipGroup.toast.deleteError'))
            toast({
                title: t('ipGroup.toast.deleteFailed'),
                description: error.message || t('ipGroup.toast.deleteError'),
                variant: "destructive",
            })
        }
    })

    return {
        deleteIPGroup: mutation.mutate,
        isLoading: mutation.isPending,
        error,
        clearError: () => setError(null),
    }
}

export const useUpdateIPGroup = () => {
    const queryClient = useQueryClient()
    const { toast } = useToast()
    const { t } = useTranslation()
    const [error, setError] = useState<string | null>(null)

    const mutation = useMutation({
        mutationFn: ({ id, data }: { id: string, data: IPGroupUpdateRequest }) =>
            ipGroupApi.updateIPGroup(id, data),
        onSuccess: () => {
            toast({
                title: t('ipGroup.toast.updateSuccess'),
                description: t('ipGroup.toast.ipGroupUpdated'),
            })
            queryClient.invalidateQueries({ queryKey: ['ipGroups'] })
        },
        onError: (error: ApiError) => {
            console.error(t('ipGroup.toast.updateFailed'), error)
            setError(error.message || t('ipGroup.toast.updateError'))
            toast({
                title: t('ipGroup.toast.updateFailed'),
                description: error.message || t('ipGroup.toast.updateError'),
                variant: "destructive",
            })
        }
    })

    return {
        updateIPGroup: mutation.mutate,
        isLoading: mutation.isPending,
        error,
        clearError: () => setError(null),
    }
}

export const useBlockIP = () => {
    const queryClient = useQueryClient()
    const { toast } = useToast()
    const { t } = useTranslation()
    const [error, setError] = useState<string | null>(null)

    const mutation = useMutation({
        mutationFn: (ip: string) => ipGroupApi.blockIP(ip),
        onSuccess: (_, ip) => {
            toast({
                title: t('ipGroup.toast.blockSuccess'),
                description: t('ipGroup.toast.ipBlocked', { ip }),
            })
            queryClient.invalidateQueries({ queryKey: ['ipGroups'] })
        },
        onError: (error: ApiError) => {
            console.error(t('ipGroup.toast.blockFailed'), error)
            setError(error.message || t('ipGroup.toast.blockError'))
            toast({
                title: t('ipGroup.toast.blockFailed'),
                description: error.message || t('ipGroup.toast.blockError'),
                variant: "destructive",
            })
        }
    })

    return {
        blockIP: (ip: string, options?: { onSettled?: () => void }) => {
            mutation.mutate(ip, {
                onSettled: options?.onSettled
            })
        },
        isLoading: mutation.isPending,
        error,
        clearError: () => setError(null),
    }
} 