import { http } from './client'

export interface Product {
  id: string
  external_id?: string | null
  name: string
  sku?: string
  barcode?: string
  category_id?: string
  unit_id?: string
  description?: string
  price: number
  cost_price: number
  min_stock: number
  currency: string
  is_archived: boolean
  created_at: string
  updated_at: string
}

export interface ProductInput {
  name: string
  sku?: string
  barcode?: string
  category_id?: string
  unit_id?: string
  description?: string
  price: number
  cost_price?: number
  min_stock?: number
  currency: string
}

export async function listProducts(
  includeArchived = false,
  categoryId?: string
): Promise<Product[]> {
  const params = new URLSearchParams()
  if (includeArchived) params.set('include_archived', 'true')
  if (categoryId) params.set('category_id', categoryId)
  const qs = params.toString()
  const url = qs ? `/products?${qs}` : '/products'
  const { data } = await http.get<{ items: Product[] }>(url)
  return data.items
}

export async function getProduct(id: string): Promise<Product> {
  const { data } = await http.get<Product>(`/products/${id}`)
  return data
}

export async function createProduct(input: ProductInput): Promise<Product> {
  const { data } = await http.post<Product>('/products', input)
  return data
}

export async function updateProduct(id: string, input: ProductInput): Promise<Product> {
  const { data } = await http.put<Product>(`/products/${id}`, input)
  return data
}

export async function archiveProduct(id: string): Promise<void> {
  await http.delete(`/products/${id}`)
}

export async function unarchiveProduct(id: string, input: ProductInput): Promise<Product> {
  const { data } = await http.put<Product>(`/products/${id}`, { ...input, is_archived: false } as any)
  return data
}