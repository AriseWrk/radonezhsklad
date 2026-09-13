<template>
  <div>
    <!-- Заголовок -->
    <div class="page-title-bar">
      <div class="page-title">
        <span>Остатки</span>
        <span class="refresh" @click="load" title="Обновить">↻</span>
      </div>
      <div class="page-actions">
        <div class="view-switch">
          <button :class="{ active: view === 'products' }" @click="view = 'products'">По товарам</button>
          <button :class="{ active: view === 'warehouses' }" @click="view = 'warehouses'">По складам</button>
        </div>
        <button class="btn" @click="toggleFilter">Фильтр</button>
        <button class="btn" @click="print">Печать</button>
      </div>
    </div>

    <!-- Панель фильтров -->
    <div v-if="showFilter" class="filter-panel">
      <div class="filter-row">
        <div class="filter-actions">
          <button class="btn-find" @click="load">Найти</button>
          <button class="btn-clear" @click="clearFilters">Очистить</button>
        </div>
        <div class="filter-field">
          <label>Остаток</label>
          <select v-model="filters.qtyMode">
            <option value="any">Любой</option>
            <option value="positive">Положительный</option>
            <option value="zero">Нулевой</option>
            <option value="negative">Отрицательный</option>
          </select>
        </div>
        <div class="filter-field">
          <label>Доступно</label>
          <select v-model="filters.availMode">
            <option value="any">Любое</option>
            <option value="nonZero">Ненулевое</option>
            <option value="zero">Нулевое</option>
          </select>
        </div>
        <div class="filter-field">
          <label>Только с ожиданием</label>
          <select v-model="filters.hasIncoming">
            <option value="no">Нет</option>
            <option value="yes">Да</option>
          </select>
        </div>
        <div class="filter-field">
          <label>Только с резервом</label>
          <select v-model="filters.hasReserve">
            <option value="no">Нет</option>
            <option value="yes">Да</option>
          </select>
        </div>
        <div class="filter-field">
          <label>Склад</label>
          <select v-model="filters.warehouseId">
            <option value="">Все</option>
            <option v-for="w in warehouses" :key="w.id" :value="w.id">{{ w.name }}</option>
          </select>
        </div>
      </div>
      <div class="filter-row wide">
        <div class="filter-field">
          <label>Товар или группа</label>
          <input v-model="filters.search" placeholder="Начните вводить название или код" />
        </div>
        <div class="filter-field">
          <label>Дней на складе от</label>
          <input v-model.number="filters.daysFrom" type="number" min="0" />
        </div>
        <div class="filter-field">
          <label>Дней на складе до</label>
          <input v-model.number="filters.daysTo" type="number" min="0" />
        </div>
        <div class="filter-field">
          <label>Артикул</label>
          <input v-model="filters.sku" placeholder="SKU" />
        </div>
        <div class="filter-field">
          <label>Поставщик</label>
          <input disabled placeholder="—" />
        </div>
      </div>
    </div>

    <div v-if="error" class="error-box">{{ error }}</div>

    <!-- Таблица -->
    <table class="ms-table">
      <thead>
        <tr>
          <th @click="sortBy('product_name')">Наименование <span v-if="sortKey === 'product_name'">{{ sortDir === 'asc' ? '↑' : '↓' }}</span></th>
          <th>Код</th>
          <th>Артикул</th>
          <th class="num" @click="sortBy('quantity')">Остаток <span v-if="sortKey === 'quantity'">{{ sortDir === 'asc' ? '↑' : '↓' }}</span></th>
          <th class="num">Неснижаемый</th>
          <th class="num">Резерв</th>
          <th class="num">Ожидание</th>
          <th class="num">Доступно</th>
          <th>Ед.</th>
          <th class="num">Дней на складе</th>
          <th class="num">Себестоимость</th>
          <th class="num">Сумма себест.</th>
          <th class="num">Цена продажи</th>
          <th class="num">Сумма продажи</th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="loading"><td colspan="14" class="muted" style="text-align:center;padding:24px">Загрузка...</td></tr>
        <tr v-else-if="filteredRows.length === 0"><td colspan="14" class="muted" style="text-align:center;padding:24px">Нет данных</td></tr>
        <tr v-for="r in paginatedRows" :key="r.product_id">
          <td class="link">{{ r.product_name }}</td>
          <td class="muted">{{ r.product_id.slice(0, 8) }}</td>
          <td class="muted">{{ r.sku || '—' }}</td>
          <td class="num">{{ formatQty(r.quantity) }}</td>
          <td class="num">{{ formatQty(r.min_stock) }}</td>
          <td class="num">{{ formatQty(r.reserve) }}</td>
          <td class="num">{{ formatQty(r.incoming) }}</td>
          <td class="num">
            {{ formatQty(r.available) }}
            <span class="cart-icon" title="В тележку">🛒</span>
          </td>
          <td>{{ r.unit_short || '—' }}</td>
          <td class="num muted">{{ r.days_on_stock ?? '—' }}</td>
          <td class="num">{{ formatMoney(r.cost_price) }}</td>
          <td class="num">{{ formatMoney(r.cost_total) }}</td>
          <td class="num red">{{ formatMoney(r.sale_price) }}</td>
          <td class="num red">{{ formatMoney(r.sale_total) }}</td>
        </tr>
      </tbody>
    </table>

    <!-- Footer -->
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
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { listStockExtended, type StockExtendedResponse, type StockExtendedRow } from '../api/stock'
import { listWarehouses, type Warehouse } from '../api/warehouses'
import { apiErrorMessage } from '../api/client'

const loading = ref(false)
const error = ref<string | null>(null)
const data = ref<StockExtendedResponse | null>(null)
const warehouses = ref<Warehouse[]>([])
const view = ref<'products' | 'warehouses'>('products')
const showFilter = ref(true)
const page = ref(1)
const perPage = 100

const filters = reactive({
  qtyMode: 'any',
  availMode: 'any',
  hasIncoming: 'no',
  hasReserve: 'no',
  warehouseId: '',
  search: '',
  sku: '',
  daysFrom: null as number | null,
  daysTo: null as number | null,
})

const sortKey = ref<keyof StockExtendedRow>('product_name')
const sortDir = ref<'asc' | 'desc'>('asc')

function toggleFilter() { showFilter.value = !showFilter.value }

function clearFilters() {
  filters.qtyMode = 'any'
  filters.availMode = 'any'
  filters.hasIncoming = 'no'
  filters.hasReserve = 'no'
  filters.warehouseId = ''
  filters.search = ''
  filters.sku = ''
  filters.daysFrom = null
  filters.daysTo = null
  load()
}

function formatQty(n: number): string {
  if (n === 0) return '0'
  return n.toLocaleString('ru-RU', { maximumFractionDigits: 3 })
}
function formatMoney(n: number): string {
  return n.toLocaleString('ru-RU', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function sortBy(key: keyof StockExtendedRow) {
  if (sortKey.value === key) {
    sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortKey.value = key
    sortDir.value = 'asc'
  }
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
    case 'negative': rows = rows.filter((r) => r.quantity < 0); break
  }
  switch (filters.availMode) {
    case 'nonZero': rows = rows.filter((r) => r.available !== 0); break
    case 'zero':    rows = rows.filter((r) => r.available === 0); break
  }
  if (filters.hasIncoming === 'yes') rows = rows.filter((r) => r.incoming > 0)
  if (filters.hasReserve === 'yes')  rows = rows.filter((r) => r.reserve > 0)

  if (filters.daysFrom != null) rows = rows.filter((r) => (r.days_on_stock ?? 0) >= filters.daysFrom!)
  if (filters.daysTo   != null) rows = rows.filter((r) => (r.days_on_stock ?? 0) <= filters.daysTo!)

  const sorted = [...rows].sort((a, b) => {
    const av = a[sortKey.value]
    const bv = b[sortKey.value]
    if (typeof av === 'number' && typeof bv === 'number') {
      return sortDir.value === 'asc' ? av - bv : bv - av
    }
    const as = String(av ?? '').toLowerCase()
    const bs = String(bv ?? '').toLowerCase()
    return sortDir.value === 'asc' ? as.localeCompare(bs, 'ru') : bs.localeCompare(as, 'ru')
  })
  return sorted
})

const filteredTotals = computed(() => {
  const t = {
    quantity: 0, min_stock: 0, reserve: 0, incoming: 0, available: 0,
    cost_total: 0, sale_total: 0,
  }
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

function print() { window.print() }

onMounted(async () => {
  try {
    warehouses.value = await listWarehouses()
  } catch { /* ignore */ }
  await load()
})
</script>

<style scoped>
/* всё в style.css глобально */
</style>