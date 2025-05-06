import type { Config } from "tailwindcss"
import tailwindcssAnimate from "tailwindcss-animate"

const config = {
    darkMode: ["class"],
    content: [
        "./pages/**/*.{ts,tsx}",
        "./components/**/*.{ts,tsx}",
        "./app/**/*.{ts,tsx}",
        "./src/**/*.{ts,tsx}",
        "*.{js,ts,jsx,tsx,mdx}",
    ],
    prefix: "",
    theme: {
        extend: {
            container: {
                center: true,
                padding: "2rem",
                screens: {
                    "2xl": "1400px",
                },
            },
            borderRadius: {
                lg: 'var(--radius)',
                md: 'calc(var(--radius) - 2px)',
                sm: 'calc(var(--radius) - 4px)'
            },
            colors: {
                background: 'hsl(var(--background))',
                foreground: 'hsl(var(--foreground))',
                card: {
                    DEFAULT: 'hsl(var(--card))',
                    foreground: 'hsl(var(--card-foreground))'
                },
                popover: {
                    DEFAULT: 'hsl(var(--popover))',
                    foreground: 'hsl(var(--popover-foreground))'
                },
                primary: {
                    DEFAULT: 'hsl(var(--primary))',
                    foreground: 'hsl(var(--primary-foreground))',
                },
                secondary: {
                    DEFAULT: 'hsl(var(--secondary))',
                    foreground: 'hsl(var(--secondary-foreground))'
                },
                muted: {
                    DEFAULT: 'hsl(var(--muted))',
                    foreground: 'hsl(var(--muted-foreground))'
                },
                accent: {
                    DEFAULT: 'hsl(var(--accent))',
                    foreground: 'hsl(var(--accent-foreground))',
                },
                destructive: {
                    DEFAULT: 'hsl(var(--destructive))',
                    foreground: 'hsl(var(--destructive-foreground))'
                },
                border: 'hsl(var(--border))',
                input: 'hsl(var(--input))',
                ring: 'hsl(var(--ring))',
                chart: {
                    '1': 'hsl(var(--chart-1))',
                    '2': 'hsl(var(--chart-2))',
                    '3': 'hsl(var(--chart-3))',
                    '4': 'hsl(var(--chart-4))',
                    '5': 'hsl(var(--chart-5))'
                },

                iconStroke: {
                    light: '##8861DB',      // 亮色背景上使用
                    DEFAULT: '#8861DB',    // 默认颜色
                    dark: '#8861DB',       // 暗色背景上使用
                    accent: '#8861DB',     // 强调色
                },
            },
            keyframes: {
                "accordion-down": {
                    from: { height: "0" },
                    to: { height: "var(--radix-accordion-content-height)" },
                },
                "accordion-up": {
                    from: { height: "var(--radix-accordion-content-height)" },
                    to: { height: "0" },
                },
                "icon-shake": {
                    "0%": { transform: "rotate(0deg)" },
                    "25%": { transform: "rotate(-12deg)" },
                    "50%": { transform: "rotate(10deg)" },
                    "75%": { transform: "rotate(-6deg)" },
                    "85%": { transform: "rotate(3deg)" },
                    "92%": { transform: "rotate(-2deg)" },
                    "100%": { transform: "rotate(0deg)" },
                },
                "float": {
                    "0%, 100%": { transform: "translateY(0) scale(1)" },
                    "50%": { transform: "translateY(-20px) scale(1.05)" },
                },
                "float-reverse": {
                    "0%, 100%": { transform: "translateY(0) scale(1)" },
                    "50%": { transform: "translateY(20px) scale(1.05)" },
                },
                "pulse-glow": {
                    "0%, 100%": { opacity: "0.6", transform: "scale(1)" },
                    "50%": { opacity: "0.8", transform: "scale(1.1)" },
                },
                "fade-in-up": {
                    "0%": { opacity: "0", transform: "translateY(20px)" },
                    "100%": { opacity: "1", transform: "translateY(0)" },
                },
                "wiggle": {
                    "0%, 100%": { transform: "rotate(-2deg)" },
                    "50%": { transform: "rotate(2deg)" },
                },
                "gradient-shift": {
                    "0%": { backgroundPosition: "0% 50%" },
                    "50%": { backgroundPosition: "100% 50%" },
                    "100%": { backgroundPosition: "0% 50%" },
                },
                "aurora": {
                    "0%": { filter: "hue-rotate(0deg) brightness(1) saturate(1.5)" },
                    "33%": { filter: "hue-rotate(60deg) brightness(1.1) saturate(1.8)" },
                    "66%": { filter: "hue-rotate(180deg) brightness(1.05) saturate(1.6)" },
                    "100%": { filter: "hue-rotate(360deg) brightness(1) saturate(1.5)" },
                },
                "border-pulsate": {
                    "0%": { boxShadow: '0 0 5px rgba(var(--primary-rgb), 0.2), 0 0 10px rgba(var(--primary-rgb), 0.1)' },
                    "100%": { boxShadow: '0 0 10px rgba(var(--primary-rgb), 0.6), 0 0 20px rgba(var(--primary-rgb), 0.4), 0 0 30px rgba(var(--primary-rgb), 0.2)' }
                },
                // 滚动条发光
                "scrollbar-glow": {
                    "0%": { boxShadow: "0 0 4px rgba(var(--primary-rgb), 0.3), 0 0 8px rgba(var(--primary-rgb), 0.1)" },
                    "50%": { boxShadow: "0 0 8px rgba(var(--primary-rgb), 0.6), 0 0 16px rgba(var(--primary-rgb), 0.3)" },
                    "100%": { boxShadow: "0 0 4px rgba(var(--primary-rgb), 0.3), 0 0 8px rgba(var(--primary-rgb), 0.1)" }
                },
                // 滚动条霓虹灯效果
                "scrollbar-neon": {
                    "0%": { boxShadow: "0 0 5px rgba(var(--primary-rgb), 0.4), 0 0 10px rgba(var(--primary-rgb), 0.2)" },
                    "50%": { boxShadow: "0 0 10px rgba(var(--primary-rgb), 0.7), 0 0 20px rgba(var(--primary-rgb), 0.4)" },
                    "100%": { boxShadow: "0 0 5px rgba(var(--primary-rgb), 0.4), 0 0 10px rgba(var(--primary-rgb), 0.2)" }
                },
                "sidebar-neon": {
                    "0%": { boxShadow: "inset 0 0 20px rgba(var(--primary-rgb), 0.2), 0 0 15px rgba(var(--primary-rgb), 0.1)" },
                    "50%": { boxShadow: "inset 0 0 30px rgba(var(--primary-rgb), 0.4), 0 0 25px rgba(var(--primary-rgb), 0.2)" },
                    "100%": { boxShadow: "inset 0 0 20px rgba(var(--primary-rgb), 0.2), 0 0 15px rgba(var(--primary-rgb), 0.1)" }
                },
                "sidebar-neon-gradient": {
                    "0%": { backgroundPosition: "0% 0%" },
                    "50%": { backgroundPosition: "100% 100%" },
                    "100%": { backgroundPosition: "0% 0%" }
                },
                // logo 弹跳
                "logo-pulse": {
                    "0%, 100%": { opacity: "0.9", transform: "scale(0.95)" },
                    "50%": { opacity: "1", transform: "scale(1.05)" },
                },
                'neon-pulse': {
                    '0%': {
                        textShadow: '0 0 7px rgba(var(--primary-rgb), 0.6), 0 0 10px rgba(var(--primary-rgb), 0.4), 0 0 15px rgba(var(--primary-rgb), 0.2)'
                    },
                    '100%': {
                        textShadow: '0 0 10px rgba(var(--primary-rgb), 0.9), 0 0 20px rgba(var(--primary-rgb), 0.6), 0 0 30px rgba(var(--primary-rgb), 0.3)'
                    }
                },
                'text-flicker': {
                    '0%, 19.999%, 22%, 62.999%, 64%, 64.999%, 70%, 100%': {
                        opacity: '1',
                        textShadow: '0 0 4px rgba(var(--primary-rgb), 0.5), 0 0 11px rgba(var(--primary-rgb), 0.3), 0 0 19px rgba(var(--primary-rgb), 0.2)'
                    },
                    '20%, 21.999%, 63%, 63.999%, 65%, 69.999%': {
                        opacity: '0.9',
                        textShadow: '0 0 4px rgba(var(--primary-rgb), 0.4), 0 0 10px rgba(var(--primary-rgb), 0.1)'
                    }
                },
                'border-neon-flow': {
                    '0%': { backgroundPosition: '0% 0%' },
                    '100%': { backgroundPosition: '100% 100%' }
                }
            },
            animation: {
                "accordion-down": "accordion-down 0.2s ease-out",
                "accordion-up": "accordion-up 0.2s ease-out",
                "icon-shake": "icon-shake 0.7s ease-out",
                "float": "float 8s ease-in-out infinite",
                "float-reverse": "float-reverse 9s ease-in-out infinite",
                "pulse-glow": "pulse-glow 4s ease-in-out infinite",
                "fade-in-up": "fade-in-up 0.5s ease-out",
                "wiggle": "wiggle 1s ease-in-out infinite",
                "gradient-shift": "gradient-shift 8s ease infinite",
                "aurora": "aurora 20s ease infinite",
                "logo-pulse": "logo-pulse 1.5s infinite ease-in-out",
                "border-pulsate": "border-pulsate 1.5s infinite alternate",
                "scrollbar-glow": "scrollbar-glow 1.5s infinite alternate",
                "scrollbar-neon": "scrollbar-neon 2s ease-in-out infinite",
                "sidebar-neon-glow": "sidebar-neon 4s ease-in-out infinite",
                "sidebar-neon-flow": "sidebar-neon-gradient 15s ease infinite, aurora 20s ease infinite",
                "neon-pulse": "neon-pulse 3s infinite alternate",
                "text-flicker": "text-flicker 5s infinite alternate",
                "border-neon-flow": "border-neon-flow 8s infinite linear",
            },
        },
    },
    plugins: [
        tailwindcssAnimate,
        function ({ addUtilities }) {
            const newUtilities = {
                // 霓虹灯文本
                '.text-shadow-glow-purple': {
                    textShadow: '0 0 8px rgba(232,223,255,0.8), 0 0 12px rgba(232,223,255,0.4)',
                },
                '.text-shadow-glow-blue': {
                    textShadow: '0 0 8px rgba(142,212,255,0.8), 0 0 12px rgba(142,212,255,0.4)',
                },
                '.text-shadow-glow-white': {
                    textShadow: '0 0 8px rgba(255,255,255,0.8), 0 0 12px rgba(255,255,255,0.4)',
                },
                '.text-shadow-primary': {
                    textShadow: '0 0 6px rgba(var(--primary-rgb), 0.7), 0 0 12px rgba(var(--primary-rgb), 0.4), 0 0 18px rgba(var(--primary-rgb), 0.2)',
                },
                '.text-shadow-primary-sm': {
                    textShadow: '0 0 4px rgba(var(--primary-rgb), 0.5), 0 0 8px rgba(var(--primary-rgb), 0.25)',
                },
                '.text-shadow-primary-bold': {
                    textShadow: '0 0 3px rgba(var(--primary-rgb), 0.9), 0 0 7px rgba(var(--primary-rgb), 0.6), 0 0 12px rgba(var(--primary-rgb), 0.3)',
                },
                '.text-shadow-primary-subtle': {
                    textShadow: '0 0 3px rgba(var(--primary-rgb), 0.4)',
                },
                '.text-shadow-none': {
                    textShadow: 'none',
                },

                // New neon text animations
                '.text-neon-pulse': {
                    animation: 'neon-pulse 3s infinite alternate'
                },
                '.text-neon-flicker': {
                    animation: 'text-flicker 5s infinite alternate'
                },

                // 霓虹灯边框效果
                '.border-glow-purple': {
                    boxShadow: '0 0 5px rgba(232,223,255,0.5), 0 0 10px rgba(232,223,255,0.3), 0 0 15px rgba(232,223,255,0.1)',
                    border: '1px solid rgba(232,223,255,0.6)',
                },
                '.border-glow-blue': {
                    boxShadow: '0 0 5px rgba(142,212,255,0.5), 0 0 10px rgba(142,212,255,0.3), 0 0 15px rgba(142,212,255,0.1)',
                    border: '1px solid rgba(142,212,255,0.6)',
                },
                '.border-glow-white': {
                    boxShadow: '0 0 5px rgba(255,255,255,0.5), 0 0 10px rgba(255,255,255,0.3), 0 0 15px rgba(255,255,255,0.1)',
                    border: '1px solid rgba(255,255,255,0.6)',
                },
                '.border-glow-primary': {
                    boxShadow: '0 0 5px rgba(var(--primary-rgb), 0.5), 0 0 10px rgba(var(--primary-rgb), 0.3), 0 0 15px rgba(var(--primary-rgb), 0.1)',
                    border: '1px solid rgba(var(--primary-rgb), 0.6)',
                },
                '.border-glow-primary-intense': {
                    boxShadow: '0 0 8px rgba(var(--primary-rgb), 0.7), 0 0 15px rgba(var(--primary-rgb), 0.5), 0 0 25px rgba(var(--primary-rgb), 0.3)',
                    border: '1px solid rgba(var(--primary-rgb), 0.8)',
                },
                '.border-glow-primary-subtle': {
                    boxShadow: '0 0 4px rgba(var(--primary-rgb), 0.3), 0 0 8px rgba(var(--primary-rgb), 0.1)',
                    border: '1px solid rgba(var(--primary-rgb), 0.4)',
                },
                '.border-glow-pulsate': {
                    animation: 'border-pulsate 1.5s infinite alternate',
                },
                '.border-glow-none': {
                    boxShadow: 'none',
                    border: 'none',
                },

                // New enhanced neon border effects
                '.border-neon-pulse': {
                    boxShadow: '0 0 5px rgba(var(--primary-rgb), 0.6), 0 0 10px rgba(var(--primary-rgb), 0.4), 0 0 15px rgba(var(--primary-rgb), 0.2)',
                    border: '1px solid rgba(var(--primary-rgb), 0.8)',
                    animation: 'border-pulsate 2.5s infinite alternate'
                },
                '.border-neon-strong': {
                    boxShadow: '0 0 10px rgba(var(--primary-rgb), 0.8), 0 0 20px rgba(var(--primary-rgb), 0.6), 0 0 30px rgba(var(--primary-rgb), 0.4), 0 0 40px rgba(var(--primary-rgb), 0.2)',
                    border: '1px solid rgba(var(--primary-rgb), 0.9)'
                },
                '.border-neon-flow': {
                    borderImageSource: 'linear-gradient(45deg, rgba(var(--primary-rgb), 0.8), rgba(var(--primary-rgb), 0.2), rgba(var(--primary-rgb), 0.8))',
                    borderImageSlice: '1',
                    animation: 'border-neon-flow 8s infinite linear',
                    backgroundSize: '200% 200%'
                },
            }
            addUtilities(newUtilities)
        },
    ],
} satisfies Config

export default config
