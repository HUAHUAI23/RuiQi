import { useState, useRef, useEffect, useMemo } from 'react'
import {
    useReactTable,
    getCoreRowModel,
    ColumnDef,
} from '@tanstack/react-table'
import { useInfiniteQuery } from '@tanstack/react-query'
import { ruleApi } from '@/api/rule'
import { MicroRule } from '@/types/rule'
import { Button } from '@/components/ui/button'
import {
    MoreHorizontal, Plus, Trash2, RefreshCcw, Pencil, ShieldCheck, ShieldOff
} from 'lucide-react'
import {
    DropdownMenu, DropdownMenuContent,
    DropdownMenuItem, DropdownMenuTrigger
} from '@/components/ui/dropdown-menu'
import { Card } from '@/components/ui/card'
import { MicroRuleDialog } from './MicroRuleDialog'
import { Loader2 } from 'lucide-react'
import { DataTable } from '@/components/table/motion-data-table'
import { DeleteMicroRuleDialog } from './DeleteMicroRuleDialog'
import { useTranslation } from 'react-i18next'
import { AnimatedIcon } from '@/components/ui/animation/components/animated-icon'
import { AnimatedButton } from '@/components/ui/animation/components/animated-button'
import { Badge } from '@/components/ui/badge'

export function MicroRuleTable() {
    const { t } = useTranslation()

    // 状态管理
    const [ruleDialogOpen, setRuleDialogOpen] = useState(false)
    const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
    const [selectedRuleId, setSelectedRuleId] = useState<string | null>(null)
    const sentinelRef = useRef<HTMLDivElement>(null)
    const [dialogMode, setDialogMode] = useState<'create' | 'update'>('create')
    const [selectedRule, setSelectedRule] = useState<MicroRule | null>(null)

    const [isRefreshAnimating, setIsRefreshAnimating] = useState(false)

    // 获取规则列表
    const {
        data,
        isLoading,
        fetchNextPage,
        hasNextPage,
        isFetchingNextPage,
        refetch
    } = useInfiniteQuery({
        queryKey: ['microRules'],
        queryFn: ({ pageParam }) => ruleApi.getMicroRules(pageParam as number, 20),
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

    // 打开创建规则对话框
    const openCreateDialog = () => {
        setDialogMode('create')
        setSelectedRule(null)
        setRuleDialogOpen(true)
    }

    // 打开更新规则对话框
    const openUpdateDialog = (rule: MicroRule) => {
        setDialogMode('update')
        setSelectedRule(rule)
        setRuleDialogOpen(true)
    }

    // 打开删除对话框
    const openDeleteDialog = (id: string) => {
        setSelectedRuleId(id)
        setDeleteDialogOpen(true)
    }

    const refreshRules = () => {
        setIsRefreshAnimating(true)
        refetch()

        // 停止动画，延迟1秒以匹配动画效果
        setTimeout(() => {
            setIsRefreshAnimating(false)
        }, 1000)
    }

    // 表格列定义
    const columns: ColumnDef<MicroRule>[] = [
        {
            accessorKey: 'name',
            header: () => <div className="font-medium py-3.5 whitespace-nowrap dark:text-shadow-glow-white dark:text-white">{t("microRule.table.name")}</div>,
            cell: ({ row }) => <div className="font-medium dark:text-shadow-glow-white">{row.original.name}</div>,
        },
        {
            accessorKey: 'type',
            header: () => <div className="font-medium py-3.5 whitespace-nowrap dark:text-shadow-glow-white dark:text-white">{t("microRule.table.type")}</div>,
            cell: ({ row }) => (
                <Badge variant={row.original.type === 'whitelist' ? 'outline' : 'destructive'} className="dark:text-shadow-glow-white">
                    {row.original.type === 'whitelist' ? t("microRule.form.whitelist") : t("microRule.form.blacklist")}
                </Badge>
            ),
        },
        {
            accessorKey: 'status',
            header: () => <div className="font-medium py-3.5 whitespace-nowrap dark:text-shadow-glow-white dark:text-white">{t("microRule.table.status")}</div>,
            cell: ({ row }) => (
                <div className="flex items-center">
                    {row.original.status === 'enabled' ? (
                        <ShieldCheck className="h-4 w-4 text-green-500 mr-1" />
                    ) : (
                        <ShieldOff className="h-4 w-4 text-gray-400 mr-1" />
                    )}
                    <span className="dark:text-shadow-glow-white">
                        {row.original.status === 'enabled' ? t("microRule.form.enabled") : t("microRule.form.disabled")}
                    </span>
                </div>
            ),
        },
        {
            accessorKey: 'priority',
            header: () => <div className="font-medium py-3.5 whitespace-nowrap dark:text-shadow-glow-white dark:text-white">{t("microRule.table.priority")}</div>,
            cell: ({ row }) => <div className="dark:text-shadow-glow-white">{row.original.priority}</div>,
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
                            className="dark:text-shadow-glow-white"
                            onClick={() => openUpdateDialog(row.original)}
                        >
                            <Pencil className="mr-2 h-4 w-4 dark:text-shadow-glow-white" />
                            {t("microRule.table.edit")}
                        </DropdownMenuItem>
                        <DropdownMenuItem
                            onClick={() => openDeleteDialog(row.original.id)}
                            className="text-red-600 dark:text-red-400 dark:text-shadow-glow-white"
                        >
                            <Trash2 className="mr-2 h-4 w-4" />
                            {t("microRule.table.delete")}
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
                    <h2 className="text-xl font-semibold text-primary dark:text-white">{t("microRule.title")}</h2>
                    <div className="flex gap-2">
                        <AnimatedButton>
                            <Button
                                variant="outline"
                                size="sm"
                                onClick={refreshRules}
                                className="flex items-center gap-2 justify-center dark:text-shadow-glow-white"
                            >
                                <AnimatedIcon animationVariant="continuous-spin" isAnimating={isRefreshAnimating} className="h-4 w-4">
                                    <RefreshCcw className="h-4 w-4" />
                                </AnimatedIcon>
                                {t("microRule.refresh")}
                            </Button>
                        </AnimatedButton>
                        <AnimatedButton>
                            <Button
                                size="sm"
                                onClick={openCreateDialog}
                                className="flex items-center gap-1 dark:text-shadow-glow-white"
                            >
                                <Plus className="h-3.5 w-3.5 dark:text-shadow-glow-white" />
                                {t("microRule.add")}
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

            {/* 规则对话框 */}
            <MicroRuleDialog
                open={ruleDialogOpen}
                onOpenChange={setRuleDialogOpen}
                mode={dialogMode}
                rule={selectedRule}
            />

            {/* 删除对话框 */}
            <DeleteMicroRuleDialog
                open={deleteDialogOpen}
                onOpenChange={setDeleteDialogOpen}
                ruleId={selectedRuleId}
                onDeleted={() => setSelectedRuleId(null)}
            />
        </>
    )
}