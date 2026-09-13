import { http } from './client'

export interface Contract {
  id: string
  number: string
  code?: string
  doc_date: string
  customer_id?: string
  organization_id?: string
  amount: number
  currency: string
  paid: number
  fulfilled: number
  comment?: string
  printed_at?: string
  sent_at?: string
  archived: boolean
  created_at: string
  updated_at: string
}

export interface ContractInput {
  number: string
  code?: string
  doc_date?: string
  customer_id?: string
  organization_id?: string
  amount?: number
  currency?: string
  paid?: number
  fulfilled?: number
  comment?: string
  archived?: boolean
}

export async function listContracts(includeArchived = false): Promise<Contract[]> {
  const url = includeArchived ? '/contracts?include_archived=true' : '/contracts'
  const { data } = await http.get<{ items: Contract[] }>(url)
  return data.items
}

export async function createContract(input: ContractInput): Promise<Contract> {
  const { data } = await http.post<Contract>('/contracts', input)
  return data
}

export async function updateContract(id: string, input: ContractInput): Promise<Contract> {
  const { data } = await http.put<Contract>(`/contracts/${id}`, input)
  return data
}

export async function deleteContract(id: string): Promise<void> {
  await http.delete(`/contracts/${id}`)
}