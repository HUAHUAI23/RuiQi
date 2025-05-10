import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from "@/components/ui/dialog"
import { IPGroupForm } from "./IPGroupForm"
import { IPGroup } from "@/types/ip-group"
import { useTranslation } from "react-i18next"
import { AnimatePresence, motion } from "motion/react"
import {
    dialogEnterExitAnimation,
    dialogContentAnimation,
    dialogContentItemAnimation,
} from "@/components/ui/animation/dialog-animation"

interface IPGroupDialogProps {
    open: boolean
    onOpenChange: (open: boolean) => void
    mode: 'create' | 'update'
    ipGroup?: IPGroup | null
}

export function IPGroupDialog({ open, onOpenChange, mode = 'create', ipGroup = null }: IPGroupDialogProps) {
    const { t } = useTranslation()

    // Define title and description based on mode
    const title = mode === 'create'
        ? t("ipGroup.dialog.createTitle")
        : t("ipGroup.dialog.editTitle")

    const description = mode === 'create'
        ? t("ipGroup.dialog.createDescription")
        : t("ipGroup.dialog.editDescription")

    // Prepare default values based on mode
    const defaultValues = mode === 'update' && ipGroup
        ? {
            name: ipGroup.name,
            items: ipGroup.items,
        }
        : {
            name: '',
            items: [],
        }

    const handleSuccess = () => {
        onOpenChange(false)
    }

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <AnimatePresence>
                {open && (
                    <motion.div {...dialogEnterExitAnimation}>
                        <DialogContent className="sm:max-w-md">
                            <motion.div {...dialogContentAnimation}>
                                <DialogHeader className="px-2">
                                    <motion.div {...dialogContentItemAnimation}>
                                        <DialogTitle className="text-xl font-semibold dark:text-shadow-glow-white dark:text-white">
                                            {title}
                                        </DialogTitle>
                                        <DialogDescription className="dark:text-shadow-glow-white">
                                            {description}
                                        </DialogDescription>
                                    </motion.div>
                                </DialogHeader>

                                <div className="p-4">
                                    <motion.div {...dialogContentItemAnimation}>
                                        <IPGroupForm
                                            mode={mode}
                                            ipGroupId={ipGroup?.id}
                                            defaultValues={defaultValues}
                                            onSuccess={handleSuccess}
                                        />
                                    </motion.div>
                                </div>
                            </motion.div>
                        </DialogContent>
                    </motion.div>
                )}
            </AnimatePresence>
        </Dialog>
    )
}