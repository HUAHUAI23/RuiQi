"use client"

import { Link, useLocation, useNavigate } from "react-router"
import { cn } from "@/lib/utils"
import { Settings, Shield, BarChart2, FileText, LogOut } from "lucide-react"
import { ROUTES } from "@/routes/constants"
import { useTranslation } from "react-i18next"
import type { TFunction } from "i18next"
import { useAuthStore } from "@/store/auth"
import { useState } from "react"

// Create sidebar config with display options
function createSidebarConfig(t: TFunction) {
    return [
        {
            title: t("sidebar.monitor"),
            icon: BarChart2,
            href: ROUTES.MONITOR,
            display: true,
        },
        {
            title: t("sidebar.logs"),
            icon: FileText,
            href: ROUTES.LOGS,
            display: true,
        },
        {
            title: t("sidebar.rules"),
            icon: Shield,
            href: ROUTES.RULES,
            display: true,
        },
        {
            title: t("sidebar.settings"),
            icon: Settings,
            href: ROUTES.SETTINGS,
            display: true,
        },
    ] as const
}

interface SidebarDisplayConfig {
    monitor?: boolean
    logs?: boolean
    rules?: boolean
    settings?: boolean
}

interface SidebarProps {
    displayConfig?: SidebarDisplayConfig
}

export function Sidebar({ displayConfig = {} }: SidebarProps) {
    const location = useLocation()
    const { t } = useTranslation()
    const navigate = useNavigate()
    const { logout } = useAuthStore()
    const [isLogoutActive, setIsLogoutActive] = useState(false)

    // Get current first level path
    const currentFirstLevelPath = "/" + location.pathname.split("/")[1]

    // Generate sidebar items with display config
    const sidebarItems = createSidebarConfig(t).map((item) => {
        // Determine which config property based on path
        let configKey: keyof SidebarDisplayConfig = "monitor"
        if (item.href === ROUTES.LOGS) configKey = "logs"
        if (item.href === ROUTES.RULES) configKey = "rules"
        if (item.href === ROUTES.SETTINGS) configKey = "settings"

        // Use config value or default
        const shouldDisplay = displayConfig[configKey] !== undefined ? displayConfig[configKey] : item.display

        return {
            ...item,
            display: shouldDisplay,
        }
    })

    const handleLogout = () => {
        setIsLogoutActive(true)

        // Visual feedback before actual logout
        setTimeout(() => {
            logout()
            navigate("/login")
        }, 300)
    }

    return (
        <div
            className="w-64 text-white flex flex-col border-r border-slate-200 dark:border-none relative overflow-hidden transition-all duration-300 bg-sidebar-gradient"
        >
            {/* 霓虹灯效果 暗色模式 */}
            <div className="absolute inset-0 dark:animate-sidebar-neon-glow pointer-events-none"></div>

            {/* Decorative background elements */}
            <div className="absolute bottom-0 left-0 w-full h-48 overflow-hidden opacity-20 dark:opacity-15 pointer-events-none">
                <div className="absolute bottom-[-10px] left-[-10px] w-20 h-20 bg-white/30 rotate-45 transform animate-float"></div>
                <div className="absolute bottom-[-5px] left-[40px] w-12 h-12 bg-white/20 rotate-12 transform animate-float-reverse"></div>
                <div className="absolute bottom-[30px] left-[80px] w-16 h-16 bg-white/25 rotate-30 transform animate-float"></div>
                <div className="absolute bottom-[10px] left-[120px] w-24 h-24 bg-white/15 rotate-20 transform animate-float-reverse"></div>
                <div className="absolute bottom-[40px] left-[180px] w-14 h-14 bg-white/20 rotate-45 transform animate-float"></div>
                <div className="absolute bottom-[-20px] left-[220px] w-20 h-20 bg-white/10 rotate-30 transform animate-float-reverse"></div>
            </div>

            {/* Logo and title */}
            <div className="flex flex-col items-center gap-2 py-6 border-b border-white/10 dark:border-white/5">
                {/* <div className="w-16 h-16 rounded-full bg-gradient-to-br from-[#A48BEA] to-[#8861DB] dark:from-[#9470DB] dark:to-[#7B4FD6] flex items-center justify-center shadow-lg animate-pulse-glow"> */}
                <div className="w-16 h-16 rounded-full flex items-center justify-center animate-pulse-glow">
                    {/* <Shield className="w-8 h-8 text-white" /> */}
                    <img src="/logo.svg" alt="logo" />
                </div>
                <div className="font-bold text-xl mt-2">
                    <span className="text-[#E8DFFF] dark:text-[#F0EBFF] text-shadow-glow-purple transition-all duration-300">RuiQi</span>
                    <span className="text-[#8ED4FF] dark:text-[#A5DEFF] text-shadow-glow-blue transition-all duration-300"> WAF</span>
                </div>
            </div>

            {/* Navigation items */}
            <div className="flex-1 py-4">
                {sidebarItems
                    .filter((item) => item.display)
                    .map((item) => {
                        const isActive = currentFirstLevelPath === item.href
                        return (
                            <Link
                                key={item.href}
                                to={item.href}
                                className={cn(
                                    "flex items-center gap-3 font-medium px-6 py-3 w-full group transition-all duration-300 relative overflow-hidden",
                                    isActive
                                        ? "bg-white/15 dark:bg-white/10 shadow-[0_0_15px_rgba(255,255,255,0.4)] dark:shadow-[0_0_15px_rgba(255,255,255,0.25)] text-white translate-x-1"
                                        : "text-white/90 hover:text-white hover:translate-x-1 hover:shadow-[0_0_10px_rgba(255,255,255,0.3)] dark:hover:shadow-[0_0_12px_rgba(255,255,255,0.2)]",
                                    "before:absolute before:content-[''] before:top-0 before:left-0 before:w-full before:h-full before:bg-gradient-to-r before:from-white/5 before:to-white/20 dark:before:from-white/5 dark:before:to-white/15 before:transition-opacity before:duration-300",
                                    isActive
                                        ? "before:opacity-100"
                                        : "before:opacity-0 hover:before:opacity-100"
                                )}
                            >
                                <span className="relative z-10 flex items-center gap-3">
                                    <item.icon className={cn(
                                        "w-5 h-5 transition-transform",
                                        isActive ? "text-white" : "group-hover:animate-icon-shake"
                                    )} />
                                    <span className={cn(
                                        "transition-all dark:text-shadow-glow-white",
                                        isActive ? "font-semibold" : "group-hover:font-medium"
                                    )}>{item.title}</span>
                                </span>
                                <div className={cn(
                                    "absolute inset-0 bg-gradient-to-r from-transparent via-white/5 to-transparent blur-sm transition-opacity duration-500",
                                    isActive ? "opacity-70" : "opacity-0 group-hover:opacity-100"
                                )}></div>
                            </Link>
                        )
                    })}
            </div>

            {/* Logout button */}
            <div className="mt-auto py-4 border-t border-white/10 dark:border-white/5 relative z-10">
                <button
                    onClick={handleLogout}
                    className={cn(
                        "flex items-center gap-3 font-medium px-6 py-3 w-full group transition-all duration-300 relative overflow-hidden",
                        isLogoutActive
                            ? "bg-white/15 dark:bg-white/10 shadow-[0_0_15px_rgba(255,255,255,0.4)] dark:shadow-[0_0_15px_rgba(255,255,255,0.25)] text-white translate-x-1"
                            : "text-white/90 hover:text-white hover:translate-x-1 hover:shadow-[0_0_10px_rgba(255,255,255,0.3)] dark:hover:shadow-[0_0_12px_rgba(255,255,255,0.2)]",
                        "before:absolute before:content-[''] before:top-0 before:left-0 before:w-full before:h-full before:bg-gradient-to-r before:from-white/5 before:to-white/20 dark:before:from-white/5 dark:before:to-white/15 before:transition-opacity before:duration-300",
                        isLogoutActive
                            ? "before:opacity-100"
                            : "before:opacity-0 hover:before:opacity-100"
                    )}
                >
                    <span className="relative z-10 flex items-center gap-3">
                        <LogOut className={cn(
                            "w-5 h-5 transition-transform",
                            isLogoutActive ? "text-white" : "group-hover:animate-icon-shake"
                        )} />
                        <span className={cn(
                            "transition-all dark:text-shadow-glow-white",
                            isLogoutActive ? "font-semibold" : "group-hover:font-medium"
                        )}>{t("sidebar.logout")}</span>
                    </span>
                    <div className={cn(
                        "absolute inset-0 bg-gradient-to-r from-transparent via-white/5 to-transparent blur-sm transition-opacity duration-500",
                        isLogoutActive ? "opacity-70" : "opacity-0 group-hover:opacity-100"
                    )}></div>
                </button>
                <div className="text-center text-xs text-white/60 dark:text-white/50 mt-4 px-4">© 2025 RuiQi WAF. All Rights Reserved.</div>
            </div>
        </div >
    )
}
