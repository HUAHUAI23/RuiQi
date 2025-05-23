export interface AppConfig {
    name: string
    directives: string
    logLevel?: string
    logFormat?: string
    logFile?: string
    transactionTTL?: number
}

export interface LimitConfig {
    enabled: boolean
    threshold: number
    statDuration: number
    blockDuration: number
    burstCount: number
    paramsCapacity: number
}

export interface FlowControlConfig {
    visitLimit: LimitConfig
    attackLimit: LimitConfig
    errorLimit: LimitConfig
}

export interface EngineConfig {
    bind: string
    useBuiltinRules: boolean
    appConfig: AppConfig[]
    flowController: FlowControlConfig
}

export interface HaproxyConfig {
    thread: number
    configBaseDir: string
    haproxyBin: string
    backupsNumber?: number
    spoeAgentAddr?: string
    spoeAgentPort?: number
}

export interface ConfigResponse {
    id: string
    name: string
    isDebug: boolean
    isResponseCheck: boolean
    engine: EngineConfig
    haproxy: HaproxyConfig
    createdAt: string
    updatedAt: string
}

export interface ConfigPatchRequest {
    name?: string
    isDebug?: boolean
    isResponseCheck?: boolean
    engine?: {
        bind?: string
        useBuiltinRules?: boolean
        appConfig?: {
            name?: string
            directives?: string
            logLevel?: string
            logFormat?: string
            logFile?: string
            transactionTTL?: number
        }[]
        flowController?: {
            visitLimit?: Partial<LimitConfig>
            attackLimit?: Partial<LimitConfig>
            errorLimit?: Partial<LimitConfig>
        }
    }
    haproxy?: {
        thread?: number
        configBaseDir?: string
        haproxyBin?: string
        backupsNumber?: number
        spoeAgentAddr?: string
        spoeAgentPort?: number
    }
}

