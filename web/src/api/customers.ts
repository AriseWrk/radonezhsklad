import { http } from './client'

export interface Customer {
  id: string
  name: string
  phone?: string
  email?: string
  address?: string
  created_at: string
  updated_at: string
}

export async function listCustomers(): Promise<Customer[]> {
  const { data } = await http.get<{ items: Customer[] }>('/customers')
  return data.items
}

export async function createCustomer(input: { name: string; phone?: string; email?: string; address?: string }): Promise<Customer> {
  const { data } = await http.post<Customer>('/customers', input)
  return data
}

export async function deleteCustomer(id: string): Promise<void> {
  await http.delete(`/customers/${id}`)
}