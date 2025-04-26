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
                float: {
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
                wiggle: {
                    "0%, 100%": { transform: "rotate(-2deg)" },
                    "50%": { transform: "rotate(2deg)" },
                },
                "gradient-shift": {
                    "0%": { backgroundPosition: "0% 50%" },
                    "50%": { backgroundPosition: "100% 50%" },
                    "100%": { backgroundPosition: "0% 50%" },
                },
                aurora: {
                    "0%": { filter: "hue-rotate(0deg) brightness(1) saturate(1.5)" },
                    "33%": { filter: "hue-rotate(60deg) brightness(1.1) saturate(1.8)" },
                    "66%": { filter: "hue-rotate(180deg) brightness(1.05) saturate(1.6)" },
                    "100%": { filter: "hue-rotate(360deg) brightness(1) saturate(1.5)" },
                },
                "logo-pulse": {
                    "0%, 100%": { opacity: "0.9", transform: "scale(0.95)" },
                    "50%": { opacity: "1", transform: "scale(1.05)" },
                },
            },
            animation: {
                "accordion-down": "accordion-down 0.2s ease-out",
                "accordion-up": "accordion-up 0.2s ease-out",
                "icon-shake": "icon-shake 0.7s ease-out",
                float: "float 8s ease-in-out infinite",
                "float-reverse": "float-reverse 9s ease-in-out infinite",
                "pulse-glow": "pulse-glow 4s ease-in-out infinite",
                "fade-in-up": "fade-in-up 0.5s ease-out",
                wiggle: "wiggle 1s ease-in-out infinite",
                "gradient-shift": "gradient-shift 8s ease infinite",
                aurora: "aurora 20s ease infinite",
                "logo-pulse": "logo-pulse 1.5s infinite ease-in-out",
            },
        },
    },
    plugins: [tailwindcssAnimate],
} satisfies Config

export default config
