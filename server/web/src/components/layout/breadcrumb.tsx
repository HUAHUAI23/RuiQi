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
              {index > 0 && <ChevronRight className="w-4 h-4 mx-2 text-primary/60" />}
              <Link
                to={`/${mainPath}/${item.path}`}
                className={cn(
                  "transition-colors duration-200",
                  index === 0
                    ? cn(
                        "text-lg font-medium",
                        currentPath === item.path ? "text-primary" : "text-slate-600 hover:text-primary",
                      )
                    : cn(
                        "text-lg",
                        currentPath === item.path
                          ? "text-primary font-medium"
                          : "text-slate-600 hover:text-primary",
                      ),
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
