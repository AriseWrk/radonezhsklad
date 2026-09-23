<template>
  <div>
    <div class="ms-title">
      <button class="ms-help" title="Справка"><MsIcon name="help" :size="14" /></button>
      <span>Приёмки</span>
      <button class="ms-refresh" @click="load" title="Обновить"><MsIcon name="refresh" :size="14" /></button>
    </div>

    <div class="ms-toolbar">
      <MsButton variant="primary" icon="plus" @click="openCreate">Приёмка</MsButton>
      <MsButton icon="filter" @click="showFilter = !showFilter">Фильтр</MsButton>
      <input v-model="search" class="ms-input" placeholder="Номер или комментарий" />
      <div class="ms-counter">{{ filtered.length }}</div>
      <select class="ms-select" v-model="filterStatus">
        <option value="">Статус: все</option>
        <option value="draft">Черновик</option>
        <option value="posted">Проведён</option>
        <option value="cancelled">Отменён</option>
      </select>
      <MsButton icon="print">Печать</MsButton>
      <MsButton variant="icon" icon="gear" title="Настройки" />
    </div>

    <div v-if="showFilter" class="ms-filter">
      <div class="filter-actions">
        <MsButton variant="green" @click="page = 1">Найти</MsButton>
        <MsButton @click="clearFilters">Очистить</MsButton>
      </div>
      <div class="filter-grid">
        <div class="filter-field">
          <label class="filter-label"><span class="dot"></span>Склад</label>
          <select v-model="filterWarehouse">
            <option value="">—</option>
            <option v-for="w in warehouses" :key="w.id" :value="w.id">{{ w.name }}</option>
          </select>
        </div>
        <div class="filter-field">
          <label class="filter-label"><span class="dot"></span>Поставщик</label>
          <select v-model="filterSupplier">
            <option value="">—</option>
            <option v-for="s in suppliers" :key="s.id" :value="s.id">{{ s.name }}</option>
          </select>
        </div>
        <div class="filter-field">
          <label class="filter-label"><span class="dot"></span>Период с</label>
          <input v-model="filterDateFrom" type="date" />
        </div>
        <div class="filter-field">
          <label class="filter-label"><span class="dot"></span>Период по</label>
          <input v-model="filterDateTo" type="date" />
        </div>
      </div>
    </div>

    <div v-if="error" class="error-box">{{ error }}</div>

    <table class="ms-table">
      <thead>
        <tr>
          <th class="chk-col"><input type="checkbox" :checked="allChecked" @change="toggleAll" /></th>
          <th style="width:80px">№</th>
          <th style="width:130px">Время</th>
          <th>На склад</th>
          <th>Контрагент</th>
          <th>Организация</th>
          <th class="num">Сумма</th>
          <th class="num">Оплачено</th>
          <th style="width:110px">Входящая дата</th>
          <th>Входящий номер</th>
          <th>Отправлено</th>
          <th>Напечатано</th>
          <th>Комментарий</th>
          <th class="actions-col"></th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="loading">
          <td colspan="14" class="muted" style="text-align:center;padding:24px">Загрузка...</td>
        </tr>
        <tr v-else-if="filtered.length === 0">
          <td colspan="14" class="muted" style="text-align:center;padding:24px">Нет приёмок</td>
        </tr>
        <tr v-else v-for="d in paginated" :key="d.id" class="clickable" @click="openEdit(d)">
          <td class="chk-col" @click.stop>
            <input type="checkbox" :checked="selected.has(d.id)" @change="toggleSelect(d.id)" />
          </td>
          <td class="mono link">{{ shortNumber(d) }}</td>
          <td class="muted">{{ formatDate(d.created_at) }}</td>
          <td>{{ warehouseName(d.warehouse_id) }}</td>
          <td>{{ supplierName(d.supplier_id) }}</td>
          <td>{{ organizationName(d.organization_id) }}</td>
          <td class="num"><strong>{{ formatMoney(d.total ?? 0) }}</strong></td>
          <td class="num">{{ formatMoney(d.paid_amount ?? 0) }}</td>
          <td class="muted">{{ d.incoming_date ? formatShortDate(d.incoming_date) : '—' }}</td>
          <td class="muted">{{ d.incoming_number || '—' }}</td>
          <td class="muted">{{ d.sent_at ? 'да' : '—' }}</td>
          <td class="muted">{{ d.printed_at ? 'да' : '—' }}</td>
          <td class="muted">{{ d.comment || '—' }}</td>
          <td class="actions-col" @click.stop>
            <button v-if="d.status === 'draft'" class="btn-link" @click="onPost(d)">Провести</button>
            <button v-if="d.status === 'posted'" class="btn-link danger-text" @click="onCancel(d)">Отменить</button>
            <span :class="['pill', d.status]" style="margin-left:6px">{{ statusLabel(d.status) }}</span>
          </td>
        </tr>
      </tbody>
    </table>

    <div class="ms-footer">
      <div class="ms-pager">
        <button :disabled="page === 1" @click="page--">◀</button>
        <span>{{ rangeFrom }}–{{ rangeTo }} из {{ filtered.length }}</span>
        <button :disabled="rangeTo >= filtered.length" @click="page++">▶</button>
      </div>
      <div class="ms-totals">
        <span>Сумма: {{ formatMoney(totalSum) }}</span>
        <span>Оплачено: {{ formatMoney(totalPaid) }}</span>
      </div>
    </div>

    <!-- Модалка создания -->
    <div v-if="showCreate" class="modal-backdrop" @click.self="closeCreate">
      <form class="card modal big" @submit.prevent="onCreate">
        <h2>Новая приёмка</h2>
        <div v-if="createError" class="error-box">{{ createError }}</div>

        <div class="grid2">
          <label>Склад*
            <select v-model="form.warehouse_id" required>
              <option value="">— выберите —</option>
              <option v-for="w in warehouses" :key="w.id" :value="w.id">{{ w.name }}</option>
            </select>
          </label>
          <label>Поставщик*
            <select v-model="form.supplier_id" required>
              <option value="">— выберите —</option>
              <option v-for="s in suppliers" :key="s.id" :value="s.id">{{ s.name }}</option>
            </select>
          </label>
        </div>

        <div class="grid2">
          <label>Организация
            <select v-model="form.organization_id">
              <option value="">— по умолчанию —</option>
              <option v-for="o in organizations" :key="o.id" :value="o.id">{{ o.name }}</option>
            </select>
          </label>
          <label>Входящий номер
            <input v-model="form.incoming_number" placeholder="Номер от поставщика" />
          </label>
        </div>

        <div class="grid2">
          <label>Входящая дата
            <input v-model="form.incoming_date" type="date" />
          </label>
          <label>Комментарий
            <input v-model="form.comment" />
          </label>
        </div>

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
    <div v-if="detail" class="modal-backdrop" @click.self="detail = null">
      <div class="card modal big">
        <h2>Приёмка {{ detail.number }}</h2>
        <div class="doc-info">
          <div><span class="lbl">Склад:</span> {{ warehouseName(detail.warehouse_id) }}</div>
          <div><span class="lbl">Поставщик:</span> {{ supplierName(detail.supplier_id) }}</div>
          <div><span class="lbl">Организация:</span> {{ organizationName(detail.organization_id) }}</div>
          <div><span class="lbl">Статус:</span> {{ statusLabel(detail.status) }}</div>
          <div><span class="lbl">Входящий №:</span> {{ detail.incoming_number || '—' }}</div>
          <div><span class="lbl">Входящая дата:</span> {{ detail.incoming_date ? formatShortDate(detail.incoming_date) : '—' }}</div>
          <div v-if="detail.comment" style="grid-column: 1/-1">
            <span class="lbl">Комментарий:</span> {{ detail.comment }}
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
            <tr v-for="it in detail.items" :key="it.id">
              <td>{{ productName(it.product_id) }}</td>
              <td class="num">{{ formatQty(it.quantity) }}</td>
              <td class="num">{{ formatMoney(it.price) }}</td>
              <td class="num">{{ formatMoney(it.quantity * it.price) }}</td>
            </tr>
          </tbody>
        </table>
        <div class="modal-actions">
          <button v-if="detail.status === 'draft'" class="primary" @click="onPost(detail)">Провести</button>
          <button v-if="detail.status === 'posted'" class="danger" @click="onCancel(detail)">Отменить</button>
          <div style="flex:1"></div>
          <button @click="detail = null">Закрыть</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import {
  listDocuments, createDocument, postDocument, cancelDocument, getDocument, type Document,
} from '../api/documents'
import { listWarehouses, type Warehouse } from '../api/warehouses'
import { listProducts, type Product } from '../api/products'
import { listSuppliers, listOrganizations, type Supplier, type Organization } from '../api/suppliers'
import { apiErrorMessage } from '../api/client'
import MsButton from '../components/MsButton.vue'
import MsIcon from '../components/MsIcon.vue'

const items = ref<Document[]>([])
const warehouses = ref<Warehouse[]>([])
const products = ref<Product[]>([])
const suppliers = ref<Supplier[]>([])
const organizations = ref<Organization[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const showFilter = ref(false)
const page = ref(1)
const perPage = 50

const search = ref('')
const filterStatus = ref('')
const filterWarehouse = ref('')
const filterSupplier = ref('')
const filterDateFrom = ref('')
const filterDateTo = ref('')

const selected = ref<Set<string>>(new Set())

const showCreate = ref(false)
const saving = ref(false)
const createError = ref<string | null>(null)
const form = reactive({
  warehouse_id: '',
  supplier_id: '',
  organization_id: '',
  incoming_number: '',
  incoming_date: '',
  comment: '',
  items: [] as Array<{ product_id: string; quantity: number; price: number }>,
})

const detail = ref<Document | null>(null)

function warehouseName(id?: string) {
  if (!id) return '—'
  return warehouses.value.find((w) => w.id === id)?.name ?? id.slice(0, 8)
}
function supplierName(id?: string) {
  if (!id) return '—'
  return suppliers.value.find((s) => s.id === id)?.name ?? '—'
}
function organizationName(id?: string) {
  if (!id) return organizations.value.find((o) => o.is_default)?.name ?? '—'
  return organizations.value.find((o) => o.id === id)?.name ?? '—'
}
function productName(id: string) {
  return products.value.find((p) => p.id === id)?.name ?? id.slice(0, 8)
}
function statusLabel(s: string) {
  return { draft: 'Черновик', posted: 'Проведён', cancelled: 'Отменён' }[s] ?? s
}
function shortNumber(d: Document) {
  // прим-1234567 → 1234567, а если UUID-формат — оставим последние 6
  const parts = d.number.split('-')
  return parts[parts.length - 1] || d.number
}
function formatDate(s: string) {
  const d = new Date(s)
  return d.toLocaleDateString('ru-RU') + ' ' + d.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })
}
function formatShortDate(s: string) {
  return new Date(s).toLocaleDateString('ru-RU')
}
function formatMoney(n: number) {
  return n.toLocaleString('ru-RU', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}
function formatQty(n: number) {
  return n.toLocaleString('ru-RU', { maximumFractionDigits: 3 })
}

const filtered = computed(() => {
  let rows = items.value
  if (search.value) {
    const q = search.value.toLowerCase()
    rows = rows.filter((d) =>
      (d.number ?? '').toLowerCase().includes(q) ||
      (d.comment ?? '').toLowerCase().includes(q)
    )
  }
  if (filterStatus.value) rows = rows.filter((d) => d.status === filterStatus.value)
  if (filterWarehouse.value) rows = rows.filter((d) => d.warehouse_id === filterWarehouse.value)
  if (filterSupplier.value) rows = rows.filter((d) => d.supplier_id === filterSupplier.value)
  if (filterDateFrom.value) {
    const t = new Date(filterDateFrom.value).getTime()
    rows = rows.filter((d) => new Date(d.created_at).getTime() >= t)
  }
  if (filterDateTo.value) {
    const t = new Date(filterDateTo.value).getTime() + 86400000 - 1
    rows = rows.filter((d) => new Date(d.created_at).getTime() <= t)
  }
  return rows
})

const paginated = computed(() => {
  const from = (page.value - 1) * perPage
  return filtered.value.slice(from, from + perPage)
})

const rangeFrom = computed(() => filtered.value.length === 0 ? 0 : (page.value - 1) * perPage + 1)
const rangeTo = computed(() => Math.min(page.value * perPage, filtered.value.length))

const totalSum = computed(() => filtered.value.reduce((s, d) => s + (d.total ?? 0), 0))
const totalPaid = computed(() => filtered.value.reduce((s, d) => s + (d.paid_amount ?? 0), 0))

const allChecked = computed(() =>
  paginated.value.length > 0 && paginated.value.every((d) => selected.value.has(d.id))
)
function toggleAll(e: Event) {
  const on = (e.target as HTMLInputElement).checked
  selected.value = on ? new Set(paginated.value.map((d) => d.id)) : new Set()
}
function toggleSelect(id: string) {
  const s = new Set(selected.value)
  s.has(id) ? s.delete(id) : s.add(id)
  selected.value = s
}

function clearFilters() {
  search.value = ''
  filterStatus.value = ''
  filterWarehouse.value = ''
  filterSupplier.value = ''
  filterDateFrom.value = ''
  filterDateTo.value = ''
  page.value = 1
}

async function load() {
  loading.value = true
  error.value = null
  try {
    items.value = await listDocuments({ type: 'receipt' })
  } catch (e) {
    error.value = apiErrorMessage(e)
  } finally {
    loading.value = false
  }
}

function openCreate() {
  form.warehouse_id = warehouses.value[0]?.id ?? ''
  form.supplier_id = ''
  form.organization_id = organizations.value.find((o) => o.is_default)?.id ?? ''
  form.incoming_number = ''
  form.incoming_date = ''
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
      type: 'receipt',
      warehouse_id: form.warehouse_id,
      supplier_id: form.supplier_id || undefined,
      organization_id: form.organization_id || undefined,
      incoming_number: form.incoming_number || undefined,
      incoming_date: form.incoming_date || undefined,
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
  if (!confirm(`Провести приёмку ${d.number}?`)) return
  try {
    await postDocument(d.id)
    await load()
    if (detail.value && detail.value.id === d.id) detail.value = null
  } catch (e) { error.value = apiErrorMessage(e) }
}

async function onCancel(d: Document) {
  if (!confirm(`Отменить приёмку ${d.number}?`)) return
  try {
    await cancelDocument(d.id)
    await load()
    if (detail.value && detail.value.id === d.id) detail.value = null
  } catch (e) { error.value = apiErrorMessage(e) }
}

async function openEdit(d: Document) {
  try { detail.value = await getDocument(d.id) } catch (e) { error.value = apiErrorMessage(e) }
}

onMounted(async () => {
  try {
    const [w, p, s, o] = await Promise.all([
      listWarehouses(), listProducts(true), listSuppliers(), listOrganizations(),
    ])
    warehouses.value = w
    products.value = p
    suppliers.value = s
    organizations.value = o
  } catch { /* ignore */ }
  await load()
})
</script>

<style scoped>
.search-input {
  padding: 6px 10px;
  font-size: 13px;
  border: 1px solid #d0d7de;
  border-radius: 4px;
  background: #fff;
  color: #1f2328;
  min-width: 200px;
}
.users-table .chk-col { width: 32px; }
tr.clickable { cursor: pointer; }

/* === МойСклад-стиль === */
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
.ms-counter {
  min-width: 40px; height: 30px; padding: 0 10px;
  display: inline-flex; align-items: center; justify-content: center;
  border: 1px solid #d0d7de; border-radius: 4px;
  background: #fff; color: #8c959f;
  font-variant-numeric: tabular-nums; font-size: 13px;
}
.ms-select {
  height: 30px; padding: 0 8px; font-size: 13px;
  border: 1px solid #d0d7de; border-radius: 4px;
  background: #fff; color: #1f2328; max-width: 160px;
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
</style>
