import axios from 'axios'

export const TOKEN_KEY = 'rs_access_token'

export const http = axios.create({
  baseURL: '/api/v1',
  timeout: 15000,
})

http.interceptors.request.use((config) => {
  const token = localStorage.getItem(TOKEN_KEY)
  if (token) {
    config.headers = config.headers ?? {}
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

http.interceptors.response.use(
  (r) => r,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem(TOKEN_KEY)
      if (!location.pathname.startsWith('/login')) {
        location.href = '/login'
      }
    }
    return Promise.reject(error)
  },
)

export function apiErrorMessage(err: unknown): string {
  const anyErr = err as any
  const data = anyErr?.response?.data
  if (data?.error?.message) return data.error.message
  if (data?.error?.details) return JSON.stringify(data.error.details)
  if (anyErr?.message) return anyErr.message
  return 'Неизвестная ошибка'
}