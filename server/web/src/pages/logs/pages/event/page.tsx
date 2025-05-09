import { useState, useEffect, useRef } from "react"
import { useNavigate } from "react-router"
import { Card } from "@/components/ui/card"
import {
    ColumnDef,
    getCoreRowModel,
    getPaginationRowModel,
    useReactTable,
} from "@tanstack/react-table"
import { DataTable } from "@/components/table/motion-data-table"
import { DataTablePagination } from "@/components/table/pagination"
import { Button } from "@/components/ui/button"
import { useTranslation } from "react-i18next"
import { AttackEventFilter } from "@/feature/log/components/AttackEventFilter"
import { AttackEventQueryFormValues } from "@/validation/log"
import { AttackEventAggregateResult } from "@/types/log"
import { useAttackEvents } from "@/feature/log/hook/useAttackEvents"
import { Badge } from "@/components/ui/badge"
import { format } from "date-fns"
import { ExternalLink, AlertTriangle, History } from "lucide-react"
import { AdvancedErrorDisplay } from "@/components/common/error/errorDisplay"
import { produce } from "immer"

export default function EventsPage() {
    const { t } = useTranslation()
    const navigate = useNavigate()
    const pollingTimerRef = useRef<number | null>(null)

    const [queryParams, setQueryParams] = useState<AttackEventQueryFormValues>({
        page: 1,
        pageSize: 10
    })

    // 轮询状态
    const [enablePolling, setEnablePolling] = useState(false)
    const [pollingInterval, setPollingInterval] = useState(30) // 默认30秒

    const { data, isLoading, error, isError, refetch } = useAttackEvents(queryParams)

    // 设置轮询
    useEffect(() => {
        // 清除现有的轮询
        if (pollingTimerRef.current !== null) {
            clearInterval(pollingTimerRef.current)
            pollingTimerRef.current = null
        }

        // 如果启用了轮询，设置新的轮询
        if (enablePolling) {
            pollingTimerRef.current = window.setInterval(() => {
                refetch()
            }, pollingInterval * 1000)
        }

        // 组件卸载时清除轮询
        return () => {
            if (pollingTimerRef.current !== null) {
                clearInterval(pollingTimerRef.current)
            }
        }
    }, [enablePolling, pollingInterval, refetch])

    const handleFilter = (values: AttackEventQueryFormValues) => {
        setQueryParams(values)
    }

    const handlePollingChange = (enabled: boolean, interval: number) => {
        setEnablePolling(enabled)
        setPollingInterval(interval)
    }

    const navigateToLogs = (domain: string, srcIp: string) => {
        const params = new URLSearchParams()
        params.append('domain', domain)
        params.append('srcIp', srcIp)

        // 如果有设置时间，也传递过去
        if (queryParams.startTime) {
            params.append('startTime', queryParams.startTime)
        }
        if (queryParams.endTime) {
            params.append('endTime', queryParams.endTime)
        }

        navigate(`/logs/protect?${params.toString()}`)
    }

    const columns: ColumnDef<AttackEventAggregateResult>[] = [
        {
            accessorKey: "domain",
            header: () => <div className="whitespace-nowrap dark:text-shadow-glow-white dark:text-white">{t('domain')}</div>,
            cell: ({ row }) => <span className="font-medium break-all dark:text-shadow-glow-white">{row.getValue("domain")}</span>
        },
        {
            accessorKey: "dstPort",
            header: () => <div className="whitespace-nowrap dark:text-shadow-glow-white dark:text-white">{t('dstPort')}</div>,
            cell: ({ row }) => <span className="dark:text-shadow-glow-white">{row.getValue("dstPort")}</span>
        },
        {
            accessorKey: "srcIp",
            header: () => <div className="whitespace-nowrap dark:text-shadow-glow-white dark:text-white">{t('srcIp')}</div>,
            cell: ({ row }) => <span className="break-all dark:text-shadow-glow-white">{row.getValue("srcIp")}</span>
        },
        {
            accessorKey: "count",
            header: () => <div className="whitespace-nowrap dark:text-shadow-glow-white dark:text-white">{t('attackCount')}</div>,
            cell: ({ row }) => (
                <Button
                    variant="link"
                    onClick={() => navigateToLogs(row.original.domain, row.original.srcIp)}
                    className="flex items-center gap-1 p-0 dark:text-shadow-glow-white"
                >
                    {row.getValue("count")}
                    <ExternalLink className="h-3 w-3 dark:text-shadow-glow-white" />
                </Button>
            )
        },
        {
            accessorKey: "firstAttackTime",
            header: () => <div className="whitespace-nowrap dark:text-shadow-glow-white dark:text-white">{t('firstAttackTime')}</div>,
            cell: ({ row }) => (
                <div className="flex flex-col">
                    <span className="dark:text-shadow-glow-white">{format(new Date(row.getValue("firstAttackTime")), "yyyy-MM-dd")}</span>
                    <span className="text-sm text-muted-foreground dark:text-shadow-glow-white">{format(new Date(row.getValue("firstAttackTime")), "HH:mm:ss")}</span>
                </div>
            )
        },
        {
            accessorKey: "lastAttackTime",
            header: () => <div className="whitespace-nowrap dark:text-shadow-glow-white dark:text-white">{t('lastAttackTime')}</div>,
            cell: ({ row }) => (
                <div className="flex flex-col">
                    <span className="dark:text-shadow-glow-white">{format(new Date(row.getValue("lastAttackTime")), "yyyy-MM-dd")}</span>
                    <span className="text-sm text-muted-foreground dark:text-shadow-glow-white">{format(new Date(row.getValue("lastAttackTime")), "HH:mm:ss")}</span>
                </div>
            )
        },
        {
            accessorKey: "isOngoing",
            header: () => <div className="whitespace-nowrap dark:text-shadow-glow-white dark:text-white">{t('status')}</div>,
            cell: ({ row }) => {
                const isOngoing = row.getValue("isOngoing")
                const minutes = row.original.durationInMinutes || 0
                const hours = Math.floor(minutes / 60)
                const remainingMinutes = Math.round(minutes % 60)
                const durationText = hours > 0
                    ? `${hours}h ${remainingMinutes}m`
                    : `${remainingMinutes}m`
                return isOngoing ? (
                    <div className="flex flex-col items-start gap-1">
                        <Badge variant="destructive" className="flex items-center gap-1 animate-pulse bg-red-500 text-white">
                            <AlertTriangle className="h-3 w-3 dark:text-shadow-glow-white" />
                            {t('ongoing')}
                        </Badge>
                        <span className="text-xs text-muted-foreground dark:text-shadow-glow-white">
                            {t('attackDuration')}: {durationText}
                        </span>
                    </div>
                ) : (
                    <div className="flex flex-col items-start gap-1">
                        <Badge variant="outline" className="flex items-center gap-1 bg-amber-400 text-amber-900 border-amber-500">
                            <History className="h-3 w-3 dark:text-shadow-glow-white" />
                            {t('attackEnded')}
                        </Badge>
                        <span className="text-xs text-amber-500 font-medium dark:text-shadow-glow-white">
                            {t('noOngoingAttack')}
                        </span>
                    </div>
                )
            }
        }
    ]

    const table = useReactTable({
        data: data?.results || [],
        columns,
        pageCount: data?.totalPages || 0,
        getCoreRowModel: getCoreRowModel(),
        getPaginationRowModel: getPaginationRowModel(),
        manualPagination: true,
        state: {
            pagination: {
                pageIndex: (queryParams.page || 1) - 1,
                pageSize: queryParams.pageSize || 10
            }
        },
        onPaginationChange: (updater) => {
            if (typeof updater === 'function') {
                const oldPagination = {
                    pageIndex: (queryParams.page || 1) - 1,
                    pageSize: queryParams.pageSize || 10
                }
                const newPagination = updater(oldPagination)

                // 使用 immer 统一处理分页变化
                setQueryParams(produce(draft => {
                    // 只有当页码改变时才更新页码
                    if (newPagination.pageIndex !== oldPagination.pageIndex) {
                        draft.page = newPagination.pageIndex + 1
                    }

                    // 只有当每页条数改变时才更新每页条数并重置页码
                    if (newPagination.pageSize !== oldPagination.pageSize) {
                        draft.pageSize = newPagination.pageSize
                        draft.page = 1 // 重置到第一页
                    }
                }))
            }
        }
    })


    return (
        <Card className="flex flex-col h-full p-0 border-none shadow-none">
            {/* 头部区域 - 固定高度 */}
            <div className="p-6 flex-shrink-0">
                <AttackEventFilter
                    onFilter={handleFilter}
                    onRefresh={refetch}
                    defaultValues={queryParams}
                    enablePolling={enablePolling}
                    pollingInterval={pollingInterval}
                    onPollingChange={handlePollingChange}
                />
            </div>

            {/* 表格区域 - 弹性高度，可滚动 */}
            <div className="px-6 flex-1 overflow-auto">
                {isError ? (
                    <AdvancedErrorDisplay error={error} onRetry={refetch} />
                ) : (
                    <DataTable
                        loadingStyle="skeleton"
                        table={table}
                        columns={columns}
                        isLoading={isLoading}
                        fixedHeader={true}
                        animatedRows={true}
                        showScrollShadows={true}
                    />
                )}
            </div>

            {/* 底部分页 - 固定高度 */}
            {!isError && <div className="py-6 px-4 flex-shrink-0">
                <DataTablePagination table={table} />
            </div>}
        </Card>
    )
} 