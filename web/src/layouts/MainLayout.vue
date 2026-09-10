<template>
  <div class="layout">
    <aside class="sidebar">
      <div class="brand">RadonezhSklad</div>
      <nav>
        <router-link to="/"           class="nav-item">Дашборд</router-link>
        <router-link to="/products"   class="nav-item">Товары</router-link>
        <router-link to="/categories" class="nav-item">Категории</router-link>
      </nav>
    </aside>
    <main class="content">
      <header class="topbar">
        <div></div>
        <div class="user">
          <span v-if="auth.userId" class="muted">{{ shortId(auth.userId) }}</span>
          <span v-if="auth.role" class="badge">{{ auth.role }}</span>
          <button class="danger" @click="onLogout">Выйти</button>
        </div>
      </header>
      <div class="page">
        <router-view />
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const router = useRouter()

function shortId(id: string) {
  return id.slice(0, 8)
}

function onLogout() {
  auth.logout()
  router.push({ name: 'login' })
}
</script>

<style scoped>
.layout {
  display: flex;
  min-height: 100vh;
}
.sidebar {
  width: 220px;
  background: #fff;
  border-right: 1px solid var(--border);
  padding: 16px 0;
  flex-shrink: 0;
}
.brand {
  padding: 0 20px 20px;
  font-weight: 700;
  font-size: 18px;
  color: var(--primary);
}
.nav-item {
  display: block;
  padding: 10px 20px;
  color: var(--text);
  font-size: 14px;
  text-decoration: none;
}
.nav-item:hover { background: #f3f4f6; text-decoration: none; }
.nav-item.router-link-exact-active {
  background: #eef4ff;
  color: var(--primary);
  font-weight: 600;
  border-right: 3px solid var(--primary);
}
.content { flex: 1; display: flex; flex-direction: column; }
.topbar {
  height: 56px;
  background: #fff;
  border-bottom: 1px solid var(--border);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
}
.user { display: flex; align-items: center; gap: 12px; }
.muted { color: var(--muted); font-size: 13px; font-family: monospace; }
.badge {
  background: #eef4ff;
  color: var(--primary);
  padding: 3px 8px;
  border-radius: 10px;
  font-size: 12px;
  font-weight: 600;
}
.page { padding: 24px; }
</style>