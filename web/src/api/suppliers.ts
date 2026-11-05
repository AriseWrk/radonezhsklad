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
  external_id?: string | null
  name: string
  full_name?: string
  inn?: string
  kpp?: string
  ogrn?: string
  okpo?: string
  legal_address?: string
  is_default: boolean
  created_at: string
}

export interface OrganizationInput {
  name: string
  full_name?: string
  inn?: string
  kpp?: string
  ogrn?: string
  okpo?: string
  legal_address?: string
  is_default?: boolean
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

export async function createOrganization(input: OrganizationInput): Promise<Organization> {
  const { data } = await http.post<Organization>('/organizations', input)
  return data
}

export async function updateOrganization(id: string, input: OrganizationInput): Promise<Organization> {
  const { data } = await http.put<Organization>(`/organizations/${id}`, input)
  return data
}

export async function deleteOrganization(id: string): Promise<void> {
  await http.delete(`/organizations/${id}`)
}