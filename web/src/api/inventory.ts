import { http } from './client'

export interface InventoryPrepareRow {
  product_id: string
  product_name: string
  sku?: string
  unit_short: string
  book_quantity: number
  cost_price: number
  sale_price: number
}

export interface InventoryPrepareResponse {
  warehouse_id: string
  items: InventoryPrepareRow[]
  count: number
}

export async function prepareInventory(warehouseId: string): Promise<InventoryPrepareResponse> {
  const { data } = await http.get<InventoryPrepareResponse>(`/inventory/prepare?warehouse_id=${warehouseId}`)
  return data
}