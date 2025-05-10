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
import { useDeleteCertificate } from '../hooks/useCertificate'
import { AnimatePresence, motion } from "motion/react"
import {
    dialogEnterExitAnimation,
    dialogContentAnimation,
    dialogHeaderAnimation,
    dialogContentItemAnimation
} from '@/components/ui/animation/dialog-animation'
import { useTranslation } from 'react-i18next'
import { AnimatedButton } from "@/components/ui/animation/components/animated-button"

interface DeleteCertificateDialogProps {
    open: boolean
    onOpenChange: (open: boolean) => void
    certificateId: string | null
    onDeleted?: () => void
}

export function DeleteCertificateDialog({
    open,
    onOpenChange,
    certificateId,
    onDeleted
}: DeleteCertificateDialogProps) {
    const { t } = useTranslation()
    // 删除证书钩子
    const { deleteCertificate, isLoading } = useDeleteCertificate()

    // 处理删除证书
    const handleDeleteCertificate = () => {
        if (!certificateId) return

        deleteCertificate(certificateId, {
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
                                        <AlertDialogTitle className="text-xl">{t("certificate.deleteDialog.confirmTitle")}</AlertDialogTitle>
                                        <AlertDialogDescription className="dark:text-shadow-glow-white">
                                            {t("certificate.deleteDialog.confirmDescription")}
                                        </AlertDialogDescription>
                                    </AlertDialogHeader>
                                </motion.div>

                                <motion.div
                                    {...dialogContentItemAnimation}
                                    className="px-6 pb-6"
                                >
                                    <AlertDialogFooter className="mt-2 flex justify-end space-x-2">
                                        <AnimatedButton>
                                            <AlertDialogCancel>{t("certificate.deleteDialog.cancel")}</AlertDialogCancel>
                                        </AnimatedButton>
                                        <AnimatedButton>
                                            <AlertDialogAction
                                                onClick={handleDeleteCertificate}
                                                disabled={isLoading}
                                                className="bg-red-600 hover:bg-red-700"
                                            >
                                                {isLoading ? t("certificate.deleteDialog.deleting") : t("certificate.deleteDialog.delete")}
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
