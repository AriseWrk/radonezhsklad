import { http } from './client'

export interface Product {
  id: string
  name: string
  sku?: string
  barcode?: string
  category_id?: string
  unit_id?: string
  description?: string
  price: number
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
  currency: string
}

export async function listProducts(includeArchived = false): Promise<Product[]> {
  const url = includeArchived ? '/products?include_archived=true' : '/products'
  const { data } = await http.get<{ items: Product[] }>(url)
  return data.items
}

export async function createProduct(input: ProductInput): Promise<Product> {
  const { data } = await http.post<Product>('/products', input)
  return data
}

export async function archiveProduct(id: string): Promise<void> {
  await http.delete(`/products/${id}`)
}