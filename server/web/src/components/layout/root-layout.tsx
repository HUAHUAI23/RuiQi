import { Outlet } from "react-router"
import { Sidebar } from "./sidebar"
import { Breadcrumb } from "./breadcrumb"

export function RootLayout() {
    return (
        <div className="flex h-screen">
            <Sidebar
                displayConfig={{
                    monitor: true,
                    logs: true,
                    rules: true,
                    settings: true,
                }}
            />
            <div className="flex-1 overflow-auto bg-gradient-to-br from-slate-50 to-white scrollbar-none">
                <main className="flex flex-col h-full">
                    <Breadcrumb />
                    <div className="flex-1 overflow-auto">
                        <Outlet />
                    </div>
                </main>
            </div>
        </div>
    )
}
