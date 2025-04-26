"use client"

import { Link, useLocation, useNavigate } from "react-router"
import { cn } from "@/lib/utils"
import { Settings, Shield, BarChart2, FileText, LogOut } from "lucide-react"
import { ROUTES } from "@/routes/constants"
import { useTranslation } from "react-i18next"
import type { TFunction } from "i18next"
import { useAuthStore } from "@/store/auth"

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
        logout()
        navigate("/login")
    }

    return (
        <div
            className="w-64 text-white flex flex-col border-r border-slate-200 relative overflow-hidden"
            style={{
                background: `linear-gradient(135deg, 
                    rgba(147, 112, 219, 0.95) 0%, 
                    rgba(138, 100, 208, 0.9) 50%, 
                    rgba(123, 79, 214, 0.95) 100%)`,
            }}
        >
            {/* Decorative background elements */}
            <div className="absolute bottom-0 left-0 w-full h-48 overflow-hidden opacity-20 pointer-events-none">
                <div className="absolute bottom-[-10px] left-[-10px] w-20 h-20 bg-white/30 rotate-45 transform animate-float"></div>
                <div className="absolute bottom-[-5px] left-[40px] w-12 h-12 bg-white/20 rotate-12 transform animate-float-reverse"></div>
                <div className="absolute bottom-[30px] left-[80px] w-16 h-16 bg-white/25 rotate-30 transform animate-float"></div>
                <div className="absolute bottom-[10px] left-[120px] w-24 h-24 bg-white/15 rotate-20 transform animate-float-reverse"></div>
                <div className="absolute bottom-[40px] left-[180px] w-14 h-14 bg-white/20 rotate-45 transform animate-float"></div>
                <div className="absolute bottom-[-20px] left-[220px] w-20 h-20 bg-white/10 rotate-30 transform animate-float-reverse"></div>
            </div>

            {/* Logo and title */}
            <div className="flex flex-col items-center gap-2 py-6 border-b border-white/10">
                <div className="w-16 h-16 rounded-full bg-gradient-to-br from-[#A48BEA] to-[#8861DB] flex items-center justify-center shadow-lg animate-pulse-glow">
                    <Shield className="w-8 h-8 text-white" />
                </div>
                <div className="font-bold text-xl mt-2">
                    <span className="text-[#E8DFFF]">RuiQi</span>
                    <span className="text-[#8ED4FF]"> WAF</span>
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
                                    "flex items-center gap-3 font-medium px-6 py-3 w-full group",
                                    isActive
                                        ? "bg-white/15 hover:bg-white/25 text-white"
                                        : "text-white/90 hover:bg-white/10 hover:text-white",
                                )}
                            >
                                <item.icon className="w-5 h-5 group-hover:animate-icon-shake" />
                                {item.title}
                            </Link>
                        )
                    })}
            </div>

            {/* Logout button */}
            <div className="mt-auto py-4 border-t border-white/10 relative z-10">
                <button
                    onClick={handleLogout}
                    className="flex items-center gap-3 font-medium text-white/90 hover:bg-white/10 hover:text-white px-6 py-3 w-full group"
                >
                    <LogOut className="w-5 h-5 group-hover:animate-icon-shake" />
                    {t("sidebar.logout")}
                </button>
                <div className="text-center text-xs text-white/60 mt-4 px-4">© 2025 RuiQi WAF. All Rights Reserved.</div>
            </div>
        </div>
    )
}
