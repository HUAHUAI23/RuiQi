import { useState, useEffect } from "react"
import { motion } from "motion/react"

// Loading component with translations
export const LoadingFallback = () => {
    const [progress, setProgress] = useState(0)

    // Simulate loading progress
    useEffect(() => {
        const timer = setInterval(() => {
            setProgress((prevProgress) => {
                // Slow down as it approaches 100%
                // 使用Math.floor确保结果是整数
                const increment = Math.floor(Math.max(1, 10 * (1 - prevProgress / 100)))
                const newProgress = Math.min(99, prevProgress + increment)
                // 确保最终结果也是整数
                return Math.floor(newProgress)
            })
        }, 200)

        return () => {
            clearInterval(timer)
        }
    }, [])

    // Define the gradient for reuse
    const purpleGradient = `linear-gradient(135deg, 
    rgba(147, 112, 219, 0.95) 0%, 
    rgba(138, 100, 208, 0.9) 50%, 
    rgba(123, 79, 214, 0.95) 100%)`

    return (
        <div className="fixed inset-0 flex flex-col items-center justify-center z-[9999]">
            <div
                className="absolute inset-0"
                style={{
                    background: purpleGradient,
                    backgroundSize: "200% 200%",
                }}
            >
                <div className="absolute inset-0 overflow-hidden">
                    <div className="absolute w-[80%] h-[80%] top-[10%] left-[10%] bg-white/10 rounded-full blur-3xl animate-float"></div>
                    <div className="absolute w-[40%] h-[40%] top-[5%] right-[15%] bg-purple-200/20 rounded-full blur-3xl animate-float-reverse"></div>
                    <div className="absolute w-[50%] h-[50%] bottom-[5%] left-[15%] bg-purple-100/20 rounded-full blur-3xl animate-pulse-glow"></div>
                </div>
            </div>

            <div className="relative z-10 flex flex-col items-center space-y-8">
                <motion.div
                    initial={{ opacity: 0, y: -20 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ duration: 0.5 }}
                    className="text-white text-2xl font-medium"
                >
                    Loading...
                </motion.div>

                <div className="w-64 h-2 bg-white/20 rounded-full overflow-hidden">
                    <motion.div
                        className="h-full rounded-full"
                        style={{
                            background: "linear-gradient(90deg, #e0c3fc 0%, #8ec5fc 100%)",
                            width: `${progress}%`,
                            boxShadow: "0 0 10px rgba(224, 195, 252, 0.7)",
                        }}
                        initial={{ width: "0%" }}
                        animate={{ width: `${progress}%` }}
                        transition={{ duration: 0.3 }}
                    />
                </div>

                <motion.div
                    className="text-white/80 text-sm"
                    animate={{ opacity: [0.5, 1, 0.5] }}
                    transition={{ duration: 2, repeat: Number.POSITIVE_INFINITY }}
                >
                    {progress}% Complete
                </motion.div>
            </div>
        </div>
    )
}