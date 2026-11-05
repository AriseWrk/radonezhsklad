<template>
  <div>
    <div class="ms-title">
      <button class="ms-help" title="Справка"><MsIcon name="help" :size="14" /></button>
      <span>Юр. лица</span>
      <button class="ms-refresh" @click="load" title="Обновить"><MsIcon name="refresh" :size="14" /></button>
    </div>

    <div class="ms-toolbar">
      <MsButton variant="primary" icon="plus" @click="openCreate">Организация</MsButton>
      <MsButton icon="filter" :class="{ active: showFilter }" @click="showFilter = !showFilter">Фильтр</MsButton>
      <input v-model="search" class="ms-input" placeholder="Наименование, ИНН, КПП" />
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
          <input v-model="filterName" />
        </div>
        <div class="filter-field"><label class="filter-label"><span class="dot"></span>Полное наименование</label>
          <input v-model="filterFullName" />
        </div>
        <div class="filter-field"><label class="filter-label"><span class="dot"></span>ИНН</label>
          <input v-model="filterInn" />
        </div>
        <div class="filter-field"><label class="filter-label"><span class="dot"></span>КПП</label>
          <input v-model="filterKpp" />
        </div>
        <div class="filter-field"><label class="filter-label"><span class="dot"></span>ОГРН</label>
          <input v-model="filterOgrn" />
        </div>
        <div class="filter-field"><label class="filter-label"><span class="dot"></span>Юр. адрес</label>
          <input v-model="filterAddress" />
        </div>
        <div class="filter-field"><label class="filter-label"><span class="dot"></span>Показывать</label>
          <select><option>Только обычные</option><option>Все</option></select>
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
          <th style="width:130px">ИНН</th>
          <th style="width:130px">КПП</th>
          <th style="width:150px">ОГРН</th>
          <th>Юр. адрес</th>
          <th style="width:100px">Основная</th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="loading"><td colspan="7" class="muted" style="text-align:center;padding:24px">Загрузка...</td></tr>
        <tr v-else-if="filtered.length === 0"><td colspan="7" class="muted" style="text-align:center;padding:24px">Нет организаций</td></tr>
        <tr v-else v-for="o in paginated" :key="o.id" class="clickable" :class="{ selected: selected.has(o.id) }" @click="openEdit(o)">
          <td class="chk-col" @click.stop>
            <input type="checkbox" :checked="selected.has(o.id)" @change="toggleSelect(o.id)" />
          </td>
          <td class="link">{{ o.name }}</td>
          <td class="muted mono">{{ o.inn || '—' }}</td>
          <td class="muted mono">{{ o.kpp || '—' }}</td>
          <td class="muted mono">{{ o.ogrn || '—' }}</td>
          <td class="muted">{{ o.legal_address || '—' }}</td>
          <td>
            <span v-if="o.is_default" class="pill default">Да</span>
            <span v-else class="muted">—</span>
          </td>
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
      <div class="ms-totals"><span>Всего: {{ filtered.length }}</span></div>
    </div>

    <div v-if="showModal" class="modal-backdrop" @click.self="closeModal">
      <form class="card modal big" @submit.prevent="onSubmit">
        <h2>{{ modal.isEdit ? 'Организация' : 'Новая организация' }}</h2>
        <div v-if="modal.error" class="error-box">{{ modal.error }}</div>

        <label>Наименование*
          <input v-model="form.name" required />
        </label>
        <label>Полное наименование
          <input v-model="form.full_name" />
        </label>
        <div class="grid2">
          <label>ИНН
            <input v-model="form.inn" />
          </label>
          <label>КПП
            <input v-model="form.kpp" />
          </label>
        </div>
        <div class="grid2">
          <label>ОГРН
            <input v-model="form.ogrn" />
          </label>
          <label>ОКПО
            <input v-model="form.okpo" />
          </label>
        </div>
        <label>Юр. адрес
          <input v-model="form.legal_address" />
        </label>
        <label class="checkbox-row">
          <input type="checkbox" v-model="form.is_default" />
          <span>Основная организация</span>
        </label>

        <div class="modal-actions">
          <button v-if="modal.isEdit" type="button" class="danger" @click="onDelete">Удалить</button>
          <div style="flex:1"></div>
          <button type="button" @click="closeModal">Отмена</button>
          <button class="primary" type="submit" :disabled="modal.saving">
            {{ modal.saving ? 'Сохранение...' : 'Сохранить' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>
<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import {
  listOrganizations, createOrganization, updateOrganization, deleteOrganization,
  type Organization, type OrganizationInput,
} from '../api/suppliers'
import { apiErrorMessage } from '../api/client'
import MsButton from '../components/MsButton.vue'
import MsIcon from '../components/MsIcon.vue'

const items = ref<Organization[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const showFilter = ref(false)
const showModal = ref(false)
const search = ref('')
const page = ref(1)
const perPage = 100

const filterName = ref('')
const filterFullName = ref('')
const filterInn = ref('')
const filterKpp = ref('')
const filterOgrn = ref('')
const filterAddress = ref('')

const sortKey = ref<'name'>('name')
const sortDir = ref<'asc' | 'desc'>('asc')

const selected = ref<Set<string>>(new Set())

const modal = reactive({ isEdit: false, saving: false, error: null as string | null, editingId: '' })
const form = reactive<OrganizationInput & { id: string }>({
  id: '',
  name: '',
  full_name: '',
  inn: '',
  kpp: '',
  ogrn: '',
  okpo: '',
  legal_address: '',
  is_default: false,
})

const filtered = computed(() => {
  let rs = items.value
  if (search.value) {
    const q = search.value.toLowerCase()
    rs = rs.filter((o) =>
      (o.name ?? '').toLowerCase().includes(q) ||
      (o.inn ?? '').toLowerCase().includes(q) ||
      (o.kpp ?? '').toLowerCase().includes(q)
    )
  }
  if (filterName.value) rs = rs.filter((o) => (o.name ?? '').toLowerCase().includes(filterName.value.toLowerCase()))
  if (filterFullName.value) rs = rs.filter((o) => (o.full_name ?? '').toLowerCase().includes(filterFullName.value.toLowerCase()))
  if (filterInn.value) rs = rs.filter((o) => (o.inn ?? '').toLowerCase().includes(filterInn.value.toLowerCase()))
  if (filterKpp.value) rs = rs.filter((o) => (o.kpp ?? '').toLowerCase().includes(filterKpp.value.toLowerCase()))
  if (filterOgrn.value) rs = rs.filter((o) => (o.ogrn ?? '').toLowerCase().includes(filterOgrn.value.toLowerCase()))
  if (filterAddress.value) rs = rs.filter((o) => (o.legal_address ?? '').toLowerCase().includes(filterAddress.value.toLowerCase()))
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
  paginated.value.length > 0 && paginated.value.every((o) => selected.value.has(o.id))
)
function toggleAll(e: Event) {
  const on = (e.target as HTMLInputElement).checked
  selected.value = on ? new Set(paginated.value.map((o) => o.id)) : new Set()
}
function toggleSelect(id: string) {
  const s = new Set(selected.value)
  s.has(id) ? s.delete(id) : s.add(id)
  selected.value = s
}
function sortBy(k: typeof sortKey.value) {
  if (sortKey.value === k) sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  else { sortKey.value = k; sortDir.value = 'asc' }
  page.value = 1
}

function clearFilters() {
  filterName.value = ''
  filterFullName.value = ''
  filterInn.value = ''
  filterKpp.value = ''
  filterOgrn.value = ''
  filterAddress.value = ''
  search.value = ''
  page.value = 1
}

async function load() {
  loading.value = true
  error.value = null
  try {
    items.value = await listOrganizations()
  } catch (e) {
    error.value = apiErrorMessage(e)
  } finally {
    loading.value = false
  }
}

function openCreate() {
  Object.assign(form, { id: '', name: '', full_name: '', inn: '', kpp: '', ogrn: '', okpo: '', legal_address: '', is_default: false })
  modal.isEdit = false
  modal.error = null
  showModal.value = true
}

function openEdit(o: Organization) {
  Object.assign(form, {
    id: o.id,
    name: o.name,
    full_name: o.full_name ?? '',
    inn: o.inn ?? '',
    kpp: o.kpp ?? '',
    ogrn: o.ogrn ?? '',
    okpo: o.okpo ?? '',
    legal_address: o.legal_address ?? '',
    is_default: o.is_default,
  })
  modal.isEdit = true
  modal.error = null
  showModal.value = true
}

function closeModal() { showModal.value = false }

async function onSubmit() {
  modal.error = null
  if (!form.name.trim()) { modal.error = 'Введите наименование'; return }
  modal.saving = true
  try {
    const payload: OrganizationInput = {
      name: form.name.trim(),
      full_name: form.full_name || undefined,
      inn: form.inn || undefined,
      kpp: form.kpp || undefined,
      ogrn: form.ogrn || undefined,
      okpo: form.okpo || undefined,
      legal_address: form.legal_address || undefined,
      is_default: form.is_default,
    }
    if (modal.isEdit) {
      await updateOrganization(form.id, payload)
    } else {
      await createOrganization(payload)
    }
    closeModal()
    await load()
  } catch (e) {
    modal.error = apiErrorMessage(e)
  } finally {
    modal.saving = false
  }
}

async function onDelete() {
  if (!confirm(`Удалить организацию «${form.name}»?`)) return
  try {
    await deleteOrganization(form.id)
    closeModal()
    await load()
  } catch (e) {
    modal.error = apiErrorMessage(e)
  }
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
.pill.default { background: #d8e8c8; color: #2d6a1e; padding: 2px 8px; border-radius: 3px; font-size: 11px; }
.checkbox-row { display: flex; align-items: center; gap: 6px; font-size: 13px; cursor: pointer; }
.checkbox-row input { width: auto; height: auto; }
.grid2 { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; }
</style>