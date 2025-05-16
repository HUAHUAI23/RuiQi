import type React from "react"

interface WaveIconProps {
    width?: number
    height?: number
    color?: string
    className?: string
}

const WaveIcon: React.FC<WaveIconProps> = ({ width = 50, height = 50, color = "#A076F9", className = "" }) => {
    return (
        <div className={`flex items-center justify-center ${className}`}>
            <svg
                width={width}
                height={height}
                viewBox="0 0 50 50"
                xmlns="http://www.w3.org/2000/svg"
                aria-label="Loading animation"
                role="img"
            >
                {/* First bar (shortest) */}
                <rect x="9" width="8" fill={color} rx="2" ry="2">
                    <animate
                        attributeName="y"
                        values="40;30;40"
                        keyTimes="0;0.5;1"
                        dur="1.2s"
                        repeatCount="indefinite"
                        begin="0s"
                    />
                    <animate
                        attributeName="height"
                        values="10;20;10"
                        keyTimes="0;0.5;1"
                        dur="1.2s"
                        repeatCount="indefinite"
                        begin="0s"
                    />
                </rect>

                {/* Second bar (longest) - opposite animation */}
                <rect x="21" width="8" fill={color} rx="2" ry="2">
                    <animate
                        attributeName="y"
                        values="20;35;20"
                        keyTimes="0;0.5;1"
                        dur="1.2s"
                        repeatCount="indefinite"
                        begin="0s"
                    />
                    <animate
                        attributeName="height"
                        values="30;15;30"
                        keyTimes="0;0.5;1"
                        dur="1.2s"
                        repeatCount="indefinite"
                        begin="0s"
                    />
                </rect>

                {/* Third bar (medium) - same as first bar but different height */}
                <rect x="33" width="8" fill={color} rx="2" ry="2">
                    <animate
                        attributeName="y"
                        values="35;25;35"
                        keyTimes="0;0.5;1"
                        dur="1.2s"
                        repeatCount="indefinite"
                        begin="0s"
                    />
                    <animate
                        attributeName="height"
                        values="15;25;15"
                        keyTimes="0;0.5;1"
                        dur="1.2s"
                        repeatCount="indefinite"
                        begin="0s"
                    />
                </rect>
            </svg>
        </div>
    )
}

export default WaveIcon
