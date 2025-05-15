import { Button } from "@/components/ui/button"
import { TimeRange } from "@/types/stats"
import { useTranslation } from "react-i18next"

interface TimeRangeSelectorProps {
    value: TimeRange
    onChange: (value: TimeRange) => void
}

export function TimeRangeSelector({ value, onChange }: TimeRangeSelectorProps) {
    const { t } = useTranslation()

    const timeRanges: { value: TimeRange; label: string }[] = [
        { value: '24h', label: t('stats.timeRange.24h') },
        { value: '7d', label: t('stats.timeRange.7d') },
        { value: '30d', label: t('stats.timeRange.30d') },
    ]

    return (
        <div className="flex items-center space-x-2">
            {timeRanges.map((range) => (
                <Button
                    key={range.value}
                    variant={value === range.value ? "default" : "outline"}
                    size="sm"
                    onClick={() => onChange(range.value)}
                    className="dark:text-shadow-glow-white"
                >
                    {range.label}
                </Button>
            ))}
        </div>
    )
}