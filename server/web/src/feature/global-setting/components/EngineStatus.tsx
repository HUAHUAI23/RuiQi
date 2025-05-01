// src/feature/global-setting/components/EngineStatus.tsx
import { RunnerStatusResponse } from '@/types/runner'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import {
    Play,
    Square,
    RefreshCw,
    X,
    RotateCw,
    CheckCircle,
    XCircle,
    Loader2,
    Activity
} from 'lucide-react'
import { useTranslation } from 'react-i18next'

interface EngineStatusProps {
    status?: RunnerStatusResponse
    isLoading: boolean
    onStart: () => void
    onStop: () => void
    onRestart: () => void
    onForceStop: () => void
    onReload: () => void
    isControlLoading: boolean
}

export function EngineStatus({
    status,
    isLoading,
    onStart,
    onStop,
    onRestart,
    onForceStop,
    onReload,
    isControlLoading
}: EngineStatusProps) {
    const isRunning = status?.isRunning || false
    const { t } = useTranslation()

    return (
        <Card className="border-none shadow-none">
            <CardContent className="p-0">
                <div className="flex items-center gap-2 mb-4">
                    <Activity className="h-5 w-5 text-primary dark:text-shadow-glow-white" />
                    <h3 className="text-lg font-medium dark:text-shadow-glow-white">{t("globalSetting.engine.status")}</h3>
                </div>

                <div className="flex items-center gap-4 mb-6">
                    <div className="text-sm font-medium dark:text-shadow-glow-white">{t("globalSetting.engine.currentStatus")}</div>
                    {isLoading ? (
                        <Badge variant="outline" className="gap-1">
                            <Loader2 className="h-3 w-3 animate-spin" />
                            <span className="dark:text-shadow-glow-white">{t("globalSetting.engine.loading")}</span>
                        </Badge>
                    ) : isRunning ? (
                        <Badge variant="outline" className="gap-1 dark:badge-neon">
                            <CheckCircle className="h-3 w-3 dark:text-shadow-glow-white" />
                            <span className="dark:text-shadow-glow-white">{t("globalSetting.engine.running")}</span>
                        </Badge>
                    ) : (
                        <Badge variant="destructive" className="gap-1 dark:badge-neon">
                            <XCircle className="h-3 w-3 dark:text-shadow-glow-white" />
                            <span className="dark:text-shadow-glow-white">{t("globalSetting.engine.stopped")}</span>
                        </Badge>
                    )}
                </div>

                <div className="flex flex-wrap gap-2">
                    <Button
                        variant="default"
                        size="sm"
                        className="gap-1 dark:text-shadow-glow-white"
                        onClick={onStart}
                        disabled={isRunning || isLoading || isControlLoading}
                    >
                        {isControlLoading ? (
                            <Loader2 className="h-4 w-4 animate-spin" />
                        ) : (
                            <Play className="h-4 w-4 dark:text-shadow-glow-white" />
                        )}
                        {isControlLoading ? t("globalSetting.engine.processing") : t("globalSetting.engine.start")}
                    </Button>
                    <Button
                        variant="outline"
                        size="sm"
                        className="gap-1 dark:text-shadow-glow-white"
                        onClick={onStop}
                        disabled={!isRunning || isLoading || isControlLoading}
                    >
                        {isControlLoading ? (
                            <Loader2 className="h-4 w-4 animate-spin" />
                        ) : (
                            <Square className="h-4 w-4 dark:text-shadow-glow-white" />
                        )}
                        {isControlLoading ? t("globalSetting.engine.processing") : t("globalSetting.engine.stop")}
                    </Button>
                    <Button
                        variant="outline"
                        size="sm"
                        className="gap-1 dark:text-shadow-glow-white"
                        onClick={onRestart}
                        disabled={!isRunning || isLoading || isControlLoading}
                    >
                        {isControlLoading ? (
                            <Loader2 className="h-4 w-4 animate-spin" />
                        ) : (
                            <RefreshCw className="h-4 w-4 dark:text-shadow-glow-white" />
                        )}
                        {isControlLoading ? t("globalSetting.engine.processing") : t("globalSetting.engine.restart")}
                    </Button>
                    <Button
                        variant="destructive"
                        size="sm"
                        className="gap-1 dark:text-shadow-glow-white"
                        onClick={onForceStop}
                        disabled={!isRunning || isLoading || isControlLoading}
                    >
                        {isControlLoading ? (
                            <Loader2 className="h-4 w-4 animate-spin" />
                        ) : (
                            <X className="h-4 w-4 dark:text-shadow-glow-white" />
                        )}
                        {isControlLoading ? t("globalSetting.engine.processing") : t("globalSetting.engine.forceStop")}
                    </Button>
                    <Button
                        variant="secondary"
                        size="sm"
                        className="gap-1 dark:text-shadow-glow-white"
                        onClick={onReload}
                        disabled={!isRunning || isLoading || isControlLoading}
                    >
                        {isControlLoading ? (
                            <Loader2 className="h-4 w-4 animate-spin" />
                        ) : (
                            <RotateCw className="h-4 w-4 dark:text-shadow-glow-white" />
                        )}
                        {isControlLoading ? t("globalSetting.engine.processing") : t("globalSetting.engine.reload")}
                    </Button>
                </div>
            </CardContent>
        </Card>
    )
}