import { Card } from "@/components/ui/card"
import { Outlet } from "react-router"
import { AnimatedRoute } from "@/components/layout/animated-route"

export function MonitorLayOut() {
    return (
        <Card className="border-none shadow-none p-0 overflow-hidden">
            <AnimatedRoute>
                <Outlet />
            </AnimatedRoute>
        </Card>
    )
}