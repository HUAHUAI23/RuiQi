import { useState, useRef, useEffect, useMemo } from 'react'
import {
    useReactTable,
    getCoreRowModel,
    ColumnDef,
} from '@tanstack/react-table'
import { useInfiniteQuery } from '@tanstack/react-query'
import { certificatesApi } from '@/api/certificate'
import { Certificate } from '@/types/certificate'
import { Button } from '@/components/ui/button'
import {
    MoreHorizontal, Plus, Trash2, RefreshCcw, Pencil
} from 'lucide-react'
import {
    DropdownMenu, DropdownMenuContent,
    DropdownMenuItem, DropdownMenuTrigger
} from '@/components/ui/dropdown-menu'
import { Card } from '@/components/ui/card'
import { CertificateDialog } from './CertificateDialog'
import { Loader2 } from 'lucide-react'
import { DataTable } from '@/components/table/motion-data-table'
import { DeleteCertificateDialog } from './DeleteCertificateDialog'
import { useTranslation } from 'react-i18next'
import { AnimatedIcon } from '@/components/ui/animation/components/animated-icon'
import { AnimatedButton } from '@/components/ui/animation/components/animated-button'

export function CertificateTable() {
    const { t } = useTranslation()

    // 状态管理
    const [certificateDialogOpen, setCertificateDialogOpen] = useState(false)
    const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
    const [selectedCertId, setSelectedCertId] = useState<string | null>(null)
    const sentinelRef = useRef<HTMLDivElement>(null)
    const [dialogMode, setDialogMode] = useState<'create' | 'update'>('create')
    const [selectedCertificate, setSelectedCertificate] = useState<Certificate | null>(null)

    const [isRefreshAnimating, setIsRefreshAnimating] = useState(false)

    // 获取证书列表
    const {
        data,
        isLoading,
        fetchNextPage,
        hasNextPage,
        isFetchingNextPage,
        refetch
    } = useInfiniteQuery({
        queryKey: ['certificates'],
        queryFn: ({ pageParam }) => certificatesApi.getCertificates(pageParam as number, 20),
        initialPageParam: 1,
        getNextPageParam: (lastPage, allPages) => {
            if (!lastPage || typeof lastPage.total === 'undefined') {
                return undefined
            }

            // 检查allPages是否存在
            if (!allPages) {
                return undefined
            }

            // 计算已加载的项目总数（累加每页的实际items长度）
            const loadedItemsCount = allPages.reduce((count, page) => {
                return count + (page.items?.length || 0)
            }, 0)

            // 如果服务器返回的总数大于已加载的数量，则还有下一页
            return lastPage.total > loadedItemsCount ? allPages.length + 1 : undefined
        },
        enabled: true,
    })

    // 扁平化分页数据
    const flatData = useMemo(() =>
        data?.pages.flatMap(page => page.items) || [],
        [data]
    )

    // 优化的无限滚动实现
    useEffect(() => {
        // 只有当有更多页面可加载时才创建观察器
        if (!hasNextPage) return

        const options = {
            // 降低threshold使其更容易触发
            threshold: 0.1,
            // 减小rootMargin以避免过早触发，但仍保持一定的预加载空间
            rootMargin: '100px 0px'
        }

        const handleObserver = (entries: IntersectionObserverEntry[]) => {
            const [entry] = entries
            if (entry.isIntersecting && hasNextPage && !isFetchingNextPage) {
                fetchNextPage()
            }
        }

        const observer = new IntersectionObserver(handleObserver, options)

        const sentinel = sentinelRef.current
        if (sentinel) {
            observer.observe(sentinel)
        }

        return () => {
            if (sentinel) {
                observer.unobserve(sentinel)
            }
            observer.disconnect()
        }
    }, [hasNextPage, isFetchingNextPage, fetchNextPage])

    // 打开创建证书对话框
    const openCreateDialog = () => {
        setDialogMode('create')
        setSelectedCertificate(null)
        setCertificateDialogOpen(true)
    }

    // 打开更新证书对话框
    const openUpdateDialog = (certificate: Certificate) => {
        setDialogMode('update')
        setSelectedCertificate(certificate)
        setCertificateDialogOpen(true)
    }

    // 打开删除对话框
    const openDeleteDialog = (id: string) => {
        setSelectedCertId(id)
        setDeleteDialogOpen(true)
    }

    // 助手函数：检查证书是否即将过期（30天内）
    const isExpiringSoon = (expireDate: string): boolean => {
        const expiryDate = new Date(expireDate)
        const now = new Date()
        const thirtyDaysInMs = 30 * 24 * 60 * 60 * 1000
        return expiryDate.getTime() - now.getTime() < thirtyDaysInMs && expiryDate > now
    }

    // 助手函数：检查证书是否已过期
    const isExpired = (expireDate: string): boolean => {
        return new Date(expireDate) < new Date()
    }

    const refreshSites = () => {
        setIsRefreshAnimating(true)
        refetch()

        // 停止动画，延迟1秒以匹配动画效果
        setTimeout(() => {
            setIsRefreshAnimating(false)
        }, 1000)
    }

    // 表格列定义
    const columns: ColumnDef<Certificate>[] = [
        {
            accessorKey: 'name',
            header: () => <div className="font-medium py-3.5 whitespace-nowrap dark:text-shadow-glow-white dark:text-white">{t("certificate.name")}</div>,
            cell: ({ row }) => <div className="font-medium dark:text-shadow-glow-white">{row.original.name}</div>,
        },
        {
            accessorKey: 'domains',
            header: () => <div className="font-medium py-3.5 whitespace-nowrap dark:text-shadow-glow-white dark:text-white">{t("certificate.domains")}</div>,
            cell: ({ row }) => (
                <div className="flex flex-wrap gap-1 max-w-xs">
                    {row.original.domains.map((domain, index) => (
                        <span key={index} className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs dark:text-shadow-glow-white">
                            {domain}
                        </span>
                    ))}
                </div>
            ),
        },
        {
            accessorKey: 'issuerAndExpiry',
            header: () => <div className="font-medium py-3.5 whitespace-nowrap dark:text-shadow-glow-white dark:text-white">{t("certificate.issuerAndExpiry")}</div>,
            cell: ({ row }) => (
                <div className="flex flex-col">
                    <span className="text-sm dark:text-shadow-glow-white">{row.original.issuerName}</span>
                    <span className="text-xs text-muted-foreground dark:text-shadow-glow-white">
                        {t("certificate.expiryDate")}{new Date(row.original.expireDate).toLocaleDateString()}
                        {isExpiringSoon(row.original.expireDate) && (
                            <span className="ml-2 px-1.5 py-0.5 bg-yellow-100 text-yellow-800 rounded text-xs dark:text-shadow-glow-white">
                                {t("certificate.expiringSoon")}
                            </span>
                        )}
                        {isExpired(row.original.expireDate) && (
                            <span className="ml-2 px-1.5 py-0.5 bg-red-100 text-red-800 rounded text-xs dark:text-shadow-glow-white">
                                {t("certificate.expired")}
                            </span>
                        )}
                    </span>
                </div>
            ),
        },
        {
            id: 'actions',
            cell: ({ row }) => (
                <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="icon" className="dark:text-shadow-glow-white">
                            <MoreHorizontal className="h-4 w-4 dark:text-shadow-glow-white" />
                        </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                        <DropdownMenuItem
                            onClick={() => openUpdateDialog(row.original)}
                        >
                            <Pencil className="mr-2 h-4 w-4 dark:text-shadow-glow-white" />
                            {t("certificate.edit")}
                        </DropdownMenuItem>
                        <DropdownMenuItem
                            onClick={() => openDeleteDialog(row.original.id)}
                            className="text-red-600 dark:text-red-400 dark:text-shadow-glow-white"
                        >
                            <Trash2 className="mr-2 h-4 w-4" />
                            {t("certificate.delete")}
                        </DropdownMenuItem>
                    </DropdownMenuContent>
                </DropdownMenu>
            ),
        },
    ]

    // 初始化表格
    const table = useReactTable({
        data: flatData,
        columns,
        getCoreRowModel: getCoreRowModel(),
    })

    return (
        <>
            <Card className="border-none shadow-none p-6 flex flex-col h-full">
                {/* 标题和操作按钮 - 固定在顶部 */}
                <div className="flex items-center justify-between mb-6">
                    <h2 className="text-xl font-semibold text-primary dark:text-white">{t("certificate.management")}</h2>
                    <div className="flex gap-2">
                        <AnimatedButton >
                            <Button
                                variant="outline"
                                size="sm"
                                onClick={refreshSites}
                                className="flex items-center gap-2 justify-center dark:text-shadow-glow-white"
                            >
                                <AnimatedIcon animationVariant="continuous-spin" isAnimating={isRefreshAnimating} className="h-4 w-4">
                                    <RefreshCcw className="h-4 w-4" />
                                </AnimatedIcon>
                                {t("certificate.refresh")}
                            </Button>
                        </AnimatedButton>
                        <AnimatedButton>
                            <Button
                                size="sm"
                                onClick={openCreateDialog}
                                className="flex items-center gap-1 dark:text-shadow-glow-white"
                            >
                                <Plus className="h-3.5 w-3.5 dark:text-shadow-glow-white" />
                                {t("certificate.add")}
                            </Button>
                        </AnimatedButton>
                    </div>
                </div>

                {/* 表格容器 - 设置固定高度和滚动 */}
                <div className="flex-1 overflow-hidden flex flex-col">
                    <div className="overflow-auto h-full">
                        <DataTable table={table}
                            loadingStyle="skeleton"
                            columns={columns}
                            isLoading={isLoading}
                            fixedHeader={true}
                            animatedRows={true}
                            showScrollShadows={true}
                        />

                        {/* 无限滚动监测元素 - 在滚动区域内 */}
                        {hasNextPage && <div
                            ref={sentinelRef}
                            className="h-5 flex justify-center items-center mt-4"
                        >
                            {isFetchingNextPage && (
                                <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
                            )}
                        </div>}
                    </div>
                </div>
            </Card>

            {/* 统一的证书对话框 */}
            <CertificateDialog
                open={certificateDialogOpen}
                onOpenChange={setCertificateDialogOpen}
                mode={dialogMode}
                certificate={selectedCertificate}
            />

            {/* 使用抽离出的删除对话框组件 */}
            <DeleteCertificateDialog
                open={deleteDialogOpen}
                onOpenChange={setDeleteDialogOpen}
                certificateId={selectedCertId}
                onDeleted={() => setSelectedCertId(null)}
            />
        </>
    )
}