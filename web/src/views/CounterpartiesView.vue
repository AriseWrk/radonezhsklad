<template>
  <div>
    <div class="page-title-bar">
      <div class="page-title">
        <span class="info-icon">ⓘ</span>
        <span>Контрагенты</span>
        <span class="refresh" @click="load" title="Обновить">↻</span>
      </div>
      <div class="page-actions">
        <button class="btn primary icon-btn" @click="openCreate">
          <span class="plus">+</span> Контрагент
        </button>
        <button class="btn" @click="showFilter = !showFilter">Фильтр</button>
        <input v-model="search" class="search-input" placeholder="Наим, тел, email, событ, коммент, код" />
        <div class="counter" :class="{ active: selected.size > 0 }">{{ selected.size }}</div>
        <select class="mini-select" :disabled="selected.size === 0">
          <option>Изменить</option>
          <option>Удалить</option>
        </select>
        <select class="mini-select" v-model="filterStatus">
          <option value="">Статус</option>
          <option value="Новый">Новый</option>
          <option value="Активный">Активный</option>
          <option value="Закрыт">Закрыт</option>
        </select>
        <button class="btn" disabled>Печать</button>
        <button class="btn" disabled>Создать задачи</button>
        <button class="btn" disabled>Рассылки</button>
        <button class="btn" disabled>Импорт</button>
        <button class="btn" @click="exportCsv">Экспорт</button>
        <button class="btn icon-only" title="Настройки">⚙</button>
      </div>
    </div>

    <div v-if="showFilter" class="filter-panel">
      <div class="filter-row">
        <div class="filter-actions">
          <button class="btn-find" @click="page = 1">Найти</button>
          <button class="btn-clear" @click="clearFilters">Очистить</button>
        </div>
        <div class="filter-field">
          <label>Архивные</label>
          <select v-model="includeArchived" @change="load">
            <option :value="false">Скрыть</option>
            <option :value="true">Показать</option>
          </select>
        </div>
        <div class="filter-field">
          <label>ИНН</label>
          <input v-model="filterInn" placeholder="ИНН" />
        </div>
        <div class="filter-field">
          <label>Группа</label>
          <input v-model="filterGroup" placeholder="Группа" />
        </div>
        <div class="filter-field">
          <label>Тип</label>
          <select v-model="filterType">
            <option value="">Все</option>
            <option value="Юридическое лицо. Россия">Юридическое лицо</option>
            <option value="Индивидуальный предприниматель. Россия">ИП</option>
          </select>
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
          <th style="width:100px">Код</th>
          <th style="width:140px" @click="sortBy('created_at')">
            Создан
            <span v-if="sortKey === 'created_at'" class="sort-arrow">{{ sortDir === 'asc' ? '↑' : '↓' }}</span>
          </th>
          <th style="width:140px">Телефон</th>
          <th style="width:80px">Факс</th>
          <th style="width:160px">E-mail</th>
          <th style="width:100px">Статус</th>
          <th>Фактический адрес</th>
          <th style="width:180px">Комментарий</th>
          <th style="width:120px">Группы</th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="loading"><td colspan="11" class="muted" style="text-align:center;padding:24px">Загрузка...</td></tr>
        <tr v-else-if="filtered.length === 0"><td colspan="11" class="muted" style="text-align:center;padding:24px">Нет контрагентов</td></tr>
        <tr
          v-else
          v-for="c in paginated"
          :key="c.id"
          class="clickable"
          :class="{ selected: selected.has(c.id) }"
          @click="openEdit(c)"
        >
          <td class="chk-col" @click.stop>
            <input type="checkbox" :checked="selected.has(c.id)" @change="toggleSelect(c.id)" />
          </td>
          <td>{{ c.name }}</td>
          <td class="muted mono">{{ c.external_code || shortId(c.id) }}</td>
          <td class="muted">{{ formatShort(c.created_at) }}</td>
          <td>{{ c.phone || '' }}</td>
          <td class="muted">{{ c.fax || '' }}</td>
          <td class="muted">{{ c.email || '' }}</td>
          <td><span class="status-badge">{{ c.status || 'Новый' }}</span></td>
          <td class="muted">{{ c.actual_address || c.address || '' }}</td>
          <td class="muted">{{ c.comment || '' }}</td>
          <td class="muted">{{ c.group_name || '' }}</td>
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
    </div>

    <!-- Модалка -->
    <div v-if="modal.open" class="modal-backdrop" @click.self="closeModal">
      <form class="card modal big" @submit.prevent="onSubmit">
        <h2>{{ modal.isEdit ? 'Карточка контрагента' : 'Новый контрагент' }}</h2>
        <div v-if="modal.error" class="error-box">{{ modal.error }}</div>

        <label>Наименование*
          <input v-model="form.name" required />
        </label>

        <label>Полное наименование
          <input v-model="form.full_name" />
        </label>

        <div class="grid3">
          <label>Фамилия
            <input v-model="form.last_name" />
          </label>
          <label>Имя
            <input v-model="form.first_name" />
          </label>
          <label>Отчество
            <input v-model="form.middle_name" />
          </label>
        </div>

        <div class="grid2">
          <label>Телефон
            <input v-model="form.phone" placeholder="+7 999 123-45-67" />
          </label>
          <label>E-mail
            <input v-model="form.email" type="email" />
          </label>
        </div>

        <div class="grid2">
          <label>Факс
            <input v-model="form.fax" />
          </label>
          <label>Код (внешний)
            <input v-model="form.external_code" />
          </label>
        </div>

        <div class="grid3">
          <label>ИНН
            <input v-model="form.inn" />
          </label>
          <label>КПП
            <input v-model="form.kpp" />
          </label>
          <label>ОКПО
            <input v-model="form.okpo" />
          </label>
        </div>

        <div class="grid2">
          <label>Юридический адрес
            <input v-model="form.legal_address" />
          </label>
          <label>Фактический адрес
            <input v-model="form.actual_address" />
          </label>
        </div>

        <div class="grid2">
          <label>Тип контрагента
            <select v-model="form.counterparty_type">
              <option value="">— не выбран —</option>
              <option value="Юридическое лицо. Россия">Юридическое лицо. Россия</option>
              <option value="Индивидуальный предприниматель. Россия">ИП. Россия</option>
              <option value="Физическое лицо. Россия">Физическое лицо. Россия</option>
            </select>
          </label>
          <label>Статус
            <select v-model="form.status">
              <option value="Новый">Новый</option>
              <option value="Активный">Активный</option>
              <option value="Закрыт">Закрыт</option>
            </select>
          </label>
        </div>

        <div class="grid2">
          <label>Группа
            <input v-model="form.group_name" />
          </label>
          <label class="chk">
            <input type="checkbox" v-model="form.archived" /> Архивный
          </label>
        </div>

        <label>Комментарий
          <textarea v-model="form.comment" rows="2"></textarea>
        </label>

        <div class="modal-actions">
          <button
            v-if="modal.isEdit"
            type="button"
            class="danger"
            @click="onDelete"
          >Удалить</button>
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
import { computed, onMounted, reactive, ref, watch } from 'vue'
import {
  listCustomers, createCustomer, updateCustomer, deleteCustomer,
  type Customer,
} from '../api/customers'
import { apiErrorMessage } from '../api/client'

const items = ref<Customer[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const showFilter = ref(false)
const page = ref(1)
const perPage = 100

const search = ref('')
const includeArchived = ref(false)
const filterStatus = ref('')
const filterInn = ref('')
const filterGroup = ref('')
const filterType = ref('')

const sortKey = ref<'name' | 'created_at'>('name')
const sortDir = ref<'asc' | 'desc'>('asc')
const selected = ref<Set<string>>(new Set())

const modal = reactive({
  open: false, isEdit: false, saving: false,
  error: null as string | null, editingId: '' as string,
})

const form = reactive({
  name: '', full_name: '', last_name: '', first_name: '', middle_name: '',
  phone: '', fax: '', email: '', address: '', legal_address: '', actual_address: '',
  inn: '', kpp: '', ogrn: '', okpo: '', external_code: '', counterparty_type: '',
  status: 'Новый', group_name: '', comment: '', archived: false,
})

function shortId(id: string) { return id.slice(0, 8) }
function formatShort(s: string) {
  const d = new Date(s)
  const day = String(d.getDate()).padStart(2, '0')
  const mon = String(d.getMonth() + 1).padStart(2, '0')
  const yr = d.getFullYear()
  const hh = String(d.getHours()).padStart(2, '0')
  const mm = String(d.getMinutes()).padStart(2, '0')
  return `${day}.${mon}.${yr} ${hh}:${mm}`
}
function sortBy(k: typeof sortKey.value) {
  if (sortKey.value === k) sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  else { sortKey.value = k; sortDir.value = 'asc' }
  page.value = 1
}

const filtered = computed(() => {
  let rows = items.value
  if (search.value) {
    const q = search.value.toLowerCase()
    rows = rows.filter((c) =>
      c.name.toLowerCase().includes(q) ||
      (c.phone ?? '').toLowerCase().includes(q) ||
      (c.email ?? '').toLowerCase().includes(q) ||
      (c.comment ?? '').toLowerCase().includes(q) ||
      (c.external_code ?? '').toLowerCase().includes(q) ||
      (c.inn ?? '').toLowerCase().includes(q)
    )
  }
  if (filterStatus.value) rows = rows.filter((c) => c.status === filterStatus.value)
  if (filterInn.value) rows = rows.filter((c) => (c.inn ?? '').includes(filterInn.value))
  if (filterGroup.value) rows = rows.filter((c) => (c.group_name ?? '').toLowerCase().includes(filterGroup.value.toLowerCase()))
  if (filterType.value) rows = rows.filter((c) => c.counterparty_type === filterType.value)

  return [...rows].sort((a, b) => {
    const av = a[sortKey.value]; const bv = b[sortKey.value]
    const as = String(av ?? ''); const bs = String(bv ?? '')
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
  paginated.value.length > 0 && paginated.value.every((c) => selected.value.has(c.id))
)
function toggleAll(e: Event) {
  const on = (e.target as HTMLInputElement).checked
  selected.value = on ? new Set(paginated.value.map((c) => c.id)) : new Set()
}
function toggleSelect(id: string) {
  const s = new Set(selected.value)
  s.has(id) ? s.delete(id) : s.add(id)
  selected.value = s
}

async function load() {
  loading.value = true
  error.value = null
  try {
    items.value = await listCustomers(includeArchived.value)
    page.value = 1
  } catch (e) {
    error.value = apiErrorMessage(e)
  } finally {
    loading.value = false
  }
}

function openCreate() {
  Object.assign(form, {
    name: '', full_name: '', last_name: '', first_name: '', middle_name: '',
    phone: '', fax: '', email: '', address: '', legal_address: '', actual_address: '',
    inn: '', kpp: '', ogrn: '', okpo: '', external_code: '', counterparty_type: '',
    status: 'Новый', group_name: '', comment: '', archived: false,
  })
  modal.open = true; modal.isEdit = false; modal.error = null; modal.editingId = ''
}

function openEdit(c: Customer) {
  Object.assign(form, {
    name: c.name,
    full_name: c.full_name ?? '',
    last_name: c.last_name ?? '',
    first_name: c.first_name ?? '',
    middle_name: c.middle_name ?? '',
    phone: c.phone ?? '',
    fax: c.fax ?? '',
    email: c.email ?? '',
    address: c.address ?? '',
    legal_address: c.legal_address ?? '',
    actual_address: c.actual_address ?? '',
    inn: c.inn ?? '',
    kpp: c.kpp ?? '',
    ogrn: c.ogrn ?? '',
    okpo: c.okpo ?? '',
    external_code: c.external_code ?? '',
    counterparty_type: c.counterparty_type ?? '',
    status: c.status || 'Новый',
    group_name: c.group_name ?? '',
    comment: c.comment ?? '',
    archived: c.archived,
  })
  modal.open = true; modal.isEdit = true; modal.error = null; modal.editingId = c.id
}

function closeModal() { modal.open = false }

async function onSubmit() {
  modal.error = null
  modal.saving = true
  try {
    const payload = { ...form }
    if (modal.isEdit) {
      await updateCustomer(modal.editingId, payload)
    } else {
      await createCustomer(payload)
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
  if (!confirm(`Удалить «${form.name}»?`)) return
  try {
    await deleteCustomer(modal.editingId)
    closeModal()
    await load()
  } catch (e) { modal.error = apiErrorMessage(e) }
}

function clearFilters() {
  search.value = ''
  filterStatus.value = ''
  filterInn.value = ''
  filterGroup.value = ''
  filterType.value = ''
  page.value = 1
}

function exportCsv() {
  const header = ['Наименование', 'Код', 'Создан', 'Телефон', 'Факс', 'E-mail', 'Статус', 'Адрес', 'Комментарий', 'Группа']
  const lines = [header.join(';')]
  for (const c of filtered.value) {
    lines.push([
      csvEsc(c.name),
      csvEsc(c.external_code ?? ''),
      formatShort(c.created_at),
      csvEsc(c.phone ?? ''),
      csvEsc(c.fax ?? ''),
      csvEsc(c.email ?? ''),
      csvEsc(c.status ?? ''),
      csvEsc(c.actual_address ?? c.address ?? ''),
      csvEsc(c.comment ?? ''),
      csvEsc(c.group_name ?? ''),
    ].join(';'))
  }
  const blob = new Blob(['\uFEFF' + lines.join('\r\n')], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `counterparties-${new Date().toISOString().slice(0,10)}.csv`
  a.click()
  URL.revokeObjectURL(url)
}
function csvEsc(s: string) {
  if (!s) return ''
  if (s.includes(';') || s.includes('"') || s.includes('\n')) return '"' + s.replace(/"/g, '""') + '"'
  return s
}

watch(includeArchived, load)
onMounted(load)
</script>

<style scoped>
.info-icon {
  display: inline-flex; align-items: center; justify-content: center;
  width: 22px; height: 22px; border: 1.5px solid #2c5d9c; border-radius: 50%;
  color: #2c5d9c; font-size: 14px; font-weight: 700; margin-right: 6px;
}
.icon-btn { display: inline-flex; align-items: center; gap: 4px; }
.icon-btn .plus { font-size: 13px; font-weight: 700; line-height: 1; }
.search-input {
  padding: 6px 10px; font-size: 13px;
  border: 1px solid #d0d7de; border-radius: 4px;
  background: #fff; color: #1f2328; min-width: 260px;
}
.counter {
  display: inline-flex; align-items: center; justify-content: center;
  min-width: 40px; height: 30px; padding: 0 10px;
  background: #fff; border: 1px solid #d0d7de; border-radius: 4px;
  font-size: 13px; color: #8c959f;
}
.counter.active { color: #2c5d9c; border-color: #2c5d9c; font-weight: 600; }
.mini-select {
  padding: 6px 10px; font-size: 13px;
  border: 1px solid #d0d7de; border-radius: 4px;
  background: #fff; color: #1f2328; max-width: 130px;
}
.mini-select:disabled { color: #8c959f; background: #f6f8fa; }
.btn.icon-only { padding: 6px 10px; font-size: 15px; }
tr.clickable { cursor: pointer; }
tr.clickable.selected { background: #eef4ff; }
.sort-arrow { color: #2c5d9c; font-size: 11px; margin-left: 4px; }
.mono { font-family: monospace; font-size: 12px; }
.status-badge {
  display: inline-block; padding: 2px 10px; border-radius: 3px;
  background: #f5a623; color: #fff; font-size: 11px; font-weight: 600;
}
.grid3 { display: grid; grid-template-columns: 1fr 1fr 1fr; gap: 12px; }
label.chk { flex-direction: row; align-items: center; gap: 8px; color: #1f2328; }
label.chk input { width: auto; }
</style>