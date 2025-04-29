import { AppRouter } from "@/routes"
import { ThemeProvider } from "@/provider/theme-provider"

function App() {
    return (
        <ThemeProvider defaultTheme="system" storageKey="waf-theme">
            <AppRouter />
        </ThemeProvider>
    )
}

export default App
