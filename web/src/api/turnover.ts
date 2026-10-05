import { http } from './client'

export interface TurnoverRow {
  product_id: string
  product_name: string
  sku?: string
  unit_short: string
  opening: number
  income: number
  outcome: number
  closing: number
}

export interface TurnoverResponse {
  items: TurnoverRow[]
  count: number
}

export async function turnoverReport(filters: {
  warehouse_id?: string
  from?: string
  to?: string
} = {}): Promise<TurnoverResponse> {
  const params = new URLSearchParams()
  if (filters.warehouse_id) params.set('warehouse_id', filters.warehouse_id)
  if (filters.from) params.set('from', filters.from)
  if (filters.to) params.set('to', filters.to)
  const qs = params.toString()
  const url = qs ? `/reports/turnover?${qs}` : '/reports/turnover'
  const { data } = await http.get<TurnoverResponse>(url)
  return data
}