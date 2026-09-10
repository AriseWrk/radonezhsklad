<template>
  <div class="login-wrap">
    <form class="card login-card" @submit.prevent="onSubmit">
      <h1>RadonezhSklad</h1>
      <p class="muted">Вход в систему</p>

      <div v-if="error" class="error-box">{{ error }}</div>

      <label>Email
        <input v-model="email" type="email" required autocomplete="username" />
      </label>
      <label>Пароль
        <input v-model="password" type="password" required autocomplete="current-password" />
      </label>

      <button class="primary" type="submit" :disabled="loading">
        {{ loading ? 'Вход...' : 'Войти' }}
      </button>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { apiErrorMessage } from '../api/client'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()

const email = ref('admin@radonezh.local')
const password = ref('qwerty123')
const loading = ref(false)
const error = ref<string | null>(null)

async function onSubmit() {
  error.value = null
  loading.value = true
  try {
    await auth.login(email.value, password.value)
    const redirect = (route.query.redirect as string) || '/'
    router.push(redirect)
  } catch (e) {
    error.value = apiErrorMessage(e)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-wrap {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}
.login-card {
  width: 360px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}
.login-card h1 { margin: 0 0 4px; font-size: 22px; }
.login-card .muted { margin: 0; color: var(--muted); font-size: 13px; }
label { display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: var(--muted); }
label input { color: var(--text); }
button.primary { margin-top: 6px; padding: 10px; font-size: 15px; }
</style>