<template>
  <div class="stock-page">
    <div class="stock-main">
      <!-- Заголовок -->
      <div class="ms-title">
        <button class="ms-help" title="Справка"><MsIcon name="help" :size="14" /></button>
        <span>Остатки</span>
        <button class="ms-refresh" @click="load" title="Обновить"><MsIcon name="refresh" :size="14" /></button>
      </div>

      <!-- Тулбар -->
      <div class="ms-toolbar">
        <div class="view-switch">
          <button :class="{ active: view === 'products' }" @click="view = 'products'">По товарам</button>
          <button :class="{ active: view === 'warehouses' }" @click="view = 'warehouses'">По складам</button>
        </div>
        <MsButton icon="filter" @click="toggleFilter">Фильтр</MsButton>
        <MsButton icon="print" @click="print">Печать</MsButton>
        <MsButton variant="icon" icon="gear" title="Настройки" />
      </div>

      <!-- Фильтры -->
      <div v-if="showFilter" class="ms-filter">
        <div class="filter-actions">
          <MsButton variant="green" @click="load">Найти</MsButton>
          <MsButton @click="clearFilters">Очистить</MsButton>
        </div>
        <div class="filter-grid">
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
            <label class="filter-label"><span class="dot"></span>Склад</label>
            <select v-model="filters.warehouseId" @change="load">
              <option value="">—</option>
              <option v-for="w in warehouses" :key="w.id" :value="w.id">{{ w.name }}</option>
            </select>
          </div>
          <div class="filter-field">
            <label class="filter-label"><span class="dot"></span>Товар</label>
            <input v-model="filters.search" placeholder="Название или артикул" />
          </div>
          <div class="filter-field">
            <label class="filter-label"><span class="dot"></span>Артикул</label>
            <input v-model="filters.sku" placeholder="SKU" />
          </div>
        </div>
      </div>

      <div v-if="error" class="error-box">{{ error }}</div>

      <!-- Таблица -->
      <table class="ms-table">
        <thead>
          <tr>
            <th @click="sortBy('product_name')">Наименование <span v-if="sortKey === 'product_name'">{{ sortDir === 'asc' ? '↑' : '↓' }}</span></th>
            <th style="width:70px">Код</th>
            <th style="width:100px">Артикул</th>
            <th class="num" style="width:90px" @click="sortBy('quantity')">Остаток <span v-if="sortKey === 'quantity'">{{ sortDir === 'asc' ? '↑' : '↓' }}</span></th>
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
          </tr>
        </thead>
        <tbody>
          <tr v-if="loading"><td colspan="14" class="muted" style="text-align:center;padding:24px">Загрузка...</td></tr>
          <tr v-else-if="filteredRows.length === 0"><td colspan="14" class="muted" style="text-align:center;padding:24px">Нет данных</td></tr>
          <tr
            v-else
            v-for="r in paginatedRows"
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
          </tr>
        </tbody>
      </table>

      <div class="ms-footer">
        <div class="ms-pager">
          <button :disabled="page === 1" @click="page--">◀</button>
          <span>{{ rangeFrom }}–{{ rangeTo }} из {{ filteredRows.length }}</span>
          <button :disabled="rangeTo >= filteredRows.length" @click="page++">▶</button>
        </div>
        <div class="ms-totals">
          <span>{{ formatQty(filteredTotals.quantity) }}</span>
          <span>{{ formatQty(filteredTotals.min_stock) }}</span>
          <span>{{ formatQty(filteredTotals.reserve) }}</span>
          <span>{{ formatQty(filteredTotals.incoming) }}</span>
          <span>{{ formatQty(filteredTotals.available) }}</span>
          <span>{{ formatMoney(filteredTotals.cost_total) }}</span>
          <span>{{ formatMoney(filteredTotals.sale_total) }}</span>
        </div>
      </div>
    </div>

    <!-- Правая панель -->
    <transition name="slide">
      <aside v-if="detail" class="detail-panel">
        <div class="panel-close" @click="closeDetail">✕</div>

        <div v-if="detailLoading" class="muted" style="padding:24px;text-align:center">Загрузка...</div>
        <template v-else-if="detail">
          <!-- Название -->
          <div class="detail-title">
            <span class="link">{{ detail.product.name }}</span>
          </div>
          <div class="detail-meta">
            <span class="muted">Код</span> <span class="mono">{{ detail.product.id.slice(0, 6) }}</span>
            <span class="muted" style="margin-left:12px">Артикул</span>
            <span class="mono">{{ detail.product.sku || '—' }}</span>
          </div>

          <!-- Сводка -->
          <div class="summary-row">
            <div class="summary-cell">
              <div class="sum-label">Доступно</div>
              <div class="sum-value">{{ formatQty(detail.summary.available) }}</div>
            </div>
            <div class="summary-cell">
              <div class="sum-label">Дней на складе</div>
              <div class="sum-value muted">—</div>
            </div>
            <div class="summary-cell">
              <div class="sum-label">Резерв</div>
              <div class="sum-value muted">{{ formatQty(detail.summary.reserve) }}</div>
            </div>
            <div class="summary-cell">
              <div class="sum-label">Ожидание</div>
              <div class="sum-value muted">{{ formatQty(detail.summary.incoming) }}</div>
            </div>
            <div class="summary-cell">
              <div class="sum-label">Остаток</div>
              <div class="sum-value">{{ formatQty(detail.summary.quantity) }}</div>
            </div>
            <div class="summary-cell">
              <div class="sum-label">Несниж. ост.</div>
              <div class="sum-value muted">{{ detail.summary.min_stock > 0 ? formatQty(detail.summary.min_stock) : '—' }}</div>
            </div>
            <div class="summary-cell">
              <div class="sum-label">Себестоимость</div>
              <div class="sum-value">{{ formatMoney(detail.summary.cost_price) }}</div>
            </div>
            <div class="summary-cell">
              <div class="sum-label">Сумма себест.</div>
              <div class="sum-value">{{ formatMoney(detail.summary.cost_sum) }}</div>
            </div>
            <div class="summary-cell">
              <div class="sum-label">Цена продажи</div>
              <div class="sum-value">{{ formatMoney(detail.summary.sale_price) }}</div>
            </div>
            <div class="summary-cell">
              <div class="sum-label">Сумма продажи</div>
              <div class="sum-value">{{ formatMoney(detail.summary.sale_sum) }}</div>
            </div>
          </div>

          <!-- Себестоимость -->
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
                  <td colspan="9" class="muted" style="text-align:center;padding:12px">
                    Нет движений
                  </td>
                </tr>
              </template>
              <tr v-if="detail.warehouses.length === 0">
                <td colspan="9" class="muted" style="text-align:center;padding:16px">
                  Нет остатков
                </td>
              </tr>
            </tbody>
          </table>
        </template>
      </aside>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import {
  listStockExtended,
  productStockDetail,
  type StockExtendedResponse,
  type StockExtendedRow,
  type ProductStockDetail,
} from '../api/stock'
import { listWarehouses, type Warehouse } from '../api/warehouses'
import { apiErrorMessage } from '../api/client'
import MsButton from '../components/MsButton.vue'
import MsIcon from '../components/MsIcon.vue'

const loading = ref(false)
const error = ref<string | null>(null)
const data = ref<StockExtendedResponse | null>(null)
const warehouses = ref<Warehouse[]>([])
const view = ref<'products' | 'warehouses'>('products')
const showFilter = ref(true)
const page = ref(1)
const perPage = 100

const selectedProductId = ref<string | null>(null)
const detail = ref<ProductStockDetail | null>(null)
const detailLoading = ref(false)

const filters = reactive({
  qtyMode: 'any',
  availMode: 'any',
  warehouseId: '',
  search: '',
  sku: '',
})

const sortKey = ref<keyof StockExtendedRow>('product_name')
const sortDir = ref<'asc' | 'desc'>('asc')

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
    receipt: 'Приёмка',
    shipment: 'Отгрузка',
    transfer: 'Перемещение',
    inventory: 'Инвентаризация',
  } as Record<string, string>)[t] ?? (t || '—')
}

function sortBy(key: keyof StockExtendedRow) {
  if (sortKey.value === key) sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  else { sortKey.value = key; sortDir.value = 'asc' }
  page.value = 1
}

const filteredRows = computed(() => {
  let rows = data.value?.items ?? []

  if (filters.search) {
    const q = filters.search.toLowerCase()
    rows = rows.filter((r) =>
      r.product_name.toLowerCase().includes(q) || (r.sku ?? '').toLowerCase().includes(q)
    )
  }
  if (filters.sku) {
    const q = filters.sku.toLowerCase()
    rows = rows.filter((r) => (r.sku ?? '').toLowerCase().includes(q))
  }
  switch (filters.qtyMode) {
    case 'positive': rows = rows.filter((r) => r.quantity > 0); break
    case 'zero':     rows = rows.filter((r) => r.quantity === 0); break
  }
  switch (filters.availMode) {
    case 'nonZero': rows = rows.filter((r) => r.available !== 0); break
    case 'zero':    rows = rows.filter((r) => r.available === 0); break
  }

  return [...rows].sort((a, b) => {
    const av = a[sortKey.value]
    const bv = b[sortKey.value]
    if (typeof av === 'number' && typeof bv === 'number') {
      return sortDir.value === 'asc' ? av - bv : bv - av
    }
    const as = String(av ?? '').toLowerCase()
    const bs = String(bv ?? '').toLowerCase()
    return sortDir.value === 'asc' ? as.localeCompare(bs, 'ru') : bs.localeCompare(as, 'ru')
  })
})

const filteredTotals = computed(() => {
  const t = { quantity: 0, min_stock: 0, reserve: 0, incoming: 0, available: 0, cost_total: 0, sale_total: 0 }
  for (const r of filteredRows.value) {
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

const paginatedRows = computed(() => {
  const from = (page.value - 1) * perPage
  return filteredRows.value.slice(from, from + perPage)
})

const rangeFrom = computed(() => filteredRows.value.length === 0 ? 0 : (page.value - 1) * perPage + 1)
const rangeTo = computed(() => Math.min(page.value * perPage, filteredRows.value.length))

async function load() {
  loading.value = true
  error.value = null
  try {
    data.value = await listStockExtended({ warehouse_id: filters.warehouseId || undefined })
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

function closeDetail() {
  selectedProductId.value = null
  detail.value = null
}

function toggleFilter() { showFilter.value = !showFilter.value }
function clearFilters() {
  filters.qtyMode = 'any'
  filters.availMode = 'any'
  filters.warehouseId = ''
  filters.search = ''
  filters.sku = ''
  load()
}
function print() { window.print() }

onMounted(async () => {
  try { warehouses.value = await listWarehouses() } catch { /* ignore */ }
  await load()
})
</script>

<style scoped>
.stock-page {
  display: flex;
  gap: 16px;
  min-height: calc(100vh - 120px);
  position: relative;
}
.stock-main {
  flex: 1;
  min-width: 0;
}
.mono { font-family: monospace; font-size: 12px; }
tr.clickable { cursor: pointer; }
tr.clickable.active { background: #eef4ff; }
tr.clickable.active td { border-bottom-color: #c8d8ec; }
.red { color: #cf222e; }

/* Правая панель */
.detail-panel {
  width: 720px;
  max-width: 55vw;
  background: #fff;
  border: 1px solid #d8dee4;
  border-radius: 6px;
  padding: 16px;
  position: sticky;
  top: 16px;
  align-self: flex-start;
  max-height: calc(100vh - 120px);
  overflow-y: auto;
  flex-shrink: 0;
}
.panel-close {
  position: sticky;
  top: 0;
  float: right;
  cursor: pointer;
  font-size: 18px;
  color: #8c959f;
  padding: 2px 8px;
  border-radius: 4px;
}
.panel-close:hover { background: #f3f4f6; color: #1f2328; }

.detail-title {
  font-size: 18px;
  margin-bottom: 4px;
}
.detail-meta {
  font-size: 12px;
  margin-bottom: 14px;
}
.detail-meta .muted { margin-right: 4px; }

.summary-row {
  display: flex;
  flex-wrap: wrap;
  gap: 12px 22px;
  margin-bottom: 20px;
  padding-bottom: 12px;
  border-bottom: 1px solid #eaeef2;
}
.summary-cell { min-width: 90px; }
.sum-label {
  font-size: 11px;
  color: #8c959f;
  margin-bottom: 3px;
  white-space: nowrap;
}
.sum-value {
  font-size: 15px;
  font-weight: 600;
  color: #1f2328;
  white-space: nowrap;
}
.sum-value.muted { color: #8c959f; font-weight: 400; }

.section-title {
  color: #cf222e;
  font-size: 14px;
  font-weight: 600;
  margin-bottom: 10px;
}

.detail-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
}
.detail-table thead th {
  background: #fff;
  border-bottom: 2px solid #d8dee4;
  padding: 6px 8px;
  text-align: left;
  font-weight: 500;
  color: #2c5d9c;
  white-space: nowrap;
}
.detail-table tbody td {
  padding: 6px 8px;
  border-bottom: 1px solid #eaeef2;
  color: #1f2328;
  vertical-align: middle;
}
.detail-table .num { text-align: right; font-variant-numeric: tabular-nums; }
.detail-table .muted { color: #8c959f; }
.detail-table .warehouse-row td {
  background: #f6f8fa;
  font-size: 12px;
}
.detail-table .link { color: #2c5d9c; cursor: pointer; }
.detail-table .link:hover { text-decoration: underline; }

/* Анимация выезда */
.slide-enter-active, .slide-leave-active { transition: transform 0.2s ease, opacity 0.2s ease; }
.slide-enter-from, .slide-leave-to { transform: translateX(20px); opacity: 0; }

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