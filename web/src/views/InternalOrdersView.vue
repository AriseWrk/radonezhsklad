<template>
  <div class="page">
    <!-- Заголовок -->
    <div class="ms-title">
      <button class="ms-help" title="Справка"><MsIcon name="help" :size="14" /></button>
      <span>Внутренние заказы</span>
      <button class="ms-refresh" @click="load" title="Обновить"><MsIcon name="refresh" :size="14" /></button>
    </div>

    <!-- Тулбар -->
    <div class="ms-toolbar">
      <MsButton variant="primary" icon="plus" @click="create">Заказ</MsButton>
      <MsButton icon="filter" @click="showFilter = !showFilter">Фильтр</MsButton>
      <input v-model="search" class="ms-input" placeholder="Номер или комментарий" />
      <div class="ms-counter" :class="{ active: selected.size > 0 }">{{ selected.size }}</div>
      <select class="ms-select" @change="onBulkAction">
        <option value="">Изменить</option>
        <option value="delete">Удалить</option>
      </select>
      <select v-model="filterStatus" class="ms-select" @change="load">
        <option value="">Статус</option>
        <option value="draft">Черновик</option>
        <option value="posted">Проведён</option>
        <option value="cancelled">Отменён</option>
      </select>
      <MsButton icon="print" @click="printSelected">Печать</MsButton>
      <MsButton variant="icon" icon="gear" title="Настройки" />
    </div>

    <!-- Фильтр-панель -->
    <div v-if="showFilter" class="ms-filter">
      <div class="filter-actions">
        <MsButton variant="green" @click="page = 1">Найти</MsButton>
        <MsButton @click="clearFilters">Очистить</MsButton>
        <MsButton variant="icon" icon="chevron-up" title="Свернуть" />
        <MsButton variant="icon" icon="chevron-down" title="Развернуть" />
        <MsButton variant="icon" icon="gear" title="Настройки" />
      </div>

      <div class="filter-grid">
        <div class="filter-field">
          <label class="filter-label"><span class="dot"></span>Период <span class="hint">вч · с · нед · мес</span></label>
          <div class="date-range">
            <input v-model="filterDateFrom" type="date" />
            <input v-model="filterDateTo" type="date" />
          </div>
        </div>
        <div class="filter-field">
          <label class="filter-label"><span class="dot"></span>Товар или группа</label>
          <select><option>—</option></select>
        </div>
        <div class="filter-field">
          <label class="filter-label"><span class="dot"></span>Склад</label>
          <select v-model="filterWarehouse" @change="load">
            <option value="">—</option>
            <option v-for="w in warehouses" :key="w.id" :value="w.id">{{ w.name }}</option>
          </select>
        </div>
        <div class="filter-field">
          <label class="filter-label"><span class="dot"></span>Проект</label>
          <select v-model="filterProject" @change="load">
            <option value="">—</option>
            <option v-for="pr in projects" :key="pr.id" :value="pr.external_id || ''">{{ pr.name }}</option>
          </select>
        </div>
        <div class="filter-field">
          <label class="filter-label"><span class="dot"></span>Организация</label>
          <select v-model="filterOrg" @change="load">
            <option value="">—</option>
            <option v-for="o in organizations" :key="o.id" :value="o.id">{{ o.name }}</option>
          </select>
        </div>

        <div class="filter-field">
          <label class="filter-label"><span class="dot"></span>Статус</label>
          <select v-model="filterStatus2" @change="load">
            <option value="">—</option>
            <option value="draft">Черновик</option>
            <option value="posted">Проведён</option>
            <option value="cancelled">Отменён</option>
          </select>
        </div>
        <div class="filter-field">
          <label class="filter-label"><span class="dot"></span>Проведено</label>
          <select><option>—</option><option>Да</option><option>Нет</option></select>
        </div>
        <div class="filter-field">
          <label class="filter-label"><span class="dot"></span>Напечатано</label>
          <select><option>—</option><option>Да</option><option>Нет</option></select>
        </div>
        <div class="filter-field">
          <label class="filter-label"><span class="dot"></span>Отправлено</label>
          <select><option>—</option><option>Да</option><option>Нет</option></select>
        </div>
        <div class="filter-field">
          <label class="filter-label"><span class="dot"></span>Владелец-сотрудник</label>
          <select><option>—</option></select>
        </div>
        <div class="filter-field">
          <label class="filter-label"><span class="dot"></span>Владелец-отдел</label>
          <select><option>—</option></select>
        </div>

        <div class="filter-field">
          <label class="filter-label"><span class="dot"></span>Общий доступ</label>
          <select><option>—</option></select>
        </div>
        <div class="filter-field">
          <label class="filter-label"><span class="dot"></span>Когда изменен <span class="hint">вч · с · нед · мес</span></label>
          <div class="date-range">
            <input v-model="filterChangedFrom" type="date" />
            <input v-model="filterChangedTo" type="date" />
          </div>
        </div>
        <div class="filter-field">
          <label class="filter-label"><span class="dot"></span>Кто изменил</label>
          <select><option>—</option></select>
        </div>
      </div>
    </div>

    <div v-if="error" class="error-box">{{ error }}</div>

    <!-- Таблица -->
    <table class="ms-table2">
      <thead>
        <tr>
          <th class="col-chk"><input type="checkbox" :checked="allChecked" @change="toggleAll" /></th>
          <th class="col-num" @click="sortBy('number')">№<span v-if="sortKey === 'number'" class="sort">{{ sortDir === 'asc' ? '▲' : '▼' }}</span></th>
          <th class="col-date" @click="sortBy('doc_date')">Время<span v-if="sortKey === 'doc_date'" class="sort">{{ sortDir === 'asc' ? '▲' : '▼' }}</span></th>
          <th>Организация</th>
          <th class="col-num-right" @click="sortBy('total')">Сумма<span v-if="sortKey === 'total'" class="sort">{{ sortDir === 'asc' ? '▲' : '▼' }}</span></th>
          <th class="col-num-right">Отгружено</th>
          <th class="col-num-right">Отправлено</th>
          <th class="col-status">Напечатано</th>
          <th>Комментарий</th>
          <th class="col-actions"><MsIcon name="gear" :size="14" /></th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="loading"><td colspan="10" class="empty">Загрузка...</td></tr>
        <tr v-else-if="filtered.length === 0"><td colspan="10" class="empty">Нет заказов</td></tr>
        <tr
          v-else
          v-for="o in paginated"
          :key="o.id"
          class="row"
          :class="{ selected: selected.has(o.id) }"
          @click="open(o)"
        >
          <td class="col-chk" @click.stop>
            <input type="checkbox" :checked="selected.has(o.id)" @change="toggleSelect(o.id)" />
          </td>
          <td class="col-num"><span class="link">{{ o.number }}</span></td>
          <td class="col-date">{{ formatDate(o.doc_date) }}</td>
          <td>{{ organizationName(o.organization_id) }}</td>
          <td class="col-num-right"><b>{{ formatMoney(o.total) }}</b></td>
          <td class="col-num-right">{{ formatMoney(o.shipped_amount ?? 0) }}</td>
          <td class="col-num-right">
            <span v-if="o.sent_at">{{ formatMoney(o.total) }}</span>
            <span v-else>0,00</span>
          </td>
          <td class="col-status">
            <span v-if="o.sent_at" class="badge">Отправлен</span>
            <span v-else-if="o.is_printed || o.printed_at" class="badge">Напечатан</span>
          </td>
          <td class="comment">{{ o.comment || '' }}</td>
          <td class="col-actions" @click.stop>
            <button class="row-menu" @click="open(o)"><MsIcon name="dots" :size="14" /></button>
          </td>
        </tr>
      </tbody>
    </table>

    <!-- Футер -->
    <div class="ms-footer2">
      <div class="pager">
        <button :disabled="page === 1" @click="page = 1" title="Первая">«</button>
        <button :disabled="page === 1" @click="page--" title="Назад">‹</button>
        <span class="range">{{ rangeFrom }}-{{ rangeTo }} из {{ filtered.length }}</span>
        <button :disabled="rangeTo >= filtered.length" @click="page++" title="Вперёд">›</button>
        <button :disabled="rangeTo >= filtered.length" @click="page = Math.ceil(filtered.length / perPage)" title="Последняя">»</button>
      </div>
      <button class="show-totals" @click="showTotals = !showTotals">
        <span class="sigma">Σ</span> {{ showTotals ? 'Скрыть итоги' : 'Показать итоги' }}
      </button>
    </div>

    <div v-if="showTotals" class="totals-panel">
      <div class="totals-row">
        <span>Всего заказов:</span><b>{{ filtered.length }}</b>
      </div>
      <div class="totals-row">
        <span>Сумма:</span><b>{{ formatMoney(totalSum) }}</b>
      </div>
      <div class="totals-row">
        <span>Отгружено:</span><b>{{ formatMoney(totalShipped) }}</b>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { listInternalOrders, deleteInternalOrder, type InternalOrder } from '../api/internalOrders'
import { listWarehouses, type Warehouse } from '../api/warehouses'
import { listOrganizations, type Organization } from '../api/suppliers'
import { listProjects, type Project } from '../api/projects'
import { apiErrorMessage } from '../api/client'
import MsButton from '../components/MsButton.vue'
import MsIcon from '../components/MsIcon.vue'

const items = ref<InternalOrder[]>([])
const warehouses = ref<Warehouse[]>([])
const organizations = ref<Organization[]>([])
const projects = ref<Project[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const showFilter = ref(true)
const showTotals = ref(false)
const page = ref(1)
const perPage = 100

const search = ref('')
const filterStatus = ref('')
const filterStatus2 = ref('')
const filterWarehouse = ref('')
const filterOrg = ref('')
const filterProject = ref('')
const filterDateFrom = ref('')
const filterDateTo = ref('')
const filterChangedFrom = ref('')
const filterChangedTo = ref('')

const sortKey = ref<'number' | 'doc_date' | 'total'>('doc_date')
const sortDir = ref<'asc' | 'desc'>('desc')

const selected = ref<Set<string>>(new Set())
const router = useRouter()

function organizationName(id?: string) {
  if (!id) return organizations.value.find((o) => o.is_default)?.name ?? '—'
  return organizations.value.find((o) => o.id === id)?.name ?? '—'
}
function formatDate(s: string) {
  const d = new Date(s)
  const dd = String(d.getDate()).padStart(2, '0')
  const mm = String(d.getMonth() + 1).padStart(2, '0')
  const yy = d.getFullYear()
  const hh = String(d.getHours()).padStart(2, '0')
  const mi = String(d.getMinutes()).padStart(2, '0')
  return `${dd}.${mm}.${yy} ${hh}:${mi}`
}
function formatMoney(n: number) {
  return (n ?? 0).toLocaleString('ru-RU', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function sortBy(k: typeof sortKey.value) {
  if (sortKey.value === k) sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  else { sortKey.value = k; sortDir.value = 'desc' }
  page.value = 1
}

const filtered = computed(() => {
  let rows = items.value
  if (search.value) {
    const q = search.value.toLowerCase()
    rows = rows.filter((o) =>
      (o.number ?? '').toLowerCase().includes(q) ||
      (o.comment ?? '').toLowerCase().includes(q)
    )
  }
  const activeStatus = filterStatus.value || filterStatus2.value
  if (activeStatus) rows = rows.filter((o) => o.status === activeStatus)
  if (filterOrg.value) rows = rows.filter((o) => o.organization_id === filterOrg.value)
  if (filterWarehouse.value) rows = rows.filter((o) => o.warehouse_id === filterWarehouse.value)
  if (filterProject.value) rows = rows.filter((o) => o.project === filterProject.value)
  if (filterDateFrom.value) {
    const t = new Date(filterDateFrom.value).getTime()
    rows = rows.filter((o) => new Date(o.doc_date).getTime() >= t)
  }
  if (filterDateTo.value) {
    const t = new Date(filterDateTo.value).getTime() + 86400000 - 1
    rows = rows.filter((o) => new Date(o.doc_date).getTime() <= t)
  }
  return [...rows].sort((a, b) => {
    const av = a[sortKey.value]; const bv = b[sortKey.value]
    if (typeof av === 'number' && typeof bv === 'number') {
      return sortDir.value === 'asc' ? av - bv : bv - av
    }
    const as = String(av ?? ''); const bs = String(bv ?? '')
    return sortDir.value === 'asc' ? as.localeCompare(bs) : bs.localeCompare(as)
  })
})

const paginated = computed(() => {
  const from = (page.value - 1) * perPage
  return filtered.value.slice(from, from + perPage)
})
const rangeFrom = computed(() => filtered.value.length === 0 ? 0 : (page.value - 1) * perPage + 1)
const rangeTo = computed(() => Math.min(page.value * perPage, filtered.value.length))

const totalSum = computed(() => filtered.value.reduce((s, o) => s + (o.total ?? 0), 0))
const totalShipped = computed(() => filtered.value.reduce((s, o) => s + (o.shipped_amount ?? 0), 0))

const allChecked = computed(() =>
  paginated.value.length > 0 && paginated.value.every((o) => selected.value.has(o.id))
)
function toggleAll(e: Event) {
  const on = (e.target as HTMLInputElement).checked
  selected.value = on ? new Set(paginated.value.map((o) => o.id)) : new Set()
}
function toggleSelect(id: string) {
  const s = new Set(selected.value); s.has(id) ? s.delete(id) : s.add(id); selected.value = s
}

async function load() {
  loading.value = true; error.value = null
  try {
    items.value = await listInternalOrders({
      status: filterStatus.value || undefined,
      warehouse_id: filterWarehouse.value || undefined,
    })
    page.value = 1
  } catch (e) { error.value = apiErrorMessage(e) }
  finally { loading.value = false }
}

function open(o: InternalOrder) { router.push(`/internal-orders/${o.id}`) }
function create() { router.push('/internal-orders/new') }
function printSelected() { window.print() }

async function onBulkAction(e: Event) {
  const v = (e.target as HTMLSelectElement).value
  ;(e.target as HTMLSelectElement).value = ''
  if (v === 'delete') await deleteSelected()
}

async function deleteSelected() {
  const ids = Array.from(selected.value)
  if (ids.length === 0) { error.value = 'Не выбрано ни одного заказа'; return }
  if (!confirm(`Удалить выбранные заказы (${ids.length})? Действие необратимо.`)) return
  const failed: string[] = []
  for (const id of ids) {
    try { await deleteInternalOrder(id) }
    catch (e) { failed.push(`${id}: ${apiErrorMessage(e)}`) }
  }
  selected.value = new Set()
  await load()
  if (failed.length > 0) error.value = `Не удалось удалить ${failed.length}: ` + failed.join('; ')
}

function clearFilters() {
  search.value = ''
  filterStatus.value = ''
  filterStatus2.value = ''
  filterWarehouse.value = ''
  filterOrg.value = ''
  filterProject.value = ''
  filterDateFrom.value = ''
  filterDateTo.value = ''
  filterChangedFrom.value = ''
  filterChangedTo.value = ''
  page.value = 1
  load()
}

onMounted(async () => {
  try {
    const [w, orgs, prj] = await Promise.all([
      listWarehouses(),
      listOrganizations(),
      listProjects().catch(() => [] as Project[]),
    ])
    warehouses.value = w
    organizations.value = orgs
    projects.value = prj
  } catch { /* ignore */ }
  await load()
})
</script>

<style scoped>
.page { font-size: 13px; }

/* Заголовок */
.ms-title {
  display: flex; align-items: center; gap: 8px;
  font-size: 20px; font-weight: 600; color: #1f2328;
  margin-bottom: 12px;
}
.ms-help {
  width: 20px; height: 20px; border-radius: 50%;
  border: 1px solid #b8c0c8; background: transparent;
  color: #57606a; cursor: pointer;
  display: inline-flex; align-items: center; justify-content: center;
  padding: 0;
}
.ms-help:hover { background: #f0f2f5; }
.ms-refresh {
  width: 24px; height: 24px; padding: 0;
  display: inline-flex; align-items: center; justify-content: center;
  border: none; background: transparent;
  color: #57606a; cursor: pointer;
}
.ms-refresh:hover { color: #2c5d9c; }

/* Тулбар */
.ms-toolbar {
  display: flex; align-items: center; gap: 6px;
  margin-bottom: 12px;
  flex-wrap: wrap;
}
.ms-input {
  flex: 1; max-width: 320px;
  height: 30px; padding: 0 10px;
  font-size: 13px;
  border: 1px solid #d0d7de; border-radius: 4px;
  background: #fff; color: #1f2328;
}
.ms-input::placeholder { color: #8c959f; }
.ms-input:focus { outline: 2px solid rgba(44,93,156,0.3); border-color: #2c5d9c; }

.ms-counter {
  min-width: 40px; height: 30px; padding: 0 10px;
  display: inline-flex; align-items: center; justify-content: center;
  border: 1px solid #d0d7de; border-radius: 4px;
  background: #fff; color: #8c959f;
  font-variant-numeric: tabular-nums;
  font-size: 13px;
}
.ms-counter.active { color: #2c5d9c; border-color: #2c5d9c; font-weight: 600; }

.ms-select {
  height: 30px; padding: 0 8px;
  font-size: 13px;
  border: 1px solid #d0d7de; border-radius: 4px;
  background: #fff; color: #1f2328;
  max-width: 160px;
}

/* Фильтр-панель */
.ms-filter {
  background: #eef1f5;
  border: 1px solid #d8dee4;
  border-radius: 4px;
  padding: 10px 14px 12px;
  margin-bottom: 12px;
}
.filter-actions {
  display: flex; align-items: center; gap: 6px;
  margin-bottom: 10px;
}
.filter-grid {
  display: grid;
  grid-template-columns: repeat(6, minmax(0, 1fr));
  gap: 10px 14px;
}
.filter-field { display: flex; flex-direction: column; gap: 3px; }
.filter-label {
  font-size: 12px; color: #57606a;
  display: flex; align-items: center; gap: 5px;
}
.filter-label .dot {
  width: 8px; height: 8px; border-radius: 50%;
  background: #2c5d9c; flex-shrink: 0;
}
.filter-label .hint {
  font-size: 10px; color: #8c959f; margin-left: 4px;
}
.filter-field input,
.filter-field select {
  height: 28px; padding: 0 8px;
  font-size: 12px;
  border: 1px solid #d0d7de; border-radius: 3px;
  background: #fff; color: #1f2328;
}
.date-range { display: flex; gap: 4px; }
.date-range input { flex: 1; min-width: 0; }

/* Таблица */
.ms-table2 {
  width: 100%;
  border-collapse: collapse;
  background: #fff;
  font-size: 13px;
}
.ms-table2 thead th {
  background: #fff;
  color: #2c5d9c;
  font-weight: 500;
  padding: 8px 10px;
  text-align: left;
  border-bottom: 1px solid #d8dee4;
  white-space: nowrap;
  cursor: pointer;
  user-select: none;
  font-size: 12px;
}
.ms-table2 thead th:hover { background: #f6f8fa; }
.ms-table2 tbody td {
  padding: 7px 10px;
  border-bottom: 1px solid #eaeef2;
  color: #1f2328;
  vertical-align: middle;
}
.ms-table2 tbody tr.row { cursor: pointer; }
.ms-table2 tbody tr.row:hover { background: #f6f8fa; }
.ms-table2 tbody tr.selected { background: #fffcc2 !important; }
.ms-table2 .col-chk { width: 30px; text-align: center; }
.ms-table2 .col-chk input { width: 13px; height: 13px; }
.ms-table2 .col-num { width: 80px; }
.ms-table2 .col-date { width: 130px; color: #57606a; }
.ms-table2 .col-num-right { width: 120px; text-align: right; font-variant-numeric: tabular-nums; }
.ms-table2 .col-status { width: 130px; }
.ms-table2 .col-actions { width: 34px; text-align: center; }
.ms-table2 .link { color: #2c5d9c; font-weight: 500; }
.ms-table2 .link:hover { text-decoration: underline; }
.ms-table2 .comment { color: #8c959f; max-width: 380px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.ms-table2 .sort { font-size: 9px; margin-left: 3px; }
.ms-table2 .empty { text-align: center; padding: 24px; color: #8c959f; }

.badge {
  display: inline-block;
  padding: 2px 10px;
  border-radius: 3px;
  background: #2196f3;
  color: #fff;
  font-size: 11px;
  font-weight: 600;
}
.row-menu {
  width: 22px; height: 22px; padding: 0;
  border: none; background: transparent;
  color: #57606a; cursor: pointer;
  border-radius: 3px;
}
.row-menu:hover { background: #eaeef2; color: #1f2328; }

/* Футер */
.ms-footer2 {
  display: flex; align-items: center; justify-content: space-between;
  padding: 6px 10px;
  font-size: 12px; color: #57606a;
  background: #fff;
  border-top: 1px solid #eaeef2;
}
.pager { display: flex; align-items: center; gap: 4px; }
.pager button {
  width: 22px; height: 22px; padding: 0;
  border: 1px solid #d0d7de; background: #fff;
  border-radius: 3px; cursor: pointer;
  font-size: 12px; color: #1f2328;
  display: inline-flex; align-items: center; justify-content: center;
}
.pager button:disabled { opacity: 0.4; cursor: default; }
.pager button:hover:not(:disabled) { background: #f6f8fa; }
.pager .range { margin: 0 6px; font-variant-numeric: tabular-nums; }
.show-totals {
  border: none; background: transparent;
  color: #2c5d9c; font-size: 12px;
  cursor: pointer; display: inline-flex; align-items: center; gap: 5px;
  padding: 4px 8px;
}
.show-totals:hover { text-decoration: underline; }
.show-totals .sigma { font-weight: 700; }

.totals-panel {
  background: #fff;
  border-top: 1px solid #eaeef2;
  padding: 10px 16px;
  font-size: 13px;
}
.totals-row { display: flex; justify-content: space-between; max-width: 420px; padding: 3px 0; }
.totals-row b { font-variant-numeric: tabular-nums; }
</style>