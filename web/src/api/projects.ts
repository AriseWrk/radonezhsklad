import { http } from './client'

export interface Project {
  id: string
  external_id?: string | null
  name: string
  archived: boolean
  created_at: string
  updated_at: string
}

export async function listProjects(includeArchived = false): Promise<Project[]> {
  const url = includeArchived ? '/projects?include_archived=true' : '/projects'
  const { data } = await http.get<{ items: Project[] }>(url)
  return data.items
}

export async function createProject(name: string): Promise<Project> {
  const { data } = await http.post<Project>('/projects', { name })
  return data
}

export async function updateProject(id: string, name: string, archived = false): Promise<Project> {
  const { data } = await http.put<Project>(`/projects/${id}`, { name, archived })
  return data
}

export async function deleteProject(id: string): Promise<void> {
  await http.delete(`/projects/${id}`)
}

// ---- Project stats ----

export interface ProjectOrderBrief {
  id: string
  number: string
  doc_date: string
  status: 'draft' | 'posted' | 'shipped' | 'closed' | 'cancelled'
  total: number
  shipped_amount: number
  item_count: number
  comment?: string | null
}

export interface ProjectSystemItem {
  system_code: string
  product_id: string
  quantity: number
  order_count: number
  last_doc_date?: string | null
  name: string
  sku?: string | null
  unit_id?: string | null
}

export async function listProjectOrders(projectId: string): Promise<ProjectOrderBrief[]> {
  const { data } = await http.get<{ items: ProjectOrderBrief[] }>(`/projects/${projectId}/internal-orders`)
  return data.items
}

export async function listProjectSystems(projectId: string): Promise<ProjectSystemItem[]> {
  const { data } = await http.get<{ items: ProjectSystemItem[] }>(`/projects/${projectId}/systems-summary`)
  return data.items
}