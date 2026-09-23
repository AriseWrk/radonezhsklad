<template>
  <div class="page">
    <!-- Заголовок -->
    <div class="ms-title">
      <button class="ms-help" title="Справка"><MsIcon name="help" :size="14" /></button>
      <span>Документы</span>
      <button class="ms-refresh" @click="load" title="Обновить"><MsIcon name="refresh" :size="14" /></button>
    </div>

    <!-- Вкладки типов -->
    <div class="ms-tabs">
      <button
        v-for="t in tabs"
        :key="t.key"
        class="tab"
        :class="{ active: activeTab === t.key }"
        @click="setTab(t.key)"
      >
        {{ t.label }}
        <span v-if="tabCount(t.key) > 0" class="tab-count">{{ tabCount(t.key) }}</span>
      </button>
    </div>

    <!-- Тулбар -->
    <div class="ms-toolbar">
      <MsButton variant="primary" icon="plus" @click="openCreate">Документ</MsButton>
      <MsButton icon="filter" @click="showFilter = !showFilter">Фильтр</MsButton>
      <input v-model="filters.search" class="ms-input" placeholder="Номер или комментарий" />
      <div class="ms-counter">{{ filtered.length }}</div>
      <MsButton icon="print" @click="print">Печать</MsButton>
      <MsButton variant="icon" icon="gear" title="Настройки" />
    </div>

    <!-- Фильтр-панель -->
    <div v-if="showFilter" class="ms-filter">
      <div class="filter-actions">
        <MsButton variant="green" @click="page = 1">Найти</MsButton>
        <MsButton @click="clearFilters">Очистить</MsButton>
      </div>
      <div class="filter-grid">
        <div class="filter-field">
          <label class="filter-label"><span class="dot"></span>Период с</label>
          <input v-model="filters.dateFrom" type="date" />
        </div>
        <div class="filter-field">
          <label class="filter-label"><span class="dot"></span>Период по</label>
          <input v-model="filters.dateTo" type="date" />
        </div>
        <div class="filter-field">
          <label class="filter-label"><span class="dot"></span>Склад</label>
          <select v-model="filters.warehouseId" @change="page = 1">
            <option value="">—</option>
            <option v-for="w in warehouses" :key="w.id" :value="w.id">{{ w.name }}</option>
          </select>
        </div>
        <div class="filter-field">
          <label class="filter-label"><span class="dot"></span>Статус</label>
          <select v-model="filters.status" @change="page = 1">
            <option value="">—</option>
            <option value="draft">Черновик</option>
            <option value="posted">Проведён</option>
            <option value="cancelled">Отменён</option>
          </select>
        </div>
        <div class="filter-field">
          <label class="filter-label"><span class="dot"></span>Сумма от</label>
          <input v-model.number="filters.sumFrom" type="number" min="0" />
        </div>
        <div class="filter-field">
          <label class="filter-label"><span class="dot"></span>Сумма до</label>
          <input v-model.number="filters.sumTo" type="number" min="0" />
        </div>
      </div>
    </div>

    <div v-if="error" class="error-box">{{ error }}</div>

    <!-- Таблица -->
    <table class="ms-table2">
      <thead>
        <tr>
          <th class="col-num" @click="sortBy('number')">№<span v-if="sortKey === 'number'" class="sort">{{ sortDir === 'asc' ? '▲' : '▼' }}</span></th>
          <th class="col-date" @click="sortBy('created_at')">Дата<span v-if="sortKey === 'created_at'" class="sort">{{ sortDir === 'asc' ? '▲' : '▼' }}</span></th>
          <th class="col-type">Тип</th>
          <th>Склад</th>
          <th>Комментарий</th>
          <th class="col-num-right">Позиций</th>
          <th class="col-num-right" @click="sortBy('total')">Сумма<span v-if="sortKey === 'total'" class="sort">{{ sortDir === 'asc' ? '▲' : '▼' }}</span></th>
          <th class="col-status">Статус</th>
          <th class="col-actions"></th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="loading"><td colspan="9" class="empty">Загрузка...</td></tr>
        <tr v-else-if="filtered.length === 0"><td colspan="9" class="empty">Нет документов</td></tr>
        <tr v-else v-for="d in paginated" :key="d.id" class="row" @click="openCard(d)">
          <td class="col-num"><span class="link">{{ d.number }}</span></td>
          <td class="col-date">{{ formatDate(d.created_at) }}</td>
          <td class="col-type">{{ typeLabel(d.type) }}</td>
          <td>{{ warehouseName(d.warehouse_id) }}</td>
          <td class="comment">{{ d.comment || '' }}</td>
          <td class="col-num-right">{{ d.items_count ?? 0 }}</td>
          <td class="col-num-right"><b>{{ formatMoney(d.total ?? 0) }}</b></td>
          <td class="col-status">
            <span class="badge" :class="'bg-' + d.status">{{ statusLabel(d.status) }}</span>
          </td>
          <td class="col-actions" @click.stop>
            <button v-if="d.status === 'draft'" class="btn-link-ms" @click="onPost(d)">Провести</button>
            <button v-if="d.status === 'posted'" class="btn-link-ms danger" @click="onCancel(d)">Отменить</button>
            <button class="row-menu" @click="openCard(d)"><MsIcon name="dots" :size="14" /></button>
          </td>
        </tr>
      </tbody>
    </table>

    <!-- Футер -->
    <div class="ms-footer2">
      <div class="pager">
        <button :disabled="page === 1" @click="page = 1">«</button>
        <button :disabled="page === 1" @click="page--">‹</button>
        <span class="range">{{ rangeFrom }}-{{ rangeTo }} из {{ filtered.length }}</span>
        <button :disabled="rangeTo >= filtered.length" @click="page++">›</button>
        <button :disabled="rangeTo >= filtered.length" @click="page = Math.ceil(filtered.length / perPage)">»</button>
      </div>
      <div class="totals-inline">
        <span>{{ filtered.length }} док.</span>
        <b>{{ formatMoney(filteredTotal) }}</b>
      </div>
    </div>

    <!-- Модалка создания -->
    <div v-if="showCreate" class="modal-backdrop" @click.self="closeCreate">
      <form class="card modal" @submit.prevent="onCreate">
        <h2>Новый документ</h2>
        <div v-if="createError" class="error-box">{{ createError }}</div>

        <label>Тип*
          <select v-model="form.type" required>
            <option value="receipt">Приёмка</option>
            <option value="shipment">Отгрузка</option>
            <option value="transfer">Перемещение</option>
          </select>
        </label>

        <label>Склад*
          <select v-model="form.warehouse_id" required>
            <option value="">— выберите —</option>
            <option v-for="w in warehouses" :key="w.id" :value="w.id">{{ w.name }}</option>
          </select>
        </label>

        <label v-if="form.type === 'transfer'">Целевой склад*
          <select v-model="form.target_warehouse_id" required>
            <option value="">— выберите —</option>
            <option v-for="w in warehouses" :key="w.id" :value="w.id">{{ w.name }}</option>
          </select>
        </label>

        <label>Комментарий
          <input v-model="form.comment" />
        </label>

        <div class="items-section">
          <div class="items-head">
            <span>Позиции</span>
            <button type="button" class="btn-link" @click="addItem">+ Добавить</button>
          </div>
          <div v-for="(it, idx) in form.items" :key="idx" class="item-row">
            <select v-model="it.product_id" required>
              <option value="">— товар —</option>
              <option v-for="p in products" :key="p.id" :value="p.id">{{ p.name }}</option>
            </select>
            <input v-model.number="it.quantity" type="number" step="0.001" min="0.001" placeholder="Кол-во" required />
            <input v-model.number="it.price" type="number" step="0.01" min="0" placeholder="Цена" />
            <button type="button" class="btn-link danger-text" @click="removeItem(idx)">×</button>
          </div>
        </div>

        <div class="modal-actions">
          <button type="button" @click="closeCreate">Отмена</button>
          <button class="primary" type="submit" :disabled="saving">
            {{ saving ? 'Сохранение...' : 'Создать' }}
          </button>
        </div>
      </form>
    </div>

    <!-- Модалка деталей -->
    <div v-if="detailDoc" class="modal-backdrop" @click.self="detailDoc = null">
      <div class="card modal big">
        <h2>Документ {{ detailDoc.number }}</h2>
        <div class="doc-info">
          <div><span class="lbl">Тип:</span> {{ typeLabel(detailDoc.type) }}</div>
          <div><span class="lbl">Статус:</span> {{ statusLabel(detailDoc.status) }}</div>
          <div><span class="lbl">Склад:</span> {{ warehouseName(detailDoc.warehouse_id) }}
            <template v-if="detailDoc.target_warehouse_id">
              → {{ warehouseName(detailDoc.target_warehouse_id) }}
            </template>
          </div>
          <div><span class="lbl">Дата:</span> {{ formatDate(detailDoc.created_at) }}</div>
          <div v-if="detailDoc.comment" style="grid-column: 1/-1">
            <span class="lbl">Комментарий:</span> {{ detailDoc.comment }}
          </div>
        </div>
        <table class="ms-table2">
          <thead>
            <tr>
              <th>Товар</th>
              <th class="col-num-right">Кол-во</th>
              <th class="col-num-right">Цена</th>
              <th class="col-num-right">Сумма</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="it in detailDoc.items" :key="it.id">
              <td>{{ productName(it.product_id) }}</td>
              <td class="col-num-right">{{ formatQty(it.quantity) }}</td>
              <td class="col-num-right">{{ formatMoney(it.price) }}</td>
              <td class="col-num-right">{{ formatMoney(it.quantity * it.price) }}</td>
            </tr>
            <tr v-if="!detailDoc.items || detailDoc.items.length === 0">
              <td colspan="4" class="empty">Нет позиций</td>
            </tr>
          </tbody>
        </table>
        <div class="modal-actions">
          <button @click="detailDoc = null">Закрыть</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useRoute } from 'vue-router'
import {
  listDocuments, createDocument, postDocument, cancelDocument,
  type Document, type DocType,
} from '../api/documents'
import { listWarehouses, type Warehouse } from '../api/warehouses'
import { listProducts, type Product } from '../api/products'
import { apiErrorMessage } from '../api/client'
import MsButton from '../components/MsButton.vue'
import MsIcon from '../components/MsIcon.vue'

type TabKey = 'all' | 'receipt' | 'shipment' | 'transfer' | 'inventory' | 'draft' | 'posted' | 'cancelled'

const items = ref<Document[]>([])
const warehouses = ref<Warehouse[]>([])
const products = ref<Product[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const showFilter = ref(false)
const page = ref(1)
const perPage = 50

const activeTab = ref<TabKey>('all')

const filters = reactive({
  dateFrom: '',
  dateTo: '',
  warehouseId: '',
  status: '',
  search: '',
  sumFrom: null as number | null,
  sumTo: null as number | null,
})

const sortKey = ref<'number' | 'created_at' | 'total'>('created_at')
const sortDir = ref<'asc' | 'desc'>('desc')

const showCreate = ref(false)
const saving = ref(false)
const createError = ref<string | null>(null)
const form = reactive({
  type: 'receipt' as DocType,
  warehouse_id: '',
  target_warehouse_id: '',
  comment: '',
  items: [] as Array<{ product_id: string; quantity: number; price: number }>,
})

const detailDoc = ref<Document | null>(null)
const route = useRoute()

const tabs: Array<{ key: TabKey; label: string }> = [
  { key: 'all',       label: 'Все' },
  { key: 'receipt',   label: 'Приёмки' },
  { key: 'shipment',  label: 'Отгрузки' },
  { key: 'transfer',  label: 'Перемещения' },
  { key: 'inventory', label: 'Инвентаризации' },
  { key: 'draft',     label: 'Черновики' },
  { key: 'posted',    label: 'Проведённые' },
  { key: 'cancelled', label: 'Отменённые' },
]

function typeLabel(t: string) {
  return { receipt: 'Приёмка', shipment: 'Отгрузка', transfer: 'Перемещение', inventory: 'Инвентаризация' }[t] ?? t
}
function statusLabel(s: string) {
  return { draft: 'Черновик', posted: 'Проведён', cancelled: 'Отменён' }[s] ?? s
}
function warehouseName(id: string) {
  return warehouses.value.find((w) => w.id === id)?.name ?? id.slice(0, 8)
}
function productName(id: string) {
  return products.value.find((p) => p.id === id)?.name ?? id.slice(0, 8)
}
function formatDate(s: string) {
  const d = new Date(s)
  return d.toLocaleDateString('ru-RU') + ' ' + d.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })
}
function formatMoney(n: number) {
  return n.toLocaleString('ru-RU', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}
function formatQty(n: number) {
  return n.toLocaleString('ru-RU', { maximumFractionDigits: 3 })
}

function tabCount(key: TabKey): number {
  if (key === 'all') return items.value.length
  if (key === 'receipt' || key === 'shipment' || key === 'transfer' || key === 'inventory') {
    return items.value.filter((d) => d.type === key).length
  }
  return items.value.filter((d) => d.status === key).length
}

function setTab(key: TabKey) {
  activeTab.value = key
  page.value = 1
}

function sortBy(k: typeof sortKey.value) {
  if (sortKey.value === k) sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  else { sortKey.value = k; sortDir.value = 'desc' }
  page.value = 1
}

const filtered = computed(() => {
  let rows = items.value

  // Вкладка
  if (activeTab.value !== 'all') {
    if (['receipt', 'shipment', 'transfer', 'inventory'].includes(activeTab.value)) {
      rows = rows.filter((d) => d.type === activeTab.value)
    } else {
      rows = rows.filter((d) => d.status === activeTab.value)
    }
  }

  // Период
  if (filters.dateFrom) {
    const t = new Date(filters.dateFrom).getTime()
    rows = rows.filter((d) => new Date(d.created_at).getTime() >= t)
  }
  if (filters.dateTo) {
    const t = new Date(filters.dateTo).getTime() + 24 * 3600 * 1000 - 1
    rows = rows.filter((d) => new Date(d.created_at).getTime() <= t)
  }

  if (filters.warehouseId) rows = rows.filter((d) => d.warehouse_id === filters.warehouseId)
  if (filters.status) rows = rows.filter((d) => d.status === filters.status)

  if (filters.search) {
    const q = filters.search.toLowerCase()
    rows = rows.filter((d) =>
      (d.number ?? '').toLowerCase().includes(q) ||
      (d.comment ?? '').toLowerCase().includes(q)
    )
  }
  if (filters.sumFrom != null) rows = rows.filter((d) => (d.total ?? 0) >= filters.sumFrom!)
  if (filters.sumTo != null) rows = rows.filter((d) => (d.total ?? 0) <= filters.sumTo!)

  return [...rows].sort((a, b) => {
    const av = a[sortKey.value]
    const bv = b[sortKey.value]
    if (typeof av === 'number' && typeof bv === 'number') {
      return sortDir.value === 'asc' ? av - bv : bv - av
    }
    const as = String(av ?? '')
    const bs = String(bv ?? '')
    return sortDir.value === 'asc' ? as.localeCompare(bs) : bs.localeCompare(as)
  })
})

const paginated = computed(() => {
  const from = (page.value - 1) * perPage
  return filtered.value.slice(from, from + perPage)
})

const rangeFrom = computed(() => filtered.value.length === 0 ? 0 : (page.value - 1) * perPage + 1)
const rangeTo = computed(() => Math.min(page.value * perPage, filtered.value.length))

const filteredTotal = computed(() =>
  filtered.value.reduce((sum, d) => sum + (d.total ?? 0), 0)
)

function clearFilters() {
  filters.dateFrom = ''
  filters.dateTo = ''
  filters.warehouseId = ''
  filters.status = ''
  filters.search = ''
  filters.sumFrom = null
  filters.sumTo = null
  page.value = 1
}

async function load() {
  loading.value = true
  error.value = null
  try {
    items.value = await listDocuments()
  } catch (e) {
    error.value = apiErrorMessage(e)
  } finally {
    loading.value = false
  }
}

function openCreate() {
  form.type = 'receipt'
  form.warehouse_id = ''
  form.target_warehouse_id = ''
  form.comment = ''
  form.items = [{ product_id: '', quantity: 1, price: 0 }]
  createError.value = null
  showCreate.value = true
}
function closeCreate() { showCreate.value = false }
function addItem() { form.items.push({ product_id: '', quantity: 1, price: 0 }) }
function removeItem(i: number) { form.items.splice(i, 1) }

async function onCreate() {
  createError.value = null
  saving.value = true
  try {
    if (form.items.length === 0) throw new Error('Добавьте хотя бы одну позицию')
    if (form.items.some((it) => !it.product_id || it.quantity <= 0)) throw new Error('Заполните все позиции')

    await createDocument({
      type: form.type,
      warehouse_id: form.warehouse_id,
      target_warehouse_id: form.type === 'transfer' ? form.target_warehouse_id : undefined,
      comment: form.comment || undefined,
      items: form.items,
    })
    closeCreate()
    await load()
  } catch (e) {
    createError.value = apiErrorMessage(e)
  } finally {
    saving.value = false
  }
}

async function onPost(d: Document) {
  if (!confirm(`Провести документ ${d.number}?`)) return
  try { await postDocument(d.id); await load() } catch (e) { error.value = apiErrorMessage(e) }
}

async function onCancel(d: Document) {
  if (!confirm(`Отменить документ ${d.number}?`)) return
  try { await cancelDocument(d.id); await load() } catch (e) { error.value = apiErrorMessage(e) }
}

const router = useRouter()
function openCard(d: Document) { router.push('/documents/' + d.id) }

function print() { window.print() }

onMounted(async () => {
  const qtype = route.query.type as string | undefined
  if (qtype && ['receipt','shipment','transfer','inventory'].includes(qtype)) {
    activeTab.value = qtype as TabKey
  }
  try {
    const [w, p] = await Promise.all([listWarehouses(), listProducts(true)])
    warehouses.value = w
    products.value = p
  } catch { /* ignore */ }
  await load()
})
</script>

<style scoped>
.page { font-size: 13px; }

.ms-title {
  display: flex; align-items: center; gap: 8px;
  font-size: 20px; font-weight: 600; color: #1f2328;
  margin-bottom: 12px;
}
.ms-help {
  width: 20px; height: 20px; border-radius: 50%;
  border: 1px solid #b8c0c8; background: transparent;
  color: #57606a; cursor: pointer;
  display: inline-flex; align-items: center; justify-content: center; padding: 0;
}
.ms-help:hover { background: #f0f2f5; }
.ms-refresh {
  width: 24px; height: 24px; padding: 0;
  display: inline-flex; align-items: center; justify-content: center;
  border: none; background: transparent; color: #57606a; cursor: pointer;
}
.ms-refresh:hover { color: #2c5d9c; }

.ms-tabs {
  display: flex; gap: 2px;
  border-bottom: 1px solid #d8dee4;
  margin-bottom: 12px; overflow-x: auto;
}
.tab {
  border: none; background: transparent;
  padding: 10px 16px; font-size: 13px; color: #57606a;
  cursor: pointer; border-bottom: 2px solid transparent;
  white-space: nowrap; display: inline-flex; align-items: center; gap: 6px;
}
.tab:hover { color: #2c5d9c; }
.tab.active { color: #2c5d9c; border-bottom-color: #2c5d9c; font-weight: 600; }
.tab-count {
  background: #eef1f5; color: #57606a;
  font-size: 11px; padding: 1px 6px; border-radius: 8px; font-weight: 500;
}
.tab.active .tab-count { background: #d8e4f0; color: #2c5d9c; }

.ms-toolbar {
  display: flex; align-items: center; gap: 6px;
  margin-bottom: 12px; flex-wrap: wrap;
}
.ms-input {
  flex: 1; max-width: 320px; height: 30px; padding: 0 10px;
  font-size: 13px; border: 1px solid #d0d7de; border-radius: 4px;
  background: #fff; color: #1f2328;
}
.ms-input::placeholder { color: #8c959f; }
.ms-input:focus { outline: 2px solid rgba(44,93,156,0.3); border-color: #2c5d9c; }
.ms-counter {
  min-width: 40px; height: 30px; padding: 0 10px;
  display: inline-flex; align-items: center; justify-content: center;
  border: 1px solid #d0d7de; border-radius: 4px;
  background: #fff; color: #8c959f;
  font-variant-numeric: tabular-nums; font-size: 13px;
}

.ms-filter {
  background: #eef1f5; border: 1px solid #d8dee4;
  border-radius: 4px; padding: 10px 14px 12px; margin-bottom: 12px;
}
.filter-actions { display: flex; align-items: center; gap: 6px; margin-bottom: 10px; }
.filter-grid {
  display: grid; grid-template-columns: repeat(6, minmax(0, 1fr));
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
.filter-field input, .filter-field select {
  height: 28px; padding: 0 8px; font-size: 12px;
  border: 1px solid #d0d7de; border-radius: 3px;
  background: #fff; color: #1f2328;
}

.ms-table2 { width: 100%; border-collapse: collapse; background: #fff; font-size: 13px; }
.ms-table2 thead th {
  background: #fff; color: #2c5d9c; font-weight: 500;
  padding: 8px 10px; text-align: left;
  border-bottom: 1px solid #d8dee4; white-space: nowrap;
  cursor: pointer; user-select: none; font-size: 12px;
}
.ms-table2 thead th:hover { background: #f6f8fa; }
.ms-table2 tbody td {
  padding: 7px 10px;
  border-bottom: 1px solid #eaeef2;
  vertical-align: middle;
}
.ms-table2 tbody tr.row { cursor: pointer; }
.ms-table2 tbody tr.row:hover { background: #f6f8fa; }
.ms-table2 .col-num { width: 90px; }
.ms-table2 .col-date { width: 140px; color: #57606a; }
.ms-table2 .col-type { width: 130px; }
.ms-table2 .col-num-right { text-align: right; font-variant-numeric: tabular-nums; }
.ms-table2 .col-status { width: 130px; }
.ms-table2 .col-actions { width: 150px; text-align: right; white-space: nowrap; }
.ms-table2 .link { color: #2c5d9c; font-weight: 500; }
.ms-table2 .link:hover { text-decoration: underline; }
.ms-table2 .comment { color: #8c959f; max-width: 300px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.ms-table2 .sort { font-size: 9px; margin-left: 3px; }
.ms-table2 .empty { text-align: center; padding: 24px; color: #8c959f; }

.badge {
  display: inline-block; padding: 2px 10px; border-radius: 3px;
  font-size: 11px; font-weight: 600; color: #fff;
}
.bg-draft     { background: #8c959f; }
.bg-posted    { background: #2196f3; }
.bg-cancelled { background: #cf222e; }

.btn-link-ms {
  border: none; background: transparent;
  color: #2c5d9c; cursor: pointer;
  font-size: 12px; padding: 4px 8px; border-radius: 3px;
}
.btn-link-ms:hover { background: #f0f2f5; }
.btn-link-ms.danger { color: #cf222e; }
.btn-link-ms.danger:hover { background: #ffecec; }
.row-menu {
  width: 22px; height: 22px; padding: 0;
  border: none; background: transparent;
  color: #57606a; cursor: pointer; border-radius: 3px;
  vertical-align: middle;
}
.row-menu:hover { background: #eaeef2; color: #1f2328; }

.ms-footer2 {
  display: flex; align-items: center; justify-content: space-between;
  padding: 6px 10px; font-size: 12px; color: #57606a;
  background: #fff; border-top: 1px solid #eaeef2;
}
.pager { display: flex; align-items: center; gap: 4px; }
.pager button {
  width: 22px; height: 22px; padding: 0;
  border: 1px solid #d0d7de; background: #fff; border-radius: 3px;
  cursor: pointer; font-size: 12px; color: #1f2328;
  display: inline-flex; align-items: center; justify-content: center;
}
.pager button:disabled { opacity: 0.4; cursor: default; }
.pager button:hover:not(:disabled) { background: #f6f8fa; }
.pager .range { margin: 0 6px; font-variant-numeric: tabular-nums; }
.totals-inline { display: flex; gap: 16px; align-items: center; font-variant-numeric: tabular-nums; }
.totals-inline b { color: #1f2328; }

/* Модалки */
.modal-backdrop {
  position: fixed; inset: 0; background: rgba(0,0,0,0.4);
  display: flex; align-items: center; justify-content: center;
  padding: 20px; z-index: 100;
}
.modal {
  background: #fff; border-radius: 4px;
  width: 480px; max-width: 100%; padding: 20px;
  display: flex; flex-direction: column; gap: 10px;
  max-height: 90vh; overflow-y: auto;
}
.modal.big { width: 720px; }
.modal h2 { margin: 0 0 4px; font-size: 18px; }
.modal label { font-size: 13px; color: #444; display: flex; flex-direction: column; gap: 4px; }
.modal-actions { display: flex; justify-content: flex-end; gap: 10px; margin-top: 6px; }
.items-section { display: flex; flex-direction: column; gap: 8px; margin-top: 6px; }
.items-head { display: flex; justify-content: space-between; align-items: center; font-size: 13px; font-weight: 500; }
.item-row { display: grid; grid-template-columns: 1fr 90px 90px 30px; gap: 6px; align-items: center; }
.item-row select, .item-row input {
  height: 28px; padding: 0 8px; font-size: 13px;
  border: 1px solid #d0d7de; border-radius: 3px; background: #fff;
}
.doc-info {
  display: grid; grid-template-columns: 1fr 1fr;
  gap: 8px 24px; font-size: 13px; margin-bottom: 12px;
}
.doc-info .lbl { color: #57606a; }
</style>
