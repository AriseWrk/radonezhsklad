import { http } from './client'

export interface Warehouse {
  id: string
  name: string
  address?: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export async function listWarehouses(): Promise<Warehouse[]> {
  const { data } = await http.get<{ items: Warehouse[] }>('/warehouses')
  return data.items
}

export async function createWarehouse(name: string, address?: string): Promise<Warehouse> {
  const body: Record<string, unknown> = { name }
  if (address) body.address = address
  const { data } = await http.post<Warehouse>('/warehouses', body)
  return data
}

export async function deleteWarehouse(id: string): Promise<void> {
  await http.delete(`/warehouses/${id}`)
}