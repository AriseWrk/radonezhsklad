import { http } from './client'

export type OrderStatus = 'draft' | 'confirmed' | 'shipped' | 'cancelled'

export interface OrderItem {
  id: string
  order_id: string
  product_id: string
  quantity: number
  price: number
}

export interface Order {
  id: string
  number: string
  customer_id: string
  warehouse_id: string
  status: OrderStatus
  total: number
  currency: string
  comment?: string
  warehouse_doc_id?: string
  created_at: string
  updated_at: string
  confirmed_at?: string
  shipped_at?: string
  cancelled_at?: string
  items?: OrderItem[]
}

export interface OrderInput {
  number?: string
  customer_id: string
  warehouse_id: string
  comment?: string
  items: Array<{ product_id: string; quantity: number; price?: number }>
}

export async function listOrders(filters?: { status?: string; customer_id?: string }): Promise<Order[]> {
  const params = new URLSearchParams()
  if (filters?.status) params.set('status', filters.status)
  if (filters?.customer_id) params.set('customer_id', filters.customer_id)
  const qs = params.toString()
  const url = qs ? `/orders?${qs}` : '/orders'
  const { data } = await http.get<{ items: Order[] }>(url)
  return data.items
}

export async function getOrder(id: string): Promise<Order> {
  const { data } = await http.get<Order>(`/orders/${id}`)
  return data
}

export async function createOrder(input: OrderInput): Promise<Order> {
  const { data } = await http.post<Order>('/orders', input)
  return data
}

export async function confirmOrder(id: string): Promise<Order> {
  const { data } = await http.post<Order>(`/orders/${id}/confirm`)
  return data
}

export async function shipOrder(id: string): Promise<Order> {
  const { data } = await http.post<Order>(`/orders/${id}/ship`)
  return data
}

export async function cancelOrder(id: string): Promise<Order> {
  const { data } = await http.post<Order>(`/orders/${id}/cancel`)
  return data
}