<template>
  <div>
    <div class="page-title-bar">
      <div class="page-title">
        <span class="info-icon">ⓘ</span>
        <span>Договоры</span>
        <span class="refresh" @click="load" title="Обновить">↻</span>
      </div>
      <div class="page-actions">
        <button class="btn primary icon-btn" @click="openCreate">
          <span class="plus">+</span> Договор
        </button>
        <button class="btn" @click="showFilter = !showFilter">Фильтр</button>
        <input v-model="search" class="search-input" placeholder="Номер или комментарий" />
        <div class="counter" :class="{ active: selected.size > 0 }">{{ selected.size }}</div>
        <select class="mini-select" :disabled="selected.size === 0">
          <option>Изменить</option>
          <option>Удалить</option>
        </select>
        <select class="mini-select" v-model="filterStatus">
          <option value="">Статус</option>
          <option value="unpaid">Не оплачен</option>
          <option value="partial">Частично оплачен</option>
          <option value="paid">Оплачен</option>
        </select>
        <button class="btn icon-only" @click="printList" title="Печать">🖨</button>
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
          <label>Контрагент</label>
          <select v-model="filterCustomer">
            <option value="">Все</option>
            <option v-for="c in customers" :key="c.id" :value="c.id">{{ c.name }}</option>
          </select>
        </div>
        <div class="filter-field">
          <label>Период с</label>
          <input v-model="filterDateFrom" type="date" />
        </div>
        <div class="filter-field">
          <label>Период по</label>
          <input v-model="filterDateTo" type="date" />
        </div>
      </div>
    </div>

    <div v-if="error" class="error-box">{{ error }}</div>

    <table class="ms-table">
      <thead>
        <tr>
          <th class="chk-col"><input type="checkbox" :checked="allChecked" @change="toggleAll" /></th>
          <th style="width:160px" @click="sortBy('number')">
            Номер
            <span v-if="sortKey === 'number'" class="sort-arrow">{{ sortDir === 'asc' ? '↑' : '↓' }}</span>
          </th>
          <th style="width:80px">Код</th>
          <th style="width:130px" @click="sortBy('doc_date')">
            Время
            <span v-if="sortKey === 'doc_date'" class="sort-arrow">{{ sortDir === 'asc' ? '↑' : '↓' }}</span>
          </th>
          <th>Контрагент</th>
          <th style="width:180px">Организация</th>
          <th class="num" style="width:120px" @click="sortBy('amount')">
            Сумма
            <span v-if="sortKey === 'amount'" class="sort-arrow">{{ sortDir === 'asc' ? '↑' : '↓' }}</span>
          </th>
          <th class="num" style="width:110px">Оплачено</th>
          <th class="num" style="width:120px">Выполнено</th>
          <th style="width:110px">Отправлено</th>
          <th style="width:110px">Напечатано</th>
          <th>Комментарий</th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="loading"><td colspan="12" class="muted" style="text-align:center;padding:24px">Загрузка...</td></tr>
        <tr v-else-if="filtered.length === 0"><td colspan="12" class="muted" style="text-align:center;padding:24px">Нет договоров</td></tr>
        <tr
          v-else
          v-for="c in paginated"
          :key="c.id"
          class="clickable"
          :class="{ selected: selected.has(c.id), 'row-warning': isWarning(c) }"
          @click="openEdit(c)"
        >
          <td class="chk-col" @click.stop>
            <input type="checkbox" :checked="selected.has(c.id)" @change="toggleSelect(c.id)" />
          </td>
          <td>{{ c.number }}</td>
          <td class="muted mono">{{ c.code || shortId(c.id) }}</td>
          <td class="muted">{{ formatDate(c.doc_date) }}</td>
          <td>{{ customerName(c.customer_id) }}</td>
          <td class="muted">{{ organizationName(c.organization_id) }}</td>
          <td class="num">{{ formatMoney(c.amount) }}</td>
          <td class="num muted">{{ formatMoney(c.paid) }}</td>
          <td class="num">{{ formatMoney(c.fulfilled) }}</td>
          <td class="muted">{{ c.sent_at ? 'да' : '' }}</td>
          <td class="muted">{{ c.printed_at ? 'да' : '' }}</td>
          <td class="muted">{{ c.comment || '' }}</td>
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
        <span>Сумма: {{ formatMoney(totalAmount) }}</span>
        <span>Оплачено: {{ formatMoney(totalPaid) }}</span>
        <span>Выполнено: {{ formatMoney(totalFulfilled) }}</span>
      </div>
    </div>

    <!-- Модалка -->
    <div v-if="modal.open" class="modal-backdrop" @click.self="closeModal">
      <form class="card modal big" @submit.prevent="onSubmit">
        <h2>{{ modal.isEdit ? 'Карточка договора' : 'Новый договор' }}</h2>
        <div v-if="modal.error" class="error-box">{{ modal.error }}</div>

        <div class="grid2">
          <label>Номер*
            <input v-model="form.number" required />
          </label>
          <label>Код (внешний)
            <input v-model="form.code" />
          </label>
        </div>

        <div class="grid2">
          <label>Дата
            <input v-model="form.doc_date" type="date" />
          </label>
          <label>Контрагент
            <select v-model="form.customer_id">
              <option value="">— не выбран —</option>
              <option v-for="c in customers" :key="c.id" :value="c.id">{{ c.name }}</option>
            </select>
          </label>
        </div>

        <div class="grid2">
          <label>Организация
            <select v-model="form.organization_id">
              <option value="">— не выбрана —</option>
              <option v-for="o in organizations" :key="o.id" :value="o.id">{{ o.name }}</option>
            </select>
          </label>
          <label>Валюта
            <input v-model="form.currency" maxlength="3" />
          </label>
        </div>

        <div class="grid3">
          <label>Сумма
            <input v-model.number="form.amount" type="number" step="0.01" min="0" />
          </label>
          <label>Оплачено
            <input v-model.number="form.paid" type="number" step="0.01" min="0" />
          </label>
          <label>Выполнено
            <input v-model.number="form.fulfilled" type="number" step="0.01" min="0" />
          </label>
        </div>

        <label>Комментарий
          <textarea v-model="form.comment" rows="2"></textarea>
        </label>

        <label class="chk">
          <input type="checkbox" v-model="form.archived" /> Архивный
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
  listContracts, createContract, updateContract, deleteContract,
  type Contract,
} from '../api/contracts'
import { listCustomers, type Customer } from '../api/customers'
import { listOrganizations, type Organization } from '../api/suppliers'
import { apiErrorMessage } from '../api/client'

const items = ref<Contract[]>([])
const customers = ref<Customer[]>([])
const organizations = ref<Organization[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const showFilter = ref(false)
const page = ref(1)
const perPage = 100

const search = ref('')
const includeArchived = ref(false)
const filterCustomer = ref('')
const filterDateFrom = ref('')
const filterDateTo = ref('')
const filterStatus = ref('')

const sortKey = ref<'number' | 'doc_date' | 'amount'>('doc_date')
const sortDir = ref<'asc' | 'desc'>('desc')
const selected = ref<Set<string>>(new Set())

const modal = reactive({
  open: false, isEdit: false, saving: false,
  error: null as string | null, editingId: '' as string,
})

const form = reactive({
  number: '', code: '', doc_date: '',
  customer_id: '', organization_id: '',
  amount: 0, currency: 'RUB', paid: 0, fulfilled: 0,
  comment: '', archived: false,
})

function shortId(id: string) { return id.slice(0, 6) }
function formatDate(s: string) {
  const d = new Date(s)
  return d.toLocaleDateString('ru-RU') + ' ' + d.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })
}
function formatMoney(n: number) {
  return n.toLocaleString('ru-RU', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}
function customerName(id?: string) {
  return id ? (customers.value.find((c) => c.id === id)?.name ?? '—') : '—'
}
function organizationName(id?: string) {
  return id ? (organizations.value.find((o) => o.id === id)?.name ?? '—') : '—'
}
function isWarning(c: Contract) {
  return c.amount > 0 && c.paid === 0 && c.fulfilled === 0
}

function sortBy(k: typeof sortKey.value) {
  if (sortKey.value === k) sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  else { sortKey.value = k; sortDir.value = 'desc' }
  page.value = 1
}

const filtered = computed(() => {
  let rows = items.value
  if (search.value) {
    const q = search.value.toLowerCase()
    rows = rows.filter((c) =>
      c.number.toLowerCase().includes(q) ||
      (c.comment ?? '').toLowerCase().includes(q)
    )
  }
  if (filterCustomer.value) rows = rows.filter((c) => c.customer_id === filterCustomer.value)
  if (filterDateFrom.value) {
    const t = new Date(filterDateFrom.value).getTime()
    rows = rows.filter((c) => new Date(c.doc_date).getTime() >= t)
  }
  if (filterDateTo.value) {
    const t = new Date(filterDateTo.value).getTime() + 86400000 - 1
    rows = rows.filter((c) => new Date(c.doc_date).getTime() <= t)
  }
  if (filterStatus.value === 'unpaid')    rows = rows.filter((c) => c.paid === 0)
  if (filterStatus.value === 'partial')   rows = rows.filter((c) => c.paid > 0 && c.paid < c.amount)
  if (filterStatus.value === 'paid')      rows = rows.filter((c) => c.paid >= c.amount && c.amount > 0)

  return [...rows].sort((a, b) => {
    const av = a[sortKey.value]; const bv = b[sortKey.value]
    if (typeof av === 'number' && typeof bv === 'number') {
      return sortDir.value === 'asc' ? av - bv : bv - av
    }
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

const totalAmount = computed(() => filtered.value.reduce((s, c) => s + c.amount, 0))
const totalPaid = computed(() => filtered.value.reduce((s, c) => s + c.paid, 0))
const totalFulfilled = computed(() => filtered.value.reduce((s, c) => s + c.fulfilled, 0))

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
    const [contracts, custs, orgs] = await Promise.all([
      listContracts(includeArchived.value),
      listCustomers(false).catch(() => []),
      listOrganizations().catch(() => []),
    ])
    items.value = contracts
    customers.value = custs
    organizations.value = orgs
    page.value = 1
  } catch (e) {
    error.value = apiErrorMessage(e)
  } finally {
    loading.value = false
  }
}

function openCreate() {
  Object.assign(form, {
    number: 'ДГ-' + new Date().getTime().toString().slice(-6),
    code: '', doc_date: new Date().toISOString().slice(0, 10),
    customer_id: '', organization_id: '',
    amount: 0, currency: 'RUB', paid: 0, fulfilled: 0,
    comment: '', archived: false,
  })
  modal.open = true; modal.isEdit = false; modal.error = null; modal.editingId = ''
}

function openEdit(c: Contract) {
  Object.assign(form, {
    number: c.number,
    code: c.code ?? '',
    doc_date: c.doc_date.slice(0, 10),
    customer_id: c.customer_id ?? '',
    organization_id: c.organization_id ?? '',
    amount: c.amount, currency: c.currency,
    paid: c.paid, fulfilled: c.fulfilled,
    comment: c.comment ?? '', archived: c.archived,
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
      await updateContract(modal.editingId, payload)
    } else {
      await createContract(payload)
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
  if (!confirm(`Удалить «${form.number}»?`)) return
  try {
    await deleteContract(modal.editingId)
    closeModal()
    await load()
  } catch (e) { modal.error = apiErrorMessage(e) }
}

function clearFilters() {
  search.value = ''
  filterCustomer.value = ''
  filterDateFrom.value = ''
  filterDateTo.value = ''
  filterStatus.value = ''
  page.value = 1
}

function printList() { window.print() }

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
  background: #fff; color: #1f2328; max-width: 150px;
}
.mini-select:disabled { color: #8c959f; background: #f6f8fa; }
.btn.icon-only { padding: 6px 10px; font-size: 15px; }
tr.clickable { cursor: pointer; }
tr.clickable.selected { background: #eef4ff; }
tr.clickable.row-warning { background: #fff8e1; }
tr.clickable.row-warning.selected { background: #fceec9; }
.mono { font-family: monospace; font-size: 12px; }
.sort-arrow { color: #2c5d9c; font-size: 11px; margin-left: 4px; }
.grid3 { display: grid; grid-template-columns: 1fr 1fr 1fr; gap: 12px; }
label.chk { flex-direction: row; align-items: center; gap: 8px; color: #1f2328; }
label.chk input { width: auto; }
</style>