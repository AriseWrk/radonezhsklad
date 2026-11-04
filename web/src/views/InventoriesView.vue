<template>
  <div>
    <div class="ms-title">
      <button class="ms-help" title="Справка"><MsIcon name="help" :size="14" /></button>
      <span>Инвентаризации</span>
      <button class="ms-refresh" @click="load" title="Обновить"><MsIcon name="refresh" :size="14" /></button>
    </div>

    <div class="ms-toolbar">
      <MsButton variant="primary" icon="plus" @click="openPickWarehouse">Инвентаризация</MsButton>
      <MsButton icon="filter" @click="showFilter = !showFilter">Фильтр</MsButton>
      <input v-model="search" class="ms-input" placeholder="Номер или комментарий" />
      <div class="ms-counter">{{ filtered.length }}</div>
      <select class="ms-select" :disabled="selected.size === 0">
        <option>Изменить</option>
        <option>Удалить</option>
      </select>
      <select class="ms-select" v-model="filterStatus">
        <option value="">Статус</option>
        <option value="draft">Черновик</option>
        <option value="posted">Проведена</option>
      </select>
      <div class="print-wrap">
        <MsButton icon="print" @click="showPrintMenu = !showPrintMenu">Печать ▾</MsButton>
        <div v-if="showPrintMenu" class="print-menu" @mouseleave="showPrintMenu = false">
          <button @click="onPrint('list')">Список инвентаризаций</button>
          <button @click="onPrint('inv')">Инвентаризация</button>
          <button @click="onPrint('inv3')">ИНВ-3</button>
          <button disabled title="В разработке">Сличительная ведомость</button>
          <button disabled title="В разработке">blank-inv-19 (1)</button>
          <button disabled title="В разработке">Список инвентаризаций по ячейкам</button>
          <button disabled title="В разработке">ИНВ-3 (инвентаризационная опись)</button>
          <button disabled title="В разработке">Инвентаризация с ячейками</button>
          <div class="print-sep"></div>
          <button disabled>Настроить...</button>
        </div>
      </div>
      <MsButton variant="icon" icon="gear" title="Настройки" />
    </div>

    <div v-if="showFilter" class="ms-filter">
      <div class="filter-actions">
        <MsButton variant="green" @click="load">Найти</MsButton>
        <MsButton @click="clearFilters">Очистить</MsButton>
      </div>
      <div class="filter-grid">
        <div class="filter-field"><label class="filter-label"><span class="dot"></span>Период</label>
          <div class="filter-dates"><input v-model="filters.from" type="date" /><span class="dash">—</span><input v-model="filters.to" type="date" /></div>
        </div>
        <div class="filter-field"><label class="filter-label"><span class="dot"></span>Товар или группа</label>
          <select disabled><option>Все</option></select>
        </div>
        <div class="filter-field"><label class="filter-label"><span class="dot"></span>Склад</label>
          <select v-model="filters.warehouse_id" @change="load">
            <option value="">—</option>
            <option v-for="w in warehouses" :key="w.id" :value="w.id">{{ w.name }}</option>
          </select>
        </div>
        <div class="filter-field"><label class="filter-label"><span class="dot"></span>Организация</label>
          <select disabled><option>Все</option></select>
        </div>
        <div class="filter-field"><label class="filter-label"><span class="dot"></span>Тип документа</label>
          <select><option>Все</option><option>Инвентаризация</option></select>
        </div>
        <div class="filter-field"><label class="filter-label"><span class="dot"></span>Показывать</label>
          <select><option>Только обычные</option><option>Все</option></select>
        </div>
        <div class="filter-field"><label class="filter-label"><span class="dot"></span>Статус</label>
          <select v-model="filterStatus">
            <option value="">Все</option>
            <option value="draft">Черновик</option>
            <option value="posted">Проведена</option>
          </select>
        </div>
        <div class="filter-field"><label class="filter-label"><span class="dot"></span>Напечатано</label>
          <select disabled><option>—</option></select>
        </div>
        <div class="filter-field"><label class="filter-label"><span class="dot"></span>Отправлено</label>
          <select disabled><option>—</option></select>
        </div>
        <div class="filter-field"><label class="filter-label"><span class="dot"></span>Владелец-сотрудник</label>
          <select disabled><option>—</option></select>
        </div>
        <div class="filter-field"><label class="filter-label"><span class="dot"></span>Владелец-отдел</label>
          <select disabled><option>—</option></select>
        </div>
        <div class="filter-field"><label class="filter-label"><span class="dot"></span>Общий доступ</label>
          <select disabled><option>—</option></select>
        </div>
        <div class="filter-field"><label class="filter-label"><span class="dot"></span>Когда изменён</label>
          <div class="filter-dates"><input type="date" disabled /><span class="dash">—</span><input type="date" disabled /></div>
        </div>
        <div class="filter-field"><label class="filter-label"><span class="dot"></span>Кто изменил</label>
          <select disabled><option>—</option></select>
        </div>
      </div>
    </div>

    <div v-if="error" class="error-box">{{ error }}</div>

    <table class="ms-table">
      <thead>
        <tr>
          <th class="chk-col"><input type="checkbox" :checked="allChecked" @change="toggleAll" /></th>
          <th style="width:130px">Тип документа</th>
          <th style="width:80px">№</th>
          <th style="width:140px">Время</th>
          <th>Со склада</th>
          <th>Организация</th>
          <th style="width:110px">Статус</th>
          <th style="width:110px">Отправлено</th>
          <th style="width:110px">Напечатано</th>
          <th>Комментарий</th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="loading"><td colspan="10" class="muted" style="text-align:center;padding:24px">Загрузка...</td></tr>
        <tr v-else-if="filtered.length === 0"><td colspan="10" class="muted" style="text-align:center;padding:24px">Нет инвентаризаций</td></tr>
        <tr v-else v-for="r in paginated" :key="r.id" class="clickable" :class="{ selected: selected.has(r.id) }" @click="open(r)">
          <td class="chk-col" @click.stop>
            <input type="checkbox" :checked="selected.has(r.id)" @change="toggleSelect(r.id)" />
          </td>
          <td>Инвентаризация</td>
          <td class="mono link">{{ shortNumber(r) }}</td>
          <td class="muted">{{ formatDate(r.doc_date) }}</td>
          <td>{{ r.warehouse_name || '—' }}</td>
          <td>—</td>
          <td>
            <span :class="['pill', r.status || 'draft']">{{ r.status === 'posted' ? 'Проведена' : 'Черновик' }}</span>
          </td>
          <td class="muted">—</td>
          <td>
            <span v-if="r.status === 'posted'" class="pill printed">Напечатан</span>
            <span v-else class="muted">—</span>
          </td>
          <td class="muted">{{ r.comment || '—' }}</td>
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
        <span>Всего: {{ filtered.length }}</span>
      </div>
    </div>

    <!-- Модалка выбора склада -->
    <div v-if="showPickWarehouse" class="modal-backdrop" @click.self="showPickWarehouse = false">
      <div class="pick-modal">
        <div class="pick-head">
          <span>Выберите склад для инвентаризации</span>
          <button class="btn-close" @click="showPickWarehouse = false">×</button>
        </div>
        <div class="pick-body">
          <div class="pick-group">Склады</div>
          <div
            v-for="w in warehouses"
            :key="w.id"
            class="pick-item"
            :class="{ active: pickedWarehouseId === w.id }"
            @click="pickedWarehouseId = w.id"
          >
            {{ w.name }}
          </div>
          <div v-if="warehouses.length === 0" class="muted" style="padding:8px 12px">Нет складов</div>
        </div>
        <div class="pick-foot">
          <button class="pick-primary" :disabled="!pickedWarehouseId" @click="pickWarehouse">Выбрать</button>
          <button class="pick-cancel" @click="showPickWarehouse = false">Отмена</button>
        </div>
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
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
const error = ref<string | null>(null)
const showFilter = ref(false)
const showPrintMenu = ref(false)
const showPickWarehouse = ref(false)
const pickedWarehouseId = ref('')
const page = ref(1)
const perPage = 100
const search = ref('')
const filterStatus = ref('')

const filters = reactive({
  warehouse_id: '',
  from: '',
  to: '',
})

const selected = ref<Set<string>>(new Set())

const filtered = computed(() => {
  let rs = rows.value
  if (filterStatus.value) rs = rs.filter((r) => (r.status || 'draft') === filterStatus.value)
  if (search.value) {
    const q = search.value.toLowerCase()
    rs = rs.filter((r) =>
      (r.number ?? '').toLowerCase().includes(q) ||
      (r.comment ?? '').toLowerCase().includes(q)
    )
  }
  return rs
})
const paginated = computed(() => {
  const from = (page.value - 1) * perPage
  return filtered.value.slice(from, from + perPage)
})
const rangeFrom = computed(() => filtered.value.length === 0 ? 0 : (page.value - 1) * perPage + 1)
const rangeTo = computed(() => Math.min(page.value * perPage, filtered.value.length))

const allChecked = computed(() =>
  paginated.value.length > 0 && paginated.value.every((r) => selected.value.has(r.id))
)
function toggleAll(e: Event) {
  const on = (e.target as HTMLInputElement).checked
  selected.value = on ? new Set(paginated.value.map((r) => r.id)) : new Set()
}
function toggleSelect(id: string) {
  const s = new Set(selected.value)
  s.has(id) ? s.delete(id) : s.add(id)
  selected.value = s
}

function shortNumber(r: Inventory) {
  return (r.number ?? '').replace(/^инв-?/i, '') || r.number
}
function formatDate(s: string) {
  if (!s) return '—'
  const d = new Date(s)
  return d.toLocaleDateString('ru-RU') + ' ' + d.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })
}

async function load() {
  loading.value = true
  error.value = null
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
  filterStatus.value = ''
  search.value = ''
  load()
}

function open(r: Inventory) {
  router.push('/inventories/' + r.id)
}

function openPickWarehouse() {
  pickedWarehouseId.value = ''
  showPickWarehouse.value = true
}
function pickWarehouse() {
  if (!pickedWarehouseId.value) return
  showPickWarehouse.value = false
  router.push({ path: '/inventory', query: { warehouse_id: pickedWarehouseId.value } })
}

function onPrint(kind: 'list' | 'inv' | 'inv3') {
  showPrintMenu.value = false
  if (kind === 'list') { window.print(); return }
  const ids = Array.from(selected.value)
  if (ids.length === 0) { alert('Выберите инвентаризацию'); return }
  const id = ids[0]
  const path = kind === 'inv' ? `/export` : kind === 'inv3' ? `/export-inv3` : `/export-inv19`
  window.open(`/api/v1/inventories/${id}${path}`, '_blank')
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
  box-shadow: 0 4px 16px rgba(0,0,0,0.12); min-width: 280px;
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

.modal-backdrop {
  position: fixed; inset: 0; background: rgba(0,0,0,0.45);
  display: flex; align-items: center; justify-content: center; z-index: 1000;
}
.pick-modal {
  background: #fff; border-radius: 6px; box-shadow: 0 10px 40px rgba(0,0,0,0.25);
  width: 100%; max-width: 520px; max-height: 80vh;
  display: flex; flex-direction: column; overflow: hidden;
}
.pick-head {
  display: flex; justify-content: space-between; align-items: center;
  padding: 14px 18px; border-bottom: 1px solid #e5e7eb;
  font-size: 15px; font-weight: 600; color: #1f2328;
}
.btn-close {
  background: none; border: none; font-size: 22px; line-height: 1;
  cursor: pointer; color: #666; padding: 0 6px;
}
.btn-close:hover { color: #000; }
.pick-body { padding: 8px 0; overflow-y: auto; flex: 1; }
.pick-group {
  padding: 6px 16px; font-size: 11px; text-transform: uppercase;
  color: #8c959f; letter-spacing: 0.5px; font-weight: 600;
}
.pick-item {
  padding: 8px 16px; font-size: 13px; color: #1f2328;
  cursor: pointer; border-radius: 3px; margin: 0 8px;
}
.pick-item:hover { background: #f6f8fa; }
.pick-item.active { background: #e8f0fb; color: #2c5d9c; font-weight: 500; }
.pick-foot {
  display: flex; gap: 8px; padding: 12px 18px; border-top: 1px solid #e5e7eb;
}
.pick-primary {
  height: 32px; padding: 0 20px; border: none; border-radius: 4px;
  background: #a8ce5e; color: #1f2328; font-weight: 600; cursor: pointer;
}
.pick-primary:hover:not(:disabled) { background: #97bd4b; }
.pick-primary:disabled { opacity: .5; cursor: default; }
.pick-cancel {
  height: 32px; padding: 0 16px; border: 1px solid #d0d7de; border-radius: 4px;
  background: #fff; color: #1f2328; cursor: pointer;
}
.pick-cancel:hover { background: #f6f8fa; }
.pill {
  display: inline-block; padding: 2px 8px; border-radius: 10px;
  font-size: 11px; font-weight: 500;
}
.pill.draft { background: #f0f2f5; color: #57606a; }
.pill.posted { background: #f5c242; color: #1f2328; }
.pill.printed { background: #4a90e2; color: #fff; }
</style>