export interface BlockedIPRecord {
    ip: string
    reason: string
    requestUri: string
    blockedAt: string
    blockedUntil: string
    isActive: boolean
    remainingTTL: number
}

export interface BlockedIPListRequest {
    page?: number
    size?: number
    ip?: string
    reason?: string
    status?: 'active' | 'expired' | 'all'
    sortBy?: 'blocked_at' | 'blocked_until' | 'ip'
    sortDir?: 'asc' | 'desc'
}

export interface BlockedIPListResponse {
    total: number
    items: BlockedIPRecord[]
    page: number
    size: number
    pages: number
}

export interface BlockedIPHourlyStats {
    hour: string
    count: number
}

export interface BlockedIPStatsResponse {
    totalBlocked: number
    activeBlocked: number
    expiredBlocked: number
    reasonStats: Record<string, number>
    last24HourStats: BlockedIPHourlyStats[]
}

export interface BlockedIPCleanupResponse {
    deletedCount: number
    message: string
}
