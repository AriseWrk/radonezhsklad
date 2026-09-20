<template>
  <div>
    <!-- Верхний тулбар -->
    <div class="page-title-bar">
      <div class="page-title">
        <span>Внутренние заказы</span>
        <span class="refresh" @click="load" title="Обновить">↻</span>
      </div>
      <div class="page-actions">
        <button class="btn primary icon-btn" @click="create">
          <span class="plus">+</span> Заказ
        </button>
        <button class="btn" @click="showFilter = !showFilter">Фильтр</button>
        <input v-model="search" class="search-input" placeholder="Номер или комментарий" />
        <div class="counter" :class="{ active: selected.size > 0 }">{{ selected.size }}</div>
        <select class="mini-select">
          <option>Изменить</option>
          <option>Удалить</option>
        </select>
        <select class="mini-select" v-model="filterStatus" @change="load">
          <option value="">Статус</option>
          <option value="draft">Черновик</option>
          <option value="posted">Проведён</option>
          <option value="cancelled">Отменён</option>
        </select>
        <button class="btn" disabled>Создать</button>
        <div class="print-dropdown">
          <button class="btn" @click="printSelected">
            🖨 Печать
          </button>
        </div>
        <button class="btn icon-only" title="Настройки">⚙</button>
      </div>
    </div>

    <!-- Фильтр-панель -->
    <div v-if="showFilter" class="filter-panel">
      <div class="filter-grid">
        <!-- Ряд 1 -->
        <div class="filter-actions-cell">
          <button class="btn-find" @click="page = 1">Найти</button>
          <button class="btn-clear" @click="clearFilters">Очистить</button>
          <div class="arrow-buttons">
            <button class="arrow-btn" title="Свернуть">▴</button>
            <button class="arrow-btn" title="Развернуть">▾</button>
            <button class="arrow-btn" title="Настройки">⚙</button>
          </div>
        </div>
        <div class="filter-field">
          <label><input type="radio" name="period" /> Период</label>
          <div class="date-range">
            <input v-model="filterDateFrom" type="date" />
            <input v-model="filterDateTo" type="date" />
            <span class="hint">вч · с · нед · мес</span>
          </div>
        </div>
        <div class="filter-field">
          <label><input type="radio" name="period" /> Товар или группа</label>
          <select><option>—</option></select>
        </div>
        <div class="filter-field">
          <label><input type="radio" name="period" /> Склад</label>
          <select v-model="filterWarehouse" @change="load">
            <option value="">—</option>
            <option v-for="w in warehouses" :key="w.id" :value="w.id">{{ w.name }}</option>
          </select>
        </div>
        <div class="filter-field">
          <label><input type="radio" name="period" /> Проект</label>
          <select><option>—</option></select>
        </div>
        <div class="filter-field">
          <label><input type="radio" name="period" /> Организация</label>
          <select v-model="filterOrg" @change="load">
            <option value="">—</option>
            <option v-for="o in organizations" :key="o.id" :value="o.id">{{ o.name }}</option>
          </select>
        </div>

        <!-- Ряд 2 -->
        <div class="filter-field">
          <label>Статус</label>
          <select v-model="filterStatus2" @change="load">
            <option value="">—</option>
            <option value="draft">Черновик</option>
            <option value="posted">Проведён</option>
            <option value="cancelled">Отменён</option>
          </select>
        </div>
        <div class="filter-field">
          <label>Проведено</label>
          <select><option>—</option><option>Да</option><option>Нет</option></select>
        </div>
        <div class="filter-field">
          <label>Напечатано</label>
          <select><option>—</option><option>Да</option><option>Нет</option></select>
        </div>
        <div class="filter-field">
          <label>Отправлено</label>
          <select><option>—</option><option>Да</option><option>Нет</option></select>
        </div>
        <div class="filter-field">
          <label>Владелец-сотрудник</label>
          <select><option>—</option></select>
        </div>
        <div class="filter-field">
          <label>Владелец-отдел</label>
          <select><option>—</option></select>
        </div>

        <!-- Ряд 3 -->
        <div class="filter-field">
          <label>Общий доступ</label>
          <select><option>—</option></select>
        </div>
        <div class="filter-field">
          <label>Когда изменен</label>
          <div class="date-range">
            <input v-model="filterChangedFrom" type="date" />
            <input v-model="filterChangedTo" type="date" />
            <span class="hint">вч · с · нед · мес</span>
          </div>
        </div>
        <div class="filter-field">
          <label>Кто изменил</label>
          <select><option>—</option></select>
        </div>
        <div class="filter-field"></div>
        <div class="filter-field"></div>
        <div class="filter-field"></div>
      </div>
    </div>

    <div v-if="error" class="error-box">{{ error }}</div>

    <!-- Таблица -->
    <table class="ms-table moysklad-orders">
      <thead>
        <tr>
          <th class="chk-col"><input type="checkbox" :checked="allChecked" @change="toggleAll" /></th>
          <th style="width:90px" @click="sortBy('number')">№ <span v-if="sortKey === 'number'">{{ sortDir === 'asc' ? '↑' : '↓' }}</span></th>
          <th style="width:150px" @click="sortBy('doc_date')">Время <span v-if="sortKey === 'doc_date'">{{ sortDir === 'asc' ? '↑' : '↓' }}</span></th>
          <th>Организация</th>
          <th class="num" style="width:130px" @click="sortBy('total')">Сумма <span v-if="sortKey === 'total'">{{ sortDir === 'asc' ? '↑' : '↓' }}</span></th>
          <th class="num" style="width:130px">Отгружено</th>
          <th class="num" style="width:130px">Отправлено</th>
          <th style="width:150px">Склад</th>
          <th style="width:130px">Напечатано</th>
          <th>Комментарий</th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="loading"><td colspan="10" class="muted" style="text-align:center;padding:24px">Загрузка...</td></tr>
        <tr v-else-if="filtered.length === 0"><td colspan="10" class="muted" style="text-align:center;padding:24px">Нет заказов</td></tr>
        <tr
          v-else
          v-for="o in paginated"
          :key="o.id"
          class="clickable"
          :class="{ selected: selected.has(o.id) }"
          @click="open(o)"
        >
          <td class="chk-col" @click.stop>
            <input type="checkbox" :checked="selected.has(o.id)" @change="toggleSelect(o.id)" />
          </td>
          <td class="link mono">{{ o.number }}</td>
          <td class="muted">{{ formatDate(o.doc_date) }}</td>
          <td>{{ organizationName(o.organization_id) }}</td>
          <td class="num"><strong>{{ formatMoney(o.total) }}</strong></td>
          <td class="num muted">{{ formatMoney(o.shipped_amount ?? 0) }}</td>
          <td class="num">
            <span v-if="o.sent_at" class="link">{{ formatMoney(o.total) }}</span>
            <span v-else class="muted">0,00</span>
          </td>
          <td class="muted">{{ o.warehouse_name || '' }}</td>
          <td>
            <span v-if="o.is_printed || o.printed_at" class="badge badge-printed">Напечатан</span>
            <span v-else-if="o.sent_at" class="badge badge-sent">Отправлен</span>
          </td>
          <td class="muted">{{ o.comment || '' }}</td>
        </tr>
      </tbody>
    </table>

    <!-- Футер -->
    <div class="ms-footer moysklad-footer">
      <div class="ms-pager">
        <button :disabled="page === 1" @click="page--">◀</button>
        <button :disabled="page === 1" @click="page = 1">↤</button>
        <span>{{ rangeFrom }}–{{ rangeTo }} из {{ filtered.length }}</span>
        <button :disabled="rangeTo >= filtered.length" @click="page = Math.ceil(filtered.length / perPage)">↦</button>
        <button :disabled="rangeTo >= filtered.length" @click="page++">▶</button>
      </div>
      <button class="show-totals" @click="showTotals = !showTotals">
        <span class="chev">{{ showTotals ? '▾' : '▸' }}</span>
        {{ showTotals ? 'Скрыть итоги' : 'Показать итоги' }}
      </button>
    </div>

    <!-- Итоги -->
    <div v-if="showTotals" class="totals-panel">
      <div class="totals-row">
        <span class="totals-label">Всего заказов:</span>
        <span class="totals-val">{{ filtered.length }}</span>
      </div>
      <div class="totals-row">
        <span class="totals-label">Сумма:</span>
        <span class="totals-val">{{ formatMoney(totalSum) }}</span>
      </div>
      <div class="totals-row">
        <span class="totals-label">Отгружено:</span>
        <span class="totals-val">{{ formatMoney(totalShipped) }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { listInternalOrders, type InternalOrder } from '../api/internalOrders'
import { listWarehouses, type Warehouse } from '../api/warehouses'
import { listOrganizations, type Organization } from '../api/suppliers'
import { apiErrorMessage } from '../api/client'

const items = ref<InternalOrder[]>([])
const warehouses = ref<Warehouse[]>([])
const organizations = ref<Organization[]>([])
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
  const day = String(d.getDate()).padStart(2, '0')
  const mon = String(d.getMonth() + 1).padStart(2, '0')
  const yr = d.getFullYear()
  const hh = String(d.getHours()).padStart(2, '0')
  const mm = String(d.getMinutes()).padStart(2, '0')
  return `${day}.${mon}.${yr} ${hh}:${mm}`
}
function formatMoney(n: number) {
  return n.toLocaleString('ru-RU', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
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

function clearFilters() {
  search.value = ''
  filterStatus.value = ''
  filterStatus2.value = ''
  filterWarehouse.value = ''
  filterOrg.value = ''
  filterDateFrom.value = ''
  filterDateTo.value = ''
  filterChangedFrom.value = ''
  filterChangedTo.value = ''
  page.value = 1
  load()
}

onMounted(async () => {
  try {
    const [w, orgs] = await Promise.all([listWarehouses(), listOrganizations()])
    warehouses.value = w
    organizations.value = orgs
  } catch { /* ignore */ }
  await load()
})
</script>

<style scoped>
/* Toolbar */
.icon-btn { display: inline-flex; align-items: center; gap: 6px; }
.icon-btn .plus { font-size: 16px; font-weight: 700; line-height: 1; }

.search-input {
  padding: 6px 10px; font-size: 13px;
  border: 1px solid #d0d7de; border-radius: 4px;
  background: #fff; color: #1f2328; min-width: 240px;
}
.counter {
  display: inline-flex;
  align-items: center; justify-content: center;
  min-width: 40px; height: 30px; padding: 0 10px;
  background: #fff; border: 1px solid #d0d7de; border-radius: 4px;
  font-size: 13px; color: #8c959f;
  font-variant-numeric: tabular-nums;
}
.counter.active { color: #2c5d9c; border-color: #2c5d9c; font-weight: 600; }

.mini-select {
  padding: 6px 10px; font-size: 13px;
  border: 1px solid #d0d7de; border-radius: 4px;
  background: #fff; color: #1f2328;
  max-width: 140px;
}

.btn.icon-only { padding: 6px 10px; font-size: 15px; }
.print-dropdown { display: inline-block; }

/* Filter panel */
.filter-panel {
  background: #eef1f5;
  border: 1px solid #d8dee4;
  border-radius: 4px;
  padding: 12px 16px;
  margin-bottom: 12px;
}
.filter-grid {
  display: grid;
  grid-template-columns: repeat(6, minmax(0, 1fr));
  gap: 10px 14px;
}
.filter-actions-cell {
  display: flex;
  align-items: flex-end;
  gap: 6px;
  padding-bottom: 4px;
}
.arrow-buttons {
  display: inline-flex;
  gap: 2px;
  margin-left: 4px;
}
.arrow-btn {
  width: 28px; height: 28px;
  padding: 0;
  border: 1px solid #d0d7de;
  background: #fff;
  border-radius: 4px;
  font-size: 12px;
  color: #57606a;
  cursor: pointer;
  line-height: 1;
}
.arrow-btn:hover { background: #f6f8fa; }

.filter-field {
  display: flex;
  flex-direction: column;
  gap: 3px;
  font-size: 12px;
  color: #57606a;
}
.filter-field > label {
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 4px;
  min-height: 16px;
}
.filter-field > label input[type="radio"] {
  margin: 0; width: auto; transform: scale(0.8);
}
.filter-field input,
.filter-field select {
  padding: 5px 8px;
  font-size: 13px;
  border: 1px solid #d0d7de;
  border-radius: 3px;
  background: #fff;
  color: #1f2328;
  width: 100%;
}
.date-range {
  display: flex;
  align-items: center;
  gap: 4px;
  position: relative;
}
.date-range input { flex: 1; min-width: 0; }
.date-range .hint {
  font-size: 10px;
  color: #8c959f;
  white-space: nowrap;
  position: absolute;
  right: 4px;
  bottom: -14px;
}

/* Table */
.moysklad-orders tbody tr.selected { background: #eef4ff; }
.moysklad-orders .mono { font-family: monospace; font-size: 12px; }
.moysklad-orders .link { color: #2c5d9c; cursor: pointer; }
.moysklad-orders .link:hover { text-decoration: underline; }
tr.clickable { cursor: pointer; }

.badge {
  display: inline-block;
  padding: 2px 10px;
  border-radius: 3px;
  font-size: 11px;
  font-weight: 600;
  color: #fff;
}
.badge-printed { background: #29aae1; }
.badge-sent    { background: #f0a020; }

/* Footer */
.moysklad-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.show-totals {
  border: none;
  background: transparent;
  color: #2c5d9c;
  font-size: 13px;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 8px;
}
.show-totals:hover { text-decoration: underline; }
.show-totals .chev { font-size: 10px; }

.totals-panel {
  background: #fff;
  border: 1px solid #d8dee4;
  border-top: none;
  padding: 12px 20px;
  font-size: 13px;
}
.totals-row {
  display: flex;
  justify-content: space-between;
  padding: 4px 0;
  max-width: 420px;
}
.totals-label { color: #57606a; }
.totals-val { font-weight: 600; color: #1f2328; font-variant-numeric: tabular-nums; }
</style>