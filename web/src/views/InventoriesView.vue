<template>
  <div class="page">
    <div class="ms-title">
      <button class="ms-help" title="Справка"><MsIcon name="help" :size="14" /></button>
      <span>Инвентаризации</span>
      <button class="ms-refresh" @click="load" title="Обновить"><MsIcon name="refresh" :size="14" /></button>
    </div>

    <div class="ms-toolbar">
      <router-link to="/inventory" class="ms-link-btn"><MsIcon name="plus" :size="16" /> Инвентаризация</router-link>
      <MsButton icon="filter" @click="showFilter = !showFilter">Фильтр</MsButton>
      <input v-model="search" class="ms-input" placeholder="Номер или комментарий" />
      <div class="ms-counter">{{ filteredRows.length }}</div>
      <MsButton icon="print">Печать</MsButton>
      <MsButton variant="icon" icon="gear" title="Настройки" />
    </div>

    <div v-if="showFilter" class="ms-filter">
      <div class="filter-actions">
        <MsButton variant="green" @click="load">Найти</MsButton>
        <MsButton @click="clearFilters">Очистить</MsButton>
      </div>
      <div class="filter-grid">
        <div class="filter-field">
          <label class="filter-label"><span class="dot"></span>Склад</label>
          <select v-model="filters.warehouse_id" @change="load">
            <option value="">—</option>
            <option v-for="w in warehouses" :key="w.id" :value="w.id">{{ w.name }}</option>
          </select>
        </div>
        <div class="filter-field">
          <label class="filter-label"><span class="dot"></span>Период с</label>
          <input v-model="filters.from" type="date" @change="load" />
        </div>
        <div class="filter-field">
          <label class="filter-label"><span class="dot"></span>Период по</label>
          <input v-model="filters.to" type="date" @change="load" />
        </div>
      </div>
    </div>

    <div v-if="error" class="error-box">{{ error }}</div>

    <table class="ms-table2">
      <thead>
        <tr>
          <th class="col-num">№</th>
          <th class="col-num">Номер</th>
          <th class="col-date">Дата</th>
          <th>Склад</th>
          <th class="col-num-right">Позиций</th>
          <th class="col-num-right">Сумма</th>
          <th>Комментарий</th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="loading"><td colspan="7" class="empty">Загрузка...</td></tr>
        <tr v-else-if="filteredRows.length === 0"><td colspan="7" class="empty">Нет данных</td></tr>
        <tr v-else v-for="(r, idx) in filteredRows" :key="r.id" class="row" @click="open(r)">
          <td class="col-num">{{ idx + 1 }}</td>
          <td class="col-num"><span class="link">{{ r.number }}</span></td>
          <td class="col-date">{{ formatDate(r.doc_date) }}</td>
          <td>{{ r.warehouse_name || '—' }}</td>
          <td class="col-num-right">{{ r.items_count }}</td>
          <td class="col-num-right"><b>{{ formatNum(r.total) }}</b></td>
          <td class="comment">{{ r.comment || '' }}</td>
        </tr>
      </tbody>
    </table>

    <div class="ms-footer2">
      <div class="pager">
        <button :disabled="page <= 1" @click="page--">‹</button>
        <span class="range">Стр. {{ page }} / {{ totalPages }}</span>
        <button :disabled="page >= totalPages" @click="page++">›</button>
      </div>
      <div class="totals-inline"><span>Всего: {{ filteredRows.length }}</span></div>
    </div>

    <div v-if="selected" class="modal-overlay" @click.self="selected = null">
      <div class="modal-card" style="max-width:900px">
        <div class="modal-head">
          <div>Инвентаризация {{ selected.number }}</div>
          <button class="btn-close" @click="selected = null">×</button>
        </div>
        <div class="modal-body">
          <p><b>Дата:</b> {{ formatDate(selected.doc_date) }}</p>
          <p><b>Склад:</b> {{ selected.warehouse_name || '—' }}</p>
          <p><b>Комментарий:</b> {{ selected.comment || '—' }}</p>
          <table class="ms-table2" style="margin-top:12px">
            <thead>
              <tr>
                <th>Товар</th>
                <th style="width:80px">Артикул</th>
                <th class="col-num-right">Учёт</th>
                <th class="col-num-right">Факт</th>
                <th class="col-num-right">Расхождение</th>
                <th class="col-num-right">Цена</th>
                <th class="col-num-right">Сумма расхождения</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="it in selected.items || []" :key="it.id">
                <td>{{ it.product_name || it.product_id.slice(0, 8) + '…' }}</td>
                <td class="muted mono">{{ it.product_sku || '—' }}</td>
                <td class="col-num-right">{{ formatNum(it.calculated_quantity) }}</td>
                <td class="col-num-right">{{ formatNum(it.quantity) }}</td>
                <td class="col-num-right" :class="{ neg: it.correction_amount < 0 }">{{ formatNum(it.correction_amount) }}</td>
                <td class="col-num-right">{{ formatNum(it.price) }}</td>
                <td class="col-num-right">{{ formatNum(it.correction_sum) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { listInventories, type Inventory } from '../api/inventories'
import { listWarehouses, type Warehouse } from '../api/warehouses'
import { apiErrorMessage } from '../api/client'
import MsButton from '../components/MsButton.vue'
import MsIcon from '../components/MsIcon.vue'

const router = useRouter()
const rows = ref<Inventory[]>([])
const warehouses = ref<Warehouse[]>([])
const loading = ref(false)
const error = ref('')
const showFilter = ref(false)
const selected = ref<Inventory | null>(null)
const page = ref(1)
const pageSize = 50
const search = ref('')

const filters = reactive({
  warehouse_id: '',
  from: '',
  to: '',
})

const filteredRows = computed(() => {
  const q = search.value.trim().toLowerCase()
  if (!q) return rows.value
  return rows.value.filter((r) =>
    (r.number ?? '').toLowerCase().includes(q) ||
    (r.comment ?? '').toLowerCase().includes(q)
  )
})
const totalPages = computed(() => Math.max(1, Math.ceil(filteredRows.value.length / pageSize)))

function formatDate(s: string) {
  if (!s) return '—'
  const d = new Date(s)
  return d.toLocaleDateString('ru-RU')
}
function formatNum(n: number | undefined) {
  if (n == null) return '—'
  return n.toLocaleString('ru-RU', { maximumFractionDigits: 3 })
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    rows.value = await listInventories({
      warehouse_id: filters.warehouse_id || undefined,
      from: filters.from || undefined,
      to: filters.to || undefined,
    })
    page.value = 1
  } catch (e) {
    error.value = apiErrorMessage(e)
  } finally {
    loading.value = false
  }
}

function clearFilters() {
  filters.warehouse_id = ''
  filters.from = ''
  filters.to = ''
  load()
}

function open(r: Inventory) {
  router.push('/inventories/' + r.id)
}

onMounted(async () => {
  try { warehouses.value = await listWarehouses() } catch { /* ignore */ }
  load()
})
</script>

<style scoped>
.page { font-size: 13px; }
.ms-title { display: flex; align-items: center; gap: 8px; font-size: 20px; font-weight: 600; color: #1f2328; margin-bottom: 12px; }
.ms-help { width: 20px; height: 20px; border-radius: 50%; border: 1px solid #b8c0c8; background: transparent; color: #57606a; cursor: pointer; display: inline-flex; align-items: center; justify-content: center; padding: 0; }
.ms-help:hover { background: #f0f2f5; }
.ms-refresh { width: 24px; height: 24px; padding: 0; display: inline-flex; align-items: center; justify-content: center; border: none; background: transparent; color: #57606a; cursor: pointer; }
.ms-refresh:hover { color: #2c5d9c; }
.ms-toolbar { display: flex; align-items: center; gap: 6px; margin-bottom: 12px; flex-wrap: wrap; }
.ms-link-btn { display: inline-flex; align-items: center; gap: 6px; padding: 0 12px; height: 30px; font-size: 13px; border: 1px solid #2c5d9c; background: #2c5d9c; color: #fff; border-radius: 4px; cursor: pointer; text-decoration: none; }
.ms-link-btn:hover { background: #234a7d; text-decoration: none; color: #fff; }
.ms-input { flex: 1; max-width: 320px; height: 30px; padding: 0 10px; font-size: 13px; border: 1px solid #d0d7de; border-radius: 4px; background: #fff; color: #1f2328; }
.ms-input::placeholder { color: #8c959f; }
.ms-counter { min-width: 40px; height: 30px; padding: 0 10px; display: inline-flex; align-items: center; justify-content: center; border: 1px solid #d0d7de; border-radius: 4px; background: #fff; color: #8c959f; font-variant-numeric: tabular-nums; font-size: 13px; }
.ms-filter { background: #eef1f5; border: 1px solid #d8dee4; border-radius: 4px; padding: 10px 14px 12px; margin-bottom: 12px; }
.filter-actions { display: flex; align-items: center; gap: 6px; margin-bottom: 10px; }
.filter-grid { display: grid; grid-template-columns: repeat(6, minmax(0, 1fr)); gap: 10px 14px; }
.filter-field { display: flex; flex-direction: column; gap: 3px; }
.filter-label { font-size: 12px; color: #57606a; display: flex; align-items: center; gap: 5px; }
.filter-label .dot { width: 8px; height: 8px; border-radius: 50%; background: #2c5d9c; flex-shrink: 0; }
.filter-field input, .filter-field select { height: 28px; padding: 0 8px; font-size: 12px; border: 1px solid #d0d7de; border-radius: 3px; background: #fff; color: #1f2328; }

.ms-table2 { width: 100%; border-collapse: collapse; background: #fff; font-size: 13px; }
.ms-table2 thead th { background: #fff; color: #2c5d9c; font-weight: 500; padding: 8px 10px; text-align: left; border-bottom: 1px solid #d8dee4; white-space: nowrap; font-size: 12px; }
.ms-table2 tbody td { padding: 7px 10px; border-bottom: 1px solid #eaeef2; vertical-align: middle; }
.ms-table2 tbody tr.row { cursor: pointer; }
.ms-table2 tbody tr.row:hover { background: #f6f8fa; }
.ms-table2 .col-num { width: 80px; }
.ms-table2 .col-date { width: 140px; color: #57606a; }
.ms-table2 .col-num-right { text-align: right; font-variant-numeric: tabular-nums; }
.ms-table2 .link { color: #2c5d9c; font-weight: 500; }
.ms-table2 .comment { color: #8c959f; max-width: 400px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.ms-table2 .empty { text-align: center; padding: 24px; color: #8c959f; }

.ms-footer2 { display: flex; align-items: center; justify-content: space-between; padding: 6px 10px; font-size: 12px; color: #57606a; background: #fff; border-top: 1px solid #eaeef2; }
.pager { display: flex; align-items: center; gap: 4px; }
.pager button { width: 22px; height: 22px; padding: 0; border: 1px solid #d0d7de; background: #fff; border-radius: 3px; cursor: pointer; font-size: 12px; color: #1f2328; }
.pager button:disabled { opacity: 0.4; cursor: default; }
.pager .range { margin: 0 6px; font-variant-numeric: tabular-nums; }
.totals-inline { font-variant-numeric: tabular-nums; }

.modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.45); display: flex; align-items: center; justify-content: center; z-index: 1000; }
.modal-card { background: #fff; border-radius: 8px; box-shadow: 0 10px 40px rgba(0,0,0,0.25); max-height: 90vh; width: 100%; max-width: 900px; display: flex; flex-direction: column; overflow: hidden; }
.modal-head { display: flex; justify-content: space-between; align-items: center; padding: 12px 18px; border-bottom: 1px solid #e5e7eb; font-weight: 600; font-size: 16px; }
.modal-body { padding: 16px 18px; overflow-y: auto; }
.modal-body p { margin: 6px 0; }
.btn-close { background: none; border: none; font-size: 24px; line-height: 1; cursor: pointer; color: #666; padding: 0 6px; }
.btn-close:hover { color: #000; }
.neg { color: #c00; }
.mono { font-family: monospace; font-size: 12px; }
.muted { color: #8c959f; }
</style>