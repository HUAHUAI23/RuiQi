import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Settings } from 'lucide-react'
import { BlockedIPTable } from '@/feature/flow-control/components/BlockedIPTable'
import { FlowControlDialog } from '@/feature/flow-control/components/FlowControlDialog'
import { useTranslation } from 'react-i18next'
import { AnimatedButton } from '@/components/ui/animation/components/animated-button'

export default function IpBlockPage() {
    const { t } = useTranslation()
    const [configDialogOpen, setConfigDialogOpen] = useState(false)

    return (
        <div className="flex flex-col h-full">
            {/* 页面标题和配置按钮 - 固定高度 */}
            <div className="flex items-center justify-end pt-6 px-6 flex-shrink-0">
                <AnimatedButton>
                    <Button
                        onClick={() => setConfigDialogOpen(true)}
                        className="flex items-center gap-2"
                    >
                        <Settings className="h-4 w-4" />
                        {t('flowControl.configSettings', '流量控制配置')}
                    </Button>
                </AnimatedButton>
            </div>

            {/* 封禁IP表格 - 弹性高度 */}
            <div className="flex-1 overflow-hidden">
                <BlockedIPTable />
            </div>

            {/* 配置对话框 */}
            <FlowControlDialog
                open={configDialogOpen}
                onOpenChange={setConfigDialogOpen}
            />
        </div>
    )
}