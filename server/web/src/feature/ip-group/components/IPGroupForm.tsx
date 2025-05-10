import { useState } from "react"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Trash2, Plus, AlertCircle, Search } from "lucide-react"
import {
    Form,
    FormControl,
    FormField,
    FormItem,
    FormLabel,
    FormMessage,
} from "@/components/ui/form"
import { useTranslation } from "react-i18next"
import { IPGroupFormValues, ipGroupFormSchema } from "@/validation/ip-group"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { useCreateIPGroup, useUpdateIPGroup } from "../hooks"
import { AnimatedContainer } from "@/components/ui/animation/components/animated-container"
import { AnimatedButton } from "@/components/ui/animation/components/animated-button"
import { ScrollArea } from "@/components/ui/scroll-area"

interface IPGroupFormProps {
    mode?: 'create' | 'update'
    ipGroupId?: string
    defaultValues?: IPGroupFormValues
    onSuccess?: () => void
}

export function IPGroupForm({
    mode = 'create',
    ipGroupId,
    defaultValues = {
        name: '',
        items: []
    },
    onSuccess
}: IPGroupFormProps) {
    const { t } = useTranslation()
    const [newIpAddress, setNewIpAddress] = useState("")
    const [ipAddressError, setIpAddressError] = useState<string | null>(null)
    const [searchQuery, setSearchQuery] = useState("")

    // API hooks
    const {
        createIPGroup,
        isLoading: isCreating,
        error: createError,
        clearError: clearCreateError
    } = useCreateIPGroup()

    const {
        updateIPGroup,
        isLoading: isUpdating,
        error: updateError,
        clearError: clearUpdateError
    } = useUpdateIPGroup()

    // Dynamic state
    const isLoading = mode === 'create' ? isCreating : isUpdating
    const error = mode === 'create' ? createError : updateError
    const clearError = mode === 'create' ? clearCreateError : clearUpdateError

    const form = useForm<IPGroupFormValues>({
        resolver: zodResolver(ipGroupFormSchema),
        defaultValues,
    })

    const ipItems = form.watch("items")
    const filteredIpItems = ipItems.filter(item =>
        item.toLowerCase().includes(searchQuery.toLowerCase())
    )

    const handleAddIpAddress = () => {
        // Validate IP address format
        const ipRegex = /^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])(\/([0-9]|[1-2][0-9]|3[0-2]))?$/
        if (!newIpAddress.trim()) {
            setIpAddressError(t("ipGroup.validation.ipItemRequired"))
            return
        }
        if (!ipRegex.test(newIpAddress.trim())) {
            setIpAddressError(t("ipGroup.validation.ipItemInvalid"))
            return
        }

        // Check for duplicates
        if (ipItems.includes(newIpAddress.trim())) {
            setIpAddressError(t("ipGroup.validation.ipItemDuplicate"))
            return
        }

        form.setValue("items", [...ipItems, newIpAddress.trim()])
        setNewIpAddress("")
        setIpAddressError(null)
    }

    const handleRemoveIpAddress = (index: number) => {
        const updatedItems = [...ipItems]
        updatedItems.splice(index, 1)
        form.setValue("items", updatedItems)
    }

    const handleKeyDown = (e: React.KeyboardEvent) => {
        if (e.key === "Enter") {
            e.preventDefault()
            handleAddIpAddress()
        }
    }

    const onSubmit = (data: IPGroupFormValues) => {
        // Clear previous errors
        if (clearError) clearError()

        if (mode === 'update' && ipGroupId) {
            updateIPGroup({
                id: ipGroupId,
                data: {
                    name: data.name,
                    items: data.items,
                }
            }, {
                onSuccess: () => {
                    onSuccess?.()
                }
            })
        } else {
            createIPGroup({
                name: data.name,
                items: data.items,
            }, {
                onSuccess: () => {
                    // Reset form for create mode
                    form.reset()
                    onSuccess?.()
                }
            })
        }
    }

    return (
        <AnimatedContainer>
            <Form {...form}>
                <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
                    {error && (
                        <Alert variant="destructive">
                            <AlertCircle className="h-4 w-4" />
                            <AlertDescription>{error}</AlertDescription>
                        </Alert>
                    )}

                    <FormField
                        control={form.control}
                        name="name"
                        render={({ field }) => (
                            <FormItem>
                                <FormLabel className="dark:text-shadow-glow-white">{t("ipGroup.form.name")}</FormLabel>
                                <FormControl>
                                    <Input className="dark:text-shadow-glow-white" placeholder={t("ipGroup.form.namePlaceholder")} {...field} />
                                </FormControl>
                                <FormMessage />
                            </FormItem>
                        )}
                    />

                    <div className="space-y-4">
                        <div>
                            <FormLabel className="dark:text-shadow-glow-white">{t("ipGroup.form.ipAddresses")}</FormLabel>
                            <div className="flex mt-2">
                                <Input
                                    placeholder={t("ipGroup.form.ipAddressPlaceholder")}
                                    value={newIpAddress}
                                    onChange={(e) => {
                                        setNewIpAddress(e.target.value)
                                        setIpAddressError(null)
                                    }}
                                    onKeyDown={handleKeyDown}
                                    className="mr-2 dark:text-shadow-glow-white"
                                />
                                <AnimatedButton>
                                    <Button
                                        type="button"
                                        onClick={handleAddIpAddress}
                                        size="icon"
                                        className="dark:text-shadow-glow-white dark:button-neon"
                                    >
                                        <Plus className="h-4 w-4 dark:icon-neon" />
                                    </Button>
                                </AnimatedButton>
                            </div>
                            {ipAddressError && (
                                <p className="text-sm font-medium text-destructive mt-2 dark:text-shadow-glow-white">{ipAddressError}</p>
                            )}
                        </div>

                        <FormField
                            control={form.control}
                            name="items"
                            render={() => (
                                <FormItem>
                                    <div className="rounded-md border dark:border-slate-700">
                                        {ipItems.length === 0 ? (
                                            <div className="py-6 text-center text-muted-foreground dark:text-slate-400 dark:text-shadow-glow-white">
                                                {t("ipGroup.form.noIpAddresses")}
                                            </div>
                                        ) : (
                                            <div className="space-y-3 p-3">
                                                <div className="relative">
                                                    <Search className="absolute left-2 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground dark:text-slate-400" />
                                                    <Input
                                                        placeholder="Search IP addresses..."
                                                        value={searchQuery}
                                                        onChange={(e) => setSearchQuery(e.target.value)}
                                                        className="pl-8 dark:bg-slate-700/50 dark:border-slate-600 dark:text-shadow-glow-white"
                                                    />
                                                </div>
                                                <ScrollArea scrollbarVariant="neon" className="h-[200px] pr-4">
                                                    <ul className="divide-y dark:divide-slate-700">
                                                        {filteredIpItems.map((item) => {
                                                            // 找出原始数组中的真实索引
                                                            const originalIndex = ipItems.indexOf(item)
                                                            return (
                                                                <li
                                                                    key={originalIndex}
                                                                    className="flex items-center justify-between py-2 px-4"
                                                                >
                                                                    <span className="font-mono dark:text-shadow-glow-white">{item}</span>
                                                                    <AnimatedButton>
                                                                        <Button
                                                                            variant="ghost"
                                                                            size="icon"
                                                                            type="button"
                                                                            onClick={() => handleRemoveIpAddress(originalIndex)}
                                                                            className="h-8 w-8 text-destructive hover:text-destructive dark:text-red-500 dark:hover:text-red-400 dark:button-neon"
                                                                        >
                                                                            <Trash2 className="h-4 w-4 dark:icon-neon" />
                                                                        </Button>
                                                                    </AnimatedButton>
                                                                </li>
                                                            )
                                                        })}
                                                    </ul>
                                                </ScrollArea>
                                            </div>
                                        )}
                                    </div>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />
                    </div>

                    <AnimatedButton>
                        <Button
                            type="submit"
                            disabled={isLoading}
                            className="w-full dark:text-shadow-glow-white dark:button-neon"
                        >
                            {isLoading
                                ? t("common.submitting")
                                : mode === 'update'
                                    ? t("common.save")
                                    : t("common.create")}
                        </Button>
                    </AnimatedButton>
                </form>
            </Form>
        </AnimatedContainer>
    )
}