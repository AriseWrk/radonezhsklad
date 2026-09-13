import { http } from './client'

export interface SalesAnalyticsRow {
  product_id: string
  sold_qty: number
  sold_sum: number
  orders_count: number
  first_sold_at?: string
  last_sold_at?: string
}

export interface SalesAnalyticsResponse {
  days: number
  items: SalesAnalyticsRow[]
  count: number
}

export async function salesAnalytics(days = 14): Promise<SalesAnalyticsResponse> {
  const { data } = await http.get<SalesAnalyticsResponse>(`/analytics/sales?days=${days}`)
  return data
}