import { http } from './client'

export type DocType = 'receipt' | 'shipment' | 'transfer' | 'writeoff' | 'inventory' | 'purchase_order' | 'invoice_in' | 'purchase_return' | 'facture_in'
export type DocStatus = 'draft' | 'posted' | 'cancelled'

export interface DocItem {
  id: string
  document_id: string
  product_id: string
  product_name?: string
  product_sku?: string
  quantity: number
  price: number
}

export interface Document {
  id: string
  type: DocType
  number: string
  status: DocStatus
  warehouse_id?: string
  target_warehouse_id?: string
  supplier_id?: string
  organization_id?: string
  incoming_number?: string
  incoming_date?: string
  paid_amount: number
  external_id?: string | null
  printed_at?: string
  sent_at?: string
  comment?: string
  created_at: string
  updated_at: string
  posted_at?: string
  cancelled_at?: string
  items?: DocItem[]
  items_count?: number
  total?: number
}

export interface DocumentInput {
  type: DocType
  number?: string
  warehouse_id?: string
  target_warehouse_id?: string
  supplier_id?: string
  organization_id?: string
  incoming_number?: string
  incoming_date?: string
  comment?: string
  items: Array<{ product_id: string; quantity: number; price?: number }>
}

export interface DocFilters {
  type?: string
  status?: string
  warehouse_id?: string
  supplier_id?: string
  organization_id?: string
  date_from?: string
  date_to?: string
  incoming_number?: string
  incoming_date_from?: string
  incoming_date_to?: string
  paid_min?: number
  paid_max?: number
  printed?: boolean
  sent?: boolean
}

export async function listDocuments(filters?: DocFilters): Promise<Document[]> {
  const params = new URLSearchParams()
  for (const [k, v] of Object.entries(filters ?? {})) {
    if (v !== undefined && v !== null && v !== '') params.set(k, String(v))
  }
  const qs = params.toString()
  const url = qs ? `/documents?${qs}` : '/documents'
  const { data } = await http.get<{ items: Document[] }>(url)
  return data.items
}

export async function getDocument(id: string): Promise<Document> {
  const { data } = await http.get<Document>(`/documents/${id}`)
  return data
}

export async function createDocument(input: DocumentInput): Promise<Document> {
  const { data } = await http.post<Document>('/documents', input)
  return data
}

export async function postDocument(id: string): Promise<Document> {
  const { data } = await http.post<Document>(`/documents/${id}/post`)
  return data
}

export async function cancelDocument(id: string): Promise<Document> {
  const { data } = await http.post<Document>(`/documents/${id}/cancel`)
  return data
}