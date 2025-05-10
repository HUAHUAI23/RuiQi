# RuiQi WAF Frontend

A modern web application interface for RuiQi Web Application Firewall (WAF) system built with React, TypeScript, and Vite.

## Features

- **Authentication System**: Secure login and password reset functionality
- **Dashboard Monitoring**: Real-time monitoring of web application security threats
- **Log Analysis**: Detailed logs for attack detection and protection events
- **Rules Management**: Configure and manage WAF security rules
- **Site Management**: Manage protected websites and domains
- **Certificate Management**: SSL/TLS certificate management
- **Multi-language Support**: Internationalization with both Chinese and English languages

## Tech Stack

- **Core**: React + TypeScript + Vite
- **Styling**: TailwindCSS + Shadcn UI
- **Routing**: React Router v7
- **State Management**: Zustand
- **Form Handling**: React Hook Form + Zod validation
- **Data Fetching**: TanStack Query (React Query v4)
- **Tables**: TanStack Table
- **Internationalization**: react-i18next with i18next-http-backend and i18next-browser-languagedetector
- **Icons**: Lucide React

## Project Structure

```
├── public/                # Static assets
│   └── locales/           # Internationalization files
│       ├── en/            # English translations
│       └── zh/            # Chinese translations
├── src/
│   ├── api/               # API services and request handling
│   ├── assets/            # Project static assets
│   ├── components/        # UI components
│   │   ├── common/        # Common shared components
│   │   ├── layout/        # Layout components
│   │   ├── table/         # Table components (TanStack Table wrappers)
│   │   └── ui/            # Shadcn UI components
│   ├── feature/           # Feature modules
│   │   └── auth/          # Authentication feature components and hooks
│   ├── hooks/             # Custom React hooks
│   ├── lib/               # Utility libraries
│   ├── pages/             # Application pages
│   │   ├── auth/          # Authentication pages
│   │   ├── logs/          # Log analysis pages
│   │   ├── monitor/       # Monitoring dashboard pages
│   │   ├── rule/          # Rules management pages
│   │   └── setting/       # Settings pages
│   ├── routes/            # Routing configuration
│   ├── store/             # Zustand state management
│   ├── types/             # TypeScript type definitions
│   ├── utils/             # Utility functions
│   ├── validation/        # Form validation schemas (Zod)
│   ├── App.tsx            # Main App component
│   ├── i18n.ts            # i18n configuration
│   ├── index.css          # Global styles
│   └── main.tsx           # Application entry point
├── .eslintrc.js           # ESLint configuration
├── components.json        # Shadcn components configuration
├── package.json           # Dependencies and scripts
├── tailwind.config.js     # Tailwind CSS configuration
├── tsconfig.json          # TypeScript configuration
└── vite.config.ts         # Vite configuration
```

## Getting Started

### Prerequisites

- Node.js 22+ 
- pnpm (recommended package manager)

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/HUAHUAI23/RuiQi.git
   cd web
   ```

2. Install dependencies:
   ```bash
   pnpm install
   ```

3. Start the development server:
   ```bash
   pnpm dev
   ```

4. Build for production:
   ```bash
   pnpm build
   ```

## Key Features

### Layout Structure
- Left sidebar for main navigation
- Right content area with breadcrumb for sub-navigation
- Different breadcrumb paths based on selected sidebar navigation

### Authentication
- Login with username/password
- Forced password reset for first-time users
- Protected routes requiring authentication

### Internationalization
- Multi-language support (English and Chinese)
- Default language detection
- Language switcher

## Contributing

1. Follow the project's code structure and naming conventions
2. Ensure code is properly formatted and commented
3. Write tests for new features
4. Create a pull request with a clear description of changes

## License

Copyright © 2025 RuiQi WAF. All rights reserved.
