import { http } from './client'

export interface ReconciliationRun {
  id: string
  status: 'running' | 'completed' | 'failed' | string
  started_at: string
  finished_at?: string | null
  triggered_by?: string | null
  triggered_by_name: string
  price_type_name: string
  error?: string | null
  total_products: number
  matched_products: number
  price_differences: number
  stock_differences: number
  missing_local_products: number
  missing_ms_products: number
  ambiguous_products: number
  unmatched_warehouses: number
  zero_stock_items: number
}

export interface ReconciliationItem {
  id: string
  run_id: string
  item_type: 'product' | 'price' | 'stock' | 'warehouse' | string
  result_status: 'matched' | 'difference' | 'missing_local' | 'missing_moysklad' | 'ambiguous' | 'unmatched' | string
  match_method?: string | null
  local_product_id?: string | null
  ms_product_id?: string | null
  sku?: string | null
  local_product_name?: string | null
  ms_product_name?: string | null
  local_warehouse_id?: string | null
  ms_store_id?: string | null
  local_warehouse_name?: string | null
  ms_store_name?: string | null
  local_price?: number | null
  ms_price?: number | null
  price_difference?: number | null
  local_stock?: number | null
  ms_stock?: number | null
  stock_difference?: number | null
  is_zero_stock: boolean
  details: Record<string, unknown>
  created_at: string
}

export interface ReconciliationItemsResponse {
  items: ReconciliationItem[]
  total: number
  limit: number
  offset: number
}

export async function listReconciliationRuns(limit = 50): Promise<ReconciliationRun[]> {
  const { data } = await http.get<{ items: ReconciliationRun[] }>('/reconciliation/runs', {
    params: { limit },
  })
  return data.items ?? []
}

export async function getActiveReconciliationRun(): Promise<ReconciliationRun | null> {
  const { data } = await http.get<{ item: ReconciliationRun | null }>('/reconciliation/runs/active')
  return data.item ?? null
}

export async function getReconciliationRun(id: string): Promise<ReconciliationRun> {
  const { data } = await http.get<{ item: ReconciliationRun }>(`/reconciliation/runs/${id}`)
  return data.item
}

export async function listReconciliationItems(
  id: string,
  filters: {
    type?: string
    status?: string
    search?: string
    limit?: number
    offset?: number
  } = {},
): Promise<ReconciliationItemsResponse> {
  const { data } = await http.get<ReconciliationItemsResponse>(
    `/reconciliation/runs/${id}/items`,
    { params: filters },
  )
  return data
}

export async function startReconciliation(): Promise<ReconciliationRun> {
  const { data } = await http.post<{ item: ReconciliationRun }>('/reconciliation/runs')
  return data.item
}

export async function exportReconciliation(id: string): Promise<Blob> {
  const { data } = await http.get(`/reconciliation/runs/${id}/export`, {
    responseType: 'blob',
    timeout: 60000,
  })
  return data as Blob
}
