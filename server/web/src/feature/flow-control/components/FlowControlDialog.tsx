import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle
} from '@/components/ui/dialog'
import { FlowControlForm } from './FlowControlForm'
import { AnimatePresence, motion } from "motion/react"
import {
    dialogEnterExitAnimation,
    dialogContentAnimation,
    dialogHeaderAnimation,
    dialogContentItemAnimation
} from '@/components/ui/animation/dialog-animation'
import { useTranslation } from 'react-i18next'

interface FlowControlDialogProps {
    open: boolean
    onOpenChange: (open: boolean) => void
}

export function FlowControlDialog({
    open,
    onOpenChange
}: FlowControlDialogProps) {
    const { t } = useTranslation()

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <AnimatePresence mode="wait">
                {open && (
                    <motion.div {...dialogEnterExitAnimation}>
                        <DialogContent className="max-w-4xl max-h-[85vh] overflow-y-auto scrollbar-neon p-0">
                            <motion.div {...dialogContentAnimation}>
                                <motion.div {...dialogHeaderAnimation}>
                                    <DialogHeader className="p-6 pb-3">
                                        <DialogTitle className="text-xl">
                                            {t('flowControl.dialog.title', '流量控制配置')}
                                        </DialogTitle>
                                        <DialogDescription className="dark:text-shadow-glow-white">
                                            {t('flowControl.dialog.description', '配置基于IP的访问频率、攻击频率和错误频率限制，保护系统免受恶意攻击')}
                                        </DialogDescription>
                                    </DialogHeader>
                                </motion.div>

                                <motion.div
                                    {...dialogContentItemAnimation}
                                    className="px-6 pb-6"
                                >
                                    <FlowControlForm
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