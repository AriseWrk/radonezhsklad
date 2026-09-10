import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import * as authApi from '../api/auth'
import { TOKEN_KEY } from '../api/client'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem(TOKEN_KEY))
  const userId = ref<string | null>(null)
  const role = ref<string | null>(null)

  const isAuthenticated = computed(() => !!token.value)

  async function login(email: string, password: string) {
    const pair = await authApi.login(email, password)
    token.value = pair.access_token
    localStorage.setItem(TOKEN_KEY, pair.access_token)
    await fetchMe()
  }

  async function fetchMe() {
    if (!token.value) return
    try {
      const m = await authApi.me()
      userId.value = m.user_id
      role.value = m.role
    } catch {
      logout()
    }
  }

  function logout() {
    token.value = null
    userId.value = null
    role.value = null
    localStorage.removeItem(TOKEN_KEY)
  }

  return { token, userId, role, isAuthenticated, login, fetchMe, logout }
})