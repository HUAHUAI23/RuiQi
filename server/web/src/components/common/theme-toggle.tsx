import { Moon, Sun } from "lucide-react"
import { useTheme } from "@/provider/theme-context"

import { Switch } from "@/components/ui/switch"

export function ThemeToggle() {
    const { theme, setTheme } = useTheme()

    const toggleTheme = () => {
        setTheme(theme === "light" ? "dark" : "light")
    }

    return (
        <div className="flex items-center space-x-2 transition-all duration-700 ease-in-out">
            <Sun
                className={`h-[1.2rem] w-[1.2rem] transition-all duration-700 ease-in-out ${theme === "dark" ? "text-[#A1A1AA] scale-75 rotate-12" : "text-foreground scale-100 rotate-0"
                    }`}
            />
            <Switch
                checked={theme === "dark"}
                onCheckedChange={toggleTheme}
                aria-label="Toggle theme"
                className="transition-all duration-700 ease-in-out hover:scale-110"
            />
            <Moon
                className={`h-[1.2rem] w-[1.2rem] transition-all duration-700 ease-in-out ${theme === "light" ? "text-[#A1A1AA] scale-75 rotate-12" : "text-foreground scale-100 rotate-0"
                    }`}
            />
        </div>
    )
}
