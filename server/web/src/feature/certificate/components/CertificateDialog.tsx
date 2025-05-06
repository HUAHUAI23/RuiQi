import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle
} from '@/components/ui/dialog'
import { CertificateForm } from './CertificateForm'
import { Certificate } from '@/types/certificate'
import { AnimatePresence, motion } from "motion/react"
import {
    dialogEnterExitAnimation,
    dialogContentAnimation,
    dialogHeaderAnimation,
    dialogContentItemAnimation
} from '@/components/ui/animation/dialog-animation'
import { useTranslation } from 'react-i18next'

interface CertificateDialogProps {
    open: boolean
    onOpenChange: (open: boolean) => void
    mode: 'create' | 'update'
    certificate?: Certificate | null // 仅在编辑模式下需要
}

export function CertificateDialog({
    open,
    onOpenChange,
    mode = 'create',
    certificate = null
}: CertificateDialogProps) {
    const { t } = useTranslation()

    // 根据模式确定标题和描述
    const title = mode === 'create' ? t("certificate.dialog.createTitle") : t("certificate.dialog.updateTitle")
    const description = mode === 'create'
        ? t("certificate.dialog.createDescription")
        : t("certificate.dialog.updateDescription")

    // 根据模式准备表单默认值
    const defaultValues = mode === 'update' && certificate
        ? {
            name: certificate.name,
            description: certificate.description,
            publicKey: certificate.publicKey,
            privateKey: certificate.privateKey,
            domains: certificate.domains,
            expireDate: certificate.expireDate,
            fingerPrint: certificate.fingerPrint,
            issuerName: certificate.issuerName
        }
        : {
            name: '',
            description: '',
            publicKey: '',
            privateKey: '',
        }

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <AnimatePresence mode="wait">
                {open && (
                    <motion.div {...dialogEnterExitAnimation}>
                        <DialogContent className="max-w-2xl max-h-[85vh] overflow-y-auto scrollbar-neon p-0">
                            <motion.div {...dialogContentAnimation}>
                                <motion.div {...dialogHeaderAnimation}>
                                    <DialogHeader className="p-6 pb-3">
                                        <DialogTitle className="text-xl">{title}</DialogTitle>
                                        <DialogDescription className="dark:text-shadow-glow-white">{description}</DialogDescription>
                                    </DialogHeader>
                                </motion.div>

                                <motion.div
                                    {...dialogContentItemAnimation}
                                    className="px-6 pb-6"
                                >
                                    <CertificateForm
                                        mode={mode}
                                        certificateId={certificate?.id}
                                        defaultValues={defaultValues}
                                        onSuccess={() => onOpenChange(false)}
                                    />
                                </motion.div>
                            </motion.div>
                        </DialogContent>
                    </motion.div>
                )}
            </AnimatePresence>
        </Dialog>
    )
} 