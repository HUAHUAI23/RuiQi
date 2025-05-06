import { useEffect } from "react"
import { useConfigQuery } from "@/feature/global-setting/hooks/useConfig"
import { useRunnerStatusQuery, useRunnerControl } from "@/feature/global-setting/hooks/useRunner"
import { EngineStatus } from "@/feature/global-setting/components/EngineStatus"
import { ConfigForm } from "@/feature/global-setting/components/ConfigForm"
import { Settings, Info } from "lucide-react"
import { AdvancedErrorDisplay } from "@/components/common/error/errorDisplay"
import { AnimatedContainer } from "@/components/ui/animation/components/animated-container"
import { useTranslation } from "react-i18next"

export default function GlobalSettingPage() {
    const { t } = useTranslation()

    // 获取配置数据
    const { config, isLoading: isConfigLoading, error: configError, refetch: refetchConfig } = useConfigQuery()

    // 获取运行器状态
    const { status, isLoading: isStatusLoading, error: statusError, refetch: refetchStatus } = useRunnerStatusQuery()

    // 运行器控制
    const {
        controlRunner,
        isLoading: isControlLoading,
        error: controlError,
        clearError: clearControlError,
    } = useRunnerControl()

    // 当页面加载时获取最新配置
    useEffect(() => {
        // 页面加载时，获取最新配置和状态
        refetchConfig()
        refetchStatus()
    }, [refetchConfig, refetchStatus])

    // 运行器控制处理函数
    const handleStart = () => {
        clearControlError()
        controlRunner("start")
    }
    const handleStop = () => {
        clearControlError()
        controlRunner("stop")
    }
    const handleRestart = () => {
        clearControlError()
        controlRunner("restart")
    }
    const handleForceStop = () => {
        clearControlError()
        controlRunner("force_stop")
    }
    const handleReload = () => {
        clearControlError()
        controlRunner("reload")
    }

    // 根据错误类型选择适当的重试函数
    const handleRetry = () => {
        if (configError) refetchConfig()
        if (statusError) refetchStatus()
    }

    return (
        <AnimatedContainer variant="smooth" className="h-full overflow-y-auto scrollbar-none p-0">
            <div className="max-w-5xl p-6 center mx-auto">
                {/* 错误处理：优先显示配置错误，其次显示状态错误 */}
                {(configError || statusError) && (
                    <AdvancedErrorDisplay error={configError || statusError} onRetry={handleRetry} />
                )}
                {controlError && <AdvancedErrorDisplay error={controlError} />}

                {/* 通用设置 */}
                <div className="bg-background rounded-xl border-none p-6 mb-6 animate-fade-in-up transition-colors duration-200">
                    <div className="space-y-3 mb-8">
                        <div className="flex items-center gap-2 pb-2 border-b border-border">
                            <Settings className="w-5 h-5 text-iconStroke dark:text-primary dark:text-shadow-glow-white" />
                            <h3 className="text-lg font-medium text-foreground text-shadow-glow-white dark:text-shadow-glow-white">{t("globalSetting.config.generalConfig")}</h3>
                        </div>
                        <div className="pl-7">
                            <p className="text-sm text-muted-foreground text-shadow-glow-white dark:text-shadow-primary-bold">{t("globalSetting.description")}</p>
                        </div>
                    </div>

                    {/* 引擎状态 */}
                    <div className="space-y-3 mb-8">
                        <div className="flex items-center gap-2 pb-2 border-b border-border">
                            <Info className="w-5 h-5 text-iconStroke dark:text-primary dark:text-shadow-glow-white" />
                            <h3 className="text-base font-medium text-foreground text-shadow-glow-white dark:text-shadow-glow-white">{t("globalSetting.engine.management")}</h3>
                        </div>
                        <div className="pl-7">
                            <EngineStatus
                                status={status}
                                isLoading={isStatusLoading}
                                onStart={handleStart}
                                onStop={handleStop}
                                onRestart={handleRestart}
                                onForceStop={handleForceStop}
                                onReload={handleReload}
                                isControlLoading={isControlLoading}
                            />
                        </div>
                    </div>
                </div>

                {/* 配置表单 */}
                <div
                    className="bg-background rounded-xl border-none p-6 mb-6 animate-fade-in-up"
                    style={{ animationDelay: "0.1s" }}
                >
                    <ConfigForm config={config} isLoading={isConfigLoading} />
                </div>
            </div>
        </AnimatedContainer>
    )
}
