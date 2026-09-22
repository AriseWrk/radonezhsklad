<template>
  <div class="card-page">
    <!-- Верхняя панель -->
    <div class="toolbar">
      <button class="btn primary" @click="save" :disabled="saving || !canEdit">{{ saving ? 'Сохранение...' : 'Сохранить' }}</button>
      <button class="btn" @click="close">Закрыть</button>
      <button class="btn" @click="print">Печать</button>
      <button class="btn" @click="onSend" :disabled="!currentId || form.status !== 'posted' || !canEdit">Отправить</button>
      <button class="btn" @click="onCancel" :disabled="!currentId || form.status !== 'posted' || !canEdit">Отменить</button>
      <button class="btn danger" @click="onDelete" :disabled="(!currentId && !isNew) || (currentId !== null && form.status !== 'draft')">Удалить</button>
      <div class="toolbar-info">
        <span>{{ userLabel }}</span>
      </div>
    </div>

    <div v-if="error" class="error-box">{{ error }}</div>

    <!-- Заголовок документа -->
    <div class="doc-header">
      <div class="doc-title">
        <span class="muted">Внутренний заказ</span>
        <span v-if="externalId" class="ext-badge" title="Импортирован из МойСклад — только просмотр">МойСклад</span>
        <span class="doc-number">№</span>
        <input v-model="form.number" class="num-input" :disabled="!canEdit" placeholder="авто" />
        <span class="muted">от</span>
        <input v-model="form.doc_date" type="datetime-local" class="date-input" :disabled="!canEdit" />
        <span class="status-label">Статус:</span>
        <select v-model="form.status" class="status-select" disabled>
          <option value="draft">Черновик</option>
          <option value="posted">Проведён</option>
          <option value="cancelled">Отменён</option>
        </select>
      </div>
      <label class="posted-check">
        <input type="checkbox" :checked="form.status === 'posted'" :disabled="!canEdit || form.status === 'cancelled'" @change="onTogglePosted" />
        Проведено
      </label>
    </div>

    <!-- Поля -->
    <div class="fields-grid">
      <div class="field">
        <label>Организация</label>
                <input
          v-model="organizationSearch"
          :disabled="!canEdit"
          placeholder="Начните вводить название..."
          @focus="showOrganizationSuggest = true"
          @blur="hideOrganizationSuggestSoon"
        />
        <div v-if="showOrganizationSuggest && organizationSuggestions.length" class="ac-dropdown">
          <div v-for="o in organizationSuggestions" :key="o.id" class="ac-item" @mousedown.prevent="selectOrganization(o)">{{ o.name }}</div>
        </div>
      </div>
      <div class="field">
        <label>Склад</label>
                <input
          v-model="warehouseSearch"
          :disabled="!canEdit"
          placeholder="Начните вводить название..."
          @focus="showWarehouseSuggest = true"
          @blur="hideWarehouseSuggestSoon"
        />
        <div v-if="showWarehouseSuggest && warehouseSuggestions.length" class="ac-dropdown">
          <div v-for="w in warehouseSuggestions" :key="w.id" class="ac-item" @mousedown.prevent="selectWarehouse(w)">{{ w.name }}</div>
        </div>
      </div>
      <div class="field">
        <label>План. дата приёмки</label>
        <input v-model="form.plan_date" type="date" :disabled="!canEdit" />
      </div>
      <div class="field">
        <label>Проект</label>
        <div class="ac-wrap">
          <input
            v-model="projectSearch"
            :disabled="!canEdit"
            placeholder="Начните вводить название..."
            @focus="showProjectSuggest = true"
            @blur="hideProjectSuggestSoon"
          />
          <button v-if="canEdit" type="button" class="ac-add" title="Создать проект" @mousedown.prevent="openCreateProject">+</button>
          <button v-if="canEdit && form.project" type="button" class="ac-clear" title="Очистить" @mousedown.prevent="clearProject">×</button>
          <div v-if="showProjectSuggest && projectSuggestions.length" class="ac-dropdown">
            <div v-for="p in projectSuggestions" :key="p.id" class="ac-item" @mousedown.prevent="selectProject(p)">{{ p.name }}</div>
          </div>
        </div>
      </div>    </div>

    <!-- Вкладки -->
    <div class="doc-tabs">
      <button class="doc-tab" :class="{ active: tab === 'main' }" @click="tab = 'main'">Главная</button>
      <button class="doc-tab" :class="{ active: tab === 'related' }" @click="tab = 'related'">Связанные документы</button>
    </div>

    <template v-if="tab === 'main'">
      <!-- Строка добавления -->
      <div class="add-row" v-if="canEdit">
        <input
          v-model="newItemSearch"
          placeholder="Добавьте позицию — введите наименование, код, штрихкод или артикул"
          class="add-input"
          @input="onSearchInput"
          @focus="showSuggest = true"
          @blur="hideSuggestSoon"
        />
        <button class="btn" @click="focusCatalog">Добавить из справочника</button>
        <button class="btn" @click="checkStock">Проверить комплектацию</button>
        <button class="btn" @click="importCsv">Импорт</button>

        <!-- Список подсказок -->
        <div v-if="showSuggest && suggestions.length" class="suggest">
          <div v-for="p in suggestions" :key="p.id" class="suggest-item" @mousedown.prevent="addItem(p)">
            <div class="suggest-name">{{ p.name }}</div>
            <div class="suggest-meta muted">
              <span v-if="p.sku">Арт. {{ p.sku }}</span>
              <span style="margin-left:12px">{{ p.price.toFixed(2) }} ₽</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Таблица позиций -->
      <table class="items-table">
        <thead>
          <tr>
            <th style="width:32px"></th>
            <th>Наименование</th>
            <th class="num" style="width:100px">Кол-во</th>
            <th class="num" style="width:100px">Доступно</th>
            <th class="num" style="width:120px">Цена</th>
            <th class="num" style="width:90px">НДС</th>
            <th class="num" style="width:120px">Сумма</th>
            <th style="width:40px"></th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="form.items.length === 0">
            <td colspan="8" class="muted" style="text-align:center;padding:24px">Добавьте позиции</td>
          </tr>
          <tr v-else v-for="(it, idx) in form.items" :key="it.product_id">
            <td class="muted num">{{ idx + 1 }}</td>
            <td>{{ productName(it.product_id) }}</td>
            <td class="num">
              <input v-model.number="it.quantity" type="number" min="0" step="0.001" class="cell-input" :disabled="!canEdit" />
            </td>
            <td class="num muted">{{ formatQty(availableFor(it.product_id)) }}</td>
            <td class="num">
              <input v-model.number="it.price" type="number" min="0" step="0.01" class="cell-input" :disabled="!canEdit" />
            </td>
            <td class="num">
              <input v-model.number="it.vat_rate" type="number" min="0" max="100" step="1" class="cell-input" :disabled="!canEdit" />
            </td>
            <td class="num">{{ formatMoney(it.quantity * it.price) }}</td>
            <td><button v-if="canEdit" class="x-btn" @click="removeItem(idx)">×</button></td>
          </tr>
        </tbody>
      </table>

      <!-- Комментарий + итоги -->
      <div class="bottom-row">
        <div class="comment-box">
          <label class="lbl">Комментарий</label>
          <textarea v-model="form.comment" rows="6" :disabled="!canEdit"></textarea>
        </div>
        <div class="totals-box">
          <div class="total-row">
            <span>Промежуточный итог:</span>
            <span class="total-num">{{ formatMoney(subtotal) }}</span>
          </div>
          <div class="total-row small">
            <label>
              <input type="checkbox" v-model="form.vat_enabled" :disabled="!canEdit" />
              НДС:
            </label>
            <span class="total-num">{{ formatMoney(vatAmount) }}</span>
          </div>
          <div class="total-row small">
            <label>
              <input type="checkbox" v-model="form.vat_included" :disabled="!canEdit" />
              Цена включает НДС
            </label>
          </div>
          <div class="total-row big">
            <span>Итого:</span>
            <span class="total-num">{{ formatMoney(subtotal) }}</span>
          </div>
        </div>
      </div>
    </template>

    <template v-else>
      <div class="card muted" style="padding:24px;text-align:center">
        Связанных документов пока нет
      </div>
    </template>

    <!-- Задачи -->
    <div class="section-block">
      <div class="section-head"><span>Задачи</span><button class="btn-link">+ Задача</button></div>
      <div class="muted" style="font-size:12px">Нет задач</div>
    </div>

    <!-- Файлы -->
    <div class="section-block">
      <div class="section-head"><span>Файлы</span><button class="btn-link">+ Файл</button></div>
      <div class="muted" style="font-size:12px">Нет файлов</div>
    </div>

    <!-- Модалка: новый проект -->
    <div v-if="showCreateProject" class="modal-backdrop" @click.self="showCreateProject = false">
      <form class="card modal" @submit.prevent="submitCreateProject">
        <h2>Новый проект</h2>
        <div v-if="projectError" class="error-box">{{ projectError }}</div>
        <label>Название</label>
        <input v-model="newProjectName" placeholder="Название проекта" autofocus />
        <div class="modal-actions">
          <button type="button" @click="showCreateProject = false">Отмена</button>
          <button class="primary" type="submit" :disabled="creatingProject">
            {{ creatingProject ? 'Сохранение...' : 'Создать' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  cancelInternalOrder,
  deleteInternalOrder,
  sendInternalOrder,
  getInternalOrder, createInternalOrder, updateInternalOrder,
  postInternalOrder, nextInternalOrderNumber,
  type IntOrderInput,
  exportInternalOrder,
} from '../api/internalOrders'
import { listProducts, type Product } from '../api/products'
import { listWarehouses, type Warehouse } from '../api/warehouses'
import { listOrganizations, type Organization } from '../api/suppliers'
import { listProjects, createProject, type Project } from '../api/projects'
import { listStockExtended } from '../api/stock'
import { apiErrorMessage } from '../api/client'
import { useAuthStore } from '../stores/auth'

interface FormItem {
  product_id: string
  quantity: number
  price: number
  vat_rate: number
}

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const products = ref<Product[]>([])
const warehouses = ref<Warehouse[]>([])
const organizations = ref<Organization[]>([])
const projects = ref<Project[]>([])
const stockMap = ref<Map<string, number>>(new Map())

const loading = ref(false)
const saving = ref(false)
const error = ref<string | null>(null)
const tab = ref<'main' | 'related'>('main')
const isNew = computed(() => route.params.id === 'new' || route.path.endsWith('/new'))

const form = reactive({
  number: '',
  doc_date: new Date().toISOString().slice(0, 16),
  status: 'draft' as string,
  organization_id: '',
  warehouse_id: '',
  plan_date: '',
  project: '',
  project_name: '',
  sent_at: undefined as string | undefined,
  comment: '',
  vat_enabled: true,
  vat_included: true,
  items: [] as FormItem[],
})

const currentId = ref<string | null>(null)
const externalId = ref<string | null>(null)
const canEdit = computed(() => isNew.value || (form.status === 'draft' && !externalId.value))

const newItemSearch = ref('')
const showSuggest = ref(false)
const suggestions = computed(() => {
  const q = newItemSearch.value.trim().toLowerCase()
  if (!q) return []
  return products.value.filter((p) => !!p.external_id).filter((p) =>
    p.name.toLowerCase().includes(q) ||
    (p.sku ?? '').toLowerCase().includes(q)
  ).slice(0, 8)
})
function hideSuggestSoon() { setTimeout(() => { showSuggest.value = false }, 150) }
function onSearchInput() { showSuggest.value = true }
function focusCatalog() { showSuggest.value = true }
function addItem(p: Product) {
  const existing = form.items.find((i) => i.product_id === p.id)
  if (existing) { existing.quantity += 1 }
  else { form.items.push({ product_id: p.id, quantity: 1, price: p.price, vat_rate: 20 }) }
  newItemSearch.value = ''
  showSuggest.value = false
}
function removeItem(idx: number) { form.items.splice(idx, 1) }

function productName(id: string) { return products.value.find((p) => p.id === id)?.name ?? '—' }
function availableFor(id: string) { return stockMap.value.get(id) ?? 0 }
function formatQty(n: number) { return n === 0 ? '0' : n.toLocaleString('ru-RU', { maximumFractionDigits: 3 }) }
function formatMoney(n: number) { return n.toLocaleString('ru-RU', { minimumFractionDigits: 2, maximumFractionDigits: 2 }) }

const subtotal = computed(() => form.items.reduce((s, i) => s + i.quantity * i.price, 0))
const vatAmount = computed(() => {
  if (!form.vat_enabled) return 0
  if (form.vat_included) return subtotal.value * 20 / 120
  return subtotal.value * 0.20
})

const userLabel = computed(() => auth.userId ? auth.userId.slice(0, 8) : '')

async function load() {
  loading.value = true; error.value = null
  try {
    const [prods, w, orgs, stock, projs] = await Promise.all([
      listProducts(true),
      listWarehouses(),
      listOrganizations(),
      listStockExtended().catch(() => ({ items: [] as any[] })),
      listProjects().catch(() => [] as Project[]),
    ])
    products.value = prods
    warehouses.value = w
    organizations.value = orgs
    projects.value = projs
    const m = new Map<string, number>()
    for (const r of (stock as any).items ?? []) m.set(r.product_id, r.quantity)
    stockMap.value = m

    if (!isNew.value && route.params.id) {
      currentId.value = route.params.id as string
      const o = await getInternalOrder(currentId.value)
      form.number = o.number
      form.doc_date = o.doc_date.slice(0, 16)
      form.status = o.status
      form.organization_id = o.organization_id ?? organizations.value.find((x) => x.is_default)?.id ?? ''
      form.warehouse_id = o.warehouse_id ?? ''
      syncWarehouseSearch()
      syncOrganizationSearch()
      form.plan_date = o.plan_date ? o.plan_date.slice(0, 10) : ''
      form.project = o.project ?? ''
      form.project_name = o.project_name ?? ''
      syncProjectSearch()
      form.comment = o.comment ?? ''
      form.vat_enabled = o.vat_enabled
      form.vat_included = o.vat_included
      externalId.value = o.external_id ?? null
      form.items = (o.items ?? []).map((it) => ({
        product_id: it.product_id, quantity: it.quantity, price: it.price, vat_rate: it.vat_rate,
      }))
    } else {
      // новосоздаваемый — получаем следующий номер и дефолтные значения
      try { form.number = await nextInternalOrderNumber() } catch { /* ignore */ }
      form.organization_id = organizations.value.find((x) => x.is_default)?.id ?? ''
      form.warehouse_id = warehouses.value[0]?.id ?? ''
      syncWarehouseSearch()
      syncOrganizationSearch()
      syncProjectSearch()
    }
  } catch (e) {
    error.value = apiErrorMessage(e)
  } finally {
    loading.value = false
  }
}

function buildInput(): IntOrderInput {
  return {
    number: form.number || undefined,
    organization_id: form.organization_id || undefined,
    warehouse_id: form.warehouse_id || undefined,
    plan_date: form.plan_date || undefined,
    project: form.project || undefined,
    comment: form.comment || undefined,
    vat_enabled: form.vat_enabled,
    vat_included: form.vat_included,
    items: form.items.map((i) => ({
      product_id: i.product_id, quantity: i.quantity, price: i.price, vat_rate: i.vat_rate,
    })),
  }
}

async function save() {
  error.value = null
  saving.value = true
  try {
    if (form.items.length === 0) throw new Error('Добавьте хотя бы одну позицию')
    if (!form.warehouse_id) throw new Error('Выберите склад из списка')
    const wh = warehouses.value.find((x) => x.id === form.warehouse_id)
    if (!wh) throw new Error('Склад не найден — выберите из списка')
    if (!wh.external_id) throw new Error('Склад не привязан к МойСклад — выберите другой')
    if (!form.organization_id) throw new Error('Выберите организацию из списка')
    const org = organizations.value.find((x) => x.id === form.organization_id)
    if (!org) throw new Error('Организация не найдена — выберите из списка')
    if (!org.external_id) throw new Error('Организация не привязана к МойСклад')
    for (const it of form.items) {
      const prod = products.value.find((x) => x.id === it.product_id)
      if (!prod) throw new Error('Товар не найден — удалите позицию и выберите из списка')
      if (!prod.external_id) throw new Error(`Товар «${prod.name}» не привязан к МойСклад`)
    }
    if (isNew.value) {
      const created = await createInternalOrder(buildInput())
      currentId.value = created.id
      form.number = created.number
      form.status = created.status
      router.replace(`/internal-orders/${created.id}`)
    } else if (currentId.value) {
      await updateInternalOrder(currentId.value, buildInput())
    }
  } catch (e) {
    error.value = apiErrorMessage(e)
  } finally {
    saving.value = false
  }
}

async function onTogglePosted(e: Event) {
  if (!currentId.value) return
  const checked = (e.target as HTMLInputElement).checked
  try {
    if (checked && form.status === 'draft') {
      await save()
      const updated = await postInternalOrder(currentId.value)
      form.status = updated.status
    }
  } catch (err) {
    error.value = apiErrorMessage(err)
  }
}

function close() { router.push('/internal-orders') }

async function onDelete() {
  if (!currentId.value) { router.push('/internal-orders'); return }
  if (!confirm('Удалить заказ? Действие необратимо.')) return
  try {
    await deleteInternalOrder(currentId.value)
    router.push('/internal-orders')
  } catch (e) {
    error.value = apiErrorMessage(e)
  }
}

async function onCancel() {
  if (!currentId.value) return
  if (!confirm('Отменить заказ?')) return
  try {
    const updated = await cancelInternalOrder(currentId.value)
    form.status = updated.status
  } catch (e) {
    error.value = apiErrorMessage(e)
  }
}

async function onSend() {
  if (!currentId.value) return
  try {
    const updated = await sendInternalOrder(currentId.value)
    form.sent_at = updated.sent_at
  } catch (e) {
    error.value = apiErrorMessage(e)
  }
}
async function print() {
  if (!currentId.value) {
    error.value = 'Сначала сохраните заказ'
    return
  }
  try {
    const { blob, filename } = await exportInternalOrder(currentId.value)
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = filename
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    setTimeout(() => URL.revokeObjectURL(url), 1500)

    // подтянем обновлённый статус (printed_at)
    const o = await getInternalOrder(currentId.value)
    form.status = o.status
  } catch (e) {
    error.value = apiErrorMessage(e)
  }
}
function checkStock() { alert('Проверка комплектации: функция в разработке') }
function importCsv() { alert('Импорт: функция в разработке') }

// === Autocomplete sklad/organizatsiya ===
const warehouseSearch = ref('')
const organizationSearch = ref('')
const showWarehouseSuggest = ref(false)
const showOrganizationSuggest = ref(false)

const warehouseSuggestions = computed(() => {
  const q = warehouseSearch.value.trim().toLowerCase()
  const list = warehouses.value.filter((w) => !!w.external_id)
  if (!q) return list.slice(0, 8)
  return list.filter((w) => w.name.toLowerCase().includes(q)).slice(0, 10)
})

const organizationSuggestions = computed(() => {
  const q = organizationSearch.value.trim().toLowerCase()
  const list = organizations.value.filter((o) => !!o.external_id)
  if (!q) return list.slice(0, 8)
  return list.filter((o) => o.name.toLowerCase().includes(q)).slice(0, 10)
})

function hideWarehouseSuggestSoon() { setTimeout(() => { showWarehouseSuggest.value = false }, 150) }
function hideOrganizationSuggestSoon() { setTimeout(() => { showOrganizationSuggest.value = false }, 150) }

function selectWarehouse(w: Warehouse) {
  form.warehouse_id = w.id
  warehouseSearch.value = w.name
  showWarehouseSuggest.value = false
}
function selectOrganization(o: Organization) {
  form.organization_id = o.id
  organizationSearch.value = o.name
  showOrganizationSuggest.value = false
}

function syncWarehouseSearch() {
  const w = warehouses.value.find((x) => x.id === form.warehouse_id)
  warehouseSearch.value = w?.name ?? ''
}
function syncOrganizationSearch() {
  const o = organizations.value.find((x) => x.id === form.organization_id)
  organizationSearch.value = o?.name ?? ''
}

watch(warehouseSearch, (val) => {
  const w = warehouses.value.find((x) => x.name === val)
  form.warehouse_id = w?.id ?? ''
})
watch(organizationSearch, (val) => {
  const o = organizations.value.find((x) => x.name === val)
  form.organization_id = o?.id ?? ''
})

// === Проект ===
const projectSearch = ref('')
const showProjectSuggest = ref(false)
const showCreateProject = ref(false)
const newProjectName = ref('')
const creatingProject = ref(false)
const projectError = ref<string | null>(null)

const projectSuggestions = computed(() => {
  const q = projectSearch.value.trim().toLowerCase()
  const list = projects.value
  if (!q) return list.slice(0, 8)
  return list.filter((p) => p.name.toLowerCase().includes(q)).slice(0, 10)
})

function hideProjectSuggestSoon() { setTimeout(() => { showProjectSuggest.value = false }, 150) }

function selectProject(p: Project) {
  form.project = p.external_id ?? ''
  form.project_name = p.name
  projectSearch.value = p.name
  showProjectSuggest.value = false
}
function clearProject() {
  form.project = ''
  form.project_name = ''
  projectSearch.value = ''
}
function syncProjectSearch() {
  projectSearch.value = form.project_name || ''
}
watch(projectSearch, (val) => {
  const p = projects.value.find((x) => x.name === val)
  form.project = p?.external_id ?? ''
  form.project_name = p?.name ?? ''
})

function openCreateProject() {
  newProjectName.value = projectSearch.value.trim()
  projectError.value = null
  showCreateProject.value = true
}
async function submitCreateProject() {
  const name = newProjectName.value.trim()
  if (!name) { projectError.value = 'Введите название'; return }
  creatingProject.value = true
  projectError.value = null
  try {
    const created = await createProject(name)
    projects.value = [...projects.value, created].sort((a,b) => a.name.localeCompare(b.name))
    selectProject(created)
    showCreateProject.value = false
  } catch (e) {
    projectError.value = apiErrorMessage(e)
  } finally {
    creatingProject.value = false
  }
}onMounted(load)
</script>

<style scoped>
.card-page { padding: 0; }

.toolbar {
  display: flex; gap: 8px; align-items: center;
  padding: 8px 0 12px;
  border-bottom: 1px solid #e1e4e8;
  margin-bottom: 12px;
}
.toolbar-info { margin-left: auto; font-size: 13px; color: #8c959f; }

.doc-header {
  display: flex; justify-content: space-between; align-items: center;
  margin-bottom: 14px;
}
.doc-title { display: flex; align-items: center; gap: 8px; font-size: 16px; }
.doc-number { font-weight: 600; }
.num-input {
  width: 100px; padding: 4px 8px;
  border: 1px solid transparent; border-radius: 4px;
  font-size: 15px; font-weight: 600; background: transparent;
}
.num-input:hover { border-color: #d0d7de; background: #fff; }
.num-input:focus { border-color: #2c5d9c; background: #fff; outline: none; }
.date-input {
  width: 180px; padding: 4px 8px;
  border: 1px solid transparent; border-radius: 4px;
  font-size: 13px; background: transparent;
}
.date-input:hover { border-color: #d0d7de; background: #fff; }
.status-label { margin-left: 12px; color: #57606a; font-size: 13px; }
.status-select {
  padding: 4px 8px; font-size: 13px;
  border: 1px solid #d0d7de; border-radius: 4px;
  background: #fff;
}
.posted-check { display: flex; align-items: center; gap: 6px; font-size: 13px; color: #1f2328; }
.posted-check input { width: auto; }
.ext-badge {
  padding: 2px 6px; font-size: 11px; font-weight: 500;
  background: #eaeef2; color: #57606a; border-radius: 3px;
  text-transform: uppercase; letter-spacing: 0.3px;
}

.fields-grid {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr 1fr;
  gap: 12px 20px;
  margin-bottom: 16px;
}
.field { display: flex; flex-direction: column; gap: 4px; font-size: 12px; color: #57606a; }
.field label { font-weight: 500; }
.field input, .field select {
  padding: 5px 8px; font-size: 13px;
  border: 1px solid #d0d7de; border-radius: 3px;
  background: #fff; color: #1f2328;
}

.doc-tabs {
  display: flex; gap: 4px;
  border-bottom: 1px solid #d8dee4;
  margin-bottom: 12px;
}
.doc-tab {
  background: transparent; border: none;
  border-bottom: 2px solid transparent;
  padding: 8px 14px; font-size: 13px;
  color: #57606a; cursor: pointer;
}
.doc-tab.active { color: #2c5d9c; border-bottom-color: #2c5d9c; font-weight: 600; }

.add-row {
  position: relative;
  display: flex; gap: 8px;
  margin-bottom: 12px;
}
.add-input {
  flex: 1;
  padding: 8px 12px;
  border: 1px solid #d0d7de;
  border-radius: 4px;
  font-size: 13px;
}
.suggest {
  position: absolute;
  top: 100%; left: 0; right: 0;
  background: #fff;
  border: 1px solid #d0d7de;
  border-radius: 4px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.1);
  z-index: 20;
  max-height: 300px;
  overflow-y: auto;
  margin-top: 2px;
}
.suggest-item {
  padding: 8px 12px;
  cursor: pointer;
  border-bottom: 1px solid #f0f2f5;
}
.suggest-item:hover { background: #f6f8fa; }
.suggest-item:last-child { border-bottom: none; }
.suggest-name { font-size: 13px; color: #1f2328; }
.suggest-meta { font-size: 11px; }

.items-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
  margin-bottom: 16px;
}
.items-table thead th {
  background: #fff;
  border-bottom: 2px solid #d8dee4;
  padding: 6px 8px;
  text-align: left;
  font-weight: 500;
  color: #2c5d9c;
  font-size: 12px;
  white-space: nowrap;
}
.items-table tbody td {
  padding: 4px 8px;
  border-bottom: 1px solid #eaeef2;
  vertical-align: middle;
}
.items-table .num { text-align: right; font-variant-numeric: tabular-nums; }
.items-table .muted { color: #8c959f; }
.cell-input {
  width: 100%;
  padding: 4px 6px;
  border: 1px solid transparent;
  border-radius: 3px;
  text-align: right;
  font-size: 13px;
  background: transparent;
  font-variant-numeric: tabular-nums;
}
.cell-input:hover { border-color: #d0d7de; background: #fff; }
.cell-input:focus { border-color: #2c5d9c; background: #fff; outline: none; }
.btn.danger {
  background: #fff;
  color: #cf222e;
  border-color: #cf222e;
}
.btn.danger:hover:not(:disabled) {
  background: #cf222e;
  color: #fff;
}

.x-btn {
  background: transparent; border: none;
  color: #8c959f; font-size: 18px; cursor: pointer;
  padding: 0 6px;
}
.x-btn:hover { color: #cf222e; }

.bottom-row {
  display: flex;
  gap: 24px;
  align-items: flex-start;
  margin-bottom: 24px;
}
.comment-box { flex: 1; display: flex; flex-direction: column; gap: 4px; }
.comment-box .lbl { font-size: 12px; color: #57606a; }
.comment-box textarea {
  padding: 8px 12px; font-size: 13px;
  border: 1px solid #d0d7de; border-radius: 4px;
  resize: vertical; font-family: inherit;
}
.totals-box { width: 360px; padding: 12px; }
.total-row {
  display: flex; justify-content: space-between;
  padding: 6px 0;
  align-items: baseline;
  font-size: 14px;
}
.total-row.small { font-size: 12px; color: #57606a; }
.total-row.small label { display: flex; align-items: center; gap: 6px; }
.total-row.big {
  border-top: 1px solid #eaeef2;
  margin-top: 6px;
  padding-top: 10px;
  font-size: 18px;
  font-weight: 700;
}
.total-num { font-variant-numeric: tabular-nums; }

.section-block {
  margin-top: 24px;
  padding-top: 16px;
  border-top: 1px solid #eaeef2;
}
.section-head {
  display: flex; align-items: center; gap: 12px;
  margin-bottom: 8px;
  font-size: 13px;
  font-weight: 600;
  color: #cf222e;
}
.section-head .btn-link {
  color: #2c5d9c; border: 1px solid #d0d7de;
  background: #fff; padding: 3px 10px;
  border-radius: 4px; font-size: 12px; cursor: pointer;
}
.section-head .btn-link:hover { background: #f6f8fa; }

.field { position: relative; }
.ac-dropdown {
  position: absolute; top: 100%; left: 0; right: 0; z-index: 20;
  background: #fff; border: 1px solid #d0d7de; border-top: none;
  border-radius: 0 0 3px 3px; max-height: 220px; overflow-y: auto;
  box-shadow: 0 4px 12px rgba(0,0,0,0.1);
}
.ac-item {
  padding: 6px 10px; font-size: 13px; cursor: pointer;
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
}
.ac-item:hover { background: #eaeef2; }

/* Проект: автокомплит + кнопка + */
.ac-wrap { position: relative; display: flex; align-items: center; gap: 4px; }
.ac-wrap > input { flex: 1; }
.ac-add, .ac-clear {
  flex: 0 0 auto;
  width: 26px; height: 26px;
  border: 1px solid #d0d7de; background: #fff;
  border-radius: 4px; cursor: pointer;
  font-size: 16px; line-height: 1;
  color: #2c5d9c;
  display: inline-flex; align-items: center; justify-content: center;
  padding: 0;
}
.ac-add:hover { background: #eaf3ff; }
.ac-clear { color: #cf222e; }
.ac-clear:hover { background: #ffecec; }

/* Модалка (для создания проекта) */
.modal-backdrop {
  position: fixed; inset: 0; background: rgba(0,0,0,0.4);
  display: flex; align-items: center; justify-content: center;
  padding: 20px; z-index: 100;
}
.modal {
  width: 420px; max-width: 100%;
  display: flex; flex-direction: column; gap: 10px;
  padding: 16px;
}
.modal h2 { margin: 0 0 4px; font-size: 18px; }
.modal label { font-size: 13px; color: #444; }
.modal-actions { display: flex; justify-content: flex-end; gap: 10px; margin-top: 6px; }
.error-box {
  background: #ffecec; color: #cf222e;
  border: 1px solid #ffb3b3; border-radius: 4px;
  padding: 6px 10px; font-size: 13px;
}
</style>