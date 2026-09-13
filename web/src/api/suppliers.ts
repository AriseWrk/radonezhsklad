import { http } from './client'

export interface Supplier {
  id: string
  name: string
  inn?: string
  phone?: string
  email?: string
  address?: string
  created_at: string
  updated_at: string
}

export interface Organization {
  id: string
  name: string
  inn?: string
  is_default: boolean
  created_at: string
}

export interface SupplierInput {
  name: string
  inn?: string
  phone?: string
  email?: string
  address?: string
}

export async function listSuppliers(): Promise<Supplier[]> {
  const { data } = await http.get<{ items: Supplier[] }>('/suppliers')
  return data.items
}

export async function createSupplier(input: SupplierInput): Promise<Supplier> {
  const { data } = await http.post<Supplier>('/suppliers', input)
  return data
}

export async function updateSupplier(id: string, input: SupplierInput): Promise<Supplier> {
  const { data } = await http.put<Supplier>(`/suppliers/${id}`, input)
  return data
}

export async function deleteSupplier(id: string): Promise<void> {
  await http.delete(`/suppliers/${id}`)
}

export async function listOrganizations(): Promise<Organization[]> {
  const { data } = await http.get<{ items: Organization[] }>('/organizations')
  return data.items
}