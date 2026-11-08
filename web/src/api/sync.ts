import { http } from './client'

export interface SyncSettings {
  enabled: boolean
  interval_minutes: number
  last_cycle_at: string | null
  last_status: string
  last_error: string
}

export interface AutoSyncStatus {
  running: boolean
  tick_count: number
  push_enabled: boolean
  pull_enabled: boolean
}

export async function getSyncSettings(): Promise<SyncSettings> {
  const { data } = await http.get<SyncSettings>('/sync/settings')
  return data
}

export async function setSyncSettings(enabled: boolean): Promise<SyncSettings> {
  const { data } = await http.put<SyncSettings>('/sync/settings', { enabled })
  return data
}

export async function getAutoSyncStatus(): Promise<AutoSyncStatus> {
  const { data } = await http.get<AutoSyncStatus>('/sync/autosync')
  return data
}