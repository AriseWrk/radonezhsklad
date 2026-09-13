import { http } from './client'

export interface StockBalance {
  id: string
  warehouse_id: string
  product_id: string
  quantity: number
  updated_at: string
}

export interface StockExtendedRow {
  product_id: string
  product_name: string
  sku?: string
  warehouse_id?: string
  quantity: number
  min_stock: number
  reserve: number
  incoming: number
  available: number
  unit_short: string
  days_on_stock?: number
  cost_price: number
  cost_total: number
  sale_price: number
  sale_total: number
}

export interface StockExtendedTotals {
  quantity: number
  min_stock: number
  reserve: number
  incoming: number
  available: number
  cost_total: number
  sale_total: number
}

export interface StockExtendedResponse {
  items: StockExtendedRow[]
  totals: StockExtendedTotals
  count: number
}

export async function listStock(warehouseId?: string, productId?: string): Promise<StockBalance[]> {
  const params = new URLSearchParams()
  if (warehouseId) params.set('warehouse_id', warehouseId)
  if (productId) params.set('product_id', productId)
  const qs = params.toString()
  const url = qs ? `/stock?${qs}` : '/stock'
  const { data } = await http.get<{ items: StockBalance[] }>(url)
  return data.items
}

export async function listStockExtended(filters: { warehouse_id?: string; search?: string } = {}): Promise<StockExtendedResponse> {
  const params = new URLSearchParams()
  if (filters.warehouse_id) params.set('warehouse_id', filters.warehouse_id)
  const qs = params.toString()
  const url = qs ? `/stock/extended?${qs}` : '/stock/extended'
  const { data } = await http.get<StockExtendedResponse>(url)
  return data
}