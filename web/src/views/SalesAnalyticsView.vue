<template>
  <div>
    <div class="page-title-bar">
      <div class="page-title">
        <span>Аналитика продаж</span>
        <span class="refresh" @click="load" title="Обновить">↻</span>
      </div>
      <div class="page-actions">
        <button class="btn" @click="showFilter = !showFilter">Фильтр</button>
        <button class="btn" @click="exportCsv">Экспорт CSV</button>
        <div class="days-switch">
          <span class="lbl">Период:</span>
          <select v-model.number="days" @change="load" class="days-select">
            <option :value="7">7 дней</option>
            <option :value="14">14 дней</option>
            <option :value="30">30 дней</option>
            <option :value="90">90 дней</option>
            <option :value="180">180 дней</option>
            <option :value="365">365 дней</option>
          </select>
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
          <label>Только проданные</label>
          <select v-model="filters.onlySold">
            <option value="yes">Да</option>
            <option value="no">Все</option>
          </select>
        </div>
        <div class="filter-field">
          <label>Сортировка</label>
          <select v-model="sortByField">
            <option value="sold_sum">По сумме</option>
            <option value="sold_qty">По количеству</option>
            <option value="profit">По прибыли</option>
            <option value="margin">По рентабельности</option>
            <option value="product_name">По названию</option>
          </select>
        </div>
      </div>
    </div>

    <div v-if="error" class="error-box">{{ error }}</div>

    <!-- KPI карточки -->
    <div class="kpi-grid">
      <div class="kpi-card">
        <div class="kpi-label">Всего продано</div>
        <div class="kpi-value">{{ formatMoney(kpi.totalSold) }}</div>
        <div class="kpi-sub">{{ kpi.totalOrders }} заказов</div>
      </div>
      <div class="kpi-card">
        <div class="kpi-label">Средний чек</div>
        <div class="kpi-value">{{ formatMoney(kpi.avgCheck) }}</div>
        <div class="kpi-sub">{{ formatQty(kpi.totalQty) }} ед. всего</div>
      </div>
      <div class="kpi-card">
        <div class="kpi-label">Прибыль</div>
        <div class="kpi-value" :class="kpi.totalProfit >= 0 ? 'green' : 'red'">
          {{ formatMoney(kpi.totalProfit) }}
        </div>
        <div class="kpi-sub">рент. {{ kpi.avgMargin.toFixed(1) }}%</div>
      </div>
      <div class="kpi-card">
        <div class="kpi-label">Топ-товар</div>
        <div class="kpi-value small">{{ kpi.topProduct || '—' }}</div>
        <div class="kpi-sub">{{ kpi.topProductSum > 0 ? formatMoney(kpi.topProductSum) : 'нет продаж' }}</div>
      </div>
    </div>

    <!-- График по дням -->
    <div class="chart-card">
      <div class="chart-header">
        <span>Продажи по дням</span>
        <span class="muted">за {{ days }} дн.</span>
      </div>
      <div v-if="daily.length === 0" class="chart-empty muted">Нет данных за период</div>
      <div v-else class="chart">
        <div v-for="d in daily" :key="d.date" class="bar-wrap">
          <div class="bar-value">{{ formatMoneyShort(d.sold_sum) }}</div>
          <div class="bar" :style="{ height: barHeight(d.sold_sum) + '%' }" :title="`${d.date}: ${formatMoney(d.sold_sum)}`"></div>
          <div class="bar-label">{{ shortDate(d.date) }}</div>
        </div>
      </div>
    </div>

    <!-- Таблица по товарам -->
    <table class="ms-table">
      <thead>
        <tr>
          <th>Наименование</th>
          <th style="width:100px">Артикул</th>
          <th style="width:70px">Ед.</th>
          <th class="num" @click="sortBy('sold_qty')">Продано <span v-if="sortKey === 'sold_qty'">{{ sortDir === 'asc' ? '↑' : '↓' }}</span></th>
          <th class="num">Сумма</th>
          <th class="num">Средняя цена</th>
          <th class="num">Себестоимость</th>
          <th class="num">Прибыль</th>
          <th class="num">Рент.</th>
          <th class="num">Заказов</th>
          <th style="width:130px">Последняя продажа</th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="loading">
          <td colspan="11" class="muted" style="text-align:center;padding:24px">Загрузка...</td>
        </tr>
        <tr v-else-if="filtered.length === 0">
          <td colspan="11" class="muted" style="text-align:center;padding:24px">Нет продаж за период</td>
        </tr>
        <tr v-else v-for="r in paginated" :key="r.product_id">
          <td class="link">{{ r.product_name }}</td>
          <td class="muted mono">{{ r.sku || '—' }}</td>
          <td>{{ r.unit_short || '—' }}</td>
          <td class="num">{{ formatQty(r.sold_qty) }}</td>
          <td class="num"><strong>{{ formatMoney(r.sold_sum) }}</strong></td>
          <td class="num muted">{{ formatMoney(r.avg_price) }}</td>
          <td class="num muted">{{ formatMoney(r.cost_sum) }}</td>
          <td class="num" :class="r.profit >= 0 ? 'diff-plus' : 'diff-minus'">
            {{ formatMoney(r.profit) }}
          </td>
          <td class="num">{{ r.margin.toFixed(1) }}%</td>
          <td class="num muted">{{ r.orders_count }}</td>
          <td class="muted">{{ r.last_sold_at ? formatShortDate(r.last_sold_at) : '—' }}</td>
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
        <span>{{ formatQty(totalQtyFiltered) }} ед.</span>
        <span>{{ formatMoney(totalSumFiltered) }}</span>
        <span :class="totalProfitFiltered >= 0 ? 'diff-plus' : 'diff-minus'">
          {{ formatMoney(totalProfitFiltered) }}
        </span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { listProducts, type Product } from '../api/products'
import { salesAnalytics, salesDaily, type SalesAnalyticsRow, type SalesDailyRow } from '../api/analytics'
import { apiErrorMessage } from '../api/client'

interface Row extends SalesAnalyticsRow {
  product_name: string
  sku?: string
  unit_short: string
  cost_sum: number
  profit: number
  margin: number
  avg_price: number
}

const loading = ref(false)
const error = ref<string | null>(null)
const days = ref(14)
const showFilter = ref(false)
const page = ref(1)
const perPage = 100

const filters = reactive({
  search: '',
  onlySold: 'yes' as 'yes' | 'no',
})

const sortKey = ref<'product_name' | 'sold_qty' | 'sold_sum' | 'profit' | 'margin'>('sold_sum')
const sortDir = ref<'asc' | 'desc'>('desc')

const sortByField = computed({
  get: () => sortKey.value,
  set: (v) => { sortKey.value = v as any; page.value = 1 },
})

const rows = ref<Row[]>([])
const daily = ref<SalesDailyRow[]>([])

const kpi = computed(() => {
  const sold = rows.value.filter((r) => r.sold_qty > 0)
  const totalSold = sold.reduce((s, r) => s + r.sold_sum, 0)
  const totalQty = sold.reduce((s, r) => s + r.sold_qty, 0)
  const totalOrders = sold.reduce((s, r) => s + r.orders_count, 0)
  const totalProfit = sold.reduce((s, r) => s + r.profit, 0)
  const avgCheck = sold.length > 0 ? totalSold / Math.max(1, totalOrders) : 0
  const avgMargin = totalSold > 0 ? (totalProfit / totalSold) * 100 : 0
  const top = [...sold].sort((a, b) => b.sold_sum - a.sold_sum)[0]
  return {
    totalSold,
    totalQty,
    totalOrders,
    totalProfit,
    avgCheck,
    avgMargin,
    topProduct: top?.product_name ?? '',
    topProductSum: top?.sold_sum ?? 0,
  }
})

function formatQty(n: number) {
  if (n === 0) return '0'
  return n.toLocaleString('ru-RU', { maximumFractionDigits: 3 })
}
function formatMoney(n: number) {
  const sign = n < 0 ? '−' : ''
  return sign + Math.abs(n).toLocaleString('ru-RU', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}
function formatMoneyShort(n: number) {
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + 'M'
  if (n >= 1_000) return (n / 1_000).toFixed(0) + 'K'
  return n.toFixed(0)
}
function formatShortDate(s: string) {
  return new Date(s).toLocaleDateString('ru-RU')
}
function shortDate(s: string) {
  const d = new Date(s)
  return d.toLocaleDateString('ru-RU', { day: '2-digit', month: '2-digit' })
}
function sortBy(k: typeof sortKey.value) {
  if (sortKey.value === k) sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  else { sortKey.value = k; sortDir.value = 'desc' }
  page.value = 1
}

const maxDailySum = computed(() => Math.max(1, ...daily.value.map((d) => d.sold_sum)))
function barHeight(v: number) { return Math.max(2, (v / maxDailySum.value) * 100) }

const filtered = computed(() => {
  let r = rows.value
  if (filters.onlySold === 'yes') r = r.filter((x) => x.sold_qty > 0)
  if (filters.search) {
    const q = filters.search.toLowerCase()
    r = r.filter((x) =>
      x.product_name.toLowerCase().includes(q) ||
      (x.sku ?? '').toLowerCase().includes(q)
    )
  }
  return [...r].sort((a, b) => {
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

const totalQtyFiltered = computed(() => filtered.value.reduce((s, r) => s + r.sold_qty, 0))
const totalSumFiltered = computed(() => filtered.value.reduce((s, r) => s + r.sold_sum, 0))
const totalProfitFiltered = computed(() => filtered.value.reduce((s, r) => s + r.profit, 0))

async function load() {
  loading.value = true
  error.value = null
  try {
    const [prods, sales, dly] = await Promise.all([
      listProducts(true),
      salesAnalytics(days.value),
      salesDaily(days.value),
    ])

    const productsMap = new Map<string, Product>()
    for (const p of prods) productsMap.set(p.id, p)

    const out: Row[] = sales.items.map((s) => {
      const p = productsMap.get(s.product_id)
      const cost = (p as any)?.cost_price ?? 0
      const costSum = s.sold_qty * cost
      const profit = s.sold_sum - costSum
      const margin = s.sold_sum > 0 ? (profit / s.sold_sum) * 100 : 0
      const avgPrice = s.sold_qty > 0 ? s.sold_sum / s.sold_qty : 0
      return {
        ...s,
        product_name: p?.name ?? '—',
        sku: p?.sku ?? undefined,
        unit_short: '',
        cost_sum: costSum,
        profit,
        margin,
        avg_price: avgPrice,
      }
    })

    rows.value = out
    daily.value = dly.items
  } catch (e) {
    error.value = apiErrorMessage(e)
  } finally {
    loading.value = false
  }
}

function clearFilters() {
  filters.search = ''
  filters.onlySold = 'yes'
  page.value = 1
}

function exportCsv() {
  const header = ['Наименование', 'Артикул', 'Продано', 'Сумма', 'Средняя цена', 'Себестоимость', 'Прибыль', 'Рентабельность %', 'Заказов', 'Последняя продажа']
  const lines = [header.join(';')]
  for (const r of filtered.value) {
    lines.push([
      csvEscape(r.product_name),
      csvEscape(r.sku ?? ''),
      r.sold_qty.toString(),
      r.sold_sum.toFixed(2),
      r.avg_price.toFixed(2),
      r.cost_sum.toFixed(2),
      r.profit.toFixed(2),
      r.margin.toFixed(1),
      r.orders_count.toString(),
      r.last_sold_at ? formatShortDate(r.last_sold_at) : '',
    ].join(';'))
  }
  const blob = new Blob(['\uFEFF' + lines.join('\r\n')], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `sales-analytics-${days.value}d-${new Date().toISOString().slice(0,10)}.csv`
  a.click()
  URL.revokeObjectURL(url)
}
function csvEscape(s: string) {
  if (s.includes(';') || s.includes('"') || s.includes('\n')) return '"' + s.replace(/"/g, '""') + '"'
  return s
}

onMounted(load)
</script>

<style scoped>
.days-switch { display: inline-flex; align-items: center; gap: 6px; font-size: 13px; color: #57606a; }
.days-select {
  padding: 5px 8px;
  border: 1px solid #d0d7de;
  border-radius: 4px;
  background: #fff;
  color: #1f2328;
  font-size: 13px;
}
.lbl { white-space: nowrap; }
.mono { font-family: monospace; font-size: 12px; }
.diff-plus  { color: #1a7f37; font-weight: 600; }
.diff-minus { color: #cf222e; font-weight: 600; }
.green { color: #1a7f37; }
.red { color: #cf222e; }

.kpi-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
  margin-bottom: 16px;
}
.kpi-card {
  background: #fff;
  border: 1px solid #d8dee4;
  border-radius: 6px;
  padding: 14px 16px;
}
.kpi-label { font-size: 12px; color: #57606a; text-transform: uppercase; letter-spacing: 0.5px; }
.kpi-value { font-size: 24px; font-weight: 700; margin-top: 6px; color: #1f2328; }
.kpi-value.small { font-size: 16px; }
.kpi-sub { font-size: 12px; color: #8c959f; margin-top: 4px; }

.chart-card {
  background: #fff;
  border: 1px solid #d8dee4;
  border-radius: 6px;
  padding: 16px;
  margin-bottom: 16px;
}
.chart-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; font-weight: 600; }
.chart-empty { padding: 32px; text-align: center; }
.chart {
  display: flex;
  align-items: flex-end;
  gap: 4px;
  height: 180px;
  padding: 8px 0;
  overflow-x: auto;
}
.bar-wrap {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: flex-end;
  min-width: 28px;
  height: 100%;
  flex-shrink: 0;
}
.bar-value {
  font-size: 10px;
  color: #8c959f;
  margin-bottom: 4px;
  white-space: nowrap;
  font-variant-numeric: tabular-nums;
}
.bar {
  width: 20px;
  background: linear-gradient(180deg, #4a7cc7 0%, #2c5d9c 100%);
  border-radius: 3px 3px 0 0;
  transition: background 0.15s;
}
.bar:hover { background: linear-gradient(180deg, #2c5d9c 0%, #1e4070 100%); }
.bar-label { font-size: 10px; color: #57606a; margin-top: 4px; white-space: nowrap; }
</style>