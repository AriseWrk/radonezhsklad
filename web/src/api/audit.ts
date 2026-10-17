import { http } from './client'

export interface AuditLog {
  id: string
  user_id?: string
  user_email?: string
  method: string
  path: string
  resource?: string
  resource_id?: string
  status: number
  request_body?: string
  client_ip?: string
  request_id?: string
  created_at: string
}

export interface AuditResponse {
  items: AuditLog[]
  total: number
  limit: number
  offset: number
}

export interface AuditFilters {
  method?: string
  resource?: string
  date_from?: string
  date_to?: string
  limit?: number
  offset?: number
}

export async function listAudit(filters: AuditFilters = {}): Promise<AuditResponse> {
  const params = new URLSearchParams()
  if (filters.method) params.set('method', filters.method)
  if (filters.resource) params.set('resource', filters.resource)
  if (filters.date_from) params.set('date_from', filters.date_from)
  if (filters.date_to) params.set('date_to', filters.date_to)
  if (filters.limit) params.set('limit', String(filters.limit))
  if (filters.offset) params.set('offset', String(filters.offset))
  const qs = params.toString()
  const url = qs ? `/audit?${qs}` : '/audit'
  const { data } = await http.get<AuditResponse>(url)
  return data
}

export interface AuditLogLine {
  service: string
  line: string
}

export interface AuditLogsResponse {
  kind: 'out' | 'err'
  service: string
  items: AuditLogLine[]
}

export async function listAuditLogs(params: { kind?: 'out' | 'err'; service?: string; lines?: number } = {}): Promise<AuditLogsResponse> {
  const p = new URLSearchParams()
  if (params.kind) p.set('kind', params.kind)
  if (params.service) p.set('service', params.service)
  if (params.lines) p.set('lines', String(params.lines))
  const qs = p.toString()
  const url = '/audit/logs' + (qs ? '?' + qs : '')
  const { data } = await http.get<AuditLogsResponse>(url)
  return data
}
