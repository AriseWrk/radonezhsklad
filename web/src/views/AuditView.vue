<template>
  <div>
    <!-- Заголовок -->
    <div class="page-title-bar">
      <div class="page-title">
        <span class="info-icon" title="История действий пользователей">ⓘ</span>
        <span>Аудит</span>
        <span class="refresh" @click="load" title="Обновить">↻</span>
      </div>
      <div class="page-actions">
        <button class="btn" @click="showFilter = !showFilter">Фильтр</button>
      </div>
    </div>

    <!-- Панель фильтров -->
    <div v-if="showFilter" class="filter-panel">
      <div class="filter-row">
        <div class="filter-actions">
          <button class="btn-find" @click="applyFilters">Найти</button>
          <button class="btn-clear" @click="clearFilters">Очистить</button>
        </div>
        <div class="filter-field">
          <label>Период с</label>
          <input v-model="filterDateFrom" type="date" />
        </div>
        <div class="filter-field">
          <label>Период по</label>
          <input v-model="filterDateTo" type="date" />
        </div>
        <div class="filter-field">
          <label>Сотрудник</label>
          <select v-model="filterUserId">
            <option value="">Все</option>
            <option v-for="u in users" :key="u.id" :value="u.id">{{ fullName(u) }}</option>
          </select>
        </div>
        <div class="filter-field">
          <label>Событие</label>
          <input v-model="filterSearch" placeholder="Например: заказ, товар" />
        </div>
      </div>
    </div>

    <div v-if="error" class="error-box">{{ error }}</div>

    <!-- Таблица -->
    <table class="ms-table audit-table">
      <thead>
        <tr>
          <th style="width:150px" @click="sortBy('created_at')">
            Время
            <span v-if="sortKey === 'created_at'" class="sort-arrow">{{ sortDir === 'asc' ? '↑' : '↓' }}</span>
          </th>
          <th style="width:230px" @click="sortBy('user')">
            Сотрудник
            <span v-if="sortKey === 'user'" class="sort-arrow">{{ sortDir === 'asc' ? '↑' : '↓' }}</span>
          </th>
          <th>Событие</th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="loading">
          <td colspan="3" class="muted" style="text-align:center;padding:24px">Загрузка...</td>
        </tr>
        <tr v-else-if="filtered.length === 0">
          <td colspan="3" class="muted" style="text-align:center;padding:24px">Нет событий</td>
        </tr>
        <tr v-else v-for="l in paginated" :key="l.id">
          <td class="muted">{{ formatDate(l.created_at) }}</td>
          <td>
            <div class="user-cell">
              <div class="avatar">{{ initials(userById(l.user_id)) }}</div>
              <span>{{ userLabel(userById(l.user_id)) }}</span>
            </div>
          </td>
          <td>
            <span v-if="describeParts(l).prefix" class="event-prefix">{{ describeParts(l).prefix }}</span>
            <a
              v-if="describeParts(l).link"
              class="event-link"
              href="#"
              @click.prevent
            >{{ describeParts(l).link }}</a>
            <span v-if="describeParts(l).suffix" class="event-suffix"> {{ describeParts(l).suffix }}</span>
          </td>
        </tr>
      </tbody>
    </table>

    <!-- Футер -->
    <div class="ms-footer">
      <div class="ms-pager">
        <button :disabled="page === 1" @click="page--">◀</button>
        <button :disabled="page === 1" @click="page = 1">↤</button>
        <span>{{ rangeFrom }}–{{ rangeTo }} из {{ filtered.length }}</span>
        <button :disabled="rangeTo >= filtered.length" @click="page = Math.ceil(filtered.length / perPage)">↦</button>
        <button :disabled="rangeTo >= filtered.length" @click="page++">▶</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { listAudit, type AuditLog } from '../api/audit'
import { listUsers, type User } from '../api/users'
import { apiErrorMessage } from '../api/client'

const items = ref<AuditLog[]>([])
const users = ref<User[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const showFilter = ref(false)
const page = ref(1)
const perPage = 100

const filterDateFrom = ref('')
const filterDateTo = ref('')
const filterUserId = ref('')
const filterSearch = ref('')

const sortKey = ref<'created_at' | 'user'>('created_at')
const sortDir = ref<'asc' | 'desc'>('desc')

const userMap = computed(() => {
  const m = new Map<string, User>()
  for (const u of users.value) m.set(u.id, u)
  return m
})

function userById(id?: string): User | null {
  if (!id) return null
  return userMap.value.get(id) ?? null
}

function fullName(u: User): string {
  const parts = [u.last_name, u.first_name, u.middle_name].filter(Boolean)
  return parts.length > 0 ? parts.join(' ') : u.email
}

function userLabel(u: User | null): string {
  if (!u) return 'Система'
  const ln = (u.last_name ?? '').trim()
  const fn = (u.first_name ?? '').trim()
  const mn = (u.middle_name ?? '').trim()
  if (ln || fn) {
    // МойСклад-стиль: «Фамилия И. О.»
    const initials = [fn.charAt(0), mn.charAt(0)].filter(Boolean).map((c) => c + '.').join(' ')
    return (ln + ' ' + initials).trim()
  }
  return u.email || 'Сотрудник'
}

function initials(u: User | null): string {
  if (!u) return '·'
  const ln = (u.last_name ?? '').trim()
  const fn = (u.first_name ?? '').trim()
  const a = ln ? ln.charAt(0) : ''
  const b = fn ? fn.charAt(0) : ''
  const s = (a + b).toUpperCase()
  return s || (u.email ? u.email.charAt(0).toUpperCase() : '·')
}

function formatDate(s: string) {
  const d = new Date(s)
  const day = String(d.getDate()).padStart(2, '0')
  const mon = String(d.getMonth() + 1).padStart(2, '0')
  const yr = d.getFullYear()
  const hh = String(d.getHours()).padStart(2, '0')
  const mm = String(d.getMinutes()).padStart(2, '0')
  return `${day}.${mon}.${yr} ${hh}:${mm}`
}

// ---------- Человекочитаемое описание события ----------

interface EventParts { prefix: string; link: string; suffix: string }

function describeParts(l: AuditLog): EventParts {
  const path = l.path || ''
  const method = (l.method || '').toUpperCase()
  const body = parseBody(l.request_body)

  const numFromBody = firstOf(body, ['number', 'incoming_number', 'name', 'id'])

  // Особые маршруты — сначала
  const special = matchSpecial(path, method, numFromBody)
  if (special) return special

  // Общие
  const resource = detectResource(path)
  if (resource) {
    const verb = verbFor(method)
    const obj = numFromBody ? `${resource.genderRu} ${numFromBody}` : ''
    const name = body && (body.name || body.title) ? `«${body.name || body.title}»` : ''

    const base = `${verb} ${resource.ru}${obj ? ' ' + obj : ''}`
    return {
      prefix: base + (name ? ' ' : ''),
      link: name ? name : '',
      suffix: '',
    }
  }

  // Непонятный путь
  return {
    prefix: `${method} ${path}`,
    link: '',
    suffix: '',
  }
}

function matchSpecial(path: string, method: string, num: string): EventParts | null {
  const lower = path.toLowerCase()

  // Внутренние заказы
  if (lower.includes('/internal-orders')) {
    if (lower.endsWith('/print') || lower.endsWith('/export')) {
      return { prefix: 'Распечатан внутренний заказ', link: num, suffix: '' }
    }
    if (lower.endsWith('/post')) {
      return { prefix: 'Проведён внутренний заказ', link: num, suffix: '' }
    }
    if (lower.endsWith('/cancel')) {
      return { prefix: 'Отменён внутренний заказ', link: num, suffix: '' }
    }
    if (method === 'POST' && /\/internal-orders\/?$/.test(lower)) {
      return { prefix: 'Создан внутренний заказ', link: num, suffix: '' }
    }
    if (method === 'PUT' || method === 'PATCH') {
      return { prefix: 'Изменён внутренний заказ', link: num, suffix: '' }
    }
    if (method === 'DELETE') {
      return { prefix: 'Удалён внутренний заказ', link: num, suffix: '' }
    }
  }

  // Документы
  if (lower.includes('/documents')) {
    if (lower.endsWith('/post')) {
      return { prefix: 'Проведён документ', link: num, suffix: '' }
    }
    if (lower.endsWith('/cancel')) {
      return { prefix: 'Отменён документ', link: num, suffix: '' }
    }
    if (method === 'POST' && /\/documents\/?$/.test(lower)) {
      return { prefix: 'Создан документ', link: num, suffix: '' }
    }
    if (method === 'DELETE') {
      return { prefix: 'Удалён документ', link: num, suffix: '' }
    }
  }

  // Заказы покупателей
  if (lower.includes('/orders/')) {
    if (lower.endsWith('/confirm')) return { prefix: 'Подтверждён заказ', link: num, suffix: '' }
    if (lower.endsWith('/ship'))    return { prefix: 'Отгружен заказ',    link: num, suffix: '' }
    if (lower.endsWith('/cancel'))  return { prefix: 'Отменён заказ',     link: num, suffix: '' }
  }

  // Экспорт
  if (lower.endsWith('/export')) {
    return { prefix: 'Экспорт выполнен', link: num, suffix: '' }
  }

  return null
}

interface Resource { ru: string; genderRu: string }
function detectResource(path: string): Resource | null {
  const lower = path.toLowerCase()
  if (lower.includes('/products'))       return { ru: 'товар',              genderRu: '' }
  if (lower.includes('/categories'))     return { ru: 'категорию',          genderRu: '' }
  if (lower.includes('/units'))          return { ru: 'единицу измерения',  genderRu: '' }
  if (lower.includes('/warehouses'))     return { ru: 'склад',              genderRu: '' }
  if (lower.includes('/suppliers'))      return { ru: 'поставщика',         genderRu: '' }
  if (lower.includes('/organizations'))  return { ru: 'организацию',        genderRu: '' }
  if (lower.includes('/customers'))      return { ru: 'покупателя',         genderRu: '' }
  if (lower.includes('/orders'))         return { ru: 'заказ',              genderRu: '' }
  if (lower.includes('/internal-orders'))return { ru: 'внутренний заказ',   genderRu: '' }
  if (lower.includes('/users'))          return { ru: 'сотрудника',         genderRu: '' }
  return null
}

function verbFor(method: string): string {
  switch (method) {
    case 'POST':   return 'Создан'
    case 'PUT':
    case 'PATCH':  return 'Изменён'
    case 'DELETE': return 'Удалён'
    default:       return method
  }
}

function parseBody(raw?: string): any {
  if (!raw) return null
  try { return JSON.parse(raw) } catch { return null }
}
function firstOf(obj: any, keys: string[]): string {
  if (!obj) return ''
  for (const k of keys) {
    if (obj[k] != null && obj[k] !== '') return String(obj[k])
  }
  return ''
}

// ---------- Фильтрация/сортировка/пагинация ----------

const filtered = computed(() => {
  let rows = items.value

  if (filterUserId.value) rows = rows.filter((r) => r.user_id === filterUserId.value)

  if (filterDateFrom.value) {
    const t = new Date(filterDateFrom.value).getTime()
    rows = rows.filter((r) => new Date(r.created_at).getTime() >= t)
  }
  if (filterDateTo.value) {
    const t = new Date(filterDateTo.value).getTime() + 86400000 - 1
    rows = rows.filter((r) => new Date(r.created_at).getTime() <= t)
  }
  if (filterSearch.value) {
    const q = filterSearch.value.toLowerCase()
    rows = rows.filter((r) => {
      const desc = describeParts(r).prefix.toLowerCase()
      return desc.includes(q)
    })
  }

  return [...rows].sort((a, b) => {
    if (sortKey.value === 'created_at') {
      const av = new Date(a.created_at).getTime()
      const bv = new Date(b.created_at).getTime()
      return sortDir.value === 'asc' ? av - bv : bv - av
    }
    // user
    const ua = userLabel(userById(a.user_id))
    const ub = userLabel(userById(b.user_id))
    return sortDir.value === 'asc' ? ua.localeCompare(ub, 'ru') : ub.localeCompare(ua, 'ru')
  })
})

const paginated = computed(() => {
  const from = (page.value - 1) * perPage
  return filtered.value.slice(from, from + perPage)
})
const rangeFrom = computed(() => filtered.value.length === 0 ? 0 : (page.value - 1) * perPage + 1)
const rangeTo = computed(() => Math.min(page.value * perPage, filtered.value.length))

function sortBy(k: typeof sortKey.value) {
  if (sortKey.value === k) sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  else { sortKey.value = k; sortDir.value = 'desc' }
  page.value = 1
}

function applyFilters() { page.value = 1 }

function clearFilters() {
  filterDateFrom.value = ''
  filterDateTo.value = ''
  filterUserId.value = ''
  filterSearch.value = ''
  page.value = 1
}

async function load() {
  loading.value = true
  error.value = null
  try {
    const [resp, us] = await Promise.all([
      listAudit({ limit: 500, offset: 0 }),
      listUsers().catch(() => [] as User[]),
    ])
    items.value = resp.items
    users.value = us
  } catch (e) {
    error.value = apiErrorMessage(e)
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.info-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 22px; height: 22px;
  border: 1.5px solid #2c5d9c;
  border-radius: 50%;
  color: #2c5d9c;
  font-size: 14px;
  font-weight: 700;
  margin-right: 6px;
}

.audit-table tbody td { vertical-align: middle; }
.audit-table tbody tr { cursor: default; }

.user-cell {
  display: flex;
  align-items: center;
  gap: 10px;
}
.avatar {
  width: 28px; height: 28px;
  border-radius: 50%;
  background: #e6ebf2;
  color: #6a7a8c;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  font-weight: 600;
  flex-shrink: 0;
}

.event-prefix { color: #1f2328; }
.event-link {
  color: #2c5d9c;
  cursor: pointer;
  text-decoration: none;
}
.event-link:hover { text-decoration: underline; }
.event-suffix { color: #1f2328; }

.sort-arrow { color: #2c5d9c; font-size: 11px; margin-left: 4px; }
</style>