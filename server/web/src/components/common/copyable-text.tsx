// src/components/common/copyable-text.tsx
import { useState } from "react"
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip"
import { Check, Copy } from "lucide-react"
import { cn } from "@/lib/utils"
import { motion, AnimatePresence } from "motion/react"
import { useTranslation } from "react-i18next"

interface CopyableTextProps {
    text: string
    className?: string
    showTooltip?: boolean
}

export function CopyableText({ text, className, showTooltip = true }: CopyableTextProps) {
    const [copied, setCopied] = useState(false)
    const { t } = useTranslation()

    const handleCopy = () => {
        navigator.clipboard.writeText(text)
        setCopied(true)
        setTimeout(() => setCopied(false), 2000)
    }

    return (
        <TooltipProvider delayDuration={300}>
            <Tooltip>
                <TooltipTrigger asChild>
                    <div
                        className={cn(
                            "group flex items-center gap-2 cursor-pointer rounded-md relative py-0.5 pr-2 pl-0.5",
                            "hover:bg-slate-50 transition-all duration-200",
                            "border border-transparent hover:border-slate-200",
                            "dark:hover:bg-slate-800/50 dark:hover:border-slate-700",
                            className
                        )}
                        onClick={handleCopy}
                    >
                        <span className="truncate">{text}</span>
                        <AnimatePresence mode="wait">
                            {copied ? (
                                <motion.div
                                    key="check"
                                    initial={{ scale: 0.8, opacity: 0 }}
                                    animate={{ scale: 1, opacity: 1 }}
                                    exit={{ scale: 0.8, opacity: 0 }}
                                    transition={{ duration: 0.15 }}
                                    className="flex items-center justify-center h-5 w-5 rounded-full bg-green-200 dark:bg-green-900/70"
                                >
                                    <Check className="h-3 w-3 text-green-500 dark:text-green-400 dark:icon-neon" strokeWidth={3} />
                                </motion.div>
                            ) : (
                                <motion.div
                                    key="copy"
                                    initial={{ scale: 0.8, opacity: 0 }}
                                    animate={{ scale: 1, opacity: 1 }}
                                    exit={{ scale: 0.8, opacity: 0 }}
                                    transition={{ duration: 0.15 }}
                                    className="flex items-center justify-center h-5 w-5 rounded-full bg-slate-100 opacity-0 group-hover:opacity-100 transition-opacity duration-200 dark:bg-slate-700"
                                >
                                    <Copy className="h-3 w-3 text-slate-500 dark:text-slate-300 dark:icon-neon" />
                                </motion.div>
                            )}
                        </AnimatePresence>
                    </div>
                </TooltipTrigger>
                {showTooltip && (
                    <TooltipContent
                        side="top"
                        className="max-w-[350px] break-all bg-white border border-slate-200 shadow-md py-2 px-3 text-sm text-slate-800 dark:bg-slate-800 dark:border-slate-700 dark:!text-slate-200 dark:text-shadow-glow-white"
                    >
                        {copied ? t("common.copied") : text}
                    </TooltipContent>
                )}
            </Tooltip>
        </TooltipProvider>
    )
}