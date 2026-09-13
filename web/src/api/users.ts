import { http } from './client'

export type UserRole = 'admin' | 'manager' | 'warehouse' | 'user'

export interface User {
  id: string
  email: string
  full_name: string
  last_name: string
  first_name: string
  middle_name: string
  phone?: string
  login?: string
  description?: string
  role: UserRole
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface UserInput {
  email: string
  password?: string
  last_name: string
  first_name: string
  middle_name?: string
  phone?: string
  login?: string
  description?: string
  role: UserRole
}

export async function listUsers(): Promise<User[]> {
  const { data } = await http.get<{ items: User[] }>('/users')
  return data.items
}

export async function createUser(input: UserInput & { password: string }): Promise<User> {
  const { data } = await http.post<User>('/users', input)
  return data
}

export async function updateUser(id: string, input: UserInput): Promise<User> {
  const { data } = await http.put<User>(`/users/${id}`, input)
  return data
}

export async function updateRole(id: string, role: string): Promise<void> {
  await http.put(`/users/${id}/role`, { role })
}

export async function updateActive(id: string, isActive: boolean): Promise<void> {
  await http.put(`/users/${id}/active`, { is_active: isActive })
}