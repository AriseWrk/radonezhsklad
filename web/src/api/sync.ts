import { http } from './client'

export type SyncStatus = 'queued' | 'running' | 'done' | 'error' | 'cancelled'

export interface SyncJob {
  id: string
  kind: string
  status: SyncStatus
  started_at: string
  ended_at?: string
  fetched: number
  total?: number
  error?: string
  last_line?: string
  triggered_by?: string
}

export interface SyncErrorItem {
  id: string
  entity: string
  local_id?: string
  ms_external_id?: string
  op: string
  attempt: number
  error: string
  created_at: string
  resolved_at?: string
}

export interface AutoSyncStatus {
  enabled: boolean
  interval_s: number
  last_tick?: string | null
  last_error?: string
  tick_count: number
  push_enabled: boolean
}

// --- jobs ---

export async function listSyncJobs(limit = 50): Promise<SyncJob[]> {
  const { data } = await http.get<{ items: SyncJob[] }>('/sync/jobs', { params: { limit } })
  return data.items ?? []
}

export async function getSyncJob(id: string): Promise<SyncJob> {
  const { data } = await http.get<SyncJob>(`/sync/jobs/${id}`)
  return data
}

// --- pull ---

export async function startPullOrders(): Promise<SyncJob> {
  const { data } = await http.post<SyncJob>('/sync/pull-orders')
  return data
}

export async function startPull(kind: string): Promise<SyncJob> {
  const { data } = await http.post<SyncJob>(`/sync/pull/${encodeURIComponent(kind)}`)
  return data
}

// --- retry ---

export async function startRetryPending(): Promise<SyncJob> {
  const { data } = await http.post<SyncJob>('/sync/retry-pending')
  return data
}

export async function retryDocument(id: string): Promise<{ id: string; status: string }> {
  const { data } = await http.post<{ id: string; status: string }>(
    `/sync/retry/document/${id}`,
  )
  return data
}

// --- errors ---

export async function listSyncErrors(limit = 100): Promise<{ items: SyncErrorItem[]; count_open: number }> {
  const { data } = await http.get<{ items: SyncErrorItem[]; count_open: number }>(
    '/sync/errors',
    { params: { limit } },
  )
  return { items: data.items ?? [], count_open: data.count_open ?? 0 }
}

// --- autosync ---

export async function getAutoSyncStatus(): Promise<AutoSyncStatus> {
  const { data } = await http.get<AutoSyncStatus>('/sync/autosync')
  return data
}

export const PULL_KINDS: { kind: string; label: string }[] = [
  { kind: 'counterparties', label: 'Контрагенты' },
  { kind: 'warehouses',     label: 'Склады' },
  { kind: 'organizations',  label: 'Организации' },
  { kind: 'projects',       label: 'Проекты' },
  { kind: 'docs-enter',     label: 'Оприходования' },
  { kind: 'docs-demand',    label: 'Отгрузки' },
  { kind: 'docs-loss',      label: 'Списания' },
  { kind: 'docs-move',      label: 'Перемещения' },
]