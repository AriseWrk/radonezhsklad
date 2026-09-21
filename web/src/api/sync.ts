import { http } from './client'

export type SyncStatus = 'queued' | 'running' | 'done' | 'error'

export interface SyncJob {
  id: string
  status: SyncStatus
  started_at: string
  ended_at?: string
  fetched: number
  total: number
  error?: string
  last_line?: string
}

export async function startPullOrders(): Promise<SyncJob> {
  const { data } = await http.post<SyncJob>('/sync/pull-orders')
  return data
}

export async function getSyncJob(id: string): Promise<SyncJob> {
  const { data } = await http.get<SyncJob>(`/sync/jobs/${id}`)
  return data
}