<template>
  <div class="page">
    <div class="ms-title">
      <button class="ms-help" title="История действий пользователей"><MsIcon name="help" :size="14" /></button>
      <span>Аудит</span>
      <button class="ms-refresh" @click="load" title="Обновить"><MsIcon name="refresh" :size="14" /></button>
    </div>

    <div class="ms-toolbar">
      <MsButton icon="filter" @click="showFilter = !showFilter">Фильтр</MsButton>
      <input v-model="filterSearch" class="ms-input" placeholder="Событие: заказ, товар..." />
      <div class="ms-counter">{{ filtered.length }}</div>
      <div class="toolbar-spacer"></div>
      <MsButton variant="icon" icon="gear" title="Настройки" />
    </div>

    <div v-if="showFilter" class="ms-filter">
      <div class="filter-actions">
        <MsButton variant="green" @click="applyFilters">Найти</MsButton>
        <MsButton @click="clearFilters">Очистить</MsButton>
      </div>
      <div class="filter-grid">
        <div class="filter-field">
          <label class="filter-label"><span class="dot"></span>Период с</label>
          <input v-model="filterDateFrom" type="date" />
        </div>
        <div class="filter-field">
          <label class="filter-label"><span class="dot"></span>Период по</label>
          <input v-model="filterDateTo" type="date" />
        </div>
        <div class="filter-field">
          <label class="filter-label"><span class="dot"></span>Сотрудник</label>
          <select v-model="filterUserId">
            <option value="">—</option>
            <option v-for="u in users" :key="u.id" :value="u.id">{{ fullName(u) }}</option>
          </select>
        </div>
      </div>
    </div>

    <div v-if="error" class="error-box">{{ error }}</div>

    <table class="ms-table2">
      <thead>
        <tr>
          <th class="col-date" @click="sortBy('created_at')">Время<span v-if="sortKey === 'created_at'" class="sort">{{ sortDir === 'asc' ? '▲' : '▼' }}</span></th>
          <th style="width:230px" @click="sortBy('user')">Сотрудник<span v-if="sortKey === 'user'" class="sort">{{ sortDir === 'asc' ? '▲' : '▼' }}</span></th>
          <th>Событие</th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="loading"><td colspan="3" class="empty">Загрузка...</td></tr>
        <tr v-else-if="filtered.length === 0"><td colspan="3" class="empty">Нет событий</td></tr>
        <tr v-else v-for="l in paginated" :key="l.id" class="row">
          <td class="col-date muted">{{ formatDate(l.created_at) }}</td>
          <td>
            <div class="user-cell">
              <div class="avatar">{{ initials(userById(l.user_id)) }}</div>
              <span>{{ userLabel(userById(l.user_id)) }}</span>
            </div>
          </td>
          <td>
            <span v-if="describeParts(l).prefix" class="event-prefix">{{ describeParts(l).prefix }}</span>
            <a v-if="describeParts(l).link" class="event-link" href="#" @click.prevent>{{ describeParts(l).link }}</a>
            <span v-if="describeParts(l).suffix" class="event-suffix"> {{ describeParts(l).suffix }}</span>
          </td>
        </tr>
      </tbody>
    </table>

    <div class="ms-footer2">
      <div class="pager">
        <button :disabled="page === 1" @click="page = 1">«</button>
        <button :disabled="page === 1" @click="page--">‹</button>
        <span class="range">{{ rangeFrom }}-{{ rangeTo }} из {{ filtered.length }}</span>
        <button :disabled="rangeTo >= filtered.length" @click="page++">›</button>
        <button :disabled="rangeTo >= filtered.length" @click="page = Math.ceil(filtered.length / perPage)">»</button>
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { listAudit, type AuditLog } from '../api/audit'
import { listUsers, type User } from '../api/users'
import { apiErrorMessage } from '../api/client'
import MsButton from '../components/MsButton.vue'
import MsIcon from '../components/MsIcon.vue'

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
.page { font-size: 13px; }
.ms-title { display: flex; align-items: center; gap: 8px; font-size: 20px; font-weight: 600; color: #1f2328; margin-bottom: 12px; }
.ms-help { width: 20px; height: 20px; border-radius: 50%; border: 1px solid #b8c0c8; background: transparent; color: #57606a; cursor: pointer; display: inline-flex; align-items: center; justify-content: center; padding: 0; }
.ms-help:hover { background: #f0f2f5; }
.ms-refresh { width: 24px; height: 24px; padding: 0; display: inline-flex; align-items: center; justify-content: center; border: none; background: transparent; color: #57606a; cursor: pointer; }
.ms-refresh:hover { color: #2c5d9c; }
.ms-toolbar { display: flex; align-items: center; gap: 6px; margin-bottom: 12px; flex-wrap: wrap; }
.toolbar-spacer { flex: 1; }
.ms-input { flex: 1; max-width: 320px; height: 30px; padding: 0 10px; font-size: 13px; border: 1px solid #d0d7de; border-radius: 4px; background: #fff; color: #1f2328; }
.ms-counter { min-width: 40px; height: 30px; padding: 0 10px; display: inline-flex; align-items: center; justify-content: center; border: 1px solid #d0d7de; border-radius: 4px; background: #fff; color: #8c959f; font-variant-numeric: tabular-nums; font-size: 13px; }
.ms-filter { background: #eef1f5; border: 1px solid #d8dee4; border-radius: 4px; padding: 10px 14px 12px; margin-bottom: 12px; }
.filter-actions { display: flex; align-items: center; gap: 6px; margin-bottom: 10px; }
.filter-grid { display: grid; grid-template-columns: repeat(6, minmax(0, 1fr)); gap: 10px 14px; }
.filter-field { display: flex; flex-direction: column; gap: 3px; }
.filter-label { font-size: 12px; color: #57606a; display: flex; align-items: center; gap: 5px; }
.filter-label .dot { width: 8px; height: 8px; border-radius: 50%; background: #2c5d9c; flex-shrink: 0; }
.filter-field input, .filter-field select { height: 28px; padding: 0 8px; font-size: 12px; border: 1px solid #d0d7de; border-radius: 3px; background: #fff; color: #1f2328; }

.ms-table2 { width: 100%; border-collapse: collapse; background: #fff; font-size: 13px; }
.ms-table2 thead th { background: #fff; color: #2c5d9c; font-weight: 500; padding: 8px 10px; text-align: left; border-bottom: 1px solid #d8dee4; white-space: nowrap; font-size: 12px; cursor: pointer; user-select: none; }
.ms-table2 thead th:hover { background: #f6f8fa; }
.ms-table2 tbody td { padding: 7px 10px; border-bottom: 1px solid #eaeef2; vertical-align: middle; }
.ms-table2 tbody tr.row:hover { background: #f6f8fa; }
.ms-table2 .col-date { width: 150px; }
.ms-table2 .sort { font-size: 9px; margin-left: 3px; }
.ms-table2 .empty { text-align: center; padding: 24px; color: #8c959f; }
.muted { color: #8c959f; }

.user-cell { display: flex; align-items: center; gap: 8px; }
.avatar { width: 24px; height: 24px; border-radius: 50%; background: #2c5d9c; color: #fff; font-size: 11px; font-weight: 600; display: inline-flex; align-items: center; justify-content: center; flex-shrink: 0; }
.event-prefix { color: #57606a; }
.event-link { color: #2c5d9c; text-decoration: none; }
.event-link:hover { text-decoration: underline; }
.event-suffix { color: #57606a; }

.ms-footer2 { display: flex; align-items: center; justify-content: space-between; padding: 6px 10px; font-size: 12px; color: #57606a; background: #fff; border-top: 1px solid #eaeef2; }
.pager { display: flex; align-items: center; gap: 4px; }
.pager button { width: 22px; height: 22px; padding: 0; border: 1px solid #d0d7de; background: #fff; border-radius: 3px; cursor: pointer; font-size: 12px; color: #1f2328; }
.pager button:disabled { opacity: 0.4; cursor: default; }
.pager .range { margin: 0 6px; font-variant-numeric: tabular-nums; }
</style>