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