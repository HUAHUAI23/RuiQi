import { useState } from "react"
import { Card } from "@/components/ui/card"
import {
    ColumnDef,
    getCoreRowModel,
    getPaginationRowModel,
    useReactTable
} from "@tanstack/react-table"
import { DataTable } from "@/components/table/motion-data-table"
import { DataTablePagination } from "@/components/table/pagination"
import { Button } from "@/components/ui/button"
import { useTranslation } from "react-i18next"
import { IPGroup } from "@/types/ip-group"
import { useIPGroups } from "@/feature/ip-group/hooks"
import { IPGroupDialog } from "@/feature/ip-group/components/IPGroupDialog"
import { DeleteIPGroupDialog } from "@/feature/ip-group/components/DeleteIPGroupDialog"
import { Plus, Pencil, Trash2 } from "lucide-react"
import { AdvancedErrorDisplay } from "@/components/common/error/errorDisplay"
import { Badge } from "@/components/ui/badge"

export default function IPGroupPage() {
    const { t } = useTranslation()
    const [page, setPage] = useState(1)
    const [pageSize, setPageSize] = useState(10)
    const { data, isLoading, isError, error, refetch } = useIPGroups(page, pageSize)

    // Dialog states
    const [dialogMode, setDialogMode] = useState<'create' | 'update'>('create')
    const [selectedIPGroup, setSelectedIPGroup] = useState<IPGroup | null>(null)
    const [ipGroupDialogOpen, setIPGroupDialogOpen] = useState(false)
    const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
    const [ipGroupToDelete, setIPGroupToDelete] = useState<IPGroup | null>(null)

    const handleOpenCreateDialog = () => {
        setDialogMode('create')
        setSelectedIPGroup(null)
        setIPGroupDialogOpen(true)
    }

    const handleOpenEditDialog = (ipGroup: IPGroup) => {
        setDialogMode('update')
        setSelectedIPGroup(ipGroup)
        setIPGroupDialogOpen(true)
    }

    const handleOpenDeleteDialog = (ipGroup: IPGroup) => {
        setIPGroupToDelete(ipGroup)
        setDeleteDialogOpen(true)
    }

    const columns: ColumnDef<IPGroup>[] = [
        {
            accessorKey: "name",
            header: () => <div className="whitespace-nowrap dark:text-shadow-glow-white dark:text-white">{t('ipGroup.table.name')}</div>,
            cell: ({ row }) => (
                <div className="font-medium dark:text-shadow-glow-white">
                    {row.getValue("name")}
                </div>
            )
        },
        {
            accessorKey: "items",
            header: () => <div className="whitespace-nowrap dark:text-shadow-glow-white dark:text-white">{t('ipGroup.table.ipAddresses')}</div>,
            cell: ({ row }) => {
                const items = row.getValue("items") as string[]
                return (
                    <div className="flex flex-wrap gap-1">
                        {items.length > 3 ? (
                            <>
                                {items.slice(0, 3).map((item, index) => (
                                    <Badge key={index} variant="outline" className="font-mono dark:border-slate-700 dark:bg-slate-800/70 dark:text-shadow-glow-white">
                                        {item}
                                    </Badge>
                                ))}
                                <Badge variant="outline" className="dark:border-slate-700 dark:bg-slate-800/70 dark:text-shadow-glow-white">
                                    +{items.length - 3}
                                </Badge>
                            </>
                        ) : (
                            items.map((item, index) => (
                                <Badge key={index} variant="outline" className="font-mono dark:border-slate-700 dark:bg-slate-800/70 dark:text-shadow-glow-white">
                                    {item}
                                </Badge>
                            ))
                        )}
                    </div>
                )
            }
        },
        {
            id: "actions",
            header: () => <div className="whitespace-nowrap dark:text-shadow-glow-white dark:text-white">{t('ipGroup.table.actions')}</div>,
            cell: ({ row }) => {
                const ipGroup = row.original
                return (
                    <div className="flex items-center gap-2">
                        <Button
                            variant="ghost"
                            size="icon"
                            onClick={() => handleOpenEditDialog(ipGroup)}
                            className="h-8 w-8 dark:text-shadow-glow-white dark:button-neon"
                        >
                            <Pencil className="h-4 w-4 dark:icon-neon" />
                        </Button>
                        <Button
                            variant="ghost"
                            size="icon"
                            onClick={() => handleOpenDeleteDialog(ipGroup)}
                            className="h-8 w-8 text-destructive hover:text-destructive dark:text-red-500 dark:hover:text-red-400 dark:button-neon"
                        >
                            <Trash2 className="h-4 w-4 dark:icon-neon" />
                        </Button>
                    </div>
                )
            }
        }
    ]

    const table = useReactTable({
        data: data?.items || [],
        columns,
        pageCount: data ? Math.ceil(data.total / pageSize) : 0,
        getCoreRowModel: getCoreRowModel(),
        getPaginationRowModel: getPaginationRowModel(),
        manualPagination: true,
        state: {
            pagination: {
                pageIndex: page - 1,
                pageSize: pageSize
            }
        },
        onPaginationChange: (updater) => {
            if (typeof updater === 'function') {
                const oldPagination = {
                    pageIndex: page - 1,
                    pageSize: pageSize
                }
                const newPagination = updater(oldPagination)

                // 只有当页码改变时才更新页码
                if (newPagination.pageIndex !== oldPagination.pageIndex) {
                    setPage(newPagination.pageIndex + 1)
                }

                // 只有当每页条数改变时才更新每页条数并重置页码
                if (newPagination.pageSize !== oldPagination.pageSize) {
                    setPageSize(newPagination.pageSize)
                    setPage(1) // 重置到第一页
                }
            }
        }
    })

    return (
        <Card className="flex flex-col h-full p-6 border-none shadow-none dark:bg-accent/10 dark:card-neon">
            {/* 头部操作栏 */}
            <div className="flex justify-between items-center mb-6">
                <h2 className="text-2xl font-bold dark:text-shadow-glow-white dark:text-white">
                    {t('ipGroup.title')}
                </h2>
                <Button
                    onClick={handleOpenCreateDialog}
                    className="dark:text-shadow-glow-white dark:button-neon"
                >
                    <Plus className="h-4 w-4 mr-2 dark:icon-neon" />
                    {t('ipGroup.createButton')}
                </Button>
            </div>

            {/* 表格区域 */}
            <div className="flex-1 overflow-auto">
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

            {/* 底部分页 */}
            {!isError && (
                <div className="mt-4">
                    <DataTablePagination table={table} />
                </div>
            )}

            {/* IP组对话框 */}
            <IPGroupDialog
                open={ipGroupDialogOpen}
                onOpenChange={setIPGroupDialogOpen}
                mode={dialogMode}
                ipGroup={selectedIPGroup}
            />

            {/* 删除确认对话框 */}
            <DeleteIPGroupDialog
                open={deleteDialogOpen}
                onOpenChange={setDeleteDialogOpen}
                ipGroup={ipGroupToDelete}
            />
        </Card>
    )
}