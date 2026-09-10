import { http } from './client'

export interface Unit {
  id: string
  code: string
  name: string
  short_name: string
}

export async function listUnits(): Promise<Unit[]> {
  const { data } = await http.get<{ items: Unit[] }>('/units')
  return data.items
}