import { http } from './client'

export type IntOrderStatus = 'draft' | 'posted' | 'cancelled'

export interface InternalOrderItem {
  id: string
  order_id: string
  product_id: string
  quantity: number
  price: number
  vat_rate: number
  sum: number
}

export interface InternalOrder {
  id: string
  number: string
  doc_date: string
  status: IntOrderStatus
  organization_id?: string
  warehouse_id?: string
  plan_date?: string
  project?: string
  project_name?: string
  comment?: string
  total: number
  shipped_amount: number
  sent_at?: string
  printed_at?: string
  owner_id?: string
  owner_dept?: string
  vat_enabled: boolean
  vat_included: boolean
  posted_at?: string
  cancelled_at?: string
  created_at: string
  updated_at: string
  items?: InternalOrderItem[]
  items_count?: number
  external_id?: string | null
}

export interface IntOrderInput {
  number?: string
  organization_id?: string
  warehouse_id?: string
  plan_date?: string
  project?: string
  project_name?: string
  comment?: string
  vat_enabled?: boolean
  vat_included?: boolean
  items: Array<{
    product_id: string
    quantity: number
    price: number
    vat_rate?: number
  }>
}

export async function listInternalOrders(filters?: { status?: string; warehouse_id?: string }): Promise<InternalOrder[]> {
  const params = new URLSearchParams()
  if (filters?.status) params.set('status', filters.status)
  if (filters?.warehouse_id) params.set('warehouse_id', filters.warehouse_id)
  const qs = params.toString()
  const url = qs ? `/internal-orders?${qs}` : '/internal-orders'
  const { data } = await http.get<{ items: InternalOrder[] }>(url)
  return data.items
}

export async function getInternalOrder(id: string): Promise<InternalOrder> {
  const { data } = await http.get<InternalOrder>(`/internal-orders/${id}`)
  return data
}

export async function createInternalOrder(input: IntOrderInput): Promise<InternalOrder> {
  const { data } = await http.post<InternalOrder>('/internal-orders', input)
  return data
}

export async function updateInternalOrder(id: string, input: IntOrderInput): Promise<InternalOrder> {
  const { data } = await http.put<InternalOrder>(`/internal-orders/${id}`, input)
  return data
}

export async function postInternalOrder(id: string): Promise<InternalOrder> {
  const { data } = await http.post<InternalOrder>(`/internal-orders/${id}/post`)
  return data
}

export async function cancelInternalOrder(id: string): Promise<InternalOrder> {
  const { data } = await http.post<InternalOrder>(`/internal-orders/${id}/cancel`)
  return data
}

export async function deleteInternalOrder(id: string): Promise<void> {
  await http.delete(`/internal-orders/${id}`)
}

export async function nextInternalOrderNumber(): Promise<string> {
  const { data } = await http.get<{ number: string }>('/internal-orders/next-number')
  return data.number
}
export async function printInternalOrder(id: string): Promise<InternalOrder> {
  const { data } = await http.post<InternalOrder>(`/internal-orders/${id}/print`)
  return data
}

export async function sendInternalOrder(id: string): Promise<InternalOrder> {
  const { data } = await http.post<InternalOrder>(`/internal-orders/${id}/send`)
  return data
}
export async function exportInternalOrder(id: string): Promise<{ blob: Blob; filename: string }> {
  const resp = await http.get(`/internal-orders/${id}/export`, { responseType: 'blob' })
  const dispo = (resp.headers['content-disposition'] as string | undefined) || ''
  let filename = `intorder1-${id}.xls`
  const m = /filename="?([^"]+)"?/.exec(dispo)
  if (m && m[1]) filename = m[1]
  return { blob: resp.data as Blob, filename }
}