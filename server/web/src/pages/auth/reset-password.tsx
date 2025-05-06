import { useEffect } from 'react'
import { useNavigate } from 'react-router'
import { PasswordResetForm } from '@/feature/auth/components/PasswordResetForm'
import useAuthStore from '@/store/auth'
import { useTranslation } from 'react-i18next'

export default function ResetPasswordPage() {
    const { isAuthenticated } = useAuthStore()
    const navigate = useNavigate()
    const { t } = useTranslation()

    useEffect(() => {
        // If not authenticated, redirect to login
        if (!isAuthenticated) {
            navigate('/login')
        }
    }, [isAuthenticated, navigate])

    return (
        <div className="min-h-screen flex flex-col items-center justify-center p-4 relative overflow-hidden bg-gradient-to-br from-purple-700 via-purple-500 to-indigo-500 before:content-[''] before:absolute before:inset-0 before:bg-[length:400%_400%] before:bg-gradient-to-br before:from-purple-600 before:via-indigo-500 before:to-pink-500 before:animate-gradient-shift before:opacity-70">
            {/* 动态光晕背景效果 */}
            <div className="absolute inset-0 bg-[size:200%_200%] animate-aurora">
                <div className="absolute inset-0 overflow-hidden">
                    <div className="absolute w-[80%] h-[80%] top-[10%] left-[10%] bg-purple-300/20 rounded-full blur-3xl animate-float"></div>
                    <div className="absolute w-[40%] h-[40%] top-[5%] right-[15%] bg-cyan-300/20 rounded-full blur-3xl animate-float-reverse"></div>
                    <div className="absolute w-[50%] h-[50%] bottom-[5%] left-[15%] bg-indigo-300/20 rounded-full blur-3xl animate-pulse-glow"></div>
                </div>
            </div>


            <div className="mb-8 z-10 animate-fade-in-up">
                <h1 className="text-3xl font-bold text-center text-white drop-shadow-md hover:animate-wiggle text-shadow-glow-white">
                    {t('sidebar.title')}
                </h1>
            </div>

            <div className="w-full max-w-md z-10 animate-fade-in-up [animation-delay:200ms]">
                <PasswordResetForm />
            </div>

            {/* 底部说明文字 */}
            <div className="mt-8 text-white/70 text-sm text-center z-10 animate-fade-in-up [animation-delay:400ms]">
                © {new Date().getFullYear()} RuiQi WAF. All rights reserved.
            </div>
        </div>
    )
} 