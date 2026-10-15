<script setup lang="ts">
import { smartFilter } from '../utils/smartSearch'
import { computed, onMounted, reactive, ref } from 'vue'
import {
  listStockExtended,
  productStockDetail,
  type StockExtendedResponse,
  type StockExtendedRow,
  type ProductStockDetail,
} from '../api/stock'
import { listWarehouses, type Warehouse } from '../api/warehouses'
import { listProducts, type Product } from '../api/products'
import { listSuppliers, type Supplier } from '../api/suppliers'
import { apiErrorMessage } from '../api/client'
import MsButton from '../components/MsButton.vue'
import MsIcon from '../components/MsIcon.vue'
import MsSmartSelect, { type SmartItem } from '../components/MsSmartSelect.vue'

const loading = ref(false)
const error = ref<string | null>(null)
const data = ref<StockExtendedResponse | null>(null)
const warehouses = ref<Warehouse[]>([])
const products = ref<Product[]>([])
const suppliers = ref<Supplier[]>([])
const view = ref<'products' | 'warehouses'>('products')
const showFilter = ref(true)
const page = ref(1)
const perPage = 100

const selectedProductId = ref<string | null>(null)
const detail = ref<ProductStockDetail | null>(null)
const detailLoading = ref(false)

const filters = reactive({
  qtyMode: 'any' as 'any' | 'positive' | 'zero',
  availMode: 'any' as 'any' | 'nonZero' | 'zero',
  onlyReserve: false,
  onlyIncoming: false,
  showMode: 'normal' as 'normal' | 'all' | 'archived',
  daysMin: '' as '' | number,
  daysMax: '' as '' | number,
  supplierId: '' as string,
  productId: '' as string,
  warehouseId: '' as string,
  sku: '',
  asOf: '' as string,
})

const sortKey = ref<keyof StockExtendedRow>('product_name')
const sortDir = ref<'asc' | 'desc'>('asc')

// ----- Умные селекты -----
const warehouseItems = computed<SmartItem[]>(() =>
  warehouses.value.map((w) => ({ id: w.id, name: w.name, sublabel: w.address || undefined }))
)
const productItems = computed<SmartItem[]>(() =>
  products.value.map((p) => ({ id: p.id, name: p.name, sublabel: p.sku || undefined }))
)
const supplierItems = computed<SmartItem[]>(() =>
  suppliers.value.map((s) => ({ id: s.id, name: s.name, sublabel: s.inn || undefined }))
)

function formatQty(n: number): string {
  if (n === 0) return '0'
  return n.toLocaleString('ru-RU', { maximumFractionDigits: 3 })
}
function formatMoney(n: number): string {
  const sign = n < 0 ? '−' : ''
  return sign + Math.abs(n).toLocaleString('ru-RU', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}
function formatDateTime(s: string): string {
  const d = new Date(s)
  return d.toLocaleDateString('ru-RU') + ' ' + d.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })
}
function typeLabel(t: string): string {
  return ({
    receipt: 'Приёмка', shipment: 'Отгрузка', transfer: 'Перемещение', inventory: 'Инвентаризация',
  } as Record<string, string>)[t] ?? (t || '—')
}
function warehouseName(id?: string): string {
  if (!id) return '—'
  return warehouses.value.find((w) => w.id === id)?.name ?? id.slice(0, 6)
}

interface ProductRow {
  product_id: string
  product_name: string
  sku?: string
  quantity: number
  min_stock: number
  reserve: number
  incoming: number
  available: number
  unit_short: string
  days_on_stock?: number
  cost_price: number
  cost_total: number
  sale_price: number
  sale_total: number
  warehouses_count: number
}
// ----- Группировка по складам -----
interface WarehouseGroup {
  warehouse_id: string
  warehouse_name: string
  rows: StockExtendedRow[]
}
const warehouseGroups = computed<WarehouseGroup[]>(() => {
  const map = new Map<string, WarehouseGroup>()
  for (const r of data.value?.items ?? []) {
    const id = r.warehouse_id ?? '__none__'
    let g = map.get(id)
    if (!g) {
      g = { warehouse_id: id, warehouse_name: warehouseName(r.warehouse_id), rows: [] }
      map.set(id, g)
    }
    g.rows.push(r)
  }
  return [...map.values()]
})

function sortBy(key: keyof StockExtendedRow) {
  if (sortKey.value === key) sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  else { sortKey.value = key; sortDir.value = 'asc' }
  page.value = 1
}

function applyCommon(rows: StockExtendedRow[]): StockExtendedRow[] {
  let r = rows
  if (filters.sku) {
    r = smartFilter(r, filters.sku, {
      getFields: (x) => [null, x.sku ?? null, null, null],
      limit: r.length,
    })
  }
  switch (filters.qtyMode) {
    case 'positive': r = r.filter((x) => x.quantity > 0); break
    case 'zero':     r = r.filter((x) => x.quantity === 0); break
  }
  switch (filters.availMode) {
    case 'nonZero': r = r.filter((x) => x.available !== 0); break
    case 'zero':    r = r.filter((x) => x.available === 0); break
  }
  if (filters.onlyReserve)  r = r.filter((x) => x.reserve !== 0)
  if (filters.onlyIncoming) r = r.filter((x) => x.incoming !== 0)
  if (filters.daysMin !== '') r = r.filter((x) => (x.days_on_stock ?? 0) >= Number(filters.daysMin))
  if (filters.daysMax !== '') r = r.filter((x) => (x.days_on_stock ?? 0) <= Number(filters.daysMax))
  if (filters.productId) {
    r = r.filter((x) => x.product_id === filters.productId)
  }
  return r
}

const filteredProductRows = computed<ProductRow[]>(() => {
  const base = applyCommon(data.value?.items ?? [])
  const map = new Map<string, ProductRow>()
  for (const r of base) {
    let row = map.get(r.product_id)
    if (!row) {
      row = {
        product_id: r.product_id, product_name: r.product_name, sku: r.sku,
        quantity: 0, min_stock: r.min_stock, reserve: 0, incoming: 0, available: 0,
        unit_short: r.unit_short, days_on_stock: undefined,
        cost_price: r.cost_price, cost_total: 0, sale_price: r.sale_price, sale_total: 0,
        warehouses_count: 0,
      }
      map.set(r.product_id, row)
    }
    if (r.warehouse_id) row.warehouses_count++
    row.quantity += r.quantity
    row.reserve += r.reserve
    row.incoming += r.incoming
    row.available += r.available
    row.cost_total += r.cost_total
    row.sale_total += r.sale_total
    if (r.days_on_stock != null && (row.days_on_stock == null || r.days_on_stock < row.days_on_stock)) {
      row.days_on_stock = r.days_on_stock
    }
  }
  const list = [...map.values()]
  return list.sort((a, b) => {
    const av = (a as any)[sortKey.value]
    const bv = (b as any)[sortKey.value]
    if (typeof av === 'number' && typeof bv === 'number') {
      return sortDir.value === 'asc' ? av - bv : bv - av
    }
    const as = String(av ?? '').toLowerCase()
    const bs = String(bv ?? '').toLowerCase()
    return sortDir.value === 'asc' ? as.localeCompare(bs, 'ru') : bs.localeCompare(as, 'ru')
  })
})

const filteredWarehouseRows = computed<StockExtendedRow[]>(() =>
  applyCommon(data.value?.items ?? [])
)

const totals = computed(() => {
  const t = { quantity: 0, min_stock: 0, reserve: 0, incoming: 0, available: 0, cost_total: 0, sale_total: 0 }
  const src: any[] = view.value === 'products'
    ? filteredProductRows.value
    : filteredWarehouseRows.value
  for (const r of src) {
    t.quantity += r.quantity
    t.min_stock += r.min_stock
    t.reserve += r.reserve
    t.incoming += r.incoming
    t.available += r.available
    t.cost_total += r.cost_total
    t.sale_total += r.sale_total
  }
  return t
})

const totalCount = computed(() =>
  view.value === 'products' ? filteredProductRows.value.length : filteredWarehouseRows.value.length
)
const paginatedProductRows = computed(() => {
  const from = (page.value - 1) * perPage
  return filteredProductRows.value.slice(from, from + perPage)
})
const rangeFrom = computed(() => totalCount.value === 0 ? 0 : (page.value - 1) * perPage + 1)
const rangeTo = computed(() => Math.min(page.value * perPage, totalCount.value))

async function load() {
  loading.value = true
  error.value = null
  try {
    data.value = await listStockExtended({
      warehouse_id: filters.warehouseId || undefined,
      product_id: filters.productId || undefined,
      qty_mode: filters.qtyMode === 'any' ? '' : filters.qtyMode,
      avail_mode: filters.availMode === 'any' ? '' : filters.availMode,
      days_min: filters.daysMin === '' ? '' : Number(filters.daysMin),
      days_max: filters.daysMax === '' ? '' : Number(filters.daysMax),
      only_incoming: filters.onlyIncoming,
      supplier_id: filters.supplierId || undefined,
      as_of: filters.asOf || undefined,
      show_mode: filters.showMode,
    })
    page.value = 1
  } catch (e) {
    error.value = apiErrorMessage(e)
  } finally {
    loading.value = false
  }
}

async function openDetail(productId: string) {
  selectedProductId.value = productId
  detail.value = null
  detailLoading.value = true
  try {
    detail.value = await productStockDetail(productId)
  } catch (e) {
    error.value = apiErrorMessage(e)
  } finally {
    detailLoading.value = false
  }
}

function closeDetail() { selectedProductId.value = null; detail.value = null }
function toggleFilter() { showFilter.value = !showFilter.value }
function clearFilters() {
  filters.qtyMode = 'any'
  filters.availMode = 'any'
  filters.onlyReserve = false
  filters.onlyIncoming = false
  filters.showMode = 'normal'
  filters.daysMin = ''
  filters.daysMax = ''
  filters.supplierId = ''
  filters.productId = ''
  filters.warehouseId = ''
  filters.sku = ''
  filters.asOf = ''
  load()
}
function print() { window.print() }

onMounted(async () => {
  try { warehouses.value = await listWarehouses() } catch { /* ignore */ }
  try { products.value = await listProducts(true) } catch { /* ignore */ }
  try { suppliers.value = await listSuppliers() } catch { /* ignore */ }
  await load()
})
</script>

<template>
  <div class="stock-page">
    <div class="stock-main">
      <div class="ms-title">
        <button class="ms-help" title="Справка"><MsIcon name="help" :size="14" /></button>
        <span>Остатки</span>
        <button class="ms-refresh" @click="load" title="Обновить"><MsIcon name="refresh" :size="14" /></button>
      </div>

      <div class="ms-toolbar">
        <div class="view-switch">
          <button :class="{ active: view === 'products' }" @click="view = 'products'; page = 1">По товарам</button>
          <button :class="{ active: view === 'warehouses' }" @click="view = 'warehouses'; page = 1">По складам</button>
        </div>
        <MsButton icon="filter" @click="toggleFilter">Фильтр</MsButton>
        <MsButton icon="print" @click="print">Печать</MsButton>
        <MsButton variant="icon" icon="gear" title="Настройки" />
      </div>

      <!-- Фильтры в стиле МС -->
      <div v-if="showFilter" class="ms-filter">
        <div class="filter-actions">
          <MsButton variant="green" @click="load">Найти</MsButton>
          <MsButton @click="clearFilters">Очистить</MsButton>
        </div>

        <div class="filter-grid filter-grid--top">
          <div class="filter-field">
            <label class="filter-label"><span class="dot"></span>На дату</label>
            <input v-model="filters.asOf" type="datetime-local" />
          </div>
          <div class="filter-field">
            <label class="filter-label"><span class="dot"></span>Остаток</label>
            <select v-model="filters.qtyMode">
              <option value="any">Любой</option>
              <option value="positive">Положительный</option>
              <option value="zero">Нулевой</option>
            </select>
          </div>
          <div class="filter-field">
            <label class="filter-label"><span class="dot"></span>Доступно</label>
            <select v-model="filters.availMode">
              <option value="any">Любое</option>
              <option value="nonZero">Ненулевое</option>
              <option value="zero">Нулевое</option>
            </select>
          </div>
          <div class="filter-field">
            <label class="filter-label"><span class="dot"></span>Весовой товар</label>
            <select disabled><option>Нет</option></select>
          </div>
          <div class="filter-field">
            <label class="filter-label"><span class="dot"></span>Только с резервом</label>
            <select v-model="filters.onlyReserve">
              <option :value="false">Нет</option>
              <option :value="true">Да</option>
            </select>
          </div>
          <div class="filter-field">
            <label class="filter-label"><span class="dot"></span>Только с ожиданием</label>
            <select v-model="filters.onlyIncoming">
              <option :value="false">Нет</option>
              <option :value="true">Да</option>
            </select>
          </div>
        </div>

        <div class="filter-grid filter-grid--mid">
          <div class="filter-field">
            <label class="filter-label"><span class="dot"></span>Показывать</label>
            <select v-model="filters.showMode">
              <option value="normal">Только обычные</option>
              <option value="all">Все</option>
              <option value="archived">Только архивные</option>
            </select>
          </div>
          <div class="filter-field">
            <label class="filter-label"><span class="dot"></span>Товар или группа</label>
            <MsSmartSelect
              v-model="filters.productId"
              :items="productItems"
              placeholder="Название или артикул"
            />
          </div>
          <div class="filter-field">
            <label class="filter-label"><span class="dot"></span>Склад</label>
            <MsSmartSelect
              v-model="filters.warehouseId"
              :items="warehouseItems"
              placeholder="Склад"
              @change="load"
            />
          </div>
          <div class="filter-field">
            <label class="filter-label"><span class="dot"></span>Дней на складе</label>
            <div class="range-row">
              <input v-model="filters.daysMin" type="number" placeholder="от" />
              <span class="range-sep">—</span>
              <input v-model="filters.daysMax" type="number" placeholder="до" />
            </div>
          </div>
          <div class="filter-field">
            <label class="filter-label"><span class="dot"></span>Поставщик</label>
            <MsSmartSelect
              v-model="filters.supplierId"
              :items="supplierItems"
              placeholder="Поставщик"
            />
          </div>
          <div class="filter-field">
            <label class="filter-label"><span class="dot"></span>Артикул</label>
            <input v-model="filters.sku" placeholder="SKU" />
          </div>
        </div>
      </div>

      <div v-if="error" class="error-box">{{ error }}</div>

      <!-- Вид: По товарам -->
      <table v-if="view === 'products'" class="ms-table">
        <thead>
          <tr>
            <th @click="sortBy('product_name')">Наименование <span v-if="sortKey === 'product_name'">{{ sortDir === 'asc' ? '↑' : '↓' }}</span></th>
            <th style="width:70px">Код</th>
            <th style="width:100px">Артикул</th>
            <th class="num" style="width:90px" @click="sortBy('quantity')">Остаток</th>
            <th class="num" style="width:100px">Несниж.</th>
            <th class="num" style="width:80px">Резерв</th>
            <th class="num" style="width:90px">Ожидание</th>
            <th class="num" style="width:90px">Доступно</th>
            <th style="width:60px">Ед.</th>
            <th class="num" style="width:100px">Дней на скл.</th>
            <th class="num" style="width:100px">Себест.</th>
            <th class="num" style="width:110px">Сумма себест.</th>
            <th class="num" style="width:100px">Цена продажи</th>
            <th class="num" style="width:120px">Сумма продажи</th>
            <th class="num" style="width:90px">Складов</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="loading"><td colspan="15" class="muted" style="text-align:center;padding:24px">Загрузка...</td></tr>
          <tr v-else-if="filteredProductRows.length === 0"><td colspan="15" class="muted" style="text-align:center;padding:24px">Нет данных</td></tr>
          <tr
            v-else
            v-for="r in paginatedProductRows"
            :key="r.product_id"
            class="clickable"
            :class="{ active: selectedProductId === r.product_id }"
            @click="openDetail(r.product_id)"
          >
            <td class="link">{{ r.product_name }}</td>
            <td class="muted mono">{{ r.product_id.slice(0, 6) }}</td>
            <td class="muted mono">{{ r.sku || '—' }}</td>
            <td class="num">{{ formatQty(r.quantity) }}</td>
            <td class="num muted">{{ r.min_stock > 0 ? formatQty(r.min_stock) : '—' }}</td>
            <td class="num muted">{{ formatQty(r.reserve) }}</td>
            <td class="num">{{ formatQty(r.incoming) }}</td>
            <td class="num"><strong>{{ formatQty(r.available) }}</strong></td>
            <td>{{ r.unit_short || '—' }}</td>
            <td class="num muted">{{ r.days_on_stock ?? '—' }}</td>
            <td class="num">{{ formatMoney(r.cost_price) }}</td>
            <td class="num">{{ formatMoney(r.cost_total) }}</td>
            <td class="num red">{{ formatMoney(r.sale_price) }}</td>
            <td class="num red">{{ formatMoney(r.sale_total) }}</td>
            <td class="num muted">{{ r.warehouses_count }}</td>
          </tr>
        </tbody>
      </table>

      <!-- Вид: По складам -->
      <table v-else class="ms-table">
        <thead>
          <tr>
            <th style="width:240px">Склад</th>
            <th>Наименование</th>
            <th style="width:100px">Артикул</th>
            <th class="num" style="width:90px">Остаток</th>
            <th class="num" style="width:100px">Несниж.</th>
            <th class="num" style="width:90px">Доступно</th>
            <th style="width:60px">Ед.</th>
            <th class="num" style="width:100px">Дней на скл.</th>
            <th class="num" style="width:110px">Сумма себест.</th>
            <th class="num" style="width:120px">Сумма продажи</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="loading"><td colspan="10" class="muted" style="text-align:center;padding:24px">Загрузка...</td></tr>
          <tr v-else-if="warehouseGroups.length === 0"><td colspan="10" class="muted" style="text-align:center;padding:24px">Нет данных</td></tr>
          <template v-else v-for="g in warehouseGroups" :key="g.warehouse_id">
            <tr class="warehouse-row">
              <td colspan="3"><strong>{{ g.warehouse_name }}</strong></td>
              <td class="num"><strong>{{ formatQty(g.rows.reduce((s, r) => s + r.quantity, 0)) }}</strong></td>
              <td></td>
              <td class="num"><strong>{{ formatQty(g.rows.reduce((s, r) => s + r.available, 0)) }}</strong></td>
              <td></td>
              <td></td>
              <td class="num"><strong>{{ formatMoney(g.rows.reduce((s, r) => s + r.cost_total, 0)) }}</strong></td>
              <td class="num"><strong>{{ formatMoney(g.rows.reduce((s, r) => s + r.sale_total, 0)) }}</strong></td>
            </tr>
            <tr
              v-for="r in g.rows"
              :key="r.product_id + ':' + (r.warehouse_id || '')"
              class="clickable"
              :class="{ active: selectedProductId === r.product_id }"
              @click="openDetail(r.product_id)"
            >
              <td></td>
              <td class="link">{{ r.product_name }}</td>
              <td class="muted mono">{{ r.sku || '—' }}</td>
              <td class="num">{{ formatQty(r.quantity) }}</td>
              <td class="num muted">{{ r.min_stock > 0 ? formatQty(r.min_stock) : '—' }}</td>
              <td class="num"><strong>{{ formatQty(r.available) }}</strong></td>
              <td>{{ r.unit_short || '—' }}</td>
              <td class="num muted">{{ r.days_on_stock ?? '—' }}</td>
              <td class="num">{{ formatMoney(r.cost_total) }}</td>
              <td class="num red">{{ formatMoney(r.sale_total) }}</td>
            </tr>
          </template>
        </tbody>
      </table>

      <div class="ms-footer">
        <div class="ms-pager">
          <button :disabled="page === 1" @click="page--">◀</button>
          <span>{{ rangeFrom }}–{{ rangeTo }} из {{ totalCount }}</span>
          <button :disabled="rangeTo >= totalCount" @click="page++">▶</button>
        </div>
        <div class="ms-totals">
          <span>{{ formatQty(totals.quantity) }}</span>
          <span>{{ formatQty(totals.min_stock) }}</span>
          <span>{{ formatQty(totals.reserve) }}</span>
          <span>{{ formatQty(totals.incoming) }}</span>
          <span>{{ formatQty(totals.available) }}</span>
          <span>{{ formatMoney(totals.cost_total) }}</span>
          <span>{{ formatMoney(totals.sale_total) }}</span>
        </div>
      </div>
    </div>

    <!-- Правая панель -->
    <transition name="slide">
      <aside v-if="detail" class="detail-panel">
        <div class="panel-close" @click="closeDetail">✕</div>
        <div v-if="detailLoading" class="muted" style="padding:24px;text-align:center">Загрузка...</div>
        <template v-else-if="detail">
          <div class="detail-title"><span class="link">{{ detail.product.name }}</span></div>
          <div class="detail-meta">
            <span class="muted">Код</span> <span class="mono">{{ detail.product.id.slice(0, 6) }}</span>
            <span class="muted" style="margin-left:12px">Артикул</span>
            <span class="mono">{{ detail.product.sku || '—' }}</span>
          </div>
          <div class="summary-row">
            <div class="summary-cell"><div class="sum-label">Доступно</div><div class="sum-value">{{ formatQty(detail.summary.available) }}</div></div>
            <div class="summary-cell"><div class="sum-label">Дней на складе</div><div class="sum-value muted">—</div></div>
            <div class="summary-cell"><div class="sum-label">Резерв</div><div class="sum-value muted">{{ formatQty(detail.summary.reserve) }}</div></div>
            <div class="summary-cell"><div class="sum-label">Ожидание</div><div class="sum-value muted">{{ formatQty(detail.summary.incoming) }}</div></div>
            <div class="summary-cell"><div class="sum-label">Остаток</div><div class="sum-value">{{ formatQty(detail.summary.quantity) }}</div></div>
            <div class="summary-cell"><div class="sum-label">Несниж. ост.</div><div class="sum-value muted">{{ detail.summary.min_stock > 0 ? formatQty(detail.summary.min_stock) : '—' }}</div></div>
            <div class="summary-cell"><div class="sum-label">Себестоимость</div><div class="sum-value">{{ formatMoney(detail.summary.cost_price) }}</div></div>
            <div class="summary-cell"><div class="sum-label">Сумма себест.</div><div class="sum-value">{{ formatMoney(detail.summary.cost_sum) }}</div></div>
            <div class="summary-cell"><div class="sum-label">Цена продажи</div><div class="sum-value">{{ formatMoney(detail.summary.sale_price) }}</div></div>
            <div class="summary-cell"><div class="sum-label">Сумма продажи</div><div class="sum-value">{{ formatMoney(detail.summary.sale_sum) }}</div></div>
          </div>

          <div class="section-title">Себестоимость</div>
          <table class="detail-table">
            <thead>
              <tr>
                <th>Склад</th>
                <th>Тип документа</th>
                <th>Номер</th>
                <th>Дата документа</th>
                <th>Дата склад. опер.</th>
                <th class="num">Доступно</th>
                <th class="num">Дней на скл.</th>
                <th class="num">Себест.</th>
                <th class="num">Сумма себест.</th>
              </tr>
            </thead>
            <tbody>
              <template v-for="g in detail.warehouses" :key="g.warehouse_id">
                <tr class="warehouse-row">
                  <td><strong>{{ g.warehouse_name }}</strong></td>
                  <td colspan="4"></td>
                  <td class="num"><strong>{{ formatQty(g.quantity) }}</strong></td>
                  <td class="num"></td>
                  <td class="num"><strong>{{ formatMoney(g.cost_price) }}</strong></td>
                  <td class="num"><strong>{{ formatMoney(g.cost_sum) }}</strong></td>
                </tr>
                <tr v-for="m in g.movements" :key="m.id">
                  <td></td>
                  <td>{{ typeLabel(m.document_type) }}</td>
                  <td class="link mono">{{ m.document_number }}</td>
                  <td class="muted">{{ m.document_created_at ? formatDateTime(m.document_created_at) : '—' }}</td>
                  <td class="muted">{{ formatDateTime(m.movement_at) }}</td>
                  <td class="num">{{ formatQty(m.quantity_delta) }}</td>
                  <td class="num muted">{{ m.days_on_stock }}</td>
                  <td class="num">{{ formatMoney(m.cost_price) }}</td>
                  <td class="num">{{ formatMoney(m.cost_sum) }}</td>
                </tr>
                <tr v-if="g.movements.length === 0">
                  <td colspan="9" class="muted" style="text-align:center;padding:12px">Нет движений</td>
                </tr>
              </template>
              <tr v-if="detail.warehouses.length === 0">
                <td colspan="9" class="muted" style="text-align:center;padding:16px">Нет остатков</td>
              </tr>
            </tbody>
          </table>
        </template>
      </aside>
    </transition>
  </div>
</template>

<style scoped>
.stock-page { display: flex; gap: 16px; min-height: calc(100vh - 120px); position: relative; }
.stock-main { flex: 1; min-width: 0; }
.mono { font-family: monospace; font-size: 12px; }
tr.clickable { cursor: pointer; }
tr.clickable.active { background: #eef4ff; }
tr.clickable.active td { border-bottom-color: #c8d8ec; }
.red { color: #cf222e; }

.detail-panel { width: 720px; max-width: 55vw; background: #fff; border: 1px solid #d8dee4; border-radius: 6px; padding: 16px; position: sticky; top: 16px; align-self: flex-start; max-height: calc(100vh - 120px); overflow-y: auto; flex-shrink: 0; }
.panel-close { position: sticky; top: 0; float: right; cursor: pointer; font-size: 18px; color: #8c959f; padding: 2px 8px; border-radius: 4px; }
.panel-close:hover { background: #f3f4f6; color: #1f2328; }
.detail-title { font-size: 18px; margin-bottom: 4px; }
.detail-meta { font-size: 12px; margin-bottom: 14px; }
.detail-meta .muted { margin-right: 4px; }
.summary-row { display: flex; flex-wrap: wrap; gap: 12px 22px; margin-bottom: 20px; padding-bottom: 12px; border-bottom: 1px solid #eaeef2; }
.summary-cell { min-width: 90px; }
.sum-label { font-size: 11px; color: #8c959f; margin-bottom: 3px; white-space: nowrap; }
.sum-value { font-size: 15px; font-weight: 600; color: #1f2328; white-space: nowrap; }
.sum-value.muted { color: #8c959f; font-weight: 400; }
.section-title { color: #cf222e; font-size: 14px; font-weight: 600; margin-bottom: 10px; }
.detail-table { width: 100%; border-collapse: collapse; font-size: 12px; }
.detail-table thead th { background: #fff; border-bottom: 2px solid #d8dee4; padding: 6px 8px; text-align: left; font-weight: 500; color: #2c5d9c; white-space: nowrap; }
.detail-table tbody td { padding: 6px 8px; border-bottom: 1px solid #eaeef2; color: #1f2328; vertical-align: middle; }
.detail-table .num { text-align: right; font-variant-numeric: tabular-nums; }
.detail-table .muted { color: #8c959f; }
.detail-table .warehouse-row td { background: #f6f8fa; font-size: 12px; }
.detail-table .link { color: #2c5d9c; cursor: pointer; }
.detail-table .link:hover { text-decoration: underline; }
.slide-enter-active, .slide-leave-active { transition: transform 0.2s ease, opacity 0.2s ease; }
.slide-enter-from, .slide-leave-to { transform: translateX(20px); opacity: 0; }

.ms-title { display: flex; align-items: center; gap: 8px; font-size: 20px; font-weight: 600; color: #1f2328; margin-bottom: 12px; }
.ms-help { width: 20px; height: 20px; border-radius: 50%; border: 1px solid #b8c0c8; background: transparent; color: #57606a; cursor: pointer; display: inline-flex; align-items: center; justify-content: center; padding: 0; }
.ms-help:hover { background: #f0f2f5; }
.ms-refresh { width: 24px; height: 24px; padding: 0; display: inline-flex; align-items: center; justify-content: center; border: none; background: transparent; color: #57606a; cursor: pointer; }
.ms-refresh:hover { color: #2c5d9c; }
.ms-toolbar { display: flex; align-items: center; gap: 6px; margin-bottom: 12px; flex-wrap: wrap; }
.view-switch { display: inline-flex; border: 1px solid #d0d7de; border-radius: 4px; overflow: hidden; margin-right: 6px; }
.view-switch button { height: 30px; padding: 0 14px; font-size: 13px; background: #fff; border: none; cursor: pointer; color: #1f2328; }
.view-switch button + button { border-left: 1px solid #d0d7de; }
.view-switch button.active { background: #2c5d9c; color: #fff; }
.ms-filter { background: #eef1f5; border: 1px solid #d8dee4; border-radius: 4px; padding: 10px 14px 12px; margin-bottom: 12px; }
.filter-actions { display: flex; align-items: center; gap: 6px; margin-bottom: 10px; }
.filter-grid { display: grid; grid-template-columns: repeat(6, minmax(0, 1fr)); gap: 10px 14px; }
.filter-grid--top { margin-bottom: 10px; }
.filter-grid--mid { grid-template-columns: repeat(6, minmax(0, 1fr)); }
.filter-field { display: flex; flex-direction: column; gap: 3px; }
.filter-label { font-size: 12px; color: #57606a; display: flex; align-items: center; gap: 5px; }
.filter-label .dot { width: 8px; height: 8px; border-radius: 50%; background: #2c5d9c; flex-shrink: 0; }
.filter-field input, .filter-field select { height: 28px; padding: 0 8px; font-size: 12px; border: 1px solid #d0d7de; border-radius: 3px; background: #fff; color: #1f2328; }
.filter-field input[type="radio"] { height: auto; width: auto; padding: 0; }
.range-row { display: flex; align-items: center; gap: 6px; }
.range-row input { width: 100%; min-width: 0; }
.range-sep { color: #8c959f; }
.error-box { background: #fff0f0; border: 1px solid #f5c1c1; color: #cf222e; padding: 8px 12px; border-radius: 4px; margin-bottom: 12px; font-size: 13px; }
.warehouse-row td { background: #f6f8fa; }
.warehouse-row td strong { font-size: 13px; }
</style>