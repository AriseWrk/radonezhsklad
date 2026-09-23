<template>
  <div class="page">
    <div class="ms-title">
      <button class="ms-help" title="Справка"><MsIcon name="help" :size="14" /></button>
      <span>Аналитика продаж</span>
      <button class="ms-refresh" @click="load" title="Обновить"><MsIcon name="refresh" :size="14" /></button>
    </div>

    <div class="ms-toolbar">
      <MsButton icon="filter" @click="showFilter = !showFilter">Фильтр</MsButton>
      <input v-model="filters.search" class="ms-input" placeholder="Название, SKU" />
      <div class="ms-counter">{{ filtered.length }}</div>
      <select v-model.number="days" @change="load" class="ms-select">
        <option :value="7">7 дней</option>
        <option :value="14">14 дней</option>
        <option :value="30">30 дней</option>
        <option :value="90">90 дней</option>
        <option :value="180">180 дней</option>
        <option :value="365">365 дней</option>
      </select>
      <MsButton icon="excel" @click="exportCsv">Экспорт</MsButton>
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
          <label class="filter-label"><span class="dot"></span>Только проданные</label>
          <select v-model="filters.onlySold">
            <option value="yes">Да</option>
            <option value="no">Все</option>
          </select>
        </div>
        <div class="filter-field">
          <label class="filter-label"><span class="dot"></span>Сортировка</label>
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

    <table class="ms-table2">
      <thead>
        <tr>
          <th>Наименование</th>
          <th class="col-num">Артикул</th>
          <th class="col-num">Ед.</th>
          <th class="col-num-right" @click="sortBy('sold_qty')">Продано<span v-if="sortKey === 'sold_qty'" class="sort">{{ sortDir === 'asc' ? '▲' : '▼' }}</span></th>
          <th class="col-num-right">Сумма</th>
          <th class="col-num-right">Ср. цена</th>
          <th class="col-num-right">Себест.</th>
          <th class="col-num-right">Прибыль</th>
          <th class="col-num-right">Рент.</th>
          <th class="col-num-right">Заказов</th>
          <th class="col-date">Последняя продажа</th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="loading"><td colspan="11" class="empty">Загрузка...</td></tr>
        <tr v-else-if="filtered.length === 0"><td colspan="11" class="empty">Нет продаж за период</td></tr>
        <tr v-else v-for="r in paginated" :key="r.product_id" class="row">
          <td class="link">{{ r.product_name }}</td>
          <td class="muted mono">{{ r.sku || '—' }}</td>
          <td>{{ r.unit_short || '—' }}</td>
          <td class="col-num-right">{{ formatQty(r.sold_qty) }}</td>
          <td class="col-num-right"><b>{{ formatMoney(r.sold_sum) }}</b></td>
          <td class="col-num-right muted">{{ formatMoney(r.avg_price) }}</td>
          <td class="col-num-right muted">{{ formatMoney(r.cost_sum) }}</td>
          <td class="col-num-right" :class="r.profit >= 0 ? 'diff-plus' : 'diff-minus'">{{ formatMoney(r.profit) }}</td>
          <td class="col-num-right">{{ r.margin.toFixed(1) }}%</td>
          <td class="col-num-right muted">{{ r.orders_count }}</td>
          <td class="col-date muted">{{ r.last_sold_at ? formatShortDate(r.last_sold_at) : '—' }}</td>
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
        <span>{{ formatQty(totalQtyFiltered) }} ед.</span>
        <span><b>{{ formatMoney(totalSumFiltered) }}</b></span>
        <span :class="totalProfitFiltered >= 0 ? 'diff-plus' : 'diff-minus'">{{ formatMoney(totalProfitFiltered) }}</span>
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { listProducts, type Product } from '../api/products'
import { salesAnalytics, salesDaily, type SalesAnalyticsRow, type SalesDailyRow } from '../api/analytics'
import { apiErrorMessage } from '../api/client'
import MsButton from '../components/MsButton.vue'
import MsIcon from '../components/MsIcon.vue'

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
.ms-select { height: 30px; padding: 0 8px; font-size: 13px; border: 1px solid #d0d7de; border-radius: 4px; background: #fff; color: #1f2328; min-width: 120px; }
.ms-filter { background: #eef1f5; border: 1px solid #d8dee4; border-radius: 4px; padding: 10px 14px 12px; margin-bottom: 12px; }
.filter-actions { display: flex; align-items: center; gap: 6px; margin-bottom: 10px; }
.filter-grid { display: grid; grid-template-columns: repeat(6, minmax(0, 1fr)); gap: 10px 14px; }
.filter-field { display: flex; flex-direction: column; gap: 3px; }
.filter-label { font-size: 12px; color: #57606a; display: flex; align-items: center; gap: 5px; }
.filter-label .dot { width: 8px; height: 8px; border-radius: 50%; background: #2c5d9c; flex-shrink: 0; }
.filter-field input, .filter-field select { height: 28px; padding: 0 8px; font-size: 12px; border: 1px solid #d0d7de; border-radius: 3px; background: #fff; color: #1f2328; }

.kpi-grid { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 12px; margin-bottom: 14px; }
.kpi-card { background: #fff; border: 1px solid #eaeef2; border-radius: 4px; padding: 12px 16px; }
.kpi-label { font-size: 11px; color: #8c959f; text-transform: uppercase; letter-spacing: 0.3px; margin-bottom: 6px; }
.kpi-value { font-size: 20px; font-weight: 600; color: #1f2328; font-variant-numeric: tabular-nums; }
.kpi-value.small { font-size: 14px; }
.kpi-value.green { color: #1a7f37; }
.kpi-value.red { color: #cf222e; }
.kpi-sub { font-size: 12px; color: #8c959f; margin-top: 2px; }

.chart-card { background: #fff; border: 1px solid #eaeef2; border-radius: 4px; padding: 12px 16px; margin-bottom: 14px; }
.chart-header { display: flex; justify-content: space-between; align-items: center; font-size: 13px; font-weight: 500; color: #1f2328; margin-bottom: 10px; }
.chart-header .muted { color: #8c959f; font-weight: 400; }
.chart-empty { padding: 32px; text-align: center; font-size: 13px; }
.chart { display: flex; align-items: flex-end; gap: 4px; height: 140px; padding-bottom: 20px; position: relative; }
.bar-wrap { flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: flex-end; height: 100%; position: relative; }
.bar-value { position: absolute; top: -2px; font-size: 9px; color: #8c959f; transform: translateY(-100%); white-space: nowrap; }
.bar { width: 100%; background: #2c5d9c; border-radius: 2px 2px 0 0; min-height: 2px; transition: background 0.1s; }
.bar:hover { background: #234a7d; }
.bar-label { position: absolute; bottom: -18px; font-size: 10px; color: #8c959f; white-space: nowrap; }

.ms-table2 { width: 100%; border-collapse: collapse; background: #fff; font-size: 13px; }
.ms-table2 thead th { background: #fff; color: #2c5d9c; font-weight: 500; padding: 8px 10px; text-align: left; border-bottom: 1px solid #d8dee4; white-space: nowrap; font-size: 12px; cursor: pointer; user-select: none; }
.ms-table2 thead th:hover { background: #f6f8fa; }
.ms-table2 tbody td { padding: 7px 10px; border-bottom: 1px solid #eaeef2; vertical-align: middle; }
.ms-table2 tbody tr.row:hover { background: #f6f8fa; }
.ms-table2 .col-num { width: 80px; }
.ms-table2 .col-date { width: 140px; }
.ms-table2 .col-num-right { text-align: right; font-variant-numeric: tabular-nums; }
.ms-table2 .link { color: #2c5d9c; font-weight: 500; }
.ms-table2 .sort { font-size: 9px; margin-left: 3px; }
.ms-table2 .empty { text-align: center; padding: 24px; color: #8c959f; }
.mono { font-family: monospace; font-size: 12px; }
.muted { color: #8c959f; }
.diff-plus { color: #1a7f37; font-weight: 600; }
.diff-minus { color: #cf222e; font-weight: 600; }

.ms-footer2 { display: flex; align-items: center; justify-content: space-between; padding: 6px 10px; font-size: 12px; color: #57606a; background: #fff; border-top: 1px solid #eaeef2; }
.pager { display: flex; align-items: center; gap: 4px; }
.pager button { width: 22px; height: 22px; padding: 0; border: 1px solid #d0d7de; background: #fff; border-radius: 3px; cursor: pointer; font-size: 12px; color: #1f2328; }
.pager button:disabled { opacity: 0.4; cursor: default; }
.pager .range { margin: 0 6px; font-variant-numeric: tabular-nums; }
.totals-inline { display: flex; gap: 16px; font-variant-numeric: tabular-nums; }
</style>