import { useState, useEffect } from 'react'
import { useForm, useFieldArray } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useQuery } from '@tanstack/react-query'
import { certificatesApi } from '@/api/certificate'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import {
    Form,
    FormControl,
    FormField,
    FormItem,
    FormLabel,
    FormMessage,
} from '@/components/ui/form'
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select'
import {
    PlusCircle,
    Trash2,
    Server,
    Shield,
    Upload,
    Info,
    AlertCircle,
    RefreshCw
} from 'lucide-react'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { siteFormSchema } from '@/validation/site'
import { CreateSiteRequest, UpdateSiteRequest, WAFMode } from '@/types/site'
import { Certificate } from '@/types/certificate'
import { CertificateDialog } from '@/feature/certificate/components/CertificateDialog'
import { InfoRow } from '@/feature/certificate/components/CertificateForm'
import { useCreateSite, useUpdateSite } from '../hooks/useSites'
import { AnimatedContainer } from '@/components/ui/animation/components/animated-container'
import { useTranslation } from 'react-i18next'
import { AnimatedButton } from '@/components/ui/animation/components/animated-button'
interface SiteFormProps {
    mode?: 'create' | 'update'
    siteId?: string
    onSuccess?: () => void
    defaultValues?: Partial<CreateSiteRequest>
}

export function SiteForm({
    mode = 'create',
    siteId,
    onSuccess,
    defaultValues = {
        name: '',
        domain: '',
        listenPort: 80,
        enableHTTPS: false,
        activeStatus: true,
        wafEnabled: false,
        wafMode: WAFMode.Observation,
        backend: {
            servers: [{ host: '', port: 80, isSSL: false }]
        },
    },
}: SiteFormProps) {
    const { t } = useTranslation()
    // 状态
    const [showCertificateDialog, setShowCertificateDialog] = useState(false)
    const [selectedCertificate, setSelectedCertificate] = useState<Certificate | null>(null)
    const [selectedCertificateId, setSelectedCertificateId] = useState<string>('')

    // API钩子
    const {
        createSite,
        isLoading: isCreating,
        error: createError,
        clearError: clearCreateError
    } = useCreateSite()

    const {
        updateSite,
        isLoading: isUpdating,
        error: updateError,
        clearError: clearUpdateError
    } = useUpdateSite()

    // 获取证书列表
    const { data: certificates, refetch: refetchCertificates } = useQuery({
        queryKey: ['certificates'],
        queryFn: () => certificatesApi.getCertificates(1, 100),
        select: (data) => data.items,
    })

    // 动态状态
    const isLoading = mode === 'create' ? isCreating : isUpdating
    const error = mode === 'create' ? createError : updateError
    const clearError = mode === 'create' ? clearCreateError : clearUpdateError

    // 表单设置
    const form = useForm<CreateSiteRequest>({
        resolver: zodResolver(siteFormSchema),
        defaultValues,
    })

    // 服务器字段数组
    const { fields: serverFields, append: appendServer, remove: removeServer } = useFieldArray({
        control: form.control,
        name: "backend.servers"
    })

    // 初始化选中的证书ID（如果有）
    useEffect(() => {
        if (defaultValues.certificate && certificates) {
            const cert = certificates.find(c =>
                c.fingerPrint === defaultValues.certificate?.fingerPrint
            )
            if (cert) {
                setSelectedCertificateId(cert.id)
                setSelectedCertificate(cert)
            }
        }
    }, [defaultValues.certificate, certificates])

    // 监听证书选择变化
    useEffect(() => {
        if (selectedCertificateId && certificates) {
            const cert = certificates.find(c => c.id === selectedCertificateId)
            if (cert) {
                setSelectedCertificate(cert)
                // 更新证书字段
                form.setValue('certificate', {
                    certName: cert.name,
                    expireDate: cert.expireDate,
                    fingerPrint: cert.fingerPrint,
                    issuerName: cert.issuerName,
                    privateKey: cert.privateKey,
                    publicKey: cert.publicKey,
                })
            }
        } else if (selectedCertificateId === '') {
            setSelectedCertificate(null)
            form.setValue('certificate', undefined)
        }
    }, [selectedCertificateId, certificates, form])

    // 处理证书选择变更
    const handleCertificateChange = (value: string) => {
        if (value === 'upload-new') {
            setShowCertificateDialog(true)
        } else {
            setSelectedCertificateId(value)
        }
    }

    // 添加新服务器
    const addServer = () => {
        appendServer({ host: '', port: 80, isSSL: false })
    }

    // 表单提交处理
    const onSubmit = (data: CreateSiteRequest) => {
        // 清除之前的错误
        if (clearError) clearError()

        if (mode === 'create') {
            createSite(data, {
                onSuccess: () => {
                    if (onSuccess) onSuccess()
                }
            })
        } else if (mode === 'update' && siteId) {
            updateSite({
                id: siteId,
                data: data as UpdateSiteRequest
            }, {
                onSuccess: () => {
                    if (onSuccess) onSuccess()
                }
            })
        }
    }

    return (
        <>
            <AnimatedContainer>
                <Form {...form}>
                    <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
                        {/* API错误提示 */}
                        {error && (
                            <Alert variant="destructive">
                                <AlertCircle className="h-4 w-4" />
                                <AlertDescription>{error}</AlertDescription>
                            </Alert>
                        )}

                        {/* 基本信息部分 */}
                        <div className="space-y-5">
                            <h3 className="text-lg font-medium">{t("site.dialog.basicInfo")}</h3>

                            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                                <FormField
                                    control={form.control}
                                    name="name"
                                    render={({ field }) => (
                                        <div className="flex flex-col gap-1.5">
                                            <FormLabel className="text-sm font-medium">{t('site.dialog.siteName')}</FormLabel>
                                            <FormControl>
                                                <Input
                                                    placeholder={t('site.dialog.siteNamePlaceholder')}
                                                    className="dark:text-shadow-glow-white rounded-md p-3 h-12"
                                                    {...field}
                                                />
                                            </FormControl>
                                            <FormMessage />
                                        </div>
                                    )}
                                />

                                <FormField
                                    control={form.control}
                                    name="domain"
                                    render={({ field }) => (
                                        <div className="flex flex-col gap-1.5">
                                            <FormLabel className="text-sm font-medium">{t('site.domain')}</FormLabel>
                                            <FormControl>
                                                <Input
                                                    placeholder={t('site.dialog.domainPlaceholder')}
                                                    className="dark:text-shadow-glow-white rounded-md p-3 h-12"
                                                    {...field}
                                                />
                                            </FormControl>
                                            <FormMessage />
                                        </div>
                                    )}
                                />

                                <FormField
                                    control={form.control}
                                    name="listenPort"
                                    render={({ field }) => (
                                        <div className="flex flex-col gap-1.5">
                                            <FormLabel className="text-sm font-medium">{t('site.listenPort')}</FormLabel>
                                            <FormControl>
                                                <Input
                                                    placeholder="80"
                                                    className="rounded-md p-3 h-12 dark:text-shadow-glow-white"
                                                    {...field}
                                                    onChange={(e) => field.onChange(Number(e.target.value))}
                                                />
                                            </FormControl>
                                            <FormMessage />
                                        </div>
                                    )}
                                />

                                <FormField
                                    control={form.control}
                                    name="activeStatus"
                                    render={({ field }) => (
                                        <div className="flex flex-col gap-1.5">
                                            <div className="text-sm font-medium dark:text-shadow-glow-white">{t('site.status')}</div>
                                            <div className="w-full flex items-center justify-between rounded-md border p-3 h-12">
                                                <FormControl>
                                                    <div className="flex items-center justify-between w-full">
                                                        <Switch
                                                            checked={field.value}
                                                            onCheckedChange={field.onChange}
                                                        />
                                                        <label className="text-xs text-muted-foreground cursor-pointer">{t('site.dialog.siteActive')}</label>
                                                    </div>
                                                </FormControl>
                                            </div>
                                        </div>
                                    )}
                                />
                            </div>
                        </div>

                        {/* HTTPS设置 */}
                        <div className="space-y-5">
                            <h3 className="text-lg font-medium">{t('site.dialog.httpsSettings')}</h3>

                            <FormField
                                control={form.control}
                                name="enableHTTPS"
                                render={({ field }) => (
                                    <div className="w-full">
                                        <div className="text-sm font-medium dark:text-shadow-glow-white">{t('site.dialog.enableHttps')}</div>
                                        <div className="text-xs text-muted-foreground mb-1 dark:text-shadow-glow-white">{t('site.dialog.httpsDescription')}</div>
                                        <div className="w-full rounded-md border p-3 flex justify-between items-center dark:border-none">
                                            <FormControl>
                                                <Switch
                                                    checked={field.value}
                                                    onCheckedChange={field.onChange}
                                                />
                                            </FormControl>
                                        </div>
                                    </div>
                                )}
                            />

                            {form.watch('enableHTTPS') && (
                                <div className="p-4 border rounded-md">
                                    <div>
                                        <FormLabel className="text-sm font-medium dark:text-shadow-glow-white">{t('site.dialog.selectCertificate')}</FormLabel>
                                        <div className="flex gap-2 mt-1">
                                            <Select
                                                value={selectedCertificateId}
                                                onValueChange={handleCertificateChange}
                                            >
                                                <SelectTrigger className="w-full dark:text-shadow-glow-white">
                                                    <SelectValue placeholder={t('site.dialog.selectCertificate')} />
                                                </SelectTrigger>
                                                <SelectContent>
                                                    {certificates?.map((cert) => (
                                                        <SelectItem key={cert.id} value={cert.id}>
                                                            <div className="flex flex-col">
                                                                <span>{cert.name}</span>
                                                                <span className="text-xs text-muted-foreground">
                                                                    {cert.domains.join(', ')}
                                                                </span>
                                                            </div>
                                                        </SelectItem>
                                                    ))}
                                                    <SelectItem value="upload-new">
                                                        <span className="flex items-center text-blue-600 dark:text-shadow-glow-white dark:text-teal-50">
                                                            <Upload className="mr-2 h-4 w-4" />
                                                            {t('site.dialog.uploadNewCert')}
                                                        </span>
                                                    </SelectItem>
                                                </SelectContent>
                                            </Select>
                                            <Button
                                                type="button"
                                                variant="outline"
                                                size="icon"
                                                onClick={() => refetchCertificates()}
                                            >
                                                <RefreshCw className="h-4 w-4" />
                                            </Button>
                                        </div>
                                        <FormMessage />
                                    </div>


                                    {selectedCertificate && (
                                        <div className="mt-4 p-4 border rounded-md bg-zinc-50 dark:bg-gray-800/10 dark:border-gray-700">
                                            <h4 className="text-sm font-medium mb-2 dark:text-shadow-glow-white">{t('site.dialog.selectedCertInfo')}</h4>
                                            <div className="space-y-2 text-sm">
                                                <InfoRow label={t('certificate.name')} value={selectedCertificate?.name || ''} />
                                                <InfoRow
                                                    label={t('certificate.issuer')}
                                                    value={selectedCertificate?.issuerName || ''}
                                                />
                                                <InfoRow
                                                    label={t('certificate.dialog.expiryDate')}
                                                    value={new Date(selectedCertificate?.expireDate || '').toLocaleDateString()}
                                                />
                                                <div className="flex">
                                                    <span className="w-24 text-muted-foreground dark:text-shadow-glow-white">{t('certificate.dialog.domains')}:</span>
                                                    <div className="flex flex-wrap gap-1">
                                                        {selectedCertificate?.domains.map((domain, index) => (
                                                            <span key={index} className="px-2 py-0.5 bg-gray-200 dark:bg-gray-700 rounded text-xs dark:text-shadow-glow-white">
                                                                {domain}
                                                            </span>
                                                        ))}
                                                    </div>
                                                </div>
                                            </div>
                                        </div>
                                    )}
                                </div>

                            )}



                        </div>

                        {/* 后端服务器 */}
                        <div className="space-y-5">
                            <div className="flex items-center justify-between">
                                <h3 className="text-lg font-medium">{t('site.dialog.backendServers')}</h3>
                                <Button
                                    type="button"
                                    variant="outline"
                                    size="sm"
                                    onClick={addServer}
                                    className="flex items-center gap-1 dark:text-shadow-glow-white"
                                >
                                    <PlusCircle className="h-4 w-4" />
                                    {t('site.dialog.addServer')}
                                </Button>
                            </div>

                            <div className="space-y-4">
                                {serverFields.map((field, index) => (
                                    <div key={field.id} className="p-4 border rounded-md">
                                        <div className="flex items-center justify-between mb-4">
                                            <div className="flex items-center">
                                                <Server className="h-4 w-4 mr-2" />
                                                <span className="font-medium">{t('site.dialog.server')} {index + 1}</span>
                                            </div>
                                            {index > 0 && (
                                                <Button
                                                    type="button"
                                                    variant="ghost"
                                                    size="sm"
                                                    onClick={() => removeServer(index)}
                                                    className="text-red-500 hover:text-red-700"
                                                >
                                                    <Trash2 className="h-4 w-4" />
                                                </Button>
                                            )}
                                        </div>

                                        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                                            <FormField
                                                control={form.control}
                                                name={`backend.servers.${index}.host`}
                                                render={({ field }) => (
                                                    <div className="flex flex-col gap-1.5 justify-between">
                                                        <FormLabel className="text-sm font-medium dark:text-shadow-glow-white">{t('site.dialog.hostAddress')}</FormLabel>
                                                        <FormControl>
                                                            <Input
                                                                placeholder={t('site.dialog.hostPlaceholder')}
                                                                className="rounded-md p-3 dark:text-shadow-glow-white"
                                                                {...field}
                                                            />
                                                        </FormControl>
                                                        <FormMessage />
                                                    </div>
                                                )}
                                            />

                                            <FormField
                                                control={form.control}
                                                name={`backend.servers.${index}.port`}
                                                render={({ field }) => (

                                                    <div className="flex flex-col gap-1.5 justify-between">
                                                        <FormLabel className="text-sm font-medium dark:text-shadow-glow-white">{t('site.dialog.port')}</FormLabel>
                                                        <FormControl>
                                                            <Input
                                                                placeholder="80"
                                                                className="rounded-md p-3 dark:text-shadow-glow-white"
                                                                {...field}
                                                                onChange={(e) => field.onChange(Number(e.target.value))}
                                                            />
                                                        </FormControl>
                                                        <FormMessage />
                                                    </div>
                                                )}
                                            />

                                            <FormField
                                                control={form.control}
                                                name={`backend.servers.${index}.isSSL`}
                                                render={({ field }) => (

                                                    <div className="flex flex-col gap-1.5">
                                                        <div className="text-sm font-medium dark:text-shadow-glow-white">{t('site.dialog.enableSsl')}</div>
                                                        <FormControl>
                                                            <div className="w-full flex items-center justify-between rounded-md border p-3">
                                                                <div className="flex items-center justify-between w-full">
                                                                    <Switch
                                                                        checked={field.value}
                                                                        onCheckedChange={field.onChange}
                                                                    />
                                                                    <label className="text-xs text-muted-foreground cursor-pointer dark:text-shadow-glow-white">{t('site.dialog.backendHttps')}</label>
                                                                </div>
                                                            </div>
                                                        </FormControl>
                                                    </div>

                                                )}
                                            />
                                        </div>
                                    </div>
                                ))}
                            </div>
                        </div>

                        {/* WAF设置 */}
                        <div className="space-y-5">
                            <h3 className="text-lg font-medium">{t('site.dialog.wafSettings')}</h3>

                            <FormField
                                control={form.control}
                                name="wafEnabled"
                                render={({ field }) => (
                                    <div className="w-full">
                                        <div className="text-sm font-medium dark:text-shadow-glow-white">{t('site.dialog.enableWaf')}</div>
                                        <div className="text-xs text-muted-foreground mb-1 dark:text-shadow-glow-white">{t('site.dialog.wafDescription')}</div>
                                        <div className="w-full rounded-md border p-3 flex justify-between items-center dark:border-none">
                                            <FormControl>
                                                <Switch
                                                    checked={field.value}
                                                    onCheckedChange={field.onChange}
                                                />
                                            </FormControl>
                                        </div>
                                    </div>
                                )}
                            />

                            {form.watch('wafEnabled') && (
                                <FormField
                                    control={form.control}
                                    name="wafMode"
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel className="text-sm font-medium dark:text-shadow-glow-white">{t('site.dialog.wafMode')}</FormLabel>
                                            <div className="flex items-center gap-2 mt-1">
                                                <FormControl>
                                                    <Select
                                                        value={field.value}
                                                        onValueChange={field.onChange}
                                                    >
                                                        <SelectTrigger className="w-full dark:text-shadow-glow-white">
                                                            <SelectValue placeholder={t('site.dialog.selectWafMode')} />
                                                        </SelectTrigger>
                                                        <SelectContent>
                                                            <SelectItem value={WAFMode.Observation}>
                                                                <div className="flex items-center">
                                                                    <Info className="mr-2 h-4 w-4 text-blue-500" />
                                                                    <div className="flex flex-col">
                                                                        <span>{t('site.dialog.observationMode')}</span>
                                                                        <span className="text-xs text-muted-foreground">
                                                                            {t('site.dialog.observationDescription')}
                                                                        </span>
                                                                    </div>
                                                                </div>
                                                            </SelectItem>
                                                            <SelectItem value={WAFMode.Protection}>
                                                                <div className="flex items-center">
                                                                    <Shield className="mr-2 h-4 w-4 text-green-500" />
                                                                    <div className="flex flex-col">
                                                                        <span>{t('site.dialog.protectionMode')}</span>
                                                                        <span className="text-xs text-muted-foreground">
                                                                            {t('site.dialog.protectionDescription')}
                                                                        </span>
                                                                    </div>
                                                                </div>
                                                            </SelectItem>
                                                        </SelectContent>
                                                    </Select>
                                                </FormControl>
                                            </div>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />
                            )}
                        </div>

                        {/* 提交按钮 */}
                        <div className="flex justify-end gap-2">
                            <AnimatedButton>
                                <Button type="submit" disabled={isLoading}>
                                    {isLoading ? t('site.dialog.submitting') : mode === 'create' ? t('site.dialog.createSite') : t('site.dialog.updateSite')}
                                </Button>
                            </AnimatedButton>
                        </div>
                    </form>
                </Form>
            </AnimatedContainer>

            {/* 证书创建对话框 */}
            <CertificateDialog
                open={showCertificateDialog || selectedCertificateId === 'upload-new'}
                onOpenChange={(open) => {
                    setShowCertificateDialog(open)
                    if (!open && selectedCertificateId === 'upload-new') {
                        setSelectedCertificateId('')
                    }
                }}
                mode="create"
                certificate={null}
            />
        </>
    )
}