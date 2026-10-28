// Shapes returned by the Go API

export interface StoredGroup {
  id: string
  name: string
  managed: boolean
  syncedAt?: string
}

export interface Member {
  id: string
  number: string
  pushName?: string
  since: string
}

export interface ReturnedMember {
  id: string
  number: string
  pushName?: string
  exits: number
  lastExit: string
  lastJoin: string
}

export interface RemovalJob {
  id: number
  createdAt: string
  requestedBy?: string
  pending: number
  ok: number
  failed: number
  canceled: number
}

export interface RemovalItem {
  id: string
  status: 'pending' | 'ok' | 'failed' | 'canceled'
  error?: string
  processedAt?: string
}

export interface RemovalDetail {
  jobId: number
  createdAt: string
  requestedBy?: string | null
  counts: Partial<Record<RemovalItem['status'], number>>
  items: RemovalItem[]
}

export interface Skipped {
  id: string
  reason: string
}

export interface CreateRemovalResponse {
  jobId: number
  queued: number
  skipped: Skipped[]
}
