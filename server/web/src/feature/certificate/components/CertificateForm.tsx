import { useState, useCallback } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import {
    Form,
    FormControl,
    FormField,
    FormItem,
    FormLabel,
    FormMessage,
} from '@/components/ui/form'
import { Upload, FileText, X, AlertCircle, Info } from 'lucide-react'
import { certificateFormSchema } from '@/validation/certificate'
import { parseCertificate, readFileAsText } from '@/utils/certificate-parser'
import { CertificateCreateRequest, ParsedCertificate } from '@/types/certificate'
import { useCreateCertificate, useUpdateCertificate } from '../hooks/useCertificate'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { AnimatedContainer } from '@/components/ui/animation/components/animated-container'
import { useTranslation } from 'react-i18next'
import { AnimatedButton } from '@/components/ui/animation/components/animated-button'
interface CertificateFormProps {
    mode?: 'create' | 'update'
    certificateId?: string
    onSuccess?: () => void
    defaultValues?: Partial<CertificateCreateRequest>
}

export function CertificateForm({
    mode = 'create',
    certificateId,
    onSuccess,
    defaultValues = {
        name: '',
        description: '',
        publicKey: '',
        privateKey: '',
    },
}: CertificateFormProps) {
    const { t } = useTranslation()

    // 状态管理
    const [parsedInfo, setParsedInfo] = useState<ParsedCertificate | null>(null)
    const [publicKeyFile, setPublicKeyFile] = useState<string | null>(null)
    const [privateKeyFile, setPrivateKeyFile] = useState<string | null>(null)
    const [parseError, setParseError] = useState<string | null>(null)

    // API钩子
    const {
        createCertificate,
        isLoading: isCreating,
        error: createError,
        clearError: clearCreateError
    } = useCreateCertificate()

    const {
        updateCertificate,
        isLoading: isUpdating,
        error: updateError,
        clearError: clearUpdateError
    } = useUpdateCertificate()

    // 动态状态
    const isLoading = mode === 'create' ? isCreating : isUpdating
    const error = mode === 'create' ? createError : updateError
    const clearError = mode === 'create' ? clearCreateError : clearUpdateError

    // 表单设置
    const form = useForm<CertificateCreateRequest>({
        resolver: zodResolver(certificateFormSchema),
        defaultValues,
    })

    // 尝试解析证书内容
    const tryParseCertificate = useCallback((content: string) => {
        if (!content) {
            setParsedInfo(null)
            setParseError(null)
            return
        }

        try {
            const parsed = parseCertificate(content)
            setParsedInfo(parsed)
            setParseError(null)

            // 如果之前有解析错误，清除相关表单错误
            form.clearErrors('publicKey')
        } catch (error) {
            console.error('证书解析错误:', error)

            // 清除已解析信息
            setParsedInfo(null)

            // 设置错误信息，但不阻止表单提交
            if (error instanceof Error) {
                setParseError(`${t("certificate.dialog.parseFailed")}${error.message}`)
            } else {
                setParseError(`${t("certificate.dialog.parseFailed")}${t("certificate.dialog.unknownError")}`)
            }
        }
    }, [form, t])

    // 处理公钥文件上传
    const handlePublicKeyFileChange = useCallback(async (e: React.ChangeEvent<HTMLInputElement>) => {
        if (e.target.files && e.target.files.length > 0) {
            const file = e.target.files[0]

            try {
                const content = await readFileAsText(file)

                // 更新表单和UI状态
                form.setValue('publicKey', content)
                setPublicKeyFile(file.name)

                // 尝试解析证书
                tryParseCertificate(content)
            } catch (error) {
                console.error('文件读取错误:', error)

                // 文件读取错误显示在表单上
                const errorMessage = error instanceof Error ? error.message : t("certificate.dialog.unknownError")
                form.setError('publicKey', {
                    type: 'manual',
                    message: `${t("certificate.dialog.fileReadFailed")}${errorMessage}`
                })
            }
        }
    }, [form, tryParseCertificate, t])

    // 处理私钥文件上传
    const handlePrivateKeyFileChange = useCallback(async (e: React.ChangeEvent<HTMLInputElement>) => {
        if (e.target.files && e.target.files.length > 0) {
            const file = e.target.files[0]

            try {
                const content = await readFileAsText(file)

                // 更新表单和UI状态
                form.setValue('privateKey', content)
                setPrivateKeyFile(file.name)
            } catch (error) {
                console.error('文件读取错误:', error)

                // 文件读取错误显示在表单上
                const errorMessage = error instanceof Error ? error.message : t("certificate.dialog.unknownError")
                form.setError('privateKey', {
                    type: 'manual',
                    message: `${t("certificate.dialog.fileReadFailed")}${errorMessage}`
                })
            }
        }
    }, [form, t])

    // 清除公钥文件
    const clearPublicKeyFile = useCallback(() => {
        setPublicKeyFile(null)
        form.setValue('publicKey', '')
        setParsedInfo(null)
        setParseError(null)
    }, [form])

    // 清除私钥文件
    const clearPrivateKeyFile = useCallback(() => {
        setPrivateKeyFile(null)
        form.setValue('privateKey', '')
    }, [form])

    // 处理公钥文本框变更
    const handlePublicKeyTextChange = useCallback((e: React.ChangeEvent<HTMLTextAreaElement>) => {
        const value = e.target.value
        tryParseCertificate(value)
    }, [tryParseCertificate])

    // 表单提交处理
    const handleFormSubmit = useCallback((data: CertificateCreateRequest) => {
        // 清除之前的错误
        if (clearError) clearError()

        // 合并解析出来的证书信息
        const finalData = {
            ...data,
            ...(parsedInfo || {}),
        }

        // 根据模式执行创建或更新操作
        if (mode === 'create') {
            createCertificate(finalData, {
                onSuccess: () => {
                    // 重置表单
                    form.reset()
                    setPublicKeyFile(null)
                    setPrivateKeyFile(null)
                    setParsedInfo(null)
                    setParseError(null)
                    // 通知父组件成功
                    if (onSuccess) onSuccess()
                }
            })
        } else if (mode === 'update' && certificateId) {
            updateCertificate({ id: certificateId, data: finalData }, {
                onSuccess: () => {
                    // 通知父组件成功
                    if (onSuccess) onSuccess()
                }
            })
        }
    }, [mode, certificateId, clearError, parsedInfo, createCertificate, updateCertificate, form, onSuccess])

    // 渲染已解析的证书信息
    const renderParsedInfo = useCallback(() => {
        if (!parsedInfo) return null

        return (
            <div className="p-4 border rounded-md bg-zinc-50 dark:bg-gray-800/10 dark:border-gray-700">
                <h3 className="text-sm font-medium mb-2">{t("certificate.dialog.parsedInfo")}</h3>
                <div className="space-y-2 text-sm">
                    <InfoRow label={t("certificate.dialog.issuer")} value={parsedInfo.issuerName} />
                    <InfoRow
                        label={t("certificate.dialog.expiryDate")}
                        value={new Date(parsedInfo.expireDate).toLocaleDateString()}
                    />
                    <InfoRow label={t("certificate.dialog.fingerprint")} value={parsedInfo.fingerPrint} className="font-mono text-xs" />
                    <div className="flex">
                        <span className="w-24 text-muted-foreground">{t("certificate.dialog.domains")}:</span>
                        <div className="flex flex-wrap gap-1">
                            {parsedInfo.domains.map((domain, index) => (
                                <span key={index} className="px-2 py-0.5 bg-gray-200 dark:bg-gray-700 rounded text-xs dark:text-shadow-glow-white">
                                    {domain}
                                </span>
                            ))}
                        </div>
                    </div>
                </div>
            </div>
        )
    }, [parsedInfo, t])

    return (
        <AnimatedContainer>
            <Form {...form}>
                <form onSubmit={form.handleSubmit(handleFormSubmit)} className="space-y-6">
                    {/* API错误提示 */}
                    {error && (
                        <Alert variant="destructive">
                            <AlertCircle className="h-4 w-4" />
                            <AlertDescription>{error}</AlertDescription>
                        </Alert>
                    )}

                    {/* 证书解析错误提示 */}
                    {parseError && (
                        <Alert variant="default" className="bg-yellow-50 border-yellow-200">
                            <Info className="h-4 w-4 text-yellow-800" />
                            <AlertDescription className="text-yellow-800">{parseError}</AlertDescription>
                        </Alert>
                    )}

                    {/* 基本信息字段 */}
                    <FormField
                        control={form.control}
                        name="name"
                        render={({ field }) => (
                            <FormItem>
                                <FormLabel className="dark:text-shadow-glow-white">{t("certificate.dialog.certName")}</FormLabel>
                                <FormControl>
                                    <Input className="dark:text-shadow-glow-white" placeholder={t("certificate.dialog.certName")} {...field} />
                                </FormControl>
                                <FormMessage />
                            </FormItem>
                        )}
                    />

                    <FormField
                        control={form.control}
                        name="description"
                        render={({ field }) => (
                            <FormItem>
                                <FormLabel className="dark:text-shadow-glow-white">{t("certificate.dialog.description")}</FormLabel>
                                <FormControl>
                                    <Textarea
                                        placeholder={t("certificate.dialog.descriptionPlaceholder")}
                                        className="resize-none dark:text-shadow-glow-white"
                                        {...field}
                                    />
                                </FormControl>
                                <FormMessage />
                            </FormItem>
                        )}
                    />

                    {/* 公钥文件上传 */}
                    <div className="space-y-2">
                        <FormLabel className="dark:text-shadow-glow-white">{t("certificate.dialog.publicKeyFile")}</FormLabel>
                        {publicKeyFile ? (
                            <FilePreview
                                filename={publicKeyFile}
                                onClear={clearPublicKeyFile}
                            />
                        ) : (
                            <FileUpload
                                label={t("certificate.dialog.uploadPublicKey")}
                                accept=".pem,.crt,.cert,.key"
                                onChange={handlePublicKeyFileChange}
                            />
                        )}
                    </div>

                    {/* 公钥内容 */}
                    <FormField
                        control={form.control}
                        name="publicKey"
                        render={({ field }) => (
                            <FormItem>
                                <FormLabel className="dark:text-shadow-glow-white">{t("certificate.dialog.publicKeyContent")}</FormLabel>
                                <FormControl>
                                    <Textarea
                                        placeholder={t("certificate.dialog.publicKeyPlaceholder")}
                                        className="font-mono text-xs h-32 dark:text-shadow-glow-white"
                                        {...field}
                                        onChange={(e) => {
                                            field.onChange(e)
                                            handlePublicKeyTextChange(e)
                                        }}
                                    />
                                </FormControl>
                                <FormMessage />
                            </FormItem>
                        )}
                    />

                    {/* 私钥文件上传 */}
                    <div className="space-y-2">
                        <FormLabel className='dark:text-shadow-glow-white'>{t("certificate.dialog.privateKeyFile")}</FormLabel>
                        {privateKeyFile ? (
                            <FilePreview
                                filename={privateKeyFile}
                                onClear={clearPrivateKeyFile}
                            />
                        ) : (
                            <FileUpload
                                label={t("certificate.dialog.uploadPrivateKey")}
                                accept=".pem,.key"
                                onChange={handlePrivateKeyFileChange}
                            />
                        )}
                    </div>

                    {/* 私钥内容 */}
                    <FormField
                        control={form.control}
                        name="privateKey"
                        render={({ field }) => (
                            <FormItem>
                                <FormLabel className="dark:text-shadow-glow-white">{t("certificate.dialog.privateKeyContent")}</FormLabel>
                                <FormControl>
                                    <Textarea
                                        placeholder={t("certificate.dialog.privateKeyPlaceholder")}
                                        className="font-mono text-xs h-32 dark:text-shadow-glow-white"
                                        {...field}
                                    />
                                </FormControl>
                                <FormMessage />
                            </FormItem>
                        )}
                    />

                    {/* 证书解析信息 */}
                    {renderParsedInfo()}

                    {/* 提交按钮 */}
                    <div className="flex justify-end">
                        <AnimatedButton>
                            <Button type="submit" disabled={isLoading}>
                                {isLoading ? t("certificate.dialog.submitting") : mode === 'create' ? t("certificate.dialog.create") : t("certificate.dialog.update")}
                            </Button>
                        </AnimatedButton>
                    </div>
                </form>
            </Form>
        </AnimatedContainer >
    )
}

// 辅助组件：信息行
interface InfoRowProps {
    label: string
    value: string
    className?: string
}

export function InfoRow({ label, value, className = '' }: InfoRowProps) {
    return (
        <div className="flex">
            <span className="w-24 flex-shrink-0 text-muted-foreground">{label}:</span>
            <span className={`break-all ${className}`}>{value}</span>
        </div>
    )
}

// 辅助组件：文件预览
interface FilePreviewProps {
    filename: string
    onClear: () => void
}

function FilePreview({ filename, onClear }: FilePreviewProps) {
    return (
        <div className="flex items-center gap-2 p-2 border rounded">
            <FileText className="h-4 w-4" />
            <span className="text-sm flex-1 truncate">{filename}</span>
            <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={onClear}
            >
                <X className="h-4 w-4" />
            </Button>
        </div>
    )
}

// 辅助组件：文件上传
interface FileUploadProps {
    label: string
    accept: string
    onChange: (e: React.ChangeEvent<HTMLInputElement>) => void
}

function FileUpload({ label, accept, onChange }: FileUploadProps) {
    const { t } = useTranslation()

    return (
        <div className="flex items-center gap-2">
            <Button
                type="button"
                variant="outline"
                size="sm"
                asChild
            >
                <label className="cursor-pointer flex items-center gap-2">
                    <Upload className="h-4 w-4" />
                    <span>{label}</span>
                    <input
                        type="file"
                        className="hidden"
                        accept={accept}
                        onChange={onChange}
                    />
                </label>
            </Button>
            <span className="text-sm text-muted-foreground">{t("certificate.dialog.orEnterContent")}</span>
        </div>
    )
}