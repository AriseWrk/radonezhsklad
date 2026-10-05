<template>
  <div class="page">
    <div class="ms-title">
      <button class="ms-help" title="Справка"><MsIcon name="help" :size="14" /></button>
      <span>Обороты</span>
      <button class="ms-refresh" @click="load" title="Обновить"><MsIcon name="refresh" :size="14" /></button>
    </div>

    <div class="ms-toolbar">
      <MsButton icon="filter" @click="showFilter = !showFilter">Фильтр</MsButton>
      <MsButton icon="print" @click="print">Печать</MsButton>
    </div>

    <div v-if="showFilter" class="ms-filter">
      <div class="filter-grid">
        <div class="filter-actions">
          <MsButton variant="green" @click="load">Найти</MsButton>
          <MsButton @click="clearFilters">Очистить</MsButton>
        </div>

        <div class="filter-field">
          <label class="filter-label"><span class="dot"></span>Период:</label>
          <div class="date-range">
            <input v-model="filters.from" type="date" />
            <span class="dash">—</span>
            <input v-model="filters.to" type="date" />
          </div>
        </div>

        <div class="filter-field">
          <label class="filter-label"><span class="dot"></span>Склад</label>
          <select v-model="filters.warehouse_id">
            <option value="">—</option>
            <option v-for="w in warehouses" :key="w.id" :value="w.id">{{ w.name }}</option>
          </select>
        </div>
      </div>
    </div>

    <div v-if="error" class="error-box">{{ error }}</div>

    <table class="ms-table2">
      <thead>
        <tr>
          <th>Наименование</th>
          <th style="width:110px">Код</th>
          <th style="width:110px">Артикул</th>
          <th style="width:70px">Ед. изм.</th>
          <th class="num">Начальный остаток</th>
          <th class="num">Приход</th>
          <th class="num">Расход</th>
          <th class="num">Конечный остаток</th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="loading">
          <td colspan="8" class="muted" style="text-align:center;padding:24px">Загрузка...</td>
        </tr>
        <tr v-else-if="rows.length === 0">
          <td colspan="8" class="muted" style="text-align:center;padding:24px">Нет данных за период</td>
        </tr>
        <tr v-for="r in rows" :key="r.product_id">
          <td>{{ r.product_name || r.product_id.slice(0, 8) }}</td>
          <td class="muted mono">{{ r.product_id.slice(0, 8) }}</td>
          <td class="muted mono">{{ r.sku || '—' }}</td>
          <td class="muted">{{ r.unit_short || '—' }}</td>
          <td class="num">{{ fmt(r.opening) }}</td>
          <td class="num">{{ fmt(r.income) }}</td>
          <td class="num">{{ fmt(r.outcome) }}</td>
          <td class="num"><b>{{ fmt(r.closing) }}</b></td>
        </tr>
      </tbody>
      <tfoot v-if="rows.length > 0">
        <tr>
          <td colspan="4" style="text-align:right"><b>Итого:</b></td>
          <td class="num"><b>{{ fmt(totals.opening) }}</b></td>
          <td class="num"><b>{{ fmt(totals.income) }}</b></td>
          <td class="num"><b>{{ fmt(totals.outcome) }}</b></td>
          <td class="num"><b>{{ fmt(totals.closing) }}</b></td>
        </tr>
      </tfoot>
    </table>
    <div v-if="rows.length > 0" class="ms-footer-count">{{ rows.length }} записей</div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { turnoverReport, type TurnoverRow } from '../api/turnover'
import { listWarehouses, type Warehouse } from '../api/warehouses'
import { apiErrorMessage } from '../api/client'
import MsButton from '../components/MsButton.vue'
import MsIcon from '../components/MsIcon.vue'

const rows = ref<TurnoverRow[]>([])
const warehouses = ref<Warehouse[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const showFilter = ref(true)

function isoDaysAgo(n: number) {
  const d = new Date()
  d.setDate(d.getDate() - n)
  return d.toISOString().slice(0, 10)
}
const filters = reactive({
  from: isoDaysAgo(30),
  to: new Date().toISOString().slice(0, 10),
  warehouse_id: '',
})

function fmt(n: number) {
  return (n ?? 0).toLocaleString('ru-RU', { maximumFractionDigits: 3 })
}

const totals = computed(() => {
  const t = { opening: 0, income: 0, outcome: 0, closing: 0 }
  for (const r of rows.value) {
    t.opening += r.opening
    t.income += r.income
    t.outcome += r.outcome
    t.closing += r.closing
  }
  return t
})

async function load() {
  loading.value = true
  error.value = null
  try {
    const data = await turnoverReport({
      warehouse_id: filters.warehouse_id || undefined,
      from: filters.from || undefined,
      to: filters.to || undefined,
    })
    rows.value = data.items
  } catch (e) {
    error.value = apiErrorMessage(e)
  } finally {
    loading.value = false
  }
}

function clearFilters() {
  filters.from = isoDaysAgo(30)
  filters.to = new Date().toISOString().slice(0, 10)
  filters.warehouse_id = ''
  load()
}

function print() { window.print() }

onMounted(async () => {
  try { warehouses.value = await listWarehouses() } catch { /* ignore */ }
  await load()
})
</script>

<style scoped>
.page { padding: 12px 16px; }

.ms-toolbar { display: flex; align-items: center; gap: 6px; margin-bottom: 12px; flex-wrap: wrap; }

.ms-filter {
  background: #eef1f5; border: 1px solid #d8dee4;
  border-radius: 4px; padding: 10px 14px 12px; margin-bottom: 12px;
}

.filter-grid {
  display: grid;
  grid-template-columns:
    200px
    minmax(220px, 1.4fr)
    minmax(200px, 1fr)
    minmax(160px, 1fr)
    minmax(160px, 1fr)
    minmax(160px, 1fr);
  gap: 10px 14px;
  align-items: start;
}

.filter-actions {
  display: flex; align-items: center; gap: 6px;
  grid-column: 1;
  padding-top: 18px;
}

.filter-field { display: flex; flex-direction: column; gap: 3px; }
.filter-field:nth-of-type(1) { grid-column: 2; }
.filter-field:nth-of-type(2) { grid-column: 3 / span 2; }

.filter-label {
  font-size: 12px; color: #2c5d9c;
  display: flex; align-items: center; gap: 5px;
}
.filter-label .dot {
  width: 6px; height: 6px; border-radius: 50%;
  background: #2c5d9c; flex-shrink: 0;
}

.date-range { display: flex; align-items: center; gap: 6px; }
.date-range .dash { color: #57606a; }
.date-range input { flex: 1 1 0; min-width: 0; }

.filter-field input, .filter-field select {
  height: 28px; padding: 0 8px; font-size: 12px;
  border: 1px solid #d0d7de; border-radius: 3px;
  background: #fff; color: #1f2328;
}

.num { text-align: right; }
.muted { color: #57606a; }
.mono { font-family: ui-monospace, SFMono-Regular, Menlo, monospace; }

.error-box {
  background: #ffebe9; border: 1px solid #ff8182; border-radius: 4px;
  padding: 8px 12px; margin-bottom: 12px; color: #82071e; font-size: 13px;
}

.ms-footer-count {
  font-size: 12px; color: #57606a;
  padding: 8px 4px 0;
}
</style>