import { http } from './client'

export interface Category {
  id: string
  name: string
  parent_id?: string
  created_at: string
  updated_at: string
}

export async function listCategories(): Promise<Category[]> {
  const { data } = await http.get<{ items: Category[] }>('/categories')
  return data.items
}

export async function createCategory(name: string, parentId?: string | null): Promise<Category> {
  const body: Record<string, unknown> = { name }
  if (parentId) body.parent_id = parentId
  const { data } = await http.post<Category>('/categories', body)
  return data
}

export async function deleteCategory(id: string): Promise<void> {
  await http.delete(`/categories/${id}`)
}