import { http } from './client'

export interface StockBalance {
  id: string
  warehouse_id: string
  product_id: string
  quantity: number
  updated_at: string
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