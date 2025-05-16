import { Activity, BarChart2, ArrowUp, ArrowDown, Shield, AlertCircle, Network, Server } from "lucide-react"
import { StatsCard } from "@/feature/stats/components/StatsCard"
import { TimeRangeSelector } from "@/feature/stats/components/TimeRangeSelector"
import { RealtimeQPSChart } from "@/feature/stats/components/charts/RealtimeQPSChart"
import { TrafficChart } from "@/feature/stats/components/charts/TrafficChart"
import { RequestsBlocksChart } from "@/feature/stats/components/charts/RequestsBlocksChart"
import { useTimeRangeSelector, useOverviewStats } from "@/feature/stats/hooks/useStats"
import { useTranslation } from "react-i18next"
import { Separator } from "@/components/ui/separator"
import { Card } from "@/components/ui/card"
import { ScrollArea } from "@/components/ui/scroll-area"

export default function StatsPage() {
    const { t } = useTranslation()
    const { timeRange, setTimeRange } = useTimeRangeSelector('24h')
    const { data: statsData, isLoading } = useOverviewStats(timeRange)

    return (
        <ScrollArea scrollbarVariant="none" className="h-full mx-auto p-4 space-y-6">
            {/* 时间范围选择器 */}
            <div className="flex justify-between items-center">
                <h1 className="text-2xl font-bold text-primary dark:text-white">
                    {t('stats.title')}
                </h1>
                <TimeRangeSelector value={timeRange} onChange={setTimeRange} />
            </div>

            <Separator className="my-4" />

            <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
                {/* 左侧指标区域 */}
                <div className="lg:col-span-7">
                    <Card className="border-none shadow-none p-4">
                        <h2 className="text-lg font-medium mb-4 dark:text-white">{t('stats.overview')}</h2>
                        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                            <StatsCard
                                title={t('stats.totalRequests')}
                                value={statsData?.totalRequests || 0}
                                icon={<Activity size={16} />}
                                loading={isLoading}
                            />
                            <StatsCard
                                title={t('stats.inboundTraffic')}
                                value={statsData?.inboundTraffic || 0}
                                icon={<ArrowDown size={16} />}
                                loading={isLoading}
                                isTraffic
                            />
                            <StatsCard
                                title={t('stats.outboundTraffic')}
                                value={statsData?.outboundTraffic || 0}
                                icon={<ArrowUp size={16} />}
                                loading={isLoading}
                                isTraffic
                            />
                            <StatsCard
                                title={t('stats.maxQPS')}
                                value={statsData?.maxQPS || 0}
                                icon={<BarChart2 size={16} />}
                                loading={isLoading}
                            />
                            <StatsCard
                                title={t('stats.4xxErrors')}
                                value={statsData?.error4xx || 0}
                                icon={<AlertCircle size={16} />}
                                change={`${statsData?.error4xxRate.toFixed(2)}%`}
                                loading={isLoading}
                            />
                            <StatsCard
                                title={t('stats.5xxErrors')}
                                value={statsData?.error5xx || 0}
                                icon={<Server size={16} />}
                                change={`${statsData?.error5xxRate.toFixed(2)}%`}
                                loading={isLoading}
                            />
                            <StatsCard
                                title={t('stats.blockCount')}
                                value={statsData?.blockCount || 0}
                                icon={<Shield size={16} />}
                                loading={isLoading}
                                link={'/logs/protect'}
                            />
                            <StatsCard
                                title={t('stats.attackIPCount')}
                                value={statsData?.attackIPCount || 0}
                                icon={<Network size={16} />}
                                loading={isLoading}
                                link={'/logs/attack'}
                            />
                        </div>
                    </Card>
                </div>

                {/* 右侧实时QPS区域 */}
                <div className="lg:col-span-5">
                    <RealtimeQPSChart />
                </div>
            </div>

            {/* 中央区域 - 流量趋势图 */}
            <TrafficChart timeRange={timeRange} />

            {/* 底部区域 - 请求与拦截折线图 */}
            <RequestsBlocksChart timeRange={timeRange} />
        </ScrollArea>
    )
}