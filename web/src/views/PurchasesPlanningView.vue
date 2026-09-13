<template>
  <div>
    <div class="page-title-bar">
      <div class="page-title">
        <span>Управление закупками</span>
        <span class="refresh" @click="load" title="Обновить">↻</span>
      </div>
      <div class="page-actions">
        <button class="btn" @click="showFilter = !showFilter">Фильтр</button>
        <button class="btn" @click="print">Печать</button>
        <div class="days-switch">
          <span class="lbl">Прогноз на</span>
          <input v-model.number="days" type="number" min="1" max="365" @change="load" class="days-input" />
          <span class="lbl">дней</span>
        </div>
      </div>
    </div>

    <div v-if="showFilter" class="filter-panel">
      <div class="filter-row">
        <div class="filter-actions">
          <button class="btn-find" @click="page = 1">Найти</button>
          <button class="btn-clear" @click="clearFilters">Очистить</button>
        </div>
        <div class="filter-field">
          <label>Товар</label>
          <input v-model="filters.search" placeholder="Название, SKU" />
        </div>
        <div class="filter-field">
          <label>Остаток</label>
          <select v-model="filters.stockMode">
            <option value="any">Любой</option>
            <option value="positive">Положительный</option>
            <option value="zero">Нулевой</option>
          </select>
        </div>
        <div class="filter-field">
          <label>Заказать</label>
          <select v-model="filters.toOrderMode">
            <option value="any">Все</option>
            <option value="only">Только с рекомендацией</option>
          </select>
        </div>
      </div>
    </div>

    <div v-if="error" class="error-box">{{ error }}</div>

    <table class="ms-table">
      <thead>
        <tr>
          <th @click="sortBy('product_name')">Наименование <span v-if="sortKey === 'product_name'">{{ sortDir === 'asc' ? '↑' : '↓' }}</span></th>
          <th style="width:80px">Код</th>
          <th style="width:100px">Артикул</th>
          <th style="width:70px">Ед. изм.</th>
          <th class="num" style="width:90px" @click="sortBy('sold_qty')">Кол-во <span v-if="sortKey === 'sold_qty'">{{ sortDir === 'asc' ? '↑' : '↓' }}</span></th>
          <th class="num" style="width:100px">Сумма</th>
          <th class="num" style="width:100px">Себестоимость</th>
          <th class="num" style="width:100px">Прибыль</th>
          <th class="num" style="width:90px">Рентабельн.</th>
          <th class="num" style="width:100px">Продаж в зак.</th>
          <th class="num" style="width:90px">Остаток</th>
          <th class="num" style="width:80px">Резерв</th>
          <th class="num" style="width:90px">Ожидание</th>
          <th class="num" style="width:90px">Доступно</th>
          <th class="num" style="width:100px">Дней на скл.</th>
          <th class="num" style="width:100px">Дней запаса</th>
          <th class="num" style="width:100px">Запас</th>
          <th class="num" style="width:100px" @click="sortBy('to_order')">Заказать <span v-if="sortKey === 'to_order'">{{ sortDir === 'asc' ? '↑' : '↓' }}</span></th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="loading">
          <td colspan="18" class="muted" style="text-align:center;padding:24px">Загрузка...</td>
        </tr>
        <tr v-else-if="filtered.length === 0">
          <td colspan="18" class="muted" style="text-align:center;padding:24px">Нет данных</td>
        </tr>
        <tr v-else v-for="r in paginated" :key="r.product_id">
          <td class="link">{{ r.product_name }}</td>
          <td class="muted mono">{{ r.product_id.slice(0, 6) }}</td>
          <td class="muted mono">{{ r.sku || '—' }}</td>
          <td>{{ r.unit_short || '—' }}</td>
          <td class="num">{{ formatQty(r.sold_qty) }}</td>
          <td class="num">{{ formatMoney(r.sold_sum) }}</td>
          <td class="num">{{ formatMoney(r.cost_sum) }}</td>
          <td class="num" :class="r.profit > 0 ? 'diff-plus' : (r.profit < 0 ? 'diff-minus' : 'muted')">
            {{ formatMoney(r.profit) }}
          </td>
          <td class="num">{{ r.margin.toFixed(1) }}%</td>
          <td class="num muted">{{ r.orders_count }}</td>
          <td class="num">{{ formatQty(r.stock) }}</td>
          <td class="num muted">{{ formatQty(0) }}</td>
          <td class="num">{{ formatQty(r.incoming) }}</td>
          <td class="num"><strong>{{ formatQty(r.available) }}</strong></td>
          <td class="num muted">{{ r.days_on_stock ?? '—' }}</td>
          <td class="num">{{ r.days_of_stock == null ? '—' : r.days_of_stock.toFixed(0) }}</td>
          <td class="num" :class="r.supply < 0 ? 'diff-minus' : (r.supply > 0 ? 'diff-plus' : 'muted')">
            {{ formatQty(r.supply) }}
          </td>
          <td class="num">
            <strong :class="r.to_order > 0 ? 'red' : 'muted'">{{ formatQty(r.to_order) }}</strong>
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
        <span>Продано: {{ formatMoney(totalSold) }}</span>
        <span>Прибыль: {{ formatMoney(totalProfit) }}</span>
        <span>К заказу: {{ formatQty(totalToOrder) }} ед.</span>
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
.days-switch {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: #57606a;
}
.days-input {
  width: 60px;
  padding: 5px 8px;
  border: 1px solid #d0d7de;
  border-radius: 4px;
  text-align: center;
  font-size: 13px;
}
.lbl { white-space: nowrap; }
.diff-plus  { color: #1a7f37; font-weight: 600; }
.diff-minus { color: #cf222e; font-weight: 600; }
.red { color: #cf222e; }
.mono { font-family: monospace; font-size: 12px; }
</style>