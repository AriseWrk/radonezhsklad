import { http } from './client'

export interface Unit {
  id: string
  code: string
  name: string
  short_name: string
}

export async function createUnit(input: UnitInput): Promise<Unit> {
  const { data } = await http.post<Unit>('/units', input)
  return data
}

export async function updateUnit(id: string, input: UnitInput): Promise<Unit> {
  const { data } = await http.put<Unit>(`/units/${id}`, input)
  return data
}

export async function deleteUnit(id: string): Promise<void> {
  await http.delete(`/units/${id}`)
}

export interface UnitInput {
  code: string
  name: string
  short_name: string
}

export async function listUnits(): Promise<Unit[]> {
  const { data } = await http.get<{ items: Unit[] }>('/units')
  return data.items
}