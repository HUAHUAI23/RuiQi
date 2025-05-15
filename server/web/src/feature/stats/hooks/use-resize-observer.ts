import { useEffect, useState, useRef } from 'react'

interface ResizeObserverEntry {
    contentRect: DOMRectReadOnly
}

interface Size {
    width: number | undefined
    height: number | undefined
}

export function useResizeObserver(
    ref: React.RefObject<HTMLElement>
): Size {
    const [size, setSize] = useState<Size>({
        width: undefined,
        height: undefined,
    })

    // ResizeObserver实例的引用
    const observerRef = useRef<ResizeObserver | null>(null)

    useEffect(() => {
        // 如果DOM元素不存在，直接返回
        if (!ref.current) return

        // 清理旧的observer
        if (observerRef.current) {
            observerRef.current.disconnect()
        }

        // 创建一个新的ResizeObserver
        const observer = new ResizeObserver((entries: ResizeObserverEntry[]) => {
            const { width, height } = entries[0].contentRect
            setSize({ width, height })
        })

        // 开始观察目标元素
        observer.observe(ref.current)
        observerRef.current = observer

        // 清理函数
        return () => {
            if (observerRef.current) {
                observerRef.current.disconnect()
                observerRef.current = null
            }
        }
    }, [ref])

    return size
}