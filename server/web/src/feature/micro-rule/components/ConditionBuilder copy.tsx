import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Plus, Trash2, ChevronDown, ChevronRight } from "lucide-react"
import { useTranslation } from "react-i18next"
import type { LogicalOperator, TargetType, MatchType, MicroRuleCreateRequest } from "@/types/rule"
import { cn } from "@/lib/utils"
import { TARGET_MATCH_TYPES } from "@/types/rule"
import { FormField, FormItem, FormLabel, FormControl, FormMessage } from "@/components/ui/form"
import type { UseFormReturn } from "react-hook-form"
import { Badge } from "@/components/ui/badge"
import { TFunction } from "i18next"

interface ConditionBuilderProps {
    form: UseFormReturn<MicroRuleCreateRequest>
    path: string
    onRemove?: () => void
    isRoot?: boolean
    showConnector?: boolean
    parentOperator?: LogicalOperator
    isLast?: boolean
}

// Extracted reusable components
const ConnectorLine = ({ className = "" }: { className?: string }) => (
    <div className={cn("border-dashed border-gray-300 dark:border-gray-700", className)} />
)

const OperatorBadge = ({ operator, onClick, className = "" }: { operator: LogicalOperator, onClick?: () => void, className?: string }) => {
    const badgeStyles =
        operator === "AND"
            ? "bg-blue-50 dark:bg-blue-900/20 text-blue-600 dark:text-blue-400 border-blue-200 dark:border-blue-800"
            : "bg-orange-50 dark:bg-orange-900/20 text-orange-600 dark:text-orange-400 border-orange-200 dark:border-orange-800"

    return (
        <Badge variant="outline" className={cn(badgeStyles, className)} onClick={onClick}>
            {operator}
        </Badge>
    )
}

const ActionButtons = ({ addSimpleCondition, addCompositeCondition, t }: { addSimpleCondition: () => void, addCompositeCondition: () => void, t: TFunction }) => (
    <div className="flex flex-wrap gap-2 mt-6 pt-2 relative z-10">
        <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={addSimpleCondition}
            className="text-teal-600 dark:text-teal-400 border-teal-300 dark:border-teal-800"
        >
            <Plus className="h-4 w-4 mr-1" />
            {t("microRule.condition.addSimpleCondition")}
        </Button>
        <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={addCompositeCondition}
            className="text-blue-600 dark:text-blue-400 border-blue-300 dark:border-blue-800"
        >
            <Plus className="h-4 w-4 mr-1" />
            {t("microRule.condition.addCompositeCondition")}
        </Button>
    </div>
)

export function ConditionBuilder({
    form,
    path,
    onRemove,
    isRoot = false,
    showConnector = false,
    parentOperator = "AND",
    isLast = false,
}: ConditionBuilderProps) {
    const { t } = useTranslation()
    const [expanded, setExpanded] = useState(true)

    // Common styles
    const shadowTextStyles = "dark:text-shadow-glow-white"

    // Get current condition type - use any to bypass TypeScript's type checking for dynamic paths
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const conditionType = (form as any).watch(`${path}.type`) || "simple"

    // If composite condition, get operator and child conditions
    const operator = conditionType === "composite"
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        ? ((form as any).watch(`${path}.operator`) || "AND") as LogicalOperator
        : "AND"

    const conditions = conditionType === "composite"
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        ? ((form as any).watch(`${path}.conditions`) || [])
        : []

    // If simple condition, get target and match type
    const target = conditionType === "simple"
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        ? ((form as any).watch(`${path}.target`) || "source_ip") as TargetType
        : "source_ip"

    // Get available match types
    const availableMatchTypes = target ? TARGET_MATCH_TYPES[target] || [] : []

    // Add simple condition
    const addSimpleCondition = () => {
        if (conditionType !== "composite") return

        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const currentConditions = (form as any).getValues(`${path}.conditions`) || []
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        form.setValue(`${path}.conditions` as any, [
            ...currentConditions,
            {
                type: "simple",
                target: "source_ip",
                match_type: "equal",
                match_value: "",
            },
        ])
    }

    // Add composite condition
    const addCompositeCondition = () => {
        if (conditionType !== "composite") return

        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const currentConditions = (form as any).getValues(`${path}.conditions`) || []
        // Use opposite operator for easier complex logic building
        const newOperator = operator === "AND" ? "OR" : "AND"

        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        form.setValue(`${path}.conditions` as any, [
            ...currentConditions,
            {
                type: "composite",
                operator: newOperator,
                conditions: [
                    {
                        type: "simple",
                        target: "source_ip",
                        match_type: "equal",
                        match_value: "",
                    },
                ],
            },
        ])
    }

    // Toggle operator
    const toggleOperator = () => {
        if (conditionType !== "composite") return
        const newOperator = operator === "AND" ? "OR" : "AND"
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        form.setValue(`${path}.operator` as any, newOperator)
    }

    // Remove child condition
    const removeCondition = (index: number) => {
        if (conditionType !== "composite") return

        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const currentConditions = [...((form as any).getValues(`${path}.conditions`) || [])]

        // If this is the last child condition and not root, remove entire composite condition
        if (currentConditions.length === 1 && !isRoot) {
            onRemove && onRemove()
            return
        }

        currentConditions.splice(index, 1)
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        form.setValue(`${path}.conditions` as any, currentConditions)
    }

    // Render simple condition
    if (conditionType === "simple") {
        return (
            <div className="relative flex mb-4">
                {/* Vertical connector line */}
                {showConnector && !isLast && <ConnectorLine className="absolute left-6 top-10 bottom-0 w-[1px] border-l z-0" />}

                {/* Operator connector */}
                {showConnector && (
                    <div className="flex items-center h-10 mr-2 z-10">
                        <OperatorBadge operator={parentOperator} className="h-8 rounded-md mr-2 w-12 justify-center" />
                        <ConnectorLine className="w-4 h-[1px] border-t" />
                    </div>
                )}

                <div className="grid grid-cols-3 gap-4 p-4 border rounded-md bg-gray-50 dark:bg-gray-800/40 dark:border-gray-700 flex-1 z-10">
                    {/* Match target */}
                    <div>
                        <FormField
                            control={form.control}
                            // eslint-disable-next-line @typescript-eslint/no-explicit-any
                            name={`${path}.target` as any}
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel className={shadowTextStyles + " text-xs"}>{t("microRule.condition.target")}</FormLabel>
                                    <Select
                                        onValueChange={(value: TargetType) => {
                                            field.onChange(value)
                                            // Reset match type
                                            // eslint-disable-next-line @typescript-eslint/no-explicit-any
                                            form.setValue(`${path}.match_type` as any, TARGET_MATCH_TYPES[value][0])
                                        }}
                                        defaultValue={field.value}
                                        value={field.value}
                                    >
                                        <FormControl>
                                            <SelectTrigger className={shadowTextStyles}>
                                                <SelectValue placeholder={t("microRule.condition.selectTarget")} />
                                            </SelectTrigger>
                                        </FormControl>
                                        <SelectContent>
                                            <SelectItem value="source_ip">{t("microRule.condition.sourceIp")}</SelectItem>
                                            <SelectItem value="url">{t("microRule.condition.url")}</SelectItem>
                                            <SelectItem value="path">{t("microRule.condition.path")}</SelectItem>
                                        </SelectContent>
                                    </Select>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />
                    </div>

                    {/* Match type */}
                    <div>
                        <FormField
                            control={form.control}
                            // eslint-disable-next-line @typescript-eslint/no-explicit-any
                            name={`${path}.match_type` as any}
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel className={shadowTextStyles + " text-xs"}>
                                        {t("microRule.condition.matchType")} <span className="text-red-500">*</span>
                                    </FormLabel>
                                    <Select onValueChange={field.onChange} defaultValue={field.value} value={field.value}>
                                        <FormControl>
                                            <SelectTrigger className={shadowTextStyles}>
                                                <SelectValue placeholder={t("microRule.condition.selectMatchType")} />
                                            </SelectTrigger>
                                        </FormControl>
                                        <SelectContent>
                                            {availableMatchTypes.map((type) => (
                                                <SelectItem key={type} value={type}>
                                                    {t(`microRule.matchTypes.${type}`)}
                                                </SelectItem>
                                            ))}
                                        </SelectContent>
                                    </Select>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />
                    </div>

                    {/* Match value and delete button */}
                    <div className="flex gap-2">
                        <div className="flex-1">
                            <FormField
                                control={form.control}
                                // eslint-disable-next-line @typescript-eslint/no-explicit-any
                                name={`${path}.match_value` as any}
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel className={shadowTextStyles + " text-xs"}>
                                            {t("microRule.condition.matchValue")} <span className="text-red-500">*</span>
                                        </FormLabel>
                                        <FormControl>
                                            <Input
                                                className={shadowTextStyles}
                                                placeholder={
                                                    target === "source_ip" ? "e.g: 192.168.10.10" : t("microRule.condition.enterMatchValue")
                                                }
                                                {...field}
                                            />
                                        </FormControl>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />
                        </div>
                        {!isRoot && (
                            <Button type="button" variant="ghost" size="icon" onClick={onRemove} className="h-8 w-8 p-0 mt-6">
                                <Trash2 className="h-4 w-4 text-red-500" />
                            </Button>
                        )}
                    </div>
                </div>
            </div>
        )
    }

    // Render composite condition
    return (
        <div className={`relative mb-6 ${showConnector ? "ml-6" : ""}`}>
            {/* Operator connector */}
            {showConnector && (
                <div className="absolute -left-16 top-4 z-10 flex items-center">
                    <OperatorBadge operator={parentOperator} className="h-8 rounded-md w-12 justify-center" />
                    <ConnectorLine className="w-4 h-[1px] border-t" />
                </div>
            )}

            {/* Vertical connector line - fixed to not overlap with buttons */}
            {conditions.length > 1 && expanded && (
                <ConnectorLine className="absolute left-6 top-16 bottom-[52px] w-[1px] border-l z-0" />
            )}

            <div className="p-4 border rounded-md border-dashed border-gray-300 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/40 z-10">
                <div className="flex items-center justify-between mb-4">
                    <div className="flex items-center">
                        <Button
                            type="button"
                            variant="ghost"
                            size="sm"
                            className={`p-1 mr-2 ${shadowTextStyles}`}
                            onClick={() => setExpanded(!expanded)}
                        >
                            {expanded ? <ChevronDown className="h-4 w-4" /> : <ChevronRight className="h-4 w-4" />}
                        </Button>
                        <div className="flex items-center gap-2">
                            <span className={`text-sm font-medium ${shadowTextStyles}`}>
                                {t("microRule.condition.conditionGroup")}
                            </span>
                            <OperatorBadge operator={operator} onClick={toggleOperator} className="h-6 px-2 font-medium" />
                        </div>
                    </div>

                    {!isRoot && (
                        <Button variant="ghost" size="sm" onClick={onRemove} className="h-8 w-8 p-0">
                            <Trash2 className="h-4 w-4 text-red-500" />
                        </Button>
                    )}
                </div>

                {expanded && (
                    <>
                        {/* Child conditions list */}
                        <div className="pl-10 mt-4 space-y-4 relative">
                            {conditions.map((condition, index) => (
                                <ConditionBuilder
                                    key={index}
                                    form={form}
                                    path={`${path}.conditions.${index}`}
                                    onRemove={() => removeCondition(index)}
                                    showConnector={index > 0}
                                    parentOperator={operator}
                                    isLast={index === conditions.length - 1}
                                />
                            ))}
                        </div>

                        {/* Action buttons - moved outside the connector's range */}
                        <ActionButtons
                            addSimpleCondition={addSimpleCondition}
                            addCompositeCondition={addCompositeCondition}
                            t={t}
                        />
                    </>
                )}
            </div>
        </div>
    )
}
