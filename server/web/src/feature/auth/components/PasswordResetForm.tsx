import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Lock } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { useResetPassword } from '../hooks'
import { passwordResetSchema, PasswordResetFormValues } from '@/validation/auth'
import useAuthStore from '@/store/auth'
import { useTranslation } from 'react-i18next'

export function PasswordResetForm() {
    const { t } = useTranslation()
    const { resetPassword, isLoading, error, clearError } = useResetPassword()
    const { user } = useAuthStore()

    const form = useForm<PasswordResetFormValues>({
        resolver: zodResolver(passwordResetSchema),
        defaultValues: {
            oldPassword: '',
            newPassword: '',
            confirmPassword: '',
        },
    })

    const onSubmit = (values: PasswordResetFormValues) => {
        clearError()
        resetPassword({
            oldPassword: values.oldPassword,
            newPassword: values.newPassword,
        })
    }

    return (
        <Card className="w-full max-w-md mx-auto shadow-xl bg-white/90 backdrop-blur-md border-0 transition-all hover:shadow-2xl duration-300">
            <CardHeader className="space-y-1 pb-2">
                <CardTitle className="text-2xl font-bold text-center text-gray-800">{t('auth.resetPassword')}</CardTitle>
                <CardDescription className="text-center text-gray-600">
                    {user?.needReset
                        ? t('auth.firstLoginReset')
                        : t('auth.enterOldAndNewPassword')}
                </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4 pt-4">
                {error && (
                    <Alert variant="destructive" className="mb-4 bg-red-50 border-red-200 animate-fade-in-up">
                        <AlertDescription className="text-red-700">{error}</AlertDescription>
                    </Alert>
                )}

                <Form {...form}>
                    <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
                        <FormField
                            control={form.control}
                            name="oldPassword"
                            render={({ field }) => (
                                <FormItem className="transition-all duration-300 animate-fade-in-up">
                                    <FormLabel className="text-gray-700">{t('auth.currentPassword')}</FormLabel>
                                    <FormControl>
                                        <div className="relative group">
                                            <Lock className="absolute left-3 top-3 h-5 w-5 text-gray-400 group-hover:text-purple-500 transition-colors duration-300" />
                                            <Input
                                                type="password"
                                                placeholder={t('auth.enterCurrentPassword')}
                                                className="pl-10 py-6 text-gray-700 bg-gray-50 border-gray-200 focus:bg-white transition-all group-hover:border-purple-300"
                                                {...field}
                                                onChange={(e) => {
                                                    clearError()
                                                    field.onChange(e)
                                                }}
                                            />
                                        </div>
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        <FormField
                            control={form.control}
                            name="newPassword"
                            render={({ field }) => (
                                <FormItem className="transition-all duration-300 animate-fade-in-up [animation-delay:150ms]">
                                    <FormLabel className="text-gray-700">{t('auth.newPassword')}</FormLabel>
                                    <FormControl>
                                        <div className="relative group">
                                            <Lock className="absolute left-3 top-3 h-5 w-5 text-gray-400 group-hover:text-purple-500 transition-colors duration-300" />
                                            <Input
                                                type="password"
                                                placeholder={t('auth.enterNewPassword')}
                                                className="pl-10 py-6 text-gray-700 bg-gray-50 border-gray-200 focus:bg-white transition-all group-hover:border-purple-300"
                                                {...field}
                                                onChange={(e) => {
                                                    clearError()
                                                    field.onChange(e)
                                                }}
                                            />
                                        </div>
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        <FormField
                            control={form.control}
                            name="confirmPassword"
                            render={({ field }) => (
                                <FormItem className="transition-all duration-300 animate-fade-in-up [animation-delay:300ms]">
                                    <FormLabel className="text-gray-700">{t('auth.confirmNewPassword')}</FormLabel>
                                    <FormControl>
                                        <div className="relative group">
                                            <Lock className="absolute left-3 top-3 h-5 w-5 text-gray-400 group-hover:text-purple-500 transition-colors duration-300" />
                                            <Input
                                                type="password"
                                                placeholder={t('auth.enterNewPasswordAgain')}
                                                className="pl-10 py-6 text-gray-700 bg-gray-50 border-gray-200 focus:bg-white transition-all group-hover:border-purple-300"
                                                {...field}
                                                onChange={(e) => {
                                                    clearError()
                                                    field.onChange(e)
                                                }}
                                            />
                                        </div>
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        <Button
                            type="submit"
                            className="w-full mt-6 py-6 !text-white bg-gradient-to-r from-purple-600 to-indigo-600 hover:from-purple-700 hover:to-indigo-700 transition-all shadow-md hover:shadow-lg animate-fade-in-up [animation-delay:450ms] hover:translate-y-[-2px] text-shadow-glow-white"
                            disabled={isLoading}
                        >
                            {isLoading ? t('auth.submitting') : t('auth.resetPassword')}
                        </Button>
                    </form>
                </Form>
            </CardContent>
            <CardFooter className="flex justify-center border-t border-gray-100 pt-4">
                <p className="text-sm text-gray-500">
                    {t('auth.passwordRequirement')}
                </p>
            </CardFooter>
        </Card>
    )
} 