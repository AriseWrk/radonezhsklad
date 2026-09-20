import { http } from './client'

export interface InventoryItem {
  id: string
  inventory_id: string
  product_id: string
  product_name?: string
  product_sku?: string
  quantity: number
  calculated_quantity: number
  correction_amount: number
  price: number
  correction_sum: number
  created_at: string
}

export interface Inventory {
  id: string
  external_id?: string
  number: string
  doc_date: string
  warehouse_id?: string
  warehouse_name?: string
  organization_id?: string
  comment?: string
  total: number
  items_count: number
  created_at: string
  updated_at: string
  items?: InventoryItem[]
}

export async function listInventories(filters?: {
  warehouse_id?: string
  from?: string
  to?: string
}): Promise<Inventory[]> {
  const params = new URLSearchParams()
  if (filters?.warehouse_id) params.set('warehouse_id', filters.warehouse_id)
  if (filters?.from) params.set('from', filters.from)
  if (filters?.to) params.set('to', filters.to)
  const qs = params.toString()
  const url = qs ? `/inventories?${qs}` : '/inventories'
  const { data } = await http.get<{ items: Inventory[] }>(url)
  return data.items
}

export async function getInventory(id: string): Promise<Inventory> {
  const { data } = await http.get<Inventory>(`/inventories/${id}`)
  return data
}


export interface InventoryNeighbors {
  prev_id: string | null
  next_id: string | null
  position: number
  total: number
}

export async function getInventoryNeighbors(id: string): Promise<InventoryNeighbors> {
  const { data } = await http.get<InventoryNeighbors>(
`
/inventories/${id}/neighbors
`
)
  return data
}
