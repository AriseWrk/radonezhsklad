<template>
  <div>
    <div class="ms-title">
      <button class="ms-help" title="Справка"><MsIcon name="help" :size="14" /></button>
      <span>{{ pageTitle }}</span>
      <button class="ms-refresh" @click="load" title="Обновить"><MsIcon name="refresh" :size="14" /></button>
    </div>

    <div class="ms-toolbar">
      <MsButton variant="primary" icon="plus" @click="openCreate">{{ docLabel }}</MsButton>
      <MsButton icon="filter" @click="showFilter = !showFilter">Фильтр</MsButton>
      <input v-model="search" class="ms-input" placeholder="Номер или комментарий" />
      <div class="ms-counter">{{ filtered.length }}</div>
      <select class="ms-select ms-select-bulk">
        <option>Изменить</option>
      </select>
      <select class="ms-select" v-model="filterStatus">
        <option value="">Статус</option>
        <option value="draft">Черновик</option>
        <option value="posted">Проведён</option>
        <option value="cancelled">Отменён</option>
      </select>
      <MsButton icon="print">Печать</MsButton>
      <select class="ms-select ms-select-edo">
        <option>ЭДО</option>
      </select>
      <MsButton variant="icon" icon="gear" title="Настройки" />
    </div>

    <div v-if="showFilter" class="ms-filter">
      <div class="filter-actions">
        <MsButton variant="green" @click="applyFilters">Найти</MsButton>
        <MsButton @click="clearFilters">Очистить</MsButton>
        <button class="filter-ico" title="Сохранить фильтр"><MsIcon name="save" :size="14" /></button>
        <button class="filter-ico" title="Настройка фильтра"><MsIcon name="gear" :size="14" /></button>
      </div>
      <div class="filter-grid">
        <div class="filter-field">
          <label class="filter-label"><span class="dot"></span>Период</label>
          <div class="filter-range">
            <select class="range-quick" v-model="filterDateQuick">
              <option value="">вс</option>
              <option value="today">сег</option>
              <option value="week">нед</option>
              <option value="month">мес</option>
            </select>
            <input v-model="filterDateFrom" type="date" />
            <span class="range-dash">—</span>
            <input v-model="filterDateTo" type="date" />
          </div>
        </div>
        <div class="filter-field">
          <label class="filter-label"><span class="dot"></span>Входящий номер</label>
          <input v-model="filterIncomingNumber" />
        </div>
        <div class="filter-field">
          <label class="filter-label"><span class="dot"></span>Входящая дата</label>
          <div class="filter-range">
            <select class="range-quick" v-model="filterIncDateQuick">
              <option value="">вс</option>
              <option value="today">сег</option>
              <option value="week">нед</option>
              <option value="month">мес</option>
            </select>
            <input v-model="filterIncomingDateFrom" type="date" />
            <span class="range-dash">—</span>
            <input v-model="filterIncomingDateTo" type="date" />
          </div>
        </div>
        <div class="filter-field">
          <label class="filter-label"><span class="dot"></span>Оплата</label>
          <div class="filter-range">
            <input v-model="filterPaidMin" type="number" placeholder="от" />
            <span class="range-dash">—</span>
            <input v-model="filterPaidMax" type="number" placeholder="до" />
          </div>
        </div>
        <div class="filter-field">
          <label class="filter-label"><span class="dot"></span>Товар или группа</label>
          <select disabled><option>—</option></select>
        </div>
        <div class="filter-field">
          <label class="filter-label"><span class="dot"></span>Склад</label>
          <MsCombo v-model="filterWarehouse" :items="warehouses" placeholder="—" />
        </div>
        <div class="filter-field">
          <label class="filter-label"><span class="dot"></span>Контрагент</label>
          <MsCombo v-model="filterSupplier" :items="suppliers" placeholder="—" />
        </div>
        <div class="filter-field">
          <label class="filter-label"><span class="dot"></span>Организация</label>
          <MsCombo v-model="filterOrganization" :items="organizations" placeholder="—" />
        </div>
        <div class="filter-field">
          <label class="filter-label"><span class="dot"></span>Статус</label>
          <select v-model="filterStatus">
            <option value="">—</option>
            <option value="draft">Черновик</option>
            <option value="posted">Проведён</option>
            <option value="cancelled">Отменён</option>
          </select>
        </div>
        <div class="filter-field">
          <label class="filter-label"><span class="dot"></span>Проведено</label>
          <select v-model="filterPosted">
            <option value="">—</option>
            <option value="posted">Да</option>
            <option value="draft">Нет</option>
          </select>
        </div>
        <div class="filter-field">
          <label class="filter-label"><span class="dot"></span>Напечатано</label>
          <select v-model="filterPrinted">
            <option value="">—</option>
            <option value="yes">Да</option>
            <option value="no">Нет</option>
          </select>
        </div>
        <div class="filter-field">
          <label class="filter-label"><span class="dot"></span>Отправлено</label>
          <select v-model="filterSent">
            <option value="">—</option>
            <option value="yes">Да</option>
            <option value="no">Нет</option>
          </select>
        </div>
      </div>
    </div>
    <div v-if="error" class="error-box">{{ error }}</div>

    <table class="ms-table">
      <thead>
        <tr>
          <th class="chk-col"><input type="checkbox" :checked="allChecked" @change="toggleAll" /></th>
          <th style="width:80px">№</th>
          <th style="width:130px">Время</th>
          <th>На склад</th>
          <th>Контрагент</th>
          <th>Организация</th>
          <th class="num">Сумма</th>
          <th class="num">Оплачено</th>
          <th style="width:110px">Входящая дата</th>
          <th>Входящий номер</th>
          <th>Отправлено</th>
          <th>Напечатано</th>
          <th>Комментарий</th>
          <th class="actions-col"></th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="loading">
          <td colspan="14" class="muted" style="text-align:center;padding:24px">Загрузка...</td>
        </tr>
        <tr v-else-if="filtered.length === 0">
          <td colspan="14" class="muted" style="text-align:center;padding:24px">Нет данных</td>
        </tr>
        <tr v-else v-for="d in paginated" :key="d.id" class="clickable" @click="openEdit(d)">
          <td class="chk-col" @click.stop>
            <input type="checkbox" :checked="selected.has(d.id)" @change="toggleSelect(d.id)" />
          </td>
          <td class="mono link">{{ shortNumber(d) }}</td>
          <td class="muted">{{ formatDate(d.created_at) }}</td>
          <td>{{ warehouseName(d.warehouse_id) }}</td>
          <td>{{ supplierName(d.supplier_id) }}</td>
          <td>{{ organizationName(d.organization_id) }}</td>
          <td class="num"><strong>{{ formatMoney(d.total ?? 0) }}</strong></td>
          <td class="num">{{ formatMoney(d.paid_amount ?? 0) }}</td>
          <td class="muted">{{ d.incoming_date ? formatShortDate(d.incoming_date) : '—' }}</td>
          <td class="muted">{{ d.incoming_number || '—' }}</td>
          <td class="muted">{{ d.sent_at ? 'да' : '—' }}</td>
          <td class="muted">{{ d.printed_at ? 'да' : '—' }}</td>
          <td class="muted">{{ d.comment || '—' }}</td>
          <td class="actions-col" @click.stop>
            <button v-if="d.status === 'draft'" class="act-link" @click="onPost(d)">Провести</button>
            <button v-if="d.status === 'posted'" class="act-link act-danger" @click="onCancel(d)">Отменить</button>
            <span class="act-status">{{ statusLabel(d.status) }}</span>
          </td>
        </tr>
      </tbody>
    </table>

    <div class="ms-footer">
      <div class="ms-pager">
        <button :disabled="page === 1" @click="page = 1" title="Первая">|‹</button>
        <button :disabled="page === 1" @click="page--" title="Назад">‹</button>
        <span class="ms-pager-info">{{ rangeFrom }}-{{ rangeTo }} из {{ filtered.length }}</span>
        <button :disabled="rangeTo >= filtered.length" @click="page++" title="Вперёд">›</button>
        <button :disabled="rangeTo >= filtered.length" @click="page = Math.ceil(filtered.length / perPage)" title="Последняя">›|</button>
      </div>
    </div>

    <!-- Модалка создания -->
    <div v-if="showCreate" class="modal-backdrop" @click.self="closeCreate">
      <form class="card modal big" @submit.prevent="onCreate">
        <h2>Новый документ: {{ meta.one }}</h2>
        <div v-if="createError" class="error-box">{{ createError }}</div>

        <div class="grid2">
          <label v-if="needsWarehouse">Склад*
            <MsCombo v-model="form.warehouse_id" :items="warehouses" placeholder="— выберите —" />
          </label>
          <label v-if="needsSupplier">Поставщик*
            <MsCombo v-model="form.supplier_id" :items="suppliers" placeholder="— выберите —" />
          </label>
        </div>

        <div class="grid2">
          <label>Организация
            <MsCombo v-model="form.organization_id" :items="organizations" placeholder="— по умолчанию —" />
          </label>
          <label>Входящий номер
            <input v-model="form.incoming_number" placeholder="Номер от поставщика" />
          </label>
        </div>

        <div class="grid2">
          <label>Входящая дата
            <input v-model="form.incoming_date" type="date" />
          </label>
          <label>Комментарий
            <input v-model="form.comment" />
          </label>
        </div>

        <div class="items-section">
          <div class="items-head">
            <span>Позиции</span>
            <button type="button" class="btn-link" @click="addItem">+ Добавить</button>
          </div>
          <div v-for="(it, idx) in form.items" :key="idx" class="item-row">
            <MsCombo v-model="it.product_id" :items="products" placeholder="— товар —" @update:model-value="(v: string) => onPickProduct(idx, v)" />
            <input v-model.number="it.quantity" type="number" step="0.001" min="0.001" placeholder="Кол-во" required />
            <input v-model.number="it.price" type="number" step="0.01" min="0" placeholder="Цена" />
            <button type="button" class="btn-link danger-text" @click="removeItem(idx)">×</button>
          </div>
        </div>

        <div class="modal-actions">
          <button type="button" @click="closeCreate">Отмена</button>
          <button class="primary" type="submit" :disabled="saving">
            {{ saving ? 'Сохранение...' : 'Создать' }}
          </button>
        </div>
      </form>
    </div>

    <!-- Модалка деталей -->
    <div v-if="detail" class="modal-backdrop" @click.self="detail = null">
      <div class="doc-card">
        <div class="dc-head">
          <div class="dc-head-left">
            <button class="dc-back" @click="detail = null" title="Закрыть"><MsIcon name="chevron-left" :size="14" /></button>
            <span class="dc-title">{{ docLabel }} № {{ shortNumber(detail) }}</span>
            <span class="dc-status" :class="detail.status">{{ statusLabel(detail.status) }}</span>
          </div>
          <div class="dc-head-right">
            <MsButton v-if="detail.status === 'draft'" variant="green" @click="onPost(detail)">Провести</MsButton>
            <MsButton v-if="detail.status === 'posted'" variant="danger" @click="onCancel(detail)">Отменить</MsButton>
            <MsButton icon="print">Печать</MsButton>
            <button class="dc-close" @click="detail = null" title="Закрыть">×</button>
          </div>
        </div>

        <div class="dc-fields">
          <div class="dc-field">
            <span class="dc-lbl"><span class="dot"></span>Организация</span>
            <div class="dc-val">{{ organizationName(detail.organization_id) }}</div>
          </div>
          <div class="dc-field" v-if="needsWarehouse || detail.warehouse_id">
            <span class="dc-lbl"><span class="dot"></span>Склад</span>
            <div class="dc-val">{{ warehouseName(detail.warehouse_id) }}</div>
          </div>
          <div class="dc-field" v-if="needsSupplier || detail.supplier_id">
            <span class="dc-lbl"><span class="dot"></span>Контрагент</span>
            <div class="dc-val">{{ supplierName(detail.supplier_id) }}</div>
          </div>
          <div class="dc-field">
            <span class="dc-lbl"><span class="dot"></span>Входящий номер</span>
            <div class="dc-val">{{ detail.incoming_number || '—' }}</div>
          </div>
          <div class="dc-field">
            <span class="dc-lbl"><span class="dot"></span>Входящая дата</span>
            <div class="dc-val">{{ detail.incoming_date ? formatShortDate(detail.incoming_date) : '—' }}</div>
          </div>
          <div class="dc-field">
            <span class="dc-lbl"><span class="dot"></span>Оплачено</span>
            <div class="dc-val">{{ formatMoney(detail.paid_amount ?? 0) }} ₽</div>
          </div>
          <div class="dc-field dc-field-wide">
            <span class="dc-lbl"><span class="dot"></span>Комментарий</span>
            <div class="dc-val">{{ detail.comment || '—' }}</div>
          </div>
        </div>

        <div class="dc-items">
          <table class="dc-table">
            <thead>
              <tr>
                <th style="width:40px">№</th>
                <th>Товар</th>
                <th class="num" style="width:100px">Кол-во</th>
                <th class="num" style="width:120px">Цена</th>
                <th class="num" style="width:130px">Сумма</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(it, i) in (detail.items || [])" :key="it.id">
                <td class="muted">{{ i + 1 }}</td>
                <td>{{ productName(it.product_id) }}</td>
                <td class="num">{{ formatQty(it.quantity) }}</td>
                <td class="num">{{ formatMoney(it.price) }}</td>
                <td class="num">{{ formatMoney(it.quantity * it.price) }}</td>
              </tr>
              <tr v-if="!detail.items || detail.items.length === 0">
                <td colspan="5" class="muted" style="text-align:center;padding:20px">Нет позиций</td>
              </tr>
            </tbody>
            <tfoot>
              <tr>
                <td colspan="4" class="num dc-total-lbl">Итого:</td>
                <td class="num dc-total-val">{{ formatMoney(detail.total ?? 0) }} ₽</td>
              </tr>
            </tfoot>
          </table>
        </div>

        <div class="dc-foot">
          <div class="dc-foot-left muted">Создан: {{ formatDate(detail.created_at) }}</div>
          <div class="dc-foot-right"><MsButton @click="detail = null">Закрыть</MsButton></div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import {
  listDocuments, createDocument, postDocument, cancelDocument, getDocument, type Document,
} from '../api/documents'
import { listWarehouses, type Warehouse } from '../api/warehouses'
import { listProducts, type Product } from '../api/products'
import { listSuppliers, listOrganizations, type Supplier, type Organization } from '../api/suppliers'
import { apiErrorMessage } from '../api/client'
import MsButton from '../components/MsButton.vue'
import MsIcon from '../components/MsIcon.vue'
import MsCombo from '../components/MsCombo.vue'

type PType = 'receipt' | 'purchase_order' | 'invoice_in' | 'purchase_return' | 'facture_in'
const route = useRoute()
const docType = (route.meta as any).docType as PType | undefined

const titles: Record<PType, { page: string; label: string; one: string }> = {
  receipt:         { page: 'Приёмки', label: 'Приёмка', one: 'Приёмка' },
  purchase_order:  { page: 'Заказы поставщикам',   label: 'Заказ',    one: 'Заказ поставщику' },
  invoice_in:      { page: 'Счета поставщиков',    label: 'Счёт',     one: 'Счёт поставщика' },
  purchase_return: { page: 'Возвраты поставщикам', label: 'Возврат',  one: 'Возврат поставщику' },
  facture_in:      { page: 'Счета-фактуры полученные', label: 'Счёт-фактура', one: 'Счёт-фактура полученный' },
}
const meta = computed(() => (docType && titles[docType]) || { page: 'Закупки', label: 'Документ', one: 'Документ' })
const pageTitle = computed(() => meta.value.page)
const docLabel = computed(() => meta.value.label)

// Требуется ли склад / поставщик
const needsWarehouse = computed(() => docType === 'purchase_return' || docType === 'receipt')
const needsSupplier = computed(() => docType === 'purchase_order' || docType === 'invoice_in' || docType === 'purchase_return' || docType === 'receipt')

const items = ref<Document[]>([])
const warehouses = ref<Warehouse[]>([])
const products = ref<Product[]>([])
const suppliers = ref<Supplier[]>([])
const organizations = ref<Organization[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const showFilter = ref(false)
const page = ref(1)
const perPage = 50

const search = ref('')
const filterStatus = ref('')
const filterWarehouse = ref('')
const filterSupplier = ref('')
const filterOrganization = ref('')
const filterDateQuick = ref('')
const filterIncDateQuick = ref('')
const filterDateFrom = ref('')
const filterDateTo = ref('')
const filterIncomingNumber = ref('')
const filterIncomingDateFrom = ref('')
const filterIncomingDateTo = ref('')
const filterPaidMin = ref('')
const filterPaidMax = ref('')
const filterPosted = ref('')
const filterPrinted = ref('')
const filterSent = ref('')

const selected = ref<Set<string>>(new Set())

const showCreate = ref(false)
const saving = ref(false)
const createError = ref<string | null>(null)
const form = reactive({
  warehouse_id: '',
  supplier_id: '',
  organization_id: '',
  incoming_number: '',
  incoming_date: '',
  comment: '',
  items: [] as Array<{ product_id: string; quantity: number; price: number }>,
})

const detail = ref<Document | null>(null)

function warehouseName(id?: string) {
  if (!id) return '—'
  return warehouses.value.find((w) => w.id === id)?.name ?? id.slice(0, 8)
}
function supplierName(id?: string) {
  if (!id) return '—'
  return suppliers.value.find((s) => s.id === id)?.name ?? '—'
}
function organizationName(id?: string) {
  if (!id) return organizations.value.find((o) => o.is_default)?.name ?? '—'
  return organizations.value.find((o) => o.id === id)?.name ?? '—'
}
function productName(id: string) {
  return products.value.find((p) => p.id === id)?.name ?? id.slice(0, 8)
}
function statusLabel(s: string) {
  return { draft: 'Черновик', posted: 'Проведён', cancelled: 'Отменён' }[s] ?? s
}
function shortNumber(d: Document) {
  const n = (d.number ?? '').trim()
  if (!n) return '—'
  // МС-стиль: длинные числовые номера (1790014183) → последние 5 цифр
  if (/^\d+$/.test(n) && n.length > 5) return n.slice(-5)
  // прим-1234567 → 1234567
  const parts = n.split('-')
  const last = parts[parts.length - 1]
  if (/^\d+$/.test(last) && last.length > 5) return last.slice(-5)
  return last || n
}
function formatDate(s: string) {
  const d = new Date(s)
  return d.toLocaleDateString('ru-RU') + ' ' + d.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })
}
function formatShortDate(s: string) {
  return new Date(s).toLocaleDateString('ru-RU')
}
function formatMoney(n: number) {
  return n.toLocaleString('ru-RU', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}
function formatQty(n: number) {
  return n.toLocaleString('ru-RU', { maximumFractionDigits: 3 })
}

const filtered = computed(() => {
  let rows = items.value
  if (search.value) {
    const q = search.value.toLowerCase()
    rows = rows.filter((d) =>
      (d.number ?? '').toLowerCase().includes(q) ||
      (d.comment ?? '').toLowerCase().includes(q)
    )
  }
  return rows
})
const paginated = computed(() => {
  const from = (page.value - 1) * perPage
  return filtered.value.slice(from, from + perPage)
})

const rangeFrom = computed(() => filtered.value.length === 0 ? 0 : (page.value - 1) * perPage + 1)
const rangeTo = computed(() => Math.min(page.value * perPage, filtered.value.length))

const totalSum = computed(() => filtered.value.reduce((s, d) => s + (d.total ?? 0), 0))
const totalPaid = computed(() => filtered.value.reduce((s, d) => s + (d.paid_amount ?? 0), 0))

const allChecked = computed(() =>
  paginated.value.length > 0 && paginated.value.every((d) => selected.value.has(d.id))
)
function toggleAll(e: Event) {
  const on = (e.target as HTMLInputElement).checked
  selected.value = on ? new Set(paginated.value.map((d) => d.id)) : new Set()
}
function toggleSelect(id: string) {
  const s = new Set(selected.value)
  s.has(id) ? s.delete(id) : s.add(id)
  selected.value = s
}

function applyFilters() {
  page.value = 1
  load()
}
function clearFilters() {
  search.value = ''
  filterStatus.value = ''
  filterWarehouse.value = ''
  filterSupplier.value = ''
  filterOrganization.value = ''
  filterDateFrom.value = ''
  filterDateTo.value = ''
  filterIncomingNumber.value = ''
  filterIncomingDateFrom.value = ''
  filterIncomingDateTo.value = ''
  filterPaidMin.value = ''
  filterPaidMax.value = ''
  filterPosted.value = ''
  filterPrinted.value = ''
  filterSent.value = ''
  page.value = 1
  load()
}

async function load() {
  loading.value = true
  error.value = null
  try {
    items.value = await listDocuments({
      type: docType,
      status: filterStatus.value || undefined,
      warehouse_id: filterWarehouse.value || undefined,
      supplier_id: filterSupplier.value || undefined,
      organization_id: filterOrganization.value || undefined,
      date_from: filterDateFrom.value || undefined,
      date_to: filterDateTo.value || undefined,
      incoming_number: filterIncomingNumber.value || undefined,
      incoming_date_from: filterIncomingDateFrom.value || undefined,
      incoming_date_to: filterIncomingDateTo.value || undefined,
      paid_min: filterPaidMin.value !== '' ? Number(filterPaidMin.value) : undefined,
      paid_max: filterPaidMax.value !== '' ? Number(filterPaidMax.value) : undefined,
      printed: filterPrinted.value ? filterPrinted.value === 'yes' : undefined,
      sent: filterSent.value ? filterSent.value === 'yes' : undefined,
    })
  } catch (e) {
    error.value = apiErrorMessage(e)
  } finally {
    loading.value = false
  }
}

function openCreate() {
  if (!docType) { createError.value = 'Не указан тип документа'; return }
  form.warehouse_id = needsWarehouse.value ? (warehouses.value[0]?.id ?? '') : ''
  form.supplier_id = ''
  form.organization_id = organizations.value.find((o) => o.is_default)?.id ?? ''
  form.incoming_number = ''
  form.incoming_date = ''
  form.comment = ''
  form.items = [{ product_id: '', quantity: 1, price: 0 }]
  createError.value = null
  showCreate.value = true
}
function closeCreate() { showCreate.value = false }
function addItem() { form.items.push({ product_id: '', quantity: 1, price: 0 }) }
function removeItem(i: number) { form.items.splice(i, 1) }
function onPickProduct(idx: number, productId: string) {
  if (!productId) return
  const p = products.value.find((x) => x.id === productId)
  if (p && p.price != null && (!form.items[idx].price || form.items[idx].price === 0)) {
    form.items[idx].price = p.price
  }
}

async function onCreate() {
  createError.value = null
  saving.value = true
  try {
    if (form.items.length === 0) throw new Error('Добавьте хотя бы одну позицию')
    if (form.items.some((it) => !it.product_id || it.quantity <= 0)) throw new Error('Заполните все позиции')
    if (needsWarehouse.value && !form.warehouse_id) throw new Error('Выберите склад')
    if (needsSupplier.value && !form.supplier_id) throw new Error('Выберите поставщика')

    await createDocument({
      type: docType as any,
      warehouse_id: form.warehouse_id || undefined,
      supplier_id: form.supplier_id || undefined,
      organization_id: form.organization_id || undefined,
      incoming_number: form.incoming_number || undefined,
      incoming_date: form.incoming_date || undefined,
      comment: form.comment || undefined,
      items: form.items,
    })
    closeCreate()
    await load()
  } catch (e) {
    createError.value = apiErrorMessage(e)
  } finally {
    saving.value = false
  }
}

async function onPost(d: Document) {
  if (!confirm(`Провести ${meta.value.one.toLowerCase()} ${d.number}?`)) return
  try {
    await postDocument(d.id)
    await load()
    if (detail.value && detail.value.id === d.id) detail.value = null
  } catch (e) { error.value = apiErrorMessage(e) }
}

async function onCancel(d: Document) {
  if (!confirm(`Отменить ${meta.value.one.toLowerCase()} ${d.number}?`)) return
  try {
    await cancelDocument(d.id)
    await load()
    if (detail.value && detail.value.id === d.id) detail.value = null
  } catch (e) { error.value = apiErrorMessage(e) }
}

async function openEdit(d: Document) {
  try { detail.value = await getDocument(d.id) } catch (e) { error.value = apiErrorMessage(e) }
}

function applyQuick(kind: 'today' | 'week' | 'month', fromRef: any, toRef: any) {
  const now = new Date()
  const y = now.getFullYear(), m = now.getMonth(), d = now.getDate()
  let from: Date, to = new Date(now)
  if (kind === 'today') from = new Date(y, m, d)
  else if (kind === 'week') from = new Date(y, m, d - 7)
  else from = new Date(y, m - 1, d)
  fromRef.value = from.toISOString().slice(0, 10)
  toRef.value = to.toISOString().slice(0, 10)
}
watch(filterDateQuick, (v) => { if (v) applyQuick(v as any, filterDateFrom, filterDateTo) })
watch(filterIncDateQuick, (v) => { if (v) applyQuick(v as any, filterIncomingDateFrom, filterIncomingDateTo) })
onMounted(async () => {
  try {
    const [w, p, s, o] = await Promise.all([
      listWarehouses(), listProducts(true), listSuppliers(), listOrganizations(),
    ])
    warehouses.value = w
    products.value = p
    suppliers.value = s
    organizations.value = o
  } catch { /* ignore */ }
  await load()
})
</script>

<style scoped>
.search-input {
  padding: 6px 10px;
  font-size: 13px;
  border: 1px solid #d0d7de;
  border-radius: 4px;
  background: #fff;
  color: #1f2328;
  min-width: 200px;
}
.users-table .chk-col { width: 32px; }
tr.clickable { cursor: pointer; }

/* === МойСклад-стиль === */
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
.ms-footer {
  display: flex; align-items: center; justify-content: space-between;
  padding: 8px 4px; margin-top: 4px;
  font-size: 12px; color: #57606a;
}
.ms-pager { display: inline-flex; align-items: center; gap: 2px; }
.ms-pager button {
  min-width: 26px; height: 26px; padding: 0 6px;
  border: 1px solid #d0d7de; background: #fff; color: #57606a;
  border-radius: 3px; cursor: pointer; font-size: 12px;
  display: inline-flex; align-items: center; justify-content: center;
}
.ms-pager button:hover:not(:disabled) { background: #f6f8fa; }
.ms-pager button:disabled { opacity: .4; cursor: default; }
.ms-pager-info { margin: 0 6px; color: #57606a; }
.ms-select-bulk, .ms-select-edo { max-width: none; }
.filter-ico {
  width: 30px; height: 30px; padding: 0;
  display: inline-flex; align-items: center; justify-content: center;
  border: 1px solid #d0d7de; border-radius: 4px;
  background: #fff; color: #57606a; cursor: pointer;
}
.filter-ico:hover { background: #f6f8fa; }
.filter-range { display: flex; align-items: center; gap: 4px; }
.filter-range input { flex: 1; min-width: 0; }
.filter-range .range-quick {
  flex: 0 0 56px; height: 28px; padding: 0 4px; font-size: 12px;
  border: 1px solid #d0d7de; border-radius: 3px; background: #fff; color: #1f2328;
}
.filter-range .range-dash { color: #8c959f; font-size: 12px; }
.filter-grid .filter-field input,
.filter-grid .filter-field select {
  height: 28px; padding: 0 8px; font-size: 12px;
  border: 1px solid #d0d7de; border-radius: 3px;
  background: #fff; color: #1f2328;
}
.ms-filter .filter-grid { background: transparent; }
.ms-filter { background: #eef1f5; }
.act-link {
  border: none; background: transparent; padding: 0; cursor: pointer;
  font-size: 12px; color: #57606a; margin-right: 10px;
}
.act-link:hover { text-decoration: underline; }
.act-link.act-danger { color: #cf222e; }
.act-status { font-size: 12px; color: #57606a; }
.actions-col { text-align: right; white-space: nowrap; }
/* === Карточка документа (МС-стиль) === */
.doc-card {
  background: #fff; border-radius: 6px; width: 92vw; max-width: 1100px; max-height: 92vh;
  display: flex; flex-direction: column; overflow: hidden;
  box-shadow: 0 12px 40px rgba(0,0,0,.25);
}
.dc-head {
  display: flex; align-items: center; justify-content: space-between;
  padding: 10px 14px; border-bottom: 1px solid #eaeef2; gap: 10px;
}
.dc-head-left { display: flex; align-items: center; gap: 10px; min-width: 0; }
.dc-head-right { display: flex; align-items: center; gap: 6px; }
.dc-back, .dc-close {
  width: 26px; height: 26px; padding: 0; border: 1px solid #d0d7de; border-radius: 4px;
  background: #fff; color: #57606a; cursor: pointer;
  display: inline-flex; align-items: center; justify-content: center;
}
.dc-close { font-size: 18px; line-height: 1; }
.dc-back:hover, .dc-close:hover { background: #f6f8fa; }
.dc-title { font-size: 16px; font-weight: 600; color: #1f2328; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.dc-status {
  font-size: 12px; padding: 2px 8px; border-radius: 10px;
  background: #eef1f5; color: #57606a;
}
.dc-status.posted { background: #e6f4ea; color: #1a7f37; }
.dc-status.cancelled { background: #fde8e8; color: #cf222e; }

.dc-fields {
  display: grid; grid-template-columns: repeat(3, minmax(0,1fr));
  gap: 10px 24px; padding: 14px 16px; background: #fafbfc;
  border-bottom: 1px solid #eaeef2;
}
.dc-field { display: flex; flex-direction: column; gap: 3px; min-width: 0; }
.dc-field-wide { grid-column: 1 / -1; }
.dc-lbl { font-size: 11px; color: #57606a; display: flex; align-items: center; gap: 5px; }
.dc-lbl .dot { width: 6px; height: 6px; border-radius: 50%; background: #2c5d9c; flex-shrink: 0; }
.dc-val { font-size: 13px; color: #1f2328; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }

.dc-items { flex: 1; overflow: auto; padding: 0 16px; }
.dc-table { width: 100%; border-collapse: collapse; font-size: 13px; }
.dc-table thead th {
  position: sticky; top: 0; background: #fff; z-index: 1;
  color: #2c5d9c; font-weight: 500; font-size: 12px;
  padding: 8px 10px; text-align: left; border-bottom: 1px solid #d8dee4; white-space: nowrap;
}
.dc-table tbody td { padding: 7px 10px; border-bottom: 1px solid #eaeef2; vertical-align: middle; }
.dc-table tbody tr:hover { background: #f6f8fa; }
.dc-table .num { text-align: right; font-variant-numeric: tabular-nums; }
.dc-table .muted { color: #8c959f; }
.dc-table tfoot td { padding: 10px; border-top: 2px solid #d8dee4; font-weight: 600; }
.dc-total-lbl { color: #57606a; font-weight: 500; }
.dc-total-val { color: #1f2328; font-size: 14px; }

.dc-foot {
  display: flex; align-items: center; justify-content: space-between;
  padding: 10px 16px; border-top: 1px solid #eaeef2; background: #fafbfc;
}
.dc-foot-left { font-size: 12px; }
.dc-foot-right { display: flex; gap: 6px; }
</style>
