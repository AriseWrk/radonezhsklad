import { http } from './client'

export interface TokenPair {
  access_token: string
  refresh_token: string
  token_type: string
  expires_in: number
}

export interface Me {
  user_id: string
  role: string
}

export async function login(email: string, password: string): Promise<TokenPair> {
  const { data } = await http.post<TokenPair>('/auth/login', { email, password })
  return data
}

export async function me(): Promise<Me> {
  const { data } = await http.get<Me>('/auth/me')
  return data
}