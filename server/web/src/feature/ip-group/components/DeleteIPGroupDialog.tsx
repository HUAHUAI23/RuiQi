import { AlertCircle } from "lucide-react"
import { useTranslation } from "react-i18next"
import { IPGroup } from "@/types/ip-group"
import { useDeleteIPGroup } from "../hooks"
import {
    AlertDialog,
    AlertDialogAction,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
} from "@/components/ui/alert-dialog"

interface DeleteIPGroupDialogProps {
    open: boolean
    onOpenChange: (open: boolean) => void
    ipGroup: IPGroup | null
}

export function DeleteIPGroupDialog({ open, onOpenChange, ipGroup }: DeleteIPGroupDialogProps) {
    const { t } = useTranslation()
    const { deleteIPGroup, isLoading: isDeleting } = useDeleteIPGroup()

    const confirmDelete = () => {
        if (ipGroup) {
            deleteIPGroup(ipGroup.id)
            onOpenChange(false)
        }
    }

    return (
        <AlertDialog open={open} onOpenChange={onOpenChange}>
            <AlertDialogContent className="dark:bg-accent/10 dark:border-slate-800 dark:card-neon">
                <AlertDialogHeader>
                    <AlertDialogTitle className="flex items-center gap-2 dark:text-shadow-glow-white dark:text-white">
                        <AlertCircle className="h-5 w-5 text-destructive dark:text-red-500 dark:icon-neon" />
                        {t('ipGroup.deleteDialog.title')}
                    </AlertDialogTitle>
                    <AlertDialogDescription>
                        {t('ipGroup.deleteDialog.description', { name: ipGroup?.name })}
                    </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                    <AlertDialogCancel className="dark:border-slate-700 dark:text-slate-300 dark:text-shadow-glow-white dark:button-neon">
                        {t('common.cancel')}
                    </AlertDialogCancel>
                    <AlertDialogAction
                        onClick={confirmDelete}
                        disabled={isDeleting}
                        className="bg-destructive text-destructive-foreground hover:bg-destructive/90 dark:bg-red-900 dark:hover:bg-red-800 dark:text-white dark:text-shadow-glow-white"
                    >
                        {t('common.delete')}
                    </AlertDialogAction>
                </AlertDialogFooter>
            </AlertDialogContent>
        </AlertDialog>
    )
} 