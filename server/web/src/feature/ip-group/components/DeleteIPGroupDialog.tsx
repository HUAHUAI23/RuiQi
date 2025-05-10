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
import { AnimatePresence, motion } from "motion/react"
import {
    dialogEnterExitAnimation,
    dialogContentAnimation,
    dialogHeaderAnimation,
    dialogContentItemAnimation
} from '@/components/ui/animation/dialog-animation'
import { AnimatedButton } from "@/components/ui/animation/components/animated-button"

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
            <AnimatePresence mode="wait">
                {open && (
                    <motion.div {...dialogEnterExitAnimation}>
                        <AlertDialogContent className="p-0 overflow-hidden dark:bg-accent/10 dark:border-slate-800 dark:card-neon">
                            <motion.div {...dialogContentAnimation}>
                                <motion.div {...dialogHeaderAnimation}>
                                    <AlertDialogHeader className="p-6 pb-3">
                                        <AlertDialogTitle className="flex items-center gap-2 dark:text-shadow-glow-white dark:text-white">
                                            <AlertCircle className="h-5 w-5 text-destructive dark:text-red-500 dark:icon-neon" />
                                            {t('ipGroup.deleteDialog.title')}
                                        </AlertDialogTitle>
                                        <AlertDialogDescription>
                                            {t('ipGroup.deleteDialog.description', { name: ipGroup?.name })}
                                        </AlertDialogDescription>
                                    </AlertDialogHeader>
                                </motion.div>

                                <motion.div
                                    {...dialogContentItemAnimation}
                                    className="px-6 pb-6"
                                >
                                    <AlertDialogFooter className="mt-2 flex justify-end space-x-2">
                                        <AnimatedButton>
                                            <AlertDialogCancel className="dark:border-slate-700 dark:text-slate-300 dark:text-shadow-glow-white dark:button-neon">
                                                {t('common.cancel')}
                                            </AlertDialogCancel>
                                        </AnimatedButton>
                                        <AnimatedButton>
                                            <AlertDialogAction
                                                onClick={confirmDelete}
                                                disabled={isDeleting}
                                                className="bg-destructive text-destructive-foreground hover:bg-destructive/90 dark:bg-red-900 dark:hover:bg-red-800 dark:text-white dark:text-shadow-glow-white"
                                            >
                                                {t('common.delete')}
                                            </AlertDialogAction>
                                        </AnimatedButton>
                                    </AlertDialogFooter>
                                </motion.div>
                            </motion.div>
                        </AlertDialogContent>
                    </motion.div>
                )}
            </AnimatePresence>
        </AlertDialog>
    )
} 