<template>
  <div class="login-wrap">
    <div class="login-bg-shape s1"></div>
    <div class="login-bg-shape s2"></div>
    <div class="login-bg-shape s3"></div>

    <div class="login-topbar">
      <div class="brand">
        <img src="/logo.png" alt="Радонеж" class="brand-logo" />
        <span class="brand-name">RadonezhSklad</span>
      </div>
      <span class="brand-help">Служба поддержки</span>
    </div>

    <div class="login-body">
      <div class="login-left">
        <form class="login-card" @submit.prevent="onSubmit">
          <div class="login-head">
            <img src="/logo.png" alt="" class="card-logo" />
            <div>
              <h1>Вход в RadonezhSklad</h1>
              <p class="muted">Система управления складом</p>
            </div>
          </div>

          <div v-if="error" class="error-box">{{ error }}</div>

          <label>Email
            <input v-model="email" type="email" required autocomplete="username" />
          </label>
          <label>Пароль
            <input v-model="password" type="password" required autocomplete="current-password" />
          </label>

          <button class="submit-btn" type="submit" :disabled="loading">
            {{ loading ? 'Вход...' : 'Войти' }}
          </button>

          <div class="login-foot">
            <span>© 2026 RadonezhSklad</span>
            <span>Русский</span>
          </div>
        </form>
      </div>

      <div class="login-right">
        <div class="promo">
          <span class="promo-badge">Склад под контролем</span>
          <h2>Всё под рукой — остатки, документы, заказы</h2>
          <p class="promo-text">
            Единая система для склада и офиса: закупки, отгрузки, инвентаризации,
            внутренние заказы, интеграция с МойСклад.
          </p>
          <ul class="promo-list">
            <li>Учёт остатков в реальном времени</li>
            <li>Синхронизация с МойСклад</li>
            <li>Отчётность и аналитика продаж</li>
          </ul>
          <div class="promo-art" aria-hidden="true">
            <div class="art-box b1"></div>
            <div class="art-box b2"></div>
            <div class="art-box b3"></div>
            <div class="art-shelf">
              <span></span><span></span><span></span>
            </div>
          </div>
        </div>
      </div>
    </div>
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
  position: relative;
  overflow: hidden;
  background:
    radial-gradient(1200px 600px at 80% 10%, #3f79c0 0%, transparent 60%),
    radial-gradient(900px 500px at 10% 90%, #244f87 0%, transparent 60%),
    linear-gradient(160deg, #2c5d9c 0%, #1e4676 60%, #173356 100%);
  color: #fff;
  display: flex;
  flex-direction: column;
}

/* Декоративные фоновые круги */
.login-bg-shape {
  position: absolute;
  border-radius: 50%;
  filter: blur(60px);
  opacity: 0.35;
  pointer-events: none;
}
.s1 { width: 400px; height: 400px; background: #7fb0ef; top: -100px; right: 10%; }
.s2 { width: 300px; height: 300px; background: #4a8cd8; bottom: -80px; left: 5%; }
.s3 { width: 260px; height: 260px; background: #a8ce5e; bottom: 20%; right: 5%; opacity: 0.18; }

.login-topbar {
  position: relative; z-index: 2;
  height: 56px;
  padding: 0 32px;
  display: flex; align-items: center; justify-content: space-between;
}
.brand { display: flex; align-items: center; gap: 10px; }
.brand-logo { height: 30px; width: auto; filter: brightness(0) invert(1); }
.brand-name { font-weight: 700; font-size: 15px; letter-spacing: 0.3px; }
.brand-help { font-size: 12px; opacity: 0.8; cursor: pointer; }
.brand-help:hover { opacity: 1; text-decoration: underline; }

.login-body {
  position: relative; z-index: 2;
  flex: 1;
  display: grid;
  grid-template-columns: minmax(0, 480px) minmax(0, 520px);
  gap: 48px;
  align-items: center;
  justify-content: center;
  padding: 24px 48px 48px;
}
@media (max-width: 960px) {
  .login-body { grid-template-columns: 1fr; padding: 16px; }
  .login-right { display: none; }
}

.login-left { display: flex; justify-content: flex-end; }
@media (max-width: 960px) {
  .login-left { justify-content: center; }
}

.login-card {
  width: 100%;
  max-width: 420px;
  background: #fff;
  color: #1f2328;
  border-radius: 10px;
  padding: 28px 32px 22px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.28);
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.login-head {
  display: flex; align-items: center; gap: 14px;
  padding-bottom: 12px;
  border-bottom: 1px solid #eaeef2;
  margin-bottom: 4px;
}
.card-logo {
  height: 44px; width: auto;
  flex-shrink: 0;
}
.login-head h1 { margin: 0; font-size: 20px; font-weight: 600; }
.login-head .muted { margin: 2px 0 0; color: #8c959f; font-size: 12px; }

label {
  display: flex; flex-direction: column; gap: 6px;
  font-size: 12px; color: #57606a;
}
label input {
  height: 36px; padding: 0 12px;
  font-size: 14px; color: #1f2328;
  border: 1px solid #d0d7de;
  border-radius: 4px;
  background: #fff;
  transition: border-color 0.1s, box-shadow 0.1s;
}
label input:focus {
  outline: none;
  border-color: #2c5d9c;
  box-shadow: 0 0 0 3px rgba(44, 93, 156, 0.15);
}

.submit-btn {
  margin-top: 6px;
  height: 40px;
  font-size: 14px;
  font-weight: 600;
  color: #fff;
  background: #2c5d9c;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  transition: background 0.1s;
}
.submit-btn:hover:not(:disabled) { background: #234a7d; }
.submit-btn:disabled { opacity: 0.6; cursor: default; }

.login-foot {
  display: flex; justify-content: space-between;
  font-size: 11px; color: #8c959f;
  margin-top: 8px;
  padding-top: 12px;
  border-top: 1px solid #eaeef2;
}

/* Промо-блок справа */
.login-right { display: flex; justify-content: flex-start; }
.promo {
  max-width: 460px;
  position: relative;
}
.promo-badge {
  display: inline-block;
  background: #a8ce5e;
  color: #173356;
  font-weight: 600;
  font-size: 11px;
  padding: 4px 12px;
  border-radius: 999px;
  letter-spacing: 0.3px;
  text-transform: uppercase;
}
.promo h2 {
  font-size: 30px;
  line-height: 1.2;
  margin: 16px 0 12px;
  color: #fff;
  font-weight: 700;
}
.promo-text { font-size: 14px; line-height: 1.5; opacity: 0.85; margin: 0 0 18px; }
.promo-list { margin: 0; padding-left: 20px; font-size: 14px; line-height: 1.9; opacity: 0.9; }
.promo-list li::marker { color: #a8ce5e; }

/* Декоративная "полка" из квадратов */
.promo-art {
  position: relative;
  margin-top: 28px;
  height: 180px;
}
.art-box {
  position: absolute;
  border-radius: 6px;
  background: linear-gradient(135deg, #d0d7de 0%, #b8c0c8 100%);
  box-shadow: 0 8px 20px rgba(0,0,0,0.25);
}
.b1 { width: 60px; height: 60px; left: 0; top: 40px; background: linear-gradient(135deg, #f0c040, #d9a520); }
.b2 { width: 80px; height: 80px; left: 76px; top: 20px; background: linear-gradient(135deg, #c9a87a, #a07a4a); }
.b3 { width: 70px; height: 70px; left: 170px; top: 30px; background: linear-gradient(135deg, #8bbf5a, #6a9d3f); }
.art-shelf {
  position: absolute;
  left: 0; right: 0; bottom: 0;
  height: 12px;
  background: rgba(255,255,255,0.15);
  border-radius: 4px;
  display: flex; align-items: center; gap: 20px; padding: 0 12px;
}
.art-shelf span {
  flex: 1; height: 4px; background: rgba(255,255,255,0.3); border-radius: 2px;
}

.error-box {
  background: #fff0f0;
  border: 1px solid #ffd0d0;
  color: #cf222e;
  padding: 10px 14px;
  border-radius: 6px;
  font-size: 13px;
}
</style>