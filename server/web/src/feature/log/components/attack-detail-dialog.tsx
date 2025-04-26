import { useState } from "react"
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { Copy, Check, AlertTriangle, Shield } from "lucide-react"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import type { AttackDetailData } from "@/types/log"
import { format } from "date-fns"
import { Badge } from "@/components/ui/badge"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { AnimatePresence, motion } from "motion/react"
import {
    dialogEnterExitAnimation,
    dialogContentAnimation,
    dialogContentItemAnimation,
} from "@/components/ui/animation/dialog-animation"
import { useTranslation } from "react-i18next"
import { CopyableText } from "@/components/common/copyable-text"

// Add global scrollbar styles
import "./scrollbar.css"
import { TabsAnimationProvider } from "@/components/ui/animation/components/tab-animation"

interface AttackDetailDialogProps {
    open: boolean
    onOpenChange: (open: boolean) => void
    data: AttackDetailData | null
}

export function AttackDetailDialog({ open, onOpenChange, data }: AttackDetailDialogProps) {
    const [copyState, setCopyState] = useState<{ [key: string]: boolean }>({})
    const [encoding, setEncoding] = useState("UTF-8")
    const [activeTab, setActiveTab] = useState("request")
    const { t } = useTranslation()

    // Define the gradient for reuse
    const purpleGradient = `linear-gradient(135deg, 
    rgba(147, 112, 219, 0.95) 0%, 
    rgba(138, 100, 208, 0.9) 50%, 
    rgba(123, 79, 214, 0.95) 100%)`

    const handleCopy = (text: string, key: string) => {
        navigator.clipboard.writeText(text).then(() => {
            setCopyState((prev) => ({ ...prev, [key]: true }))
            setTimeout(() => setCopyState((prev) => ({ ...prev, [key]: false })), 2000)
        })
    }

    if (!data) return null

    // 构建curl命令
    const curlCommand = `curl -X GET "${data.target}"`

    // 为了演示，假设规则ID > 1000 的是高危规则
    const isHighRisk = data.ruleId > 1000

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <AnimatePresence>
                {open && (
                    <motion.div {...dialogEnterExitAnimation}>
                        <DialogContent className="sm:max-w-[90vw] lg:max-w-[75vw] xl:max-w-[65vw] max-h-[90vh] w-full p-0 gap-0 overflow-hidden">
                            <motion.div {...dialogContentAnimation}>
                                <DialogHeader
                                    className="px-6 py-4"
                                    style={{
                                        background: purpleGradient,
                                    }}
                                >
                                    <motion.div {...dialogContentItemAnimation}>
                                        <div className="flex items-center gap-2">
                                            <DialogTitle className="text-xl font-semibold flex items-center gap-2 text-white">
                                                {isHighRisk && <AlertTriangle className="h-5 w-5 text-yellow-300" />}
                                                {t("attackDetail.title")}
                                            </DialogTitle>
                                            {isHighRisk && (
                                                <Badge variant="destructive" className="ml-2 bg-purple-300 text-purple-800 hover:bg-purple-300">
                                                    {t("attackDetail.highRiskAttack")}
                                                </Badge>
                                            )}
                                        </div>
                                    </motion.div>
                                </DialogHeader>

                                <ScrollArea className="px-4 py-2 h-[calc(90vh-6rem)] scrollbar-thin scrollbar-thumb-purple-300 scrollbar-track-transparent hover:scrollbar-thumb-purple-400 transition-colors duration-200">
                                    <div className="space-y-2 p-0 max-w-full max-h-full">
                                        {/* 攻击概述卡片 */}
                                        <motion.div {...dialogContentItemAnimation}>
                                            <Card className="p-6 bg-card border-none shadow-none rounded-sm bg-gradient-to-r from-purple-50 to-white">
                                                <h3 className="text-lg font-semibold mb-4 flex items-center gap-2 text-card-foreground">
                                                    <Shield className="h-5 w-5 text-purple-600" />
                                                    {t("attackDetail.overview")}
                                                </h3>
                                                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                                                    <div className="space-y-4">
                                                        <div>
                                                            <span className="text-muted-foreground text-sm block mb-1">
                                                                {t("attackDetail.target")}
                                                            </span>
                                                            <div className="font-medium truncate text-card-foreground">
                                                                <CopyableText text={data.target} className="font-medium text-card-foreground" />
                                                            </div>
                                                        </div>
                                                        <div>
                                                            <span className="text-muted-foreground text-sm block mb-1">
                                                                {t("attackDetail.message")}
                                                            </span>
                                                            <div className="font-medium text-card-foreground">{data.message}</div>
                                                        </div>
                                                        <div>
                                                            <span className="text-muted-foreground text-sm block mb-1">{t("requestId")}</span>
                                                            <div className="font-mono text-sm flex items-center gap-1 text-card-foreground">
                                                                {data.requestId}
                                                                <Button
                                                                    variant="ghost"
                                                                    size="icon"
                                                                    className="h-6 w-6 text-muted-foreground hover:text-card-foreground"
                                                                    onClick={() => handleCopy(data.requestId, "requestId")}
                                                                >
                                                                    {copyState["requestId"] ? (
                                                                        <Check className="h-3 w-3" />
                                                                    ) : (
                                                                        <Copy className="h-3 w-3" />
                                                                    )}
                                                                </Button>
                                                            </div>
                                                        </div>
                                                    </div>
                                                    <div className="space-y-4">
                                                        <div>
                                                            <span className="text-muted-foreground text-sm block mb-1">{t("ruleId")}</span>
                                                            <div className="font-medium flex items-center gap-2 text-card-foreground">
                                                                {data.ruleId}
                                                                {/* <Button
                                                                    variant="outline"
                                                                    size="sm"
                                                                    className="h-7 text-xs border-border hover:bg-accent"
                                                                >
                                                                    {t("attackDetail.viewRuleDetail")}
                                                                    <ArrowUpRight className="h-3 w-3 ml-1" />
                                                                </Button> */}
                                                            </div>
                                                        </div>
                                                        <div>
                                                            <span className="text-muted-foreground text-sm block mb-1">
                                                                {t("attackDetail.attackTime")}
                                                            </span>
                                                            <div className="font-medium text-card-foreground">
                                                                {format(new Date(data.createdAt), "yyyy-MM-dd HH:mm:ss")}
                                                            </div>
                                                        </div>
                                                    </div>
                                                </div>
                                            </Card>
                                        </motion.div>

                                        {/* 载荷信息 */}
                                        <motion.div {...dialogContentItemAnimation}>
                                            <Card className="p-6 bg-card border-none shadow-none">
                                                <h3 className="text-lg font-semibold mb-4 text-card-foreground">
                                                    {t("attackDetail.detectedPayload")}
                                                </h3>
                                                <div className="bg-muted rounded-md p-4 border-none bg-zinc-100">
                                                    <div className="flex items-center justify-between">
                                                        <code className="text-sm break-all font-mono text-card-foreground whitespace-pre-wrap break-words block w-full overflow-hidden">
                                                            {data.payload}
                                                        </code>
                                                        <Button
                                                            variant="ghost"
                                                            size="icon"
                                                            onClick={() => handleCopy(data.payload, "payload")}
                                                            className="text-muted-foreground hover:text-card-foreground"
                                                        >
                                                            {copyState["payload"] ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
                                                        </Button>
                                                    </div>
                                                </div>
                                            </Card>
                                        </motion.div>

                                        {/* 来源和目标信息 */}
                                        <motion.div {...dialogContentItemAnimation}>
                                            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                                                {/* 攻击来源 */}
                                                <Card className="p-6 bg-card border-none shadow-none bg-gradient-to-r from-purple-100 to-white rounded-sm">
                                                    <h3 className="text-lg font-semibold mb-4 text-card-foreground">
                                                        {t("attackDetail.attackSource")}
                                                    </h3>
                                                    <div className="space-y-4">
                                                        <div>
                                                            <span className="text-muted-foreground text-sm block mb-1">{t("srcIp")}</span>
                                                            <div className="font-medium flex items-center justify-between text-card-foreground">
                                                                <span className="break-all font-mono">{data.srcIp}</span>
                                                                {/* <Button
                                                                    variant="destructive"
                                                                    size="sm"
                                                                    className="h-7 text-xs bg-destructive text-destructive-foreground hover:bg-destructive/90"
                                                                >
                                                                    {t("attackDetail.blockThisIp")}
                                                                </Button> */}
                                                            </div>
                                                        </div>
                                                        <div>
                                                            <span className="text-muted-foreground text-sm block mb-1">{t("srcPort")}</span>
                                                            <div className="font-medium font-mono text-card-foreground">{data.srcPort}</div>
                                                        </div>
                                                    </div>
                                                </Card>

                                                {/* 目标信息 */}
                                                <Card className="p-6 bg-card border-none shadow-none bg-gradient-to-r from-purple-100 to-white rounded-sm">
                                                    <h3 className="text-lg font-semibold mb-4 text-card-foreground">
                                                        {t("attackDetail.targetInfo")}
                                                    </h3>
                                                    <div className="space-y-4">
                                                        <div>
                                                            <span className="text-muted-foreground text-sm block mb-1">{t("dstIp")}</span>
                                                            <div className="font-medium font-mono break-all text-card-foreground">{data.dstIp}</div>
                                                        </div>
                                                        <div>
                                                            <span className="text-muted-foreground text-sm block mb-1">{t("dstPort")}</span>
                                                            <div className="font-medium font-mono text-card-foreground">{data.dstPort}</div>
                                                        </div>
                                                    </div>
                                                </Card>
                                            </div>
                                        </motion.div>

                                        {/* 请求详情选项卡 */}
                                        <motion.div {...dialogContentItemAnimation}>
                                            <Card className="p-6 bg-card border-none shadow-none rounded-sm">
                                                <Tabs defaultValue="request" className="w-full" onValueChange={(value) => setActiveTab(value)}>
                                                    <div className="flex justify-between items-center mb-4">
                                                        <h3 className="text-lg font-semibold text-card-foreground">
                                                            {t("attackDetail.technicalDetails")}
                                                        </h3>
                                                        <div className="flex items-center gap-2">
                                                            <Button
                                                                variant="outline"
                                                                size="sm"
                                                                onClick={() => handleCopy(curlCommand, "curl")}
                                                                className="flex items-center gap-1 h-8 border-border hover:bg-purple-100"
                                                            >
                                                                {copyState["curl"] ? <Check className="h-3 w-3" /> : <Copy className="h-3 w-3" />}
                                                                {t("attackDetail.copyCurl")}
                                                            </Button>

                                                            <Select value={encoding} onValueChange={setEncoding}>
                                                                <SelectTrigger className="w-[110px] h-8 border-purple-200">
                                                                    <SelectValue placeholder={t("attackDetail.encoding")} />
                                                                </SelectTrigger>
                                                                <SelectContent>
                                                                    <SelectItem value="UTF-8">UTF-8</SelectItem>
                                                                    <SelectItem value="GBK">GBK</SelectItem>
                                                                    <SelectItem value="ISO-8859-1">ISO-8859-1</SelectItem>
                                                                </SelectContent>
                                                            </Select>
                                                        </div>
                                                    </div>

                                                    <TabsList className="mb-3 w-full bg-purple-100 transition-all duration-300 ease-in-out">
                                                        <TabsTrigger
                                                            value="request"
                                                            className="flex-1 transition-all duration-300 ease-in-out transform hover:scale-[1.02] text-purple-800"
                                                            style={{
                                                                background: activeTab === "request" ? purpleGradient : "transparent",
                                                                color: activeTab === "request" ? "white" : "inherit",
                                                            }}
                                                        >
                                                            {t("attackDetail.request")}
                                                        </TabsTrigger>
                                                        <TabsTrigger
                                                            value="response"
                                                            className="flex-1 transition-all duration-300 ease-in-out transform hover:scale-[1.02] text-purple-800"
                                                            style={{
                                                                background: activeTab === "response" ? purpleGradient : "transparent",
                                                                color: activeTab === "response" ? "white" : "inherit",
                                                            }}
                                                        >
                                                            {t("attackDetail.response")}
                                                        </TabsTrigger>
                                                        <TabsTrigger
                                                            value="logs"
                                                            className="flex-1 transition-all duration-300 ease-in-out transform hover:scale-[1.02] text-purple-800"
                                                            style={{
                                                                background: activeTab === "logs" ? purpleGradient : "transparent",
                                                                color: activeTab === "logs" ? "white" : "inherit",
                                                            }}
                                                        >
                                                            {t("attackDetail.logs")}
                                                        </TabsTrigger>
                                                    </TabsList>

                                                    <TabsAnimationProvider currentView={activeTab} animationVariant="slide">
                                                        {activeTab === "request" ? (
                                                            <TabsContent value="request" forceMount>
                                                                <div className="border rounded-md overflow-hidden bg-muted/10 border-border">
                                                                    <div className="flex justify-end p-2 bg-muted border-b border-border">
                                                                        <Button
                                                                            variant="ghost"
                                                                            size="sm"
                                                                            className="h-7 text-muted-foreground hover:text-card-foreground"
                                                                            onClick={() => handleCopy(data.request, "requestCopy")}
                                                                        >
                                                                            {copyState["requestCopy"] ? (
                                                                                <Check className="h-3 w-3 mr-1" />
                                                                            ) : (
                                                                                <Copy className="h-3 w-3 mr-1" />
                                                                            )}
                                                                            {t("attackDetail.copyAll")}
                                                                        </Button>
                                                                    </div>
                                                                    <ScrollArea className="scrollbar-thin scrollbar-thumb-purple-300 scrollbar-track-transparent hover:scrollbar-thumb-purple-400 transition-colors duration-200 max-h-[300px]">
                                                                        <pre className="p-4 text-sm whitespace-pre-wrap font-mono text-card-foreground bg-background">
                                                                            <code className="text-sm break-all font-mono text-card-foreground whitespace-pre-wrap break-words block w-full overflow-hidden">
                                                                                {data.request}
                                                                            </code>
                                                                        </pre>
                                                                    </ScrollArea>
                                                                </div>
                                                            </TabsContent>
                                                        ) : activeTab === "response" ? (
                                                            <TabsContent value="response" forceMount>
                                                                <div className="border rounded-md overflow-hidden bg-muted/10 border-border">
                                                                    <div className="flex justify-end p-2 bg-muted border-b border-border">
                                                                        <Button
                                                                            variant="ghost"
                                                                            size="sm"
                                                                            className="h-7 text-muted-foreground hover:text-card-foreground"
                                                                            onClick={() => handleCopy(data.response, "responseCopy")}
                                                                        >
                                                                            {copyState["responseCopy"] ? (
                                                                                <Check className="h-3 w-3 mr-1" />
                                                                            ) : (
                                                                                <Copy className="h-3 w-3 mr-1" />
                                                                            )}
                                                                            {t("attackDetail.copyAll")}
                                                                        </Button>
                                                                    </div>
                                                                    <ScrollArea className="scrollbar-thin scrollbar-thumb-purple-300 scrollbar-track-transparent hover:scrollbar-thumb-purple-400 transition-colors duration-200 max-h-[300px]">
                                                                        <pre className="p-4 text-sm whitespace-pre-wrap font-mono text-card-foreground bg-background">
                                                                            <code className="text-sm break-all font-mono text-card-foreground whitespace-pre-wrap break-words block w-full overflow-hidden">
                                                                                {data.response ? data.response : t("attackDetail.noResponse")}
                                                                            </code>
                                                                        </pre>
                                                                    </ScrollArea>
                                                                </div>
                                                            </TabsContent>
                                                        ) : (
                                                            <TabsContent value="logs" forceMount>
                                                                <div className="border rounded-md overflow-hidden bg-muted/10 border-border">
                                                                    <div className="flex justify-end p-2 bg-muted border-b border-border">
                                                                        <Button
                                                                            variant="ghost"
                                                                            size="sm"
                                                                            className="h-7 text-muted-foreground hover:text-card-foreground"
                                                                            onClick={() => handleCopy(data.logs, "logsCopy")}
                                                                        >
                                                                            {copyState["logsCopy"] ? (
                                                                                <Check className="h-3 w-3 mr-1" />
                                                                            ) : (
                                                                                <Copy className="h-3 w-3 mr-1" />
                                                                            )}
                                                                            {t("attackDetail.copyAll")}
                                                                        </Button>
                                                                    </div>
                                                                    <ScrollArea className="scrollbar-thin scrollbar-thumb-purple-300 scrollbar-track-transparent hover:scrollbar-thumb-purple-400 transition-colors duration-200 max-h-[300px]">
                                                                        <pre className="p-4 text-sm whitespace-pre-wrap font-mono text-card-foreground bg-background">
                                                                            <code className="text-sm break-all font-mono text-card-foreground whitespace-pre-wrap break-words block w-full overflow-hidden">
                                                                                {data.logs}
                                                                            </code>
                                                                        </pre>
                                                                    </ScrollArea>
                                                                </div>
                                                            </TabsContent>
                                                        )}
                                                    </TabsAnimationProvider>
                                                </Tabs>
                                            </Card>
                                        </motion.div>
                                    </div>
                                </ScrollArea>
                            </motion.div>
                        </DialogContent>
                    </motion.div>
                )}
            </AnimatePresence>
        </Dialog>
    )
}
