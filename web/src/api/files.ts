import { http } from './client'

export interface FileMeta {
  name: string
  category: string
  size_bytes: number
  uploaded_by?: string
  uploaded_by_name?: string
  created_at: string
}

export async function listFiles(): Promise<FileMeta[]> {
  const { data } = await http.get<{ items: FileMeta[] }>('/files')
  return data.items
}

export async function uploadFile(file: File, category: string): Promise<FileMeta> {
  const fd = new FormData()
  fd.append('file', file)
  fd.append('category', category)
  const { data } = await http.post<FileMeta>('/files', fd, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
  return data
}

export async function deleteFile(name: string): Promise<void> {
  await http.delete(`/files/${encodeURIComponent(name)}`)
}

export function downloadFileUrl(name: string): string {
  return `/api/v1/files/${encodeURIComponent(name)}/download`
}
export async function updateFileCategory(name: string, category: string): Promise<void> {
  await http.patch(`/files/${encodeURIComponent(name)}`, { category })
}