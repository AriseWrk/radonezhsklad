<template>
  <div class="warehouses-layout">
    <aside class="wh-panel">
      <div class="cat-section">Склады</div>
      <div
        class="cat-item"
        :class="{ active: selectedWarehouseId === '' }"
        @click="selectWarehouse('')"
      >
        <span>Все склады</span>
        <span class="cat-count">{{ items.length }}</span>
      </div>
      <div
        v-for="w in items"
        :key="w.id"
        class="cat-item"
        :class="{ active: selectedWarehouseId === w.id }"
        @click="selectWarehouse(w.id)"
      >
        <span class="cat-name">{{ w.name }}</span>
      </div>
    </aside>

    <div class="wh-main">
      <div class="ms-title">
        <button class="ms-help" title="Справка"><MsIcon name="help" :size="14" /></button>
        <span>Склады</span>
        <button class="ms-refresh" @click="load" title="Обновить"><MsIcon name="refresh" :size="14" /></button>
      </div>

      <div class="ms-toolbar">
        <MsButton variant="primary" icon="plus" @click="showCreate = true">Склад</MsButton>
        <MsButton icon="filter" :class="{ active: showFilter }" @click="showFilter = !showFilter">Фильтр</MsButton>
        <input v-model="search" class="ms-input" placeholder="Наименование или код" />
        <div class="ms-counter">{{ filtered.length }}</div>
        <select class="ms-select" :disabled="selected.size === 0">
          <option>Изменить</option>
          <option>Удалить</option>
        </select>
        <MsButton variant="icon" icon="gear" title="Настройки" />
      </div>

      <div v-if="showFilter" class="ms-filter">
        <div class="filter-actions">
          <MsButton variant="green" @click="page = 1">Найти</MsButton>
          <MsButton @click="clearFilters">Очистить</MsButton>
        </div>
        <div class="filter-grid">
          <div class="filter-field"><label class="filter-label"><span class="dot"></span>Наименование</label>
            <input v-model="filterName" placeholder="" />
          </div>
          <div class="filter-field"><label class="filter-label"><span class="dot"></span>Код</label>
            <input v-model="filterCode" placeholder="" />
          </div>
          <div class="filter-field"><label class="filter-label"><span class="dot"></span>Адрес</label>
            <input v-model="filterAddress" placeholder="" />
          </div>
          <div class="filter-field"><label class="filter-label"><span class="dot"></span>Показывать</label>
            <select><option>Только обычные</option><option>Все</option></select>
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
        </div>
      </div>

      <div v-if="error" class="error-box">{{ error }}</div>

      <table class="ms-table">
        <thead>
          <tr>
            <th class="chk-col"><input type="checkbox" :checked="allChecked" @change="toggleAll" /></th>
            <th @click="sortBy('name')">
              Наименование
              <span v-if="sortKey === 'name'" class="sort-arrow">{{ sortDir === 'asc' ? '↑' : '↓' }}</span>
            </th>
            <th style="width:120px">Код</th>
            <th>Адрес</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="loading"><td colspan="4" class="muted" style="text-align:center;padding:24px">Загрузка...</td></tr>
          <tr v-else-if="filtered.length === 0"><td colspan="4" class="muted" style="text-align:center;padding:24px">Нет складов</td></tr>
          <tr v-else v-for="w in paginated" :key="w.id" class="clickable" :class="{ selected: selected.has(w.id) }" @click="onEdit(w)">
            <td class="chk-col" @click.stop>
              <input type="checkbox" :checked="selected.has(w.id)" @change="toggleSelect(w.id)" />
            </td>
            <td class="link">{{ w.name }}</td>
            <td class="muted mono">{{ shortId(w.id) }}</td>
            <td class="muted">{{ w.address || '—' }}</td>
          </tr>
        </tbody>
      </table>

      <div class="ms-footer">
        <div class="ms-pager">
          <button :disabled="page === 1" @click="page--">◀</button>
          <button :disabled="page === 1" @click="page = 1">↤</button>
          <span>{{ rangeFrom }}–{{ rangeTo }} из {{ filtered.length }}</span>
          <button :disabled="rangeTo >= filtered.length" @click="page = Math.ceil(filtered.length / perPage)">↦</button>
          <button :disabled="rangeTo >= filtered.length" @click="page++">▶</button>
        </div>
        <div class="ms-totals">
          <span>Всего: {{ filtered.length }}</span>
        </div>
      </div>
    </div>

    <div v-if="showCreate" class="modal-backdrop" @click.self="showCreate = false">
      <form class="card modal" @submit.prevent="onCreate">
        <h2>Новый склад</h2>
        <div v-if="createError" class="error-box">{{ createError }}</div>
        <label>Наименование*
          <input v-model="form.name" required />
        </label>
        <label>Адрес
          <input v-model="form.address" />
        </label>
        <div class="modal-actions">
          <button type="button" @click="showCreate = false">Отмена</button>
          <button class="primary" type="submit" :disabled="saving">
            {{ saving ? 'Сохранение...' : 'Создать' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>
<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { listWarehouses, createWarehouse, deleteWarehouse, type Warehouse } from '../api/warehouses'
import { apiErrorMessage } from '../api/client'
import MsButton from '../components/MsButton.vue'
import MsIcon from '../components/MsIcon.vue'

const items = ref<Warehouse[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const showCreate = ref(false)
const showFilter = ref(false)
const saving = ref(false)
const createError = ref<string | null>(null)
const search = ref('')
const page = ref(1)
const perPage = 100
const selectedWarehouseId = ref('')

const filterName = ref('')
const filterCode = ref('')
const filterAddress = ref('')

const sortKey = ref<'name'>('name')
const sortDir = ref<'asc' | 'desc'>('asc')

const selected = ref<Set<string>>(new Set())

const form = reactive({ name: '', address: '' })

function shortId(id: string) { return id.slice(0, 8) }

const filtered = computed(() => {
  let rs = items.value
  if (selectedWarehouseId.value) rs = rs.filter((w) => w.id === selectedWarehouseId.value)
  if (search.value) {
    const q = search.value.toLowerCase()
    rs = rs.filter((w) =>
      (w.name ?? '').toLowerCase().includes(q) ||
      (w.address ?? '').toLowerCase().includes(q)
    )
  }
  if (filterName.value) rs = rs.filter((w) => (w.name ?? '').toLowerCase().includes(filterName.value.toLowerCase()))
  if (filterCode.value) rs = rs.filter((w) => (w.id ?? '').toLowerCase().includes(filterCode.value.toLowerCase()))
  if (filterAddress.value) rs = rs.filter((w) => (w.address ?? '').toLowerCase().includes(filterAddress.value.toLowerCase()))
  return [...rs].sort((a, b) => {
    const as = String(a.name ?? '').toLowerCase()
    const bs = String(b.name ?? '').toLowerCase()
    return sortDir.value === 'asc' ? as.localeCompare(bs, 'ru') : bs.localeCompare(as, 'ru')
  })
})

const paginated = computed(() => {
  const from = (page.value - 1) * perPage
  return filtered.value.slice(from, from + perPage)
})
const rangeFrom = computed(() => filtered.value.length === 0 ? 0 : (page.value - 1) * perPage + 1)
const rangeTo = computed(() => Math.min(page.value * perPage, filtered.value.length))

const allChecked = computed(() =>
  paginated.value.length > 0 && paginated.value.every((w) => selected.value.has(w.id))
)
function toggleAll(e: Event) {
  const on = (e.target as HTMLInputElement).checked
  selected.value = on ? new Set(paginated.value.map((w) => w.id)) : new Set()
}
function toggleSelect(id: string) {
  const s = new Set(selected.value)
  s.has(id) ? s.delete(id) : s.add(id)
  selected.value = s
}
function selectWarehouse(id: string) {
  selectedWarehouseId.value = id
  page.value = 1
}
function sortBy(k: typeof sortKey.value) {
  if (sortKey.value === k) sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  else { sortKey.value = k; sortDir.value = 'asc' }
  page.value = 1
}

function clearFilters() {
  filterName.value = ''
  filterCode.value = ''
  filterAddress.value = ''
  search.value = ''
  selectedWarehouseId.value = ''
  page.value = 1
}

async function load() {
  loading.value = true
  error.value = null
  try {
    items.value = await listWarehouses()
  } catch (e) {
    error.value = apiErrorMessage(e)
  } finally {
    loading.value = false
  }
}

async function onCreate() {
  createError.value = null
  saving.value = true
  try {
    await createWarehouse(form.name, form.address || undefined)
    showCreate.value = false
    form.name = ''; form.address = ''
    await load()
  } catch (e) {
    createError.value = apiErrorMessage(e)
  } finally {
    saving.value = false
  }
}

async function onDelete(w: Warehouse) {
  if (!confirm(`Удалить склад «${w.name}»?`)) return
  try { await deleteWarehouse(w.id); await load() }
  catch (e) { error.value = apiErrorMessage(e) }
}

function onEdit(w: Warehouse) {
  // Пока без отдельной карточки — открываем удаление/редактирование в будущем
  void w
}

onMounted(load)
</script>
<style scoped>
.warehouses-layout {
  display: grid;
  grid-template-columns: 240px 1fr;
  gap: 16px;
  min-height: calc(100vh - 120px);
}
.wh-panel {
  background: #fff;
  border: 1px solid #d8dee4;
  border-radius: 4px;
  padding: 8px 0;
  height: fit-content;
  position: sticky;
  top: 16px;
  max-height: calc(100vh - 140px);
  overflow-y: auto;
}
.cat-section {
  padding: 10px 16px 4px;
  font-size: 11px;
  text-transform: uppercase;
  color: #8c959f;
  letter-spacing: 0.5px;
  font-weight: 600;
}
.cat-item {
  display: flex; justify-content: space-between; align-items: center;
  padding: 6px 16px; font-size: 13px; color: #1f2328; cursor: pointer;
  gap: 8px;
}
.cat-item:hover { background: #f6f8fa; }
.cat-item.active { background: #eef4ff; color: #2c5d9c; font-weight: 600; }
.cat-name { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.cat-count { font-size: 11px; color: #8c959f; background: #f0f2f5; padding: 1px 6px; border-radius: 8px; flex-shrink: 0; }
.cat-item.active .cat-count { background: #d8e4f0; color: #2c5d9c; }
.wh-main { min-width: 0; }

.ms-title { display: flex; align-items: center; gap: 8px; font-size: 20px; font-weight: 600; color: #1f2328; margin-bottom: 12px; }
.ms-help { width: 20px; height: 20px; border-radius: 50%; border: 1px solid #b8c0c8; background: transparent; color: #57606a; cursor: pointer; display: inline-flex; align-items: center; justify-content: center; padding: 0; }
.ms-help:hover { background: #f0f2f5; }
.ms-refresh { width: 24px; height: 24px; padding: 0; display: inline-flex; align-items: center; justify-content: center; border: none; background: transparent; color: #57606a; cursor: pointer; }
.ms-refresh:hover { color: #2c5d9c; }
.ms-toolbar { display: flex; align-items: center; gap: 6px; margin-bottom: 12px; flex-wrap: wrap; }
.ms-input { flex: 1; max-width: 320px; height: 30px; padding: 0 10px; font-size: 13px; border: 1px solid #d0d7de; border-radius: 4px; background: #fff; color: #1f2328; }
.ms-input::placeholder { color: #8c959f; }
.ms-counter { min-width: 40px; height: 30px; padding: 0 10px; display: inline-flex; align-items: center; justify-content: center; border: 1px solid #d0d7de; border-radius: 4px; background: #fff; color: #8c959f; font-variant-numeric: tabular-nums; font-size: 13px; }
.ms-select { height: 30px; padding: 0 8px; font-size: 13px; border: 1px solid #d0d7de; border-radius: 4px; background: #fff; color: #1f2328; max-width: 140px; }
.ms-filter { background: #eef1f5; border: 1px solid #d8dee4; border-radius: 4px; padding: 10px 14px 12px; margin-bottom: 12px; }
.filter-actions { display: flex; align-items: center; gap: 6px; margin-bottom: 10px; }
.filter-grid { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 10px 14px; }
.filter-field { display: flex; flex-direction: column; gap: 3px; }
.filter-label { font-size: 12px; color: #57606a; display: flex; align-items: center; gap: 5px; }
.filter-label .dot { width: 8px; height: 8px; border-radius: 50%; background: #2c5d9c; flex-shrink: 0; }
.filter-field input, .filter-field select { height: 28px; padding: 0 8px; font-size: 12px; border: 1px solid #d0d7de; border-radius: 3px; background: #fff; color: #1f2328; }
.filter-field input:disabled, .filter-field select:disabled { background: #f6f8fa; color: #8c959f; }

.ms-table { width: 100%; border-collapse: collapse; background: #fff; font-size: 13px; }
.ms-table thead th { background: #fff; color: #2c5d9c; font-weight: 500; padding: 8px 10px; text-align: left; border-bottom: 1px solid #d8dee4; white-space: nowrap; font-size: 12px; }
.ms-table tbody td { padding: 7px 10px; border-bottom: 1px solid #eaeef2; vertical-align: middle; }
.ms-table tbody tr.clickable { cursor: pointer; }
.ms-table tbody tr.clickable:hover { background: #f6f8fa; }
.ms-table tbody tr.selected { background: #eef4ff; }
.ms-table .chk-col { width: 32px; }
.ms-table .link { color: #2c5d9c; font-weight: 500; }
.ms-table .muted { color: #57606a; }
.ms-table .mono { font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 12px; }
.sort-arrow { color: #2c5d9c; font-size: 11px; margin-left: 4px; }

.ms-footer { display: flex; align-items: center; justify-content: space-between; padding: 8px 10px; font-size: 12px; color: #57606a; background: #fff; border-top: 1px solid #eaeef2; }
.ms-pager { display: flex; align-items: center; gap: 4px; }
.ms-pager button { width: 22px; height: 22px; padding: 0; border: 1px solid #d0d7de; background: #fff; border-radius: 3px; cursor: pointer; font-size: 12px; color: #1f2328; }
.ms-pager button:disabled { opacity: 0.4; cursor: default; }
.ms-pager span { margin: 0 6px; font-variant-numeric: tabular-nums; }
.ms-totals { display: flex; gap: 16px; font-variant-numeric: tabular-nums; }

.error-box { background: #ffebe9; border: 1px solid #ff8182; border-radius: 4px; padding: 8px 12px; margin-bottom: 12px; color: #82071e; font-size: 13px; }

.modal-backdrop { position: fixed; inset: 0; background: rgba(0,0,0,0.4); display: flex; align-items: center; justify-content: center; padding: 20px; z-index: 1000; }
.modal { background: #fff; border-radius: 4px; width: 420px; max-width: 100%; padding: 20px; display: flex; flex-direction: column; gap: 10px; }
.modal h2 { margin: 0 0 4px; font-size: 18px; }
.modal label { font-size: 13px; color: #444; display: flex; flex-direction: column; gap: 4px; }
.modal-actions { display: flex; justify-content: flex-end; gap: 10px; margin-top: 6px; }
</style>