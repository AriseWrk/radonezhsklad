<template>
  <div>
    <!-- Заголовок -->
    <div class="page-title-bar">
      <div class="page-title">
        <span>Документы</span>
        <span class="refresh" @click="load" title="Обновить">↻</span>
      </div>
      <div class="page-actions">
        <button class="btn primary" @click="openCreate">+ Документ</button>
        <button class="btn" @click="showFilter = !showFilter">Фильтр</button>
        <button class="btn" @click="print">Печать</button>
      </div>
    </div>

    <!-- Вкладки по типу / статусу -->
    <div class="doc-tabs">
      <button
        v-for="t in tabs"
        :key="t.key"
        class="doc-tab"
        :class="{ active: activeTab === t.key }"
        @click="setTab(t.key)"
      >
        {{ t.label }}
        <span v-if="tabCount(t.key) > 0" class="tab-count">{{ tabCount(t.key) }}</span>
      </button>
    </div>

    <!-- Панель фильтров -->
    <div v-if="showFilter" class="filter-panel">
      <div class="filter-row">
        <div class="filter-actions">
          <button class="btn-find" @click="page = 1">Найти</button>
          <button class="btn-clear" @click="clearFilters">Очистить</button>
        </div>
        <div class="filter-field">
          <label>Период с</label>
          <input v-model="filters.dateFrom" type="date" />
        </div>
        <div class="filter-field">
          <label>Период по</label>
          <input v-model="filters.dateTo" type="date" />
        </div>
        <div class="filter-field">
          <label>Склад</label>
          <select v-model="filters.warehouseId" @change="page = 1">
            <option value="">Все</option>
            <option v-for="w in warehouses" :key="w.id" :value="w.id">{{ w.name }}</option>
          </select>
        </div>
        <div class="filter-field">
          <label>Статус</label>
          <select v-model="filters.status" @change="page = 1">
            <option value="">Все</option>
            <option value="draft">Черновик</option>
            <option value="posted">Проведён</option>
            <option value="cancelled">Отменён</option>
          </select>
        </div>
      </div>
      <div class="filter-row wide">
        <div class="filter-field">
          <label>Поиск</label>
          <input v-model="filters.search" placeholder="Номер или комментарий" />
        </div>
        <div class="filter-field">
          <label>Сумма от</label>
          <input v-model.number="filters.sumFrom" type="number" min="0" />
        </div>
        <div class="filter-field">
          <label>Сумма до</label>
          <input v-model.number="filters.sumTo" type="number" min="0" />
        </div>
      </div>
    </div>

    <div v-if="error" class="error-box">{{ error }}</div>

    <!-- Таблица -->
    <table class="ms-table">
      <thead>
        <tr>
          <th class="num" @click="sortBy('number')">№ <span v-if="sortKey === 'number'">{{ sortDir === 'asc' ? '↑' : '↓' }}</span></th>
          <th @click="sortBy('created_at')">Дата <span v-if="sortKey === 'created_at'">{{ sortDir === 'asc' ? '↑' : '↓' }}</span></th>
          <th>Тип</th>
          <th>Склад</th>
          <th>Комментарий</th>
          <th class="num">Позиций</th>
          <th class="num" @click="sortBy('total')">Сумма <span v-if="sortKey === 'total'">{{ sortDir === 'asc' ? '↑' : '↓' }}</span></th>
          <th>Статус</th>
          <th class="actions-col"></th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="loading">
          <td colspan="9" class="muted" style="text-align:center;padding:24px">Загрузка...</td>
        </tr>
        <tr v-else-if="filtered.length === 0">
          <td colspan="9" class="muted" style="text-align:center;padding:24px">Нет документов</td>
        </tr>
        <tr v-else v-for="d in paginated" :key="d.id">
          <td class="num mono">{{ d.number }}</td>
          <td>{{ formatDate(d.created_at) }}</td>
          <td>{{ typeLabel(d.type) }}</td>
          <td>{{ warehouseName(d.warehouse_id) }}</td>
          <td class="muted">{{ d.comment || '—' }}</td>
          <td class="num">{{ d.items_count ?? 0 }}</td>
          <td class="num">{{ formatMoney(d.total ?? 0) }}</td>
          <td><span :class="['pill', d.status]">{{ statusLabel(d.status) }}</span></td>
          <td class="actions-col">
            <button v-if="d.status === 'draft'" class="btn-link" @click="onPost(d)">Провести</button>
            <button v-if="d.status === 'posted'" class="btn-link danger-text" @click="onCancel(d)">Отменить</button>
            <button class="btn-link" @click="openCard(d)">Открыть</button>
          </td>
        </tr>
      </tbody>
    </table>

    <!-- Футер -->
    <div class="ms-footer">
      <div class="ms-pager">
        <button :disabled="page === 1" @click="page--">◀</button>
        <button :disabled="page === 1" @click="page--">↤</button>
        <span>{{ rangeFrom }}–{{ rangeTo }} из {{ filtered.length }}</span>
        <button :disabled="rangeTo >= filtered.length" @click="page++">↦</button>
        <button :disabled="rangeTo >= filtered.length" @click="page++">▶</button>
      </div>
      <div class="ms-totals">
        <span>{{ filtered.length }} док.</span>
        <span>{{ formatMoney(filteredTotal) }}</span>
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
        <table class="ms-table">
          <thead>
            <tr>
              <th>Товар</th>
              <th class="num">Кол-во</th>
              <th class="num">Цена</th>
              <th class="num">Сумма</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="it in detailDoc.items" :key="it.id">
              <td>{{ productName(it.product_id) }}</td>
              <td class="num">{{ formatQty(it.quantity) }}</td>
              <td class="num">{{ formatMoney(it.price) }}</td>
              <td class="num">{{ formatMoney(it.quantity * it.price) }}</td>
            </tr>
            <tr v-if="!detailDoc.items || detailDoc.items.length === 0">
              <td colspan="4" class="muted" style="text-align:center;padding:16px">Нет позиций</td>
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
/* всё в style.css глобально */
</style>