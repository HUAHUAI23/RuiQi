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
import { useDeleteMicroRule } from '../hooks/useMicroRule'
import { AnimatePresence, motion } from "motion/react"
import {
    dialogEnterExitAnimation,
    dialogContentAnimation,
    dialogHeaderAnimation,
    dialogContentItemAnimation
} from '@/components/ui/animation/dialog-animation'
import { useTranslation } from 'react-i18next'
import { AnimatedButton } from "@/components/ui/animation/components/animated-button"

interface DeleteMicroRuleDialogProps {
    open: boolean
    onOpenChange: (open: boolean) => void
    ruleId: string | null
    onDeleted?: () => void
}

export function DeleteMicroRuleDialog({
    open,
    onOpenChange,
    ruleId,
    onDeleted
}: DeleteMicroRuleDialogProps) {
    const { t } = useTranslation()
    // 删除规则钩子
    const { deleteMicroRule, isLoading } = useDeleteMicroRule()

    // 处理删除规则
    const handleDeleteMicroRule = () => {
        if (!ruleId) return

        deleteMicroRule(ruleId, {
            onSettled: () => {
                onOpenChange(false)
                onDeleted?.()
            }
        })
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
                                        <AlertDialogTitle className="text-xl">{t("microRule.deleteDialog.confirmTitle")}</AlertDialogTitle>
                                        <AlertDialogDescription className="dark:text-shadow-glow-white">
                                            {t("microRule.deleteDialog.confirmDescription")}
                                        </AlertDialogDescription>
                                    </AlertDialogHeader>
                                </motion.div>

                                <motion.div
                                    {...dialogContentItemAnimation}
                                    className="px-6 pb-6"
                                >
                                    <AlertDialogFooter className="mt-2 flex justify-end space-x-2">
                                        <AnimatedButton>
                                            <AlertDialogCancel>{t("microRule.deleteDialog.cancel")}</AlertDialogCancel>
                                        </AnimatedButton>
                                        <AnimatedButton>
                                            <AlertDialogAction
                                                onClick={handleDeleteMicroRule}
                                                disabled={isLoading}
                                                className="bg-red-600 hover:bg-red-700"
                                            >
                                                {isLoading ? t("microRule.deleteDialog.deleting") : t("microRule.deleteDialog.delete")}
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