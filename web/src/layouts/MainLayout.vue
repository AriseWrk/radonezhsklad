<template>
  <div class="app">
    <header class="topbar">
      <div class="logo" @click="go('/')">
        <img src="/logo.png" alt="Радонеж" class="logo-image" />
      </div>

      <nav class="top-nav">
        <router-link
          v-for="s in visibleSections"
          :key="s.key"
          :to="s.path"
          class="top-nav-item"
          :class="{ active: currentSection === s.key }"
        >
          <MsIcon :name="s.icon" :size="22" />
          <span class="top-nav-label">{{ s.label }}</span>
        </router-link>
      </nav>

      <div class="top-right">
        <button class="sync-btn" @click="openSync" title="Синхронизировать с МойСклад">
          <MsIcon name="refresh" :size="16" />
          <span>Синхронизировать</span>
        </button>
        <button class="top-icon" title="Чат"><MsIcon name="chat" :size="18" /></button>
        <button class="top-icon" title="Уведомления"><MsIcon name="bell" :size="18" /></button>
        <button class="top-icon" title="Помощь"><MsIcon name="help" :size="18" /></button>
        <div class="user-block">
          <div class="avatar">{{ avatarLetter }}</div>
          <div class="user-info">
            <div class="user-name">{{ auth.userId ? shortId(auth.userId) : 'Пользователь' }}</div>
            <div class="user-role" :class="'role-' + auth.role">{{ roleLabel(auth.role) }}</div>
          </div>
          <button class="logout-btn" @click="onLogout" title="Выйти">
            <MsIcon name="close" :size="14" />
          </button>
        </div>
      </div>
    </header>

    <div v-if="currentSubTabs.length" class="subtabs">
      <router-link
        v-for="t in currentSubTabs"
        :key="t.name"
        :to="t.path"
        class="subtab"
        :class="{ active: isTabActive(t) }"
      >
        {{ t.label }}
      </router-link>
    </div>

    <main class="page-content">
      <router-view />
    </main>

    <SyncModal ref="syncRef" />
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore, type Role } from '../stores/auth'
import SyncModal from '../components/SyncModal.vue'
import MsIcon from '../components/MsIcon.vue'

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()

const syncRef = ref<InstanceType<typeof SyncModal> | null>(null)
function openSync() { syncRef.value?.open() }

interface Section {
  key: string
  label: string
  icon: string
  path: string
  roles?: Role[]
}
interface SubTab {
  name: string
  label: string
  path: string
  match: string
  exact?: boolean
  roles?: Role[]
}

const sections: Section[] = [
  { key: 'company',   label: 'Компания', icon: 'building', path: '/' },
  { key: 'purchases', label: 'Закупки',  icon: 'cart',     path: '/purchases/planning', roles: ['admin', 'manager', 'warehouse'] },
  { key: 'sales',     label: 'Продажи',  icon: 'send',     path: '/orders',             roles: ['admin', 'manager'] },
  { key: 'crm',       label: 'CRM',      icon: 'user',     path: '/counterparties' },
  { key: 'products',  label: 'Товары',   icon: 'package',  path: '/products' },
  { key: 'stock',     label: 'Склад',    icon: 'building', path: '/stock',              roles: ['admin', 'manager', 'warehouse'] },
]

const subTabsMap: Record<string, SubTab[]> = {
  company: [
    { name: 'dashboard', label: 'Показатели',  path: '/',        match: '/',      exact: true, roles: ['admin','manager','warehouse','user'] },
    { name: 'users',     label: 'Сотрудники',  path: '/users',   match: '/users',              roles: ['admin'] },
    { name: 'audit',     label: 'Аудит',       path: '/audit',   match: '/audit',              roles: ['admin'] },
  ],
  crm: [
    { name: 'counterparties', label: 'Контрагенты', path: '/counterparties', match: '/counterparties', roles: ['admin','manager','warehouse','user'] },
    { name: 'contracts',      label: 'Договоры',    path: '/contracts',      match: '/contracts',      roles: ['admin','manager'] },
  ],
  products: [
    { name: 'products',   label: 'Товары',    path: '/products',   match: '/products',   roles: ['admin','manager','warehouse','user'] },
    { name: 'categories', label: 'Категории', path: '/categories', match: '/categories', roles: ['admin','manager','warehouse','user'] },
  ],
  purchases: [
    { name: 'purchases-planning', label: 'Управление закупками', path: '/purchases/planning', match: '/purchases/planning', roles: ['admin','manager','warehouse'] },
    { name: 'purchases-receipts', label: 'Приёмки',              path: '/purchases/receipts', match: '/purchases/receipts', roles: ['admin','manager','warehouse'] },
    { name: 'purchases-suppliers',label: 'Поставщики',           path: '/suppliers',          match: '/suppliers',          roles: ['admin','manager','warehouse'] },
    { name: 'inventory',          label: 'Инвентаризации',       path: '/inventory',          match: '/inventory',          roles: ['admin','manager','warehouse'] },
    { name: 'warehouses',         label: 'Склады',               path: '/warehouses',         match: '/warehouses',         roles: ['admin','warehouse'] },
  ],
  sales: [
    { name: 'orders',          label: 'Заказы',     path: '/orders',          match: '/orders',          roles: ['admin','manager'] },
    { name: 'customers',       label: 'Покупатели', path: '/customers',       match: '/customers',       roles: ['admin','manager'] },
    { name: 'sales-analytics', label: 'Аналитика',  path: '/sales/analytics', match: '/sales/analytics', roles: ['admin','manager'] },
  ],
  stock: [
    { name: 'stock',           label: 'Остатки',           path: '/stock',           match: '/stock',           roles: ['admin','manager','warehouse'] },
    { name: 'internal-orders', label: 'Внутренние заказы', path: '/internal-orders', match: '/internal-orders', roles: ['admin','manager','warehouse'] },
    { name: 'documents',       label: 'Документы',         path: '/documents',       match: '/documents',       roles: ['admin','manager','warehouse'] },
    { name: 'inventories',     label: 'Инвентаризации',    path: '/inventories',     match: '/inventories',     roles: ['admin','manager','warehouse'] },
    { name: 'warehouses',      label: 'Склады',            path: '/warehouses',      match: '/warehouses',      roles: ['admin','warehouse'] },
  ],
}

const visibleSections = computed(() =>
  sections.filter((s) => !s.roles || auth.can(...s.roles))
)

const currentSection = computed(() => {
  const p = route.path
  if (p.startsWith('/products') || p.startsWith('/categories')) return 'products'
  if (p.startsWith('/counterparties') || p.startsWith('/customers')) return 'crm'
  if (p.startsWith('/orders') || p.startsWith('/customers') || p.startsWith('/sales')) return 'sales'
  if (p.startsWith('/stock') || p.startsWith('/internal-orders')) return 'stock'
  if (p === '/inventory' || p === '/inventories') return 'purchases'
  if (p.startsWith('/users') || p.startsWith('/audit')) return 'company'
  return 'company'
})

const currentSubTabs = computed(() => {
  const tabs = subTabsMap[currentSection.value] || []
  return tabs.filter((t) => !t.roles || auth.can(...t.roles))
})

function isTabActive(t: SubTab): boolean {
  if (t.exact) return route.path === t.path
  return route.path === t.path || route.path.startsWith(t.match + '/') || route.path === t.match
}

function shortId(id: string) { return id.slice(0, 8) }
const avatarLetter = computed(() => (auth.userId ? auth.userId.slice(0, 1).toUpperCase() : '?'))
function roleLabel(r: Role | null) {
  if (!r) return ''
  return { admin: 'Админ', manager: 'Менеджер', warehouse: 'Кладовщик', user: 'Пользователь' }[r]
}
function go(path: string) { router.push(path) }
function onLogout() {
  auth.logout()
  router.push({ name: 'login' })
}
</script>

<style scoped>
* { box-sizing: border-box; }
.app { min-height: 100vh; display: flex; flex-direction: column; background: #f4f5f6; }

.topbar {
  height: 48px;
  background: #2c5d9c;
  display: flex;
  align-items: center;
  padding: 0 12px 0 16px;
  color: #fff;
  flex-shrink: 0;
}
.logo {
  display: flex; align-items: center; gap: 8px;
  margin-right: 16px;
  cursor: pointer;
}
.logo-image {
  height: 32px; width: auto; display: block;
  filter: brightness(0) invert(1);
}

.top-nav {
  display: flex;
  height: 100%;
  flex: 1;
  overflow-x: auto;
}
.top-nav-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4px 14px;
  color: rgba(255,255,255,0.85);
  font-size: 11px;
  text-decoration: none;
  gap: 2px;
  transition: background 0.1s;
  white-space: nowrap;
}
.top-nav-item:hover { background: rgba(255,255,255,0.1); color: #fff; text-decoration: none; }
.top-nav-item.active { background: rgba(255,255,255,0.18); color: #fff; }
.top-nav-label { font-size: 11px; line-height: 1; }

.top-right {
  display: flex; align-items: center; gap: 4px;
  margin-left: 12px;
}
.sync-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 5px 10px;
  background: rgba(255,255,255,0.12);
  border: 1px solid rgba(255,255,255,0.25);
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  color: #fff;
  margin-right: 4px;
}
.sync-btn:hover { background: rgba(255,255,255,0.22); }

.top-icon {
  width: 32px; height: 32px;
  display: inline-flex; align-items: center; justify-content: center;
  border: none; background: transparent;
  color: #fff; cursor: pointer;
  border-radius: 4px;
}
.top-icon:hover { background: rgba(255,255,255,0.12); }

.user-block {
  display: flex; align-items: center; gap: 8px;
  padding-left: 12px;
  margin-left: 6px;
  border-left: 1px solid rgba(255,255,255,0.2);
}
.avatar {
  width: 32px; height: 32px;
  border-radius: 50%;
  background: rgba(255,255,255,0.22);
  color: #fff;
  font-weight: 600;
  display: inline-flex; align-items: center; justify-content: center;
  font-size: 14px;
}
.user-info { display: flex; flex-direction: column; line-height: 1.2; }
.user-name { font-size: 12px; font-weight: 600; }
.user-role { font-size: 10px; opacity: 0.85; }
.logout-btn {
  display: inline-flex; align-items: center; justify-content: center;
  width: 26px; height: 26px;
  background: transparent;
  color: rgba(255,255,255,0.8);
  border: none;
  border-radius: 4px;
  cursor: pointer;
}
.logout-btn:hover { background: rgba(255,255,255,0.15); color: #fff; }

.subtabs {
  background: #fff;
  border-bottom: 1px solid #e1e4e8;
  display: flex;
  padding: 0 16px;
  gap: 4px;
  overflow-x: auto;
  flex-shrink: 0;
}
.subtab {
  padding: 11px 14px;
  font-size: 13px;
  color: #57606a;
  text-decoration: none;
  border-bottom: 2px solid transparent;
  white-space: nowrap;
}
.subtab:hover { color: #2c5d9c; text-decoration: none; }
.subtab.active {
  color: #2c5d9c;
  border-bottom-color: #2c5d9c;
  font-weight: 600;
}

.page-content {
  flex: 1;
  padding: 16px 24px 32px;
  overflow-y: auto;
}

.role-admin     { color: #ffcdd2; }
.role-manager   { color: #bbdefb; }
.role-warehouse { color: #ffe082; }
.role-user      { color: #e0e0e0; }
</style>