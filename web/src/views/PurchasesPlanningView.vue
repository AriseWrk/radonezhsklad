<template>
  <div class="page">
    <div class="ms-title">
      <button class="ms-help" title="Справка"><MsIcon name="help" :size="14" /></button>
      <span>Управление закупками</span>
      <button class="ms-refresh" @click="load" title="Обновить"><MsIcon name="refresh" :size="14" /></button>
    </div>

    <div class="ms-toolbar">
      <MsButton icon="filter" @click="showFilter = !showFilter">Фильтр</MsButton>
      <input v-model="filters.search" class="ms-input" placeholder="Название, SKU" />
      <div class="ms-counter">{{ filtered.length }}</div>
      <div class="days-switch">
        <span class="lbl">Прогноз на</span>
        <input v-model.number="days" type="number" min="1" max="365" @change="load" class="days-input" />
        <span class="lbl">дней</span>
      </div>
      <MsButton icon="print" @click="print">Печать</MsButton>
      <div class="toolbar-spacer"></div>
      <MsButton variant="icon" icon="gear" title="Настройки" />
    </div>

    <div v-if="showFilter" class="ms-filter">
      <div class="filter-actions">
        <MsButton variant="green" @click="page = 1">Найти</MsButton>
        <MsButton @click="clearFilters">Очистить</MsButton>
      </div>
      <div class="filter-grid">
        <div class="filter-field">
          <label class="filter-label"><span class="dot"></span>Остаток</label>
          <select v-model="filters.stockMode">
            <option value="any">Любой</option>
            <option value="positive">Положительный</option>
            <option value="zero">Нулевой</option>
          </select>
        </div>
        <div class="filter-field">
          <label class="filter-label"><span class="dot"></span>Заказать</label>
          <select v-model="filters.toOrderMode">
            <option value="any">Все</option>
            <option value="only">Только с рекомендацией</option>
          </select>
        </div>
      </div>
    </div>

    <div v-if="error" class="error-box">{{ error }}</div>

    <table class="ms-table2">
      <thead>
        <tr>
          <th @click="sortBy('product_name')">Наименование<span v-if="sortKey === 'product_name'" class="sort">{{ sortDir === 'asc' ? '▲' : '▼' }}</span></th>
          <th class="col-num">Артикул</th>
          <th class="col-num">Ед.</th>
          <th class="col-num-right" @click="sortBy('sold_qty')">Кол-во<span v-if="sortKey === 'sold_qty'" class="sort">{{ sortDir === 'asc' ? '▲' : '▼' }}</span></th>
          <th class="col-num-right">Сумма</th>
          <th class="col-num-right">Себест.</th>
          <th class="col-num-right">Прибыль</th>
          <th class="col-num-right">Рент.</th>
          <th class="col-num-right">Остаток</th>
          <th class="col-num-right">Ожид.</th>
          <th class="col-num-right">Доступно</th>
          <th class="col-num-right">Дней зап.</th>
          <th class="col-num-right">Запас</th>
          <th class="col-num-right" @click="sortBy('to_order')">Заказать<span v-if="sortKey === 'to_order'" class="sort">{{ sortDir === 'asc' ? '▲' : '▼' }}</span></th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="loading"><td colspan="14" class="empty">Загрузка...</td></tr>
        <tr v-else-if="filtered.length === 0"><td colspan="14" class="empty">Нет данных</td></tr>
        <tr v-else v-for="r in paginated" :key="r.product_id" class="row">
          <td class="link">{{ r.product_name }}</td>
          <td class="col-num muted mono">{{ r.sku || '—' }}</td>
          <td class="col-num">{{ r.unit_short || '—' }}</td>
          <td class="col-num-right">{{ formatQty(r.sold_qty) }}</td>
          <td class="col-num-right">{{ formatMoney(r.sold_sum) }}</td>
          <td class="col-num-right">{{ formatMoney(r.cost_sum) }}</td>
          <td class="col-num-right" :class="r.profit > 0 ? 'diff-plus' : (r.profit < 0 ? 'diff-minus' : 'muted')">{{ formatMoney(r.profit) }}</td>
          <td class="col-num-right">{{ r.margin.toFixed(1) }}%</td>
          <td class="col-num-right">{{ formatQty(r.stock) }}</td>
          <td class="col-num-right">{{ formatQty(r.incoming) }}</td>
          <td class="col-num-right"><b>{{ formatQty(r.available) }}</b></td>
          <td class="col-num-right">{{ r.days_of_stock == null ? '—' : r.days_of_stock.toFixed(0) }}</td>
          <td class="col-num-right" :class="r.supply < 0 ? 'diff-minus' : (r.supply > 0 ? 'diff-plus' : 'muted')">{{ formatQty(r.supply) }}</td>
          <td class="col-num-right">
            <strong :class="r.to_order > 0 ? 'red' : 'muted'">{{ formatQty(r.to_order) }}</strong>
          </td>
        </tr>
      </tbody>
    </table>

    <div class="ms-footer2">
      <div class="pager">
        <button :disabled="page === 1" @click="page--">‹</button>
        <span class="range">{{ rangeFrom }}-{{ rangeTo }} из {{ filtered.length }}</span>
        <button :disabled="rangeTo >= filtered.length" @click="page++">›</button>
      </div>
      <div class="totals-inline">
        <span>Продано: <b>{{ formatMoney(totalSold) }}</b></span>
        <span>Прибыль: <b>{{ formatMoney(totalProfit) }}</b></span>
        <span>К заказу: <b>{{ formatQty(totalToOrder) }}</b></span>
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { listProducts, type Product } from '../api/products'
import { listStockExtended, type StockExtendedRow } from '../api/stock'
import { salesAnalytics, type SalesAnalyticsRow } from '../api/analytics'
import { apiErrorMessage } from '../api/client'
import MsButton from '../components/MsButton.vue'
import MsIcon from '../components/MsIcon.vue'

interface PlanningRow {
  product_id: string
  product_name: string
  sku?: string
  unit_short: string
  // продажи
  sold_qty: number
  sold_sum: number
  cost_sum: number
  profit: number
  margin: number
  orders_count: number
  // склад
  stock: number
  incoming: number
  available: number
  days_on_stock?: number
  days_of_stock: number | null
  supply: number
  to_order: number
}

const loading = ref(false)
const error = ref<string | null>(null)
const days = ref(14)
const showFilter = ref(false)
const page = ref(1)
const perPage = 100

const filters = reactive({
  search: '',
  stockMode: 'any' as 'any' | 'positive' | 'zero',
  toOrderMode: 'any' as 'any' | 'only',
})

const sortKey = ref<'product_name' | 'sold_qty' | 'to_order'>('sold_qty')
const sortDir = ref<'asc' | 'desc'>('desc')

const allRows = ref<PlanningRow[]>([])

function formatQty(n: number) {
  if (n === 0) return '0'
  return n.toLocaleString('ru-RU', { maximumFractionDigits: 3 })
}
function formatMoney(n: number) {
  const sign = n < 0 ? '−' : ''
  return sign + Math.abs(n).toLocaleString('ru-RU', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}
function sortBy(k: typeof sortKey.value) {
  if (sortKey.value === k) sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  else { sortKey.value = k; sortDir.value = 'desc' }
  page.value = 1
}

async function load() {
  loading.value = true
  error.value = null
  try {
    const [prods, stock, sales] = await Promise.all([
      listProducts(true),
      listStockExtended(),
      salesAnalytics(days.value),
    ])

    // карты по product_id
    const productsMap = new Map<string, Product>()
    for (const p of prods) productsMap.set(p.id, p)

    const stockMap = new Map<string, StockExtendedRow>()
    for (const s of stock.items) stockMap.set(s.product_id, s)

    const salesMap = new Map<string, SalesAnalyticsRow>()
    for (const s of sales.items) salesMap.set(s.product_id, s)

    // объединяем ключи
    const allIds = new Set<string>()
    for (const p of prods) allIds.add(p.id)

    const rows: PlanningRow[] = []
    for (const id of allIds) {
      const p = productsMap.get(id)!
      const s = stockMap.get(id)
      const sl = salesMap.get(id)

      const soldQty  = sl?.sold_qty ?? 0
      const soldSum  = sl?.sold_sum ?? 0
      const costPrice = (p as any).cost_price ?? 0
      const costSum  = soldQty * costPrice
      const profit   = soldSum - costSum
      const margin   = soldSum > 0 ? (profit / soldSum) * 100 : 0

      const stockQty = s?.quantity ?? 0
      const incoming = s?.incoming ?? 0
      const available = stockQty + incoming

      const avgDaily = days.value > 0 ? soldQty / days.value : 0
      const daysOfStock = avgDaily > 0 ? available / avgDaily : null

      // Запас: сколько "сверх" ожидаемого спроса
      const expectedDemand = avgDaily * days.value
      const supply = available - expectedDemand

      // Заказать: до уровня ожидаемого спроса
      const toOrder = Math.max(0, expectedDemand - available)

      rows.push({
        product_id: id,
        product_name: p.name,
        sku: p.sku ?? undefined,
        unit_short: s?.unit_short ?? '',
        sold_qty: soldQty,
        sold_sum: soldSum,
        cost_sum: costSum,
        profit,
        margin,
        orders_count: sl?.orders_count ?? 0,
        stock: stockQty,
        incoming,
        available,
        days_on_stock: s?.days_on_stock,
        days_of_stock: daysOfStock,
        supply,
        to_order: toOrder,
      })
    }

    allRows.value = rows
  } catch (e) {
    error.value = apiErrorMessage(e)
  } finally {
    loading.value = false
  }
}

const filtered = computed(() => {
  let rows = allRows.value
  if (filters.search) {
    const q = filters.search.toLowerCase()
    rows = rows.filter((r) =>
      r.product_name.toLowerCase().includes(q) ||
      (r.sku ?? '').toLowerCase().includes(q)
    )
  }
  if (filters.stockMode === 'positive') rows = rows.filter((r) => r.stock > 0)
  if (filters.stockMode === 'zero')     rows = rows.filter((r) => r.stock === 0)
  if (filters.toOrderMode === 'only')   rows = rows.filter((r) => r.to_order > 0)

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

const paginated = computed(() => {
  const from = (page.value - 1) * perPage
  return filtered.value.slice(from, from + perPage)
})

const rangeFrom = computed(() => filtered.value.length === 0 ? 0 : (page.value - 1) * perPage + 1)
const rangeTo = computed(() => Math.min(page.value * perPage, filtered.value.length))

const totalSold = computed(() => filtered.value.reduce((s, r) => s + r.sold_sum, 0))
const totalProfit = computed(() => filtered.value.reduce((s, r) => s + r.profit, 0))
const totalToOrder = computed(() => filtered.value.reduce((s, r) => s + r.to_order, 0))

function clearFilters() {
  filters.search = ''
  filters.stockMode = 'any'
  filters.toOrderMode = 'any'
  page.value = 1
}
function print() { window.print() }

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
.ms-input { flex: 1; max-width: 260px; height: 30px; padding: 0 10px; font-size: 13px; border: 1px solid #d0d7de; border-radius: 4px; background: #fff; color: #1f2328; }
.ms-counter { min-width: 40px; height: 30px; padding: 0 10px; display: inline-flex; align-items: center; justify-content: center; border: 1px solid #d0d7de; border-radius: 4px; background: #fff; color: #8c959f; font-variant-numeric: tabular-nums; font-size: 13px; }
.days-switch { display: inline-flex; align-items: center; gap: 6px; font-size: 13px; color: #57606a; margin: 0 6px; }
.days-input { width: 64px; height: 30px; padding: 0 8px; font-size: 13px; border: 1px solid #d0d7de; border-radius: 4px; }
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
.ms-table2 .col-num { width: 80px; }
.ms-table2 .col-num-right { text-align: right; font-variant-numeric: tabular-nums; }
.ms-table2 .link { color: #2c5d9c; font-weight: 500; }
.ms-table2 .sort { font-size: 9px; margin-left: 3px; }
.ms-table2 .empty { text-align: center; padding: 24px; color: #8c959f; }
.mono { font-family: monospace; font-size: 12px; }
.muted { color: #8c959f; }
.diff-plus { color: #1a7f37; font-weight: 600; }
.diff-minus { color: #cf222e; font-weight: 600; }
.red { color: #cf222e; }

.ms-footer2 { display: flex; align-items: center; justify-content: space-between; padding: 6px 10px; font-size: 12px; color: #57606a; background: #fff; border-top: 1px solid #eaeef2; }
.pager { display: flex; align-items: center; gap: 4px; }
.pager button { width: 22px; height: 22px; padding: 0; border: 1px solid #d0d7de; background: #fff; border-radius: 3px; cursor: pointer; font-size: 12px; color: #1f2328; }
.pager button:disabled { opacity: 0.4; cursor: default; }
.pager .range { margin: 0 6px; font-variant-numeric: tabular-nums; }
.totals-inline { display: flex; gap: 16px; font-variant-numeric: tabular-nums; }
</style>