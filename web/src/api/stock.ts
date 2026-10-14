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

export interface ProductStockMovement {
  id: string
  document_id?: string
  document_type: string
  document_number: string
  document_created_at?: string
  movement_at: string
  quantity_delta: number
  days_on_stock: number
  cost_price: number
  cost_sum: number
}

export interface ProductStockWarehouseGroup {
  warehouse_id: string
  warehouse_name: string
  quantity: number
  cost_price: number
  cost_sum: number
  movements: ProductStockMovement[]
}

export interface ProductStockDetail {
  product: {
    id: string
    name: string
    sku?: string
    unit_short: string
    cost_price: number
    price: number
    min_stock: number
    currency: string
  }
  summary: {
    available: number
    reserve: number
    incoming: number
    quantity: number
    min_stock: number
    cost_price: number
    cost_sum: number
    sale_price: number
    sale_sum: number
  }
  warehouses: ProductStockWarehouseGroup[]
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

export interface StockExtendedFilters {
  warehouse_id?: string
  product_id?: string
  qty_mode?: 'positive' | 'zero' | ''
  avail_mode?: 'nonZero' | 'zero' | ''
  days_min?: number | ''
  days_max?: number | ''
  only_incoming?: boolean
  supplier_id?: string
  as_of?: string
  show_mode?: string
}

export async function listStockExtended(filters: StockExtendedFilters = {}): Promise<StockExtendedResponse> {
  const params = new URLSearchParams()
  if (filters.warehouse_id) params.set('warehouse_id', filters.warehouse_id)
  if (filters.product_id) params.set('product_id', filters.product_id)
  if (filters.qty_mode) params.set('qty_mode', filters.qty_mode)
  if (filters.avail_mode) params.set('avail_mode', filters.avail_mode)
  if (filters.days_min !== '' && filters.days_min != null) params.set('days_min', String(filters.days_min))
  if (filters.days_max !== '' && filters.days_max != null) params.set('days_max', String(filters.days_max))
  if (filters.only_incoming) params.set('only_incoming', 'true')
  if (filters.supplier_id) params.set('supplier_id', filters.supplier_id)
  if (filters.as_of) params.set('as_of', filters.as_of)
  if (filters.show_mode) params.set('show_mode', filters.show_mode)
  const qs = params.toString()
  const url = qs ? `/stock/extended?${qs}` : '/stock/extended'
  const { data } = await http.get<StockExtendedResponse>(url)
  return data
}

export async function productStockDetail(productId: string): Promise<ProductStockDetail> {
  const { data } = await http.get<ProductStockDetail>(`/stock/product/${productId}`)
  return data
}
