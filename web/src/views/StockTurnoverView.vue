<template>
  <div>
    <div class="ms-title">
      <button class="ms-help" title="Справка"><MsIcon name="help" :size="14" /></button>
      <span>Обороты</span>
      <button class="ms-refresh" @click="load" title="Обновить"><MsIcon name="refresh" :size="14" /></button>
    </div>

    <div class="ms-toolbar">
      <MsButton icon="filter" :class="{ active: showFilter }" @click="showFilter = !showFilter">Фильтр</MsButton>
      <div class="print-wrap">
        <MsButton icon="print" @click="showPrintMenu = !showPrintMenu">Печать ▾</MsButton>
        <div v-if="showPrintMenu" class="print-menu" @mouseleave="showPrintMenu = false">
          <button @click="onPrint('main')">Обороты</button>
          <button disabled title="В разработке">Обороты по партиям</button>
          <button disabled title="В разработке">Обороты по товарам</button>
          <div class="print-sep"></div>
          <button disabled>Настроить...</button>
        </div>
      </div>
    </div>

    <div v-if="showFilter" class="ms-filter">
      <div class="filter-actions">
        <MsButton variant="green" @click="load">Найти</MsButton>
        <MsButton @click="clearFilters">Очистить</MsButton>
      </div>
      <div class="filter-grid">
        <div class="filter-field">
          <label class="filter-label"><span class="dot"></span>Период</label>
          <div class="filter-dates">
            <input v-model="filters.from" type="date" />
            <span class="dash">—</span>
            <input v-model="filters.to" type="date" />
          </div>
        </div>
        <div class="filter-field"><label class="filter-label"><span class="dot"></span>Показывать товары без движения</label>
          <select><option>Нет</option><option>Да</option></select>
        </div>
        <div class="filter-field"><label class="filter-label"><span class="dot"></span>Показывать архивные</label>
          <select><option>Нет</option><option>Да</option></select>
        </div>
        <div class="filter-field"><label class="filter-label"><span class="dot"></span>Товар или группа</label>
          <select disabled><option>Все</option></select>
        </div>
        <div class="filter-field"><label class="filter-label"><span class="dot"></span>Склад</label>
          <select v-model="filters.warehouse_id">
            <option value="">—</option>
            <option v-for="w in warehouses" :key="w.id" :value="w.id">{{ w.name }}</option>
          </select>
        </div>
        <div class="filter-field"><label class="filter-label"><span class="dot"></span>Точка продаж</label>
          <input disabled placeholder="—" />
        </div>
        <div class="filter-field"><label class="filter-label"><span class="dot"></span>Контрагент</label>
          <select disabled><option>Все</option></select>
        </div>
        <div class="filter-field"><label class="filter-label"><span class="dot"></span>Группа контрагента</label>
          <select disabled><option>Все</option></select>
        </div>
        <div class="filter-field"><label class="filter-label"><span class="dot"></span>Договор</label>
          <select disabled><option>Все</option></select>
        </div>
        <div class="filter-field"><label class="filter-label"><span class="dot"></span>Поставщик</label>
          <select disabled><option>Все</option></select>
        </div>
        <div class="filter-field"><label class="filter-label"><span class="dot"></span>Организация</label>
          <select disabled><option>Все</option></select>
        </div>
        <div class="filter-field"><label class="filter-label"><span class="dot"></span>Тип документа</label>
          <select disabled><option>Все</option></select>
        </div>
      </div>
    </div>

    <div v-if="error" class="error-box">{{ error }}</div>

    <table class="ms-table turnover-table">
      <thead>
        <tr class="header-group">
          <th rowspan="2">Наименование</th>
          <th rowspan="2" style="width:100px">Код</th>
          <th rowspan="2" style="width:110px">Артикул</th>
          <th rowspan="2" style="width:70px">Ед. изм.</th>
          <th colspan="2" class="group-head">Приход / Расход</th>
        </tr>
        <tr class="header-cols">
          <th class="num" style="width:120px">Кол-во</th>
          <th class="num" style="width:120px">Кол-во</th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="loading"><td colspan="6" class="muted" style="text-align:center;padding:24px">Загрузка...</td></tr>
        <tr v-else-if="rows.length === 0"><td colspan="6" class="muted" style="text-align:center;padding:24px">Нет данных за период</td></tr>
        <tr v-for="r in rows" :key="r.product_id" class="row">
          <td class="link">{{ r.product_name || r.product_id.slice(0, 8) }}</td>
          <td class="muted mono">{{ r.product_id.slice(0, 8) }}</td>
          <td class="muted mono">{{ r.sku || '—' }}</td>
          <td class="muted">{{ r.unit_short || '—' }}</td>
          <td class="num">{{ fmt(r.income) }}</td>
          <td class="num">{{ fmt(r.outcome) }}</td>
        </tr>
      </tbody>
      <tfoot v-if="rows.length > 0">
        <tr>
          <td colspan="4" style="text-align:right"><b>Итого:</b></td>
          <td class="num"><b>{{ fmt(totals.income) }}</b></td>
          <td class="num"><b>{{ fmt(totals.outcome) }}</b></td>
        </tr>
      </tfoot>
    </table>

    <div class="ms-footer">
      <div class="ms-pager">
        <span>1–{{ rows.length }} из {{ rows.length }}</span>
      </div>
      <div class="ms-totals">
        <span>Всего: {{ rows.length }}</span>
      </div>
    </div>
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
const showPrintMenu = ref(false)

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
  const t = { income: 0, outcome: 0 }
  for (const r of rows.value) {
    t.income += r.income
    t.outcome += r.outcome
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

function onPrint(kind: 'main') {
  showPrintMenu.value = false
  if (kind === 'main') window.print()
}

onMounted(async () => {
  try { warehouses.value = await listWarehouses() } catch { /* ignore */ }
  await load()
})
</script>

<style scoped>
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
.filter-dates { display: flex; align-items: center; gap: 4px; }
.filter-dates input { flex: 1; min-width: 0; }
.dash { color: #8c959f; font-size: 12px; flex-shrink: 0; }
tr.clickable { cursor: pointer; }
.btn-link { background: none; border: none; color: #2c5d9c; cursor: pointer; padding: 0; font-size: 13px; }
.btn-link:hover { text-decoration: underline; }
.danger-text { color: #cf222e; }
.pill.printed { background: #4a90e2; color: #fff; padding: 2px 8px; border-radius: 3px; font-size: 11px; display: inline-block; }
.print-wrap { position: relative; display: inline-block; }
.print-menu {
  position: absolute; top: 32px; left: 0; z-index: 100;
  background: #fff; border: 1px solid #d0d7de; border-radius: 4px;
  box-shadow: 0 4px 16px rgba(0,0,0,0.12); min-width: 220px;
  padding: 4px 0;
}
.print-menu button {
  display: block; width: 100%; text-align: left;
  padding: 8px 14px; border: none; background: transparent;
  font-size: 13px; color: #1f2328; cursor: pointer;
}
.print-menu button:hover:not(:disabled) { background: #f6f8fa; }
.print-menu button:disabled { color: #b8c0c8; cursor: default; }
.print-sep { height: 1px; background: #eaeef2; margin: 4px 0; }
.turnover-table thead th { font-size: 12px; }
.turnover-table thead tr.header-group th { border-bottom: none; padding-bottom: 4px; }
.turnover-table thead tr.header-cols th { padding-top: 0; border-bottom: 1px solid #d8dee4; }
.turnover-table .group-head { text-align: center; color: #57606a; font-weight: 500; }
.turnover-table tbody tr.row td { font-size: 13px; }
.turnover-table .link { color: #2c5d9c; font-weight: 500; }
</style>