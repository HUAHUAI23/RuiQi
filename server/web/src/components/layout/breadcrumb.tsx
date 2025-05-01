import { useLocation, Link } from "react-router"
import { useBreadcrumbMap } from "@/routes/config"
import type { RoutePath } from "@/routes/constants"
import { cn } from "@/lib/utils"
import { ChevronRight } from "lucide-react"

export function Breadcrumb() {
    const location = useLocation()
    const breadcrumbMap = useBreadcrumbMap()
    const [mainPath, subPath] = location.pathname.split("/").filter(Boolean)
    const config = breadcrumbMap[`/${mainPath}` as RoutePath]

    if (!config) return null

    const currentPath = subPath || config.defaultPath

    return (
        <div className="bg-white dark:bg-background border-b border-slate-100 dark:border-background shadow-sm">
            <div className="px-6 py-4">
                <div className="flex items-center gap-2">
                    {config.items.map((item, index) => (
                        <div key={item.path} className="flex items-center">
                            {index > 0 && <ChevronRight className="w-4 h-4 mx-2 text-primary text-shadow-primary" />}
                            <Link
                                to={`/${mainPath}/${item.path}`}
                                className={cn(
                                    "transition-all duration-300",
                                    index === 0
                                        ? (currentPath === item.path
                                            ? "text-lg font-medium text-[#E8DFFF] text-shadow-glow-purple dark:text-primary dark:text-shadow-glow-white"
                                            : "text-lg font-medium text-slate-600 hover:text-[#E8DFFF] hover:text-shadow-glow-purple dark:text-white dark:font-semibold dark:hover:text-[#d4b8ff] dark:hover:text-shadow-glow-white")
                                        : (currentPath === item.path
                                            ? "text-lg text-primary font-medium text-[#E8DFFF] text-shadow-glow-purple dark:text-primary dark:text-shadow-glow-white"
                                            : "text-lg text-slate-600 hover:text-[#E8DFFF] hover:text-shadow-glow-purple dark:text-white dark:font-semibold dark:hover:text-[#d4b8ff] dark:hover:text-shadow-glow-white")
                                )}
                            >
                                {item.title}
                            </Link>
                        </div>
                    ))}
                </div>
            </div>
        </div>
    )
}
