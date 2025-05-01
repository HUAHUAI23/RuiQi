// src/feature/site/components/SiteGrid.tsx
import { useRef, useEffect, useMemo } from 'react'
import { useInfiniteQuery } from '@tanstack/react-query'
import { siteApi } from '@/api/site'
import { Site, WAFMode } from '@/types/site'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card } from '@/components/ui/card'
import {
    MoreHorizontal,
    Pencil,
    Trash2,
    Shield,
    ShieldAlert,
    Server,
    Globe,
    CheckCircle,
    XCircle,
    LinkIcon,
    Loader2
} from 'lucide-react'
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger
} from '@/components/ui/dropdown-menu'
import { Skeleton } from '@/components/ui/skeleton'
import { motion, AnimatePresence } from "motion/react"
import Tilt from 'react-parallax-tilt'
import {
    gridItemAnimation,
    layoutAnimationProps
} from '@/components/ui/animation/grid-animation'
import { useTranslation } from 'react-i18next'

interface SiteGridProps {
    onEdit: (site: Site) => void
    onDelete: (id: string) => void
}

export function SiteGrid({ onEdit, onDelete }: SiteGridProps) {
    const { t } = useTranslation()

    // 引用用于无限滚动
    const sentinelRef = useRef<HTMLDivElement>(null)

    // 每页数据条数
    const PAGE_SIZE = 20

    // 获取站点列表
    const {
        data,
        isLoading,
        fetchNextPage,
        hasNextPage,
        isFetchingNextPage,
    } = useInfiniteQuery({
        queryKey: ['sites'],
        queryFn: ({ pageParam }) => siteApi.getSites(pageParam as number, PAGE_SIZE),
        initialPageParam: 1,
        getNextPageParam: (lastPage, allPages) => {
            // 优化判断逻辑：使用实际获取的数据总量，而不是假设每页恰好有PAGE_SIZE条
            const fetchedItemsCount = allPages.reduce((total, page) => total + page.items.length, 0)
            return fetchedItemsCount < lastPage.total ? allPages.length + 1 : undefined
        },
    })

    // 扁平化分页数据
    const sites = useMemo(() => {
        return data?.pages.flatMap(page => page.items) || []
    }, [data])

    // 优化的无限滚动实现
    useEffect(() => {
        // 只有当有更多页面可加载时才创建观察器
        if (!hasNextPage) return

        const options = {
            // 降低threshold使其更容易触发
            threshold: 0.1,
            // 减小rootMargin以避免过早触发，但仍保持一定的预加载空间
            rootMargin: '100px 0px'
        }

        const handleObserver = (entries: IntersectionObserverEntry[]) => {
            const [entry] = entries
            if (entry.isIntersecting && hasNextPage && !isFetchingNextPage) {
                fetchNextPage()
            }
        }

        const observer = new IntersectionObserver(handleObserver, options)

        const sentinel = sentinelRef.current
        if (sentinel) {
            observer.observe(sentinel)
        }

        return () => {
            if (sentinel) {
                observer.unobserve(sentinel)
            }
            observer.disconnect()
        }
    }, [hasNextPage, isFetchingNextPage, fetchNextPage])

    // 站点卡片组件
    const SiteCard = ({ site }: { site: Site }) => {
        const isInactive = !site.activeStatus

        return (
            <motion.div {...gridItemAnimation}>
                <Tilt
                    className="h-full"
                    tiltMaxAngleX={5}
                    tiltMaxAngleY={5}
                    perspective={1200}
                    transitionSpeed={400}
                    glareEnable={true}
                    glareMaxOpacity={0.1}
                    glareColor="#ffffff"
                    glarePosition="all"
                    scale={1.02}
                >
                    <Card
                        className={`
        p-5 rounded-md shadow-none h-full relative group transition-all duration-200 dark:card-neon
        ${isInactive || !site.wafEnabled
                                ? 'bg-gradient-to-r from-slate-100 to-white dark:from-zinc-800/60 dark:to-accent/40'
                                : site.wafMode === WAFMode.Protection
                                    ? 'bg-gradient-to-r from-green-50 to-white dark:from-green-950/20 dark:to-accent/50'
                                    : site.wafMode === WAFMode.Observation
                                        ? 'bg-gradient-to-r from-amber-50 to-white dark:from-amber-900/20 dark:to-accent/50'
                                        : 'bg-gradient-to-r from-slate-100 to-white dark:from-zinc-800/60 dark:to-accent/40'
                            }
    `}
                    >
                        <div className="flex justify-between items-start mb-4">
                            <div className={`flex flex-col ${isInactive ? 'text-gray-400 dark:text-gray-500' : 'text-slate-700 dark:text-slate-200'}`}>
                                <h3 className="font-medium text-lg dark:text-shadow-glow-white">{site.name}</h3>
                                <div className="flex items-center text-sm mt-1">
                                    <Globe className="h-3.5 w-3.5 mr-1 dark:text-shadow-glow-white" />
                                    <span className="dark:text-shadow-glow-white">{site.domain}:{site.listenPort}</span>
                                </div>
                            </div>

                            <div>
                                <DropdownMenu>
                                    <DropdownMenuTrigger asChild>
                                        <Button variant="ghost" size="icon" className="rounded-full bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 dark:button-neon">
                                            <MoreHorizontal className="h-4 w-4 dark:icon-neon" />
                                        </Button>
                                    </DropdownMenuTrigger>
                                    <DropdownMenuContent align="end">
                                        <DropdownMenuItem className="dark:text-shadow-glow-white"
                                            onClick={() => onEdit(site)}
                                        >
                                            <Pencil className="mr-2 h-4 w-4 dark:icon-neon" />
                                            {t("site.edit")}
                                        </DropdownMenuItem>
                                        <DropdownMenuItem
                                            onClick={() => onDelete(site.id)}
                                            className="text-red-600 dark:text-red-400 dark:text-shadow-glow-white"
                                        >
                                            <Trash2 className="mr-2 h-4 w-4 dark:icon-neon" />
                                            {t("site.delete")}
                                        </DropdownMenuItem>
                                    </DropdownMenuContent>
                                </DropdownMenu>
                            </div>
                        </div>

                        <div className="space-y-4">
                            {/* 状态信息 */}
                            <div className="flex flex-wrap gap-2">
                                {/* 站点状态 */}
                                {site.activeStatus ? (
                                    <Badge variant="outline" className="flex items-center gap-1 bg-green-300 border-green-300 text-green-700 dark:bg-green-900/50 dark:border-green-800 dark:text-green-300 rounded-full px-3 py-1 dark:badge-neon">
                                        <CheckCircle className="h-3 w-3 text-green-600 dark:text-green-300 dark:icon-neon" />
                                        <span className="font-medium dark:text-shadow-glow-white">{t("site.active")}</span>
                                    </Badge>
                                ) : (
                                    <Badge variant="outline" className="flex items-center gap-1 bg-gray-200 border-gray-200 text-gray-700 dark:bg-gray-800/70 dark:border-gray-700 dark:text-gray-300 rounded-full px-3 py-1 dark:badge-neon">
                                        <XCircle className="h-3 w-3 text-gray-600 dark:text-gray-300 dark:icon-neon" />
                                        <span className="font-medium dark:text-shadow-glow-white">{t("site.inactive")}</span>
                                    </Badge>
                                )}

                                {/* HTTPS状态 */}
                                {site.enableHTTPS ? (
                                    <Badge variant="outline" className={`flex items-center gap-1 rounded-full px-3 py-1 dark:badge-neon ${isInactive
                                        ? 'bg-gray-200 border-gray-200 text-gray-700 dark:bg-gray-800/70 dark:border-gray-700 dark:text-gray-300'
                                        : 'bg-blue-50 border-blue-200 text-blue-700 dark:bg-blue-900/30 dark:border-blue-800/60 dark:text-blue-300'
                                        }`}>
                                        <LinkIcon className="h-3 w-3 dark:icon-neon" />
                                        <span className="font-medium dark:text-shadow-glow-white">HTTPS</span>
                                    </Badge>
                                ) : null}

                                {/* WAF状态 */}
                                {site.wafEnabled && (
                                    site.wafMode === WAFMode.Protection ? (
                                        <Badge variant="outline" className={`flex items-center gap-1 rounded-full px-3 py-1 dark:badge-neon ${isInactive
                                            ? 'bg-gray-200 border-gray-200 text-gray-700 dark:bg-gray-800/70 dark:border-gray-700 dark:text-gray-300'
                                            : 'bg-sky-300 border-sky-300 text-sky-700 dark:bg-sky-900/40 dark:border-sky-800/70 dark:text-sky-300'
                                            }`}>
                                            <Shield className="h-3 w-3 text-sky-700 dark:text-sky-300 dark:icon-neon" />
                                            <span className="font-medium dark:text-shadow-glow-white">{t("site.dialog.protectionMode")}</span>
                                        </Badge>

                                    ) : (
                                        <Badge variant="outline" className={`flex items-center gap-1 rounded-full px-3 py-1 dark:badge-neon ${isInactive
                                            ? 'bg-gray-200 border-gray-200 text-gray-700 dark:bg-gray-800/70 dark:border-gray-700 dark:text-gray-300'
                                            : 'bg-yellow-300 border-yellow-300 text-yellow-700 dark:bg-yellow-900/40 dark:border-yellow-800/70 dark:text-yellow-300'
                                            }`}>
                                            <ShieldAlert className="h-3 w-3 text-yellow-700 dark:text-yellow-300 dark:icon-neon" />
                                            <span className="font-medium dark:text-shadow-glow-white">{t("site.dialog.observationMode")}</span>
                                        </Badge>
                                    )
                                )}
                            </div>

                            {/* 上游服务器信息 */}
                            <div className={`space-y-1 ${isInactive ? 'text-gray-400 dark:text-gray-500' : 'text-slate-700 dark:text-slate-300'}`}>
                                <div className="text-sm font-medium dark:text-shadow-glow-white">{t("site.card.upstreamServers")}</div>
                                <div className="space-y-1">
                                    {site.backend.servers.map((server, index) => (
                                        <div key={index} className="flex items-center gap-1 text-xs pl-2">
                                            <Server className="h-3 w-3 dark:icon-neon" />
                                            <span className="dark:text-shadow-glow-white">
                                                {server.isSSL ? 'https://' : 'http://'}
                                                {server.host}:{server.port}
                                            </span>
                                        </div>
                                    ))}
                                </div>
                            </div>

                            {/* 如果有证书，显示证书信息 */}
                            {site.enableHTTPS && site.certificate && (
                                <div className={`space-y-1 ${isInactive ? 'text-gray-400 dark:text-gray-500' : 'text-slate-700 dark:text-slate-300'}`}>
                                    <div className="text-sm font-medium dark:text-shadow-glow-white">{t("site.card.certInfo")}</div>
                                    <div className="text-xs pl-2">
                                        <span className="dark:text-shadow-glow-white">{site.certificate.certName}</span>
                                        <span className="text-muted-foreground ml-2 dark:text-shadow-glow-white">
                                            ({site.certificate.issuerName})
                                        </span>
                                    </div>
                                </div>
                            )}
                        </div>
                    </Card>
                </Tilt>
            </motion.div>
        )
    }

    // 卡片加载骨架屏
    const SiteCardSkeleton = () => (
        <Card className="p-5 rounded-md dark:border-neon-pulse">
            <div className="flex justify-between items-start mb-4">
                <div>
                    <Skeleton className="h-6 w-32 mb-2 dark:text-shadow-glow-white" />
                    <Skeleton className="h-4 w-48 dark:text-shadow-glow-white" />
                </div>
                <Skeleton className="h-8 w-8 rounded-full dark:text-shadow-glow-white" />
            </div>

            <div className="space-y-4">
                <div className="flex flex-wrap gap-2">
                    <Skeleton className="h-6 w-16 dark:text-shadow-glow-white" />
                    <Skeleton className="h-6 w-16 dark:text-shadow-glow-white" />
                    <Skeleton className="h-6 w-16 dark:text-shadow-glow-white" />
                </div>

                <div>
                    <Skeleton className="h-4 w-24 mb-2 dark:text-shadow-glow-white" />
                    <Skeleton className="h-4 w-full mb-1 dark:text-shadow-glow-white" />
                    <Skeleton className="h-4 w-full dark:text-shadow-glow-white" />
                </div>
            </div>
        </Card>
    )

    // 显示加载状态
    if (isLoading) {
        return (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {Array(6).fill(0).map((_, index) => (
                    <SiteCardSkeleton key={index} />
                ))}
            </div>
        )
    }

    return (
        <div>
            {sites.length === 0 ? (
                <div className="text-center py-10 text-muted-foreground dark:text-shadow-glow-white">
                    {t("site.noData")}
                </div>
            ) : (
                <motion.div
                    className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4"
                    layout // 启用自动布局动画
                >
                    <AnimatePresence mode="popLayout"> {/* 使用AnimatePresence处理元素的添加/删除动画 */}
                        {sites.map(site => (
                            <motion.div
                                key={site.id}
                                layoutId={`site-card-${site.id}`} // 使用唯一ID跟踪元素位置
                                {...layoutAnimationProps} // 使用布局动画配置
                                whileInView={{
                                    opacity: 1,
                                    scale: 1,
                                    y: 0,
                                    transition: {
                                        type: "spring",
                                        damping: 20,
                                        stiffness: 250
                                    }
                                }}
                                viewport={{
                                    once: false,
                                    margin: "-5% 0px"
                                }}
                                className="h-full"
                            >
                                <SiteCard site={site} />
                            </motion.div>
                        ))}
                    </AnimatePresence>
                </motion.div>
            )}

            {/* 无限滚动监测元素，只在有更多数据时显示 */}
            {hasNextPage && (
                <div
                    ref={sentinelRef}
                    className="h-5 flex justify-center items-center mt-4"
                >
                    {isFetchingNextPage && (
                        <Loader2 className="h-4 w-4 animate-spin text-muted-foreground dark:icon-neon" />
                    )}
                </div>
            )}
        </div>
    )
}