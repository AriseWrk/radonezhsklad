import { http } from './client'

export interface Customer {
  id: string
  name: string
  full_name?: string
  last_name?: string
  first_name?: string
  middle_name?: string
  phone?: string
  fax?: string
  email?: string
  address?: string
  legal_address?: string
  actual_address?: string
  inn?: string
  kpp?: string
  ogrn?: string
  okpo?: string
  external_code?: string
  counterparty_type?: string
  status: string
  group_name?: string
  comment?: string
  archived: boolean
  created_at: string
  updated_at: string
}

export interface CustomerInput {
  name: string
  full_name?: string
  last_name?: string
  first_name?: string
  middle_name?: string
  phone?: string
  fax?: string
  email?: string
  address?: string
  legal_address?: string
  actual_address?: string
  inn?: string
  kpp?: string
  ogrn?: string
  okpo?: string
  external_code?: string
  counterparty_type?: string
  status?: string
  group_name?: string
  comment?: string
  archived?: boolean
}

export async function listCustomers(includeArchived = false): Promise<Customer[]> {
  const url = includeArchived ? '/customers?include_archived=true' : '/customers'
  const { data } = await http.get<{ items: Customer[] }>(url)
  return data.items
}

export async function createCustomer(input: CustomerInput): Promise<Customer> {
  const { data } = await http.post<Customer>('/customers', input)
  return data
}

export async function updateCustomer(id: string, input: CustomerInput): Promise<Customer> {
  const { data } = await http.put<Customer>(`/customers/${id}`, input)
  return data
}

export async function deleteCustomer(id: string): Promise<void> {
  await http.delete(`/customers/${id}`)
}