import { http } from './client'

// ВНИМАНИЕ: модуль временный. При переходе в прод синхронизация с МойСклад
// удаляется — вместе с этим файлом, кнопкой в DocumentsView и роутом /sync/pull.
export type MspullKind =
  | 'docs-enter'
  | 'docs-demand'
  | 'docs-loss'
  | 'docs-move'
  | 'docs-purchaseorder'
  | 'docs-invoicein'
  | 'docs-purchasereturn'
  | 'docs-facturein'
  | 'counterparties'
  | 'warehouses'
  | 'organizations'
  | 'projects'

export interface SyncJobShort {
  id: string
  kind: string
  status: 'queued' | 'running' | 'done' | 'error'
  error?: string | null
}

export async function startPull(kind: MspullKind): Promise<SyncJobShort> {
  const { data } = await http.post<SyncJobShort>('/sync/pull/' + kind)
  return data
}

export async function getSyncJob(id: string): Promise<SyncJobShort> {
  const { data } = await http.get<SyncJobShort>('/sync/jobs/' + id)
  return data
}

export async function pullDocsForToday(onProgress?: (kind: string) => void): Promise<void> {
  const kinds: MspullKind[] = ['docs-enter', 'docs-demand', 'docs-loss', 'docs-move']
  for (const k of kinds) {
    onProgress?.(k)
    const job = await startPull(k)
    for (let i = 0; i < 300; i++) {
      await new Promise((r) => setTimeout(r, 500))
      const j = await getSyncJob(job.id)
      if (j.status === 'done') break
      if (j.status === 'error') throw new Error('MS pull ' + k + ' failed: ' + (j.error ?? ''))
    }
  }
}