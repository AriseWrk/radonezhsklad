<template>
  <div class="card-page">
    <!-- Заголовок -->
    <div class="ms-doc-head">
      <div class="ms-title">
        <button class="ms-help" title="Справка"><MsIcon name="help" :size="14" /></button>
        <span>Внутренний заказ</span>
        <button class="ms-refresh" @click="close" title="Обновить"><MsIcon name="refresh" :size="14" /></button>
        <span v-if="externalId" class="ext-badge" title="Импортирован из МойСклад">МойСклад</span>
      </div>
      <div class="ms-status">
        <span class="st" :class="'st-' + form.status">{{ statusLabel(form.status) }}</span>
        <label class="posted-check">
          <input type="checkbox" :checked="form.status === 'posted'" :disabled="!canEdit || form.status === 'cancelled'" @change="onTogglePosted" />
          Проведено
        </label>
      </div>
    </div>

    <!-- Тулбар -->
    <div class="ms-toolbar">
      <MsButton variant="primary" icon="plus" @click="close">Создать</MsButton>
      <MsButton icon="save" @click="save" :disabled="saving || !canEdit">{{ saving ? 'Сохранение...' : 'Сохранить' }}</MsButton>
      <MsButton icon="post" @click="onTogglePosted" :disabled="!canEdit || form.status === 'cancelled'">Провести</MsButton>
      <MsButton icon="print" @click="print">Печать</MsButton>
      <MsButton icon="send" @click="onSend" :disabled="!currentId || form.status !== 'posted' || !canEdit">Отправить</MsButton>
      <MsButton icon="close" @click="onCancel" :disabled="!currentId || form.status !== 'posted' || !canEdit">Отменить</MsButton>
      <MsButton variant="icon" icon="trash" title="Удалить" @click="onDelete" :disabled="(!currentId && !isNew) || (currentId !== null && form.status !== 'draft')" />
      <div class="toolbar-spacer"></div>
      <MsButton variant="icon" icon="close" @click="close" title="Закрыть" />
    </div>

    <div v-if="error" class="error-box">{{ error }}</div>

    <!-- Реквизиты -->
    <div class="ms-doc-fields">
      <div class="doc-left">
        <div class="row">
          <label>Номер</label>
          <input v-model="form.number" class="num-input" :disabled="!canEdit" placeholder="авто" />
        </div>
        <div class="row">
          <label>Дата</label>
          <input v-model="form.doc_date" type="datetime-local" class="date-input" :disabled="!canEdit" />
        </div>
        <div class="row">
          <label>Организация</label>
          <div class="ac-wrap">
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
        </div>
        <div class="row">
          <label>Склад</label>
          <div class="ac-wrap">
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
        </div>
      </div>

      <div class="doc-right">
        <div class="row">
          <label>План. дата приёмки</label>
          <input v-model="form.plan_date" type="date" :disabled="!canEdit" />
        </div>
        <div class="row">
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
        </div>
        <div class="row">
          <label>Владелец</label>
          <span class="muted">{{ userLabel }}</span>
        </div>
      </div>
    </div>

    <!-- Вкладки -->
    <div class="ms-tabs">
      <button class="tab" :class="{ active: tab === 'main' }" @click="tab = 'main'">Главная</button>
      <button class="tab" :class="{ active: tab === 'related' }" @click="tab = 'related'">Связанные документы</button>
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
        <MsButton @click="focusCatalog">Из справочника</MsButton>
        <MsButton @click="checkStock">Проверить</MsButton>
        <MsButton @click="importCsv">Импорт</MsButton>
        <div v-if="showSuggest && suggestions.length" class="suggest">
          <div v-for="p in suggestions" :key="p.id" class="suggest-item" @mousedown.prevent="addItem(p)">
            <div class="suggest-row">
              <div class="suggest-name">{{ p.name }}</div>
              <div class="suggest-stock" :class="{ 'out': availableFor(p.id) <= 0 }">
                {{ availableFor(p.id) > 0 ? 'Остаток: ' + formatQty(availableFor(p.id)) : 'Нет в наличии' }}
              </div>
            </div>
            <div class="suggest-meta muted">
              <span v-if="p.sku">Арт. {{ p.sku }}</span>
              <span style="margin-left:12px">{{ p.price.toFixed(2) }} ₽</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Таблица позиций -->
      <table class="ms-items">
        <thead>
          <tr>
            <th class="col-num">#</th>
            <th>Наименование</th>
            <th class="col-num-right">Кол-во</th>
            <th class="col-num-right">Доступно</th>
            <th class="col-num-right">Цена</th>
            <th class="col-num-right">НДС</th>
            <th class="col-num-right">Сумма</th>
            <th class="col-x"></th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="form.items.length === 0">
            <td colspan="8" class="empty">Добавьте позиции</td>
          </tr>
          <tr v-else v-for="(it, idx) in form.items" :key="it.product_id">
            <td class="col-num">{{ idx + 1 }}</td>
            <td>{{ productName(it.product_id) }}</td>
            <td class="col-num-right">
              <input v-model.number="it.quantity" type="number" min="0" step="0.001" class="cell-input" :disabled="!canEdit" />
            </td>
            <td class="col-num-right muted">{{ formatQty(availableFor(it.product_id)) }}</td>
            <td class="col-num-right">
              <input v-model.number="it.price" type="number" min="0" step="0.01" class="cell-input" :disabled="!canEdit" />
            </td>
            <td class="col-num-right">
              <input v-model.number="it.vat_rate" type="number" min="0" max="100" step="1" class="cell-input" :disabled="!canEdit" />
            </td>
            <td class="col-num-right">{{ formatMoney(it.quantity * it.price) }}</td>
            <td class="col-x"><button v-if="canEdit" class="x-btn" @click="removeItem(idx)">×</button></td>
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
      <div class="empty-block">Связанных документов пока нет</div>
    </template>

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
import MsButton from '../components/MsButton.vue'
import MsIcon from '../components/MsIcon.vue'

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

function statusLabel(s: string) {
  return { draft: 'Черновик', posted: 'Проведён', cancelled: 'Отменён' }[s] || s
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
.suggest-row { display: flex; justify-content: space-between; align-items: baseline; gap: 12px; }
.suggest-name { font-size: 13px; color: #1f2328; }
.suggest-meta { font-size: 11px; }
.suggest-stock { font-size: 12px; color: #15803d; white-space: nowrap; }
.suggest-stock.out { color: #b91c1c; }

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

/* === Карточка внутреннего заказа (МойСклад-стиль) === */
.card-page { font-size: 13px; }

.ms-doc-head {
  display: flex; align-items: center; justify-content: space-between;
  margin-bottom: 10px;
}
.ms-title {
  display: flex; align-items: center; gap: 8px;
  font-size: 20px; font-weight: 600; color: #1f2328;
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
.ext-badge {
  background: #eaf3ff; color: #2c5d9c; border: 1px solid #c8dcf0;
  font-size: 11px; padding: 1px 8px; border-radius: 3px;
  font-weight: 500;
}
.ms-status { display: flex; align-items: center; gap: 12px; }
.st {
  display: inline-block; padding: 3px 12px;
  border-radius: 3px; font-size: 12px; font-weight: 500;
}
.st-draft     { background: #eef1f5; color: #57606a; }
.st-posted    { background: #d8e8c8; color: #2d6a1e; }
.st-cancelled { background: #f5d5d5; color: #a01c1c; }
.posted-check {
  display: inline-flex; align-items: center; gap: 5px;
  font-size: 12px; color: #57606a; cursor: pointer;
}
.posted-check input { width: 14px; height: 14px; }

.ms-toolbar {
  display: flex; align-items: center; gap: 6px;
  margin-bottom: 12px; flex-wrap: wrap;
  padding-bottom: 10px;
  border-bottom: 1px solid #eaeef2;
}
.toolbar-spacer { flex: 1; }

/* Реквизиты документа */
.ms-doc-fields {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 6px 40px;
  margin-bottom: 14px;
  max-width: 1100px;
}
.doc-left, .doc-right { display: flex; flex-direction: column; gap: 6px; }
.row {
  display: grid;
  grid-template-columns: 160px 1fr;
  align-items: center;
  gap: 10px;
}
.row > label { font-size: 13px; color: #57606a; }
.row > input,
.row > .ac-wrap > input,
.row > select {
  height: 28px; padding: 0 8px; font-size: 13px;
  border: 1px solid #d0d7de; border-radius: 3px;
  background: #fff; color: #1f2328; width: 100%;
}
.row > input:disabled,
.row > .ac-wrap > input:disabled { background: #f6f8fa; color: #57606a; cursor: not-allowed; }
.num-input { max-width: 140px; }
.date-input { max-width: 200px; }

/* Вкладки */
.ms-tabs {
  display: flex; gap: 2px;
  border-bottom: 1px solid #d8dee4;
  margin-bottom: 12px;
}
.tab {
  border: none; background: transparent;
  padding: 10px 16px;
  font-size: 13px; color: #57606a;
  cursor: pointer;
  border-bottom: 2px solid transparent;
}
.tab:hover { color: #2c5d9c; }
.tab.active { color: #2c5d9c; border-bottom-color: #2c5d9c; font-weight: 600; }

/* Добавление позиции */
.add-row {
  display: flex; align-items: center; gap: 8px;
  margin-bottom: 10px; position: relative;
}
.add-input {
  flex: 1; height: 30px; padding: 0 10px;
  font-size: 13px;
  border: 1px solid #d0d7de; border-radius: 4px;
  background: #fff; color: #1f2328;
}
.add-input::placeholder { color: #8c959f; }
.add-input:focus { outline: 2px solid rgba(44,93,156,0.3); border-color: #2c5d9c; }
.suggest {
  position: absolute; top: 100%; left: 0; right: 0; z-index: 20;
  background: #fff; border: 1px solid #d0d7de;
  border-radius: 0 0 4px 4px;
  max-height: 260px; overflow-y: auto;
  box-shadow: 0 4px 12px rgba(0,0,0,0.1);
}
.suggest-item {
  padding: 8px 12px; cursor: pointer;
  border-bottom: 1px solid #f0f2f5;
}
.suggest-item:hover { background: #f6f8fa; }
.suggest-row { display: flex; justify-content: space-between; align-items: baseline; gap: 12px; }
.suggest-name { font-size: 13px; color: #1f2328; }
.suggest-meta { font-size: 11px; margin-top: 2px; }
.suggest-stock { font-size: 12px; color: #15803d; white-space: nowrap; }
.suggest-stock.out { color: #b91c1c; }

/* Таблица позиций */
.ms-items {
  width: 100%; border-collapse: collapse;
  background: #fff; font-size: 13px;
  margin-bottom: 14px;
}
.ms-items thead th {
  background: #fff; color: #2c5d9c;
  font-weight: 500; padding: 8px 10px;
  text-align: left; border-bottom: 1px solid #d8dee4;
  white-space: nowrap; font-size: 12px;
}
.ms-items tbody td {
  padding: 6px 10px;
  border-bottom: 1px solid #eaeef2;
  vertical-align: middle;
}
.ms-items .col-num { width: 40px; color: #57606a; text-align: center; }
.ms-items .col-num-right { text-align: right; font-variant-numeric: tabular-nums; }
.ms-items .col-x { width: 32px; text-align: center; }
.ms-items .cell-input {
  width: 90px; height: 26px; padding: 0 6px;
  font-size: 13px; text-align: right;
  border: 1px solid #d0d7de; border-radius: 3px;
  background: #fff;
}
.ms-items .cell-input:disabled { background: #f6f8fa; color: #8c959f; }
.ms-items .empty { text-align: center; padding: 24px; color: #8c959f; }
.x-btn {
  width: 22px; height: 22px; padding: 0;
  border: none; background: transparent;
  color: #8c959f; cursor: pointer;
  font-size: 16px; line-height: 1;
  border-radius: 3px;
}
.x-btn:hover { background: #ffecec; color: #cf222e; }

/* Комментарий + итоги */
.bottom-row {
  display: grid;
  grid-template-columns: 1fr 380px;
  gap: 24px;
  align-items: start;
  margin-top: 12px;
}
.comment-box { display: flex; flex-direction: column; gap: 4px; }
.comment-box .lbl { font-size: 12px; color: #57606a; }
.comment-box textarea {
  font-family: inherit; font-size: 13px;
  padding: 8px 10px;
  border: 1px solid #d0d7de; border-radius: 3px;
  background: #fff; color: #1f2328;
  resize: vertical;
}
.totals-box {
  background: #f6f8fa; border: 1px solid #eaeef2;
  border-radius: 4px; padding: 12px 16px;
  display: flex; flex-direction: column; gap: 6px;
}
.total-row {
  display: flex; justify-content: space-between; align-items: center;
  font-size: 13px; color: #1f2328;
}
.total-row.small { font-size: 12px; color: #57606a; }
.total-row.small label { display: inline-flex; align-items: center; gap: 6px; }
.total-row.big {
  font-size: 15px; font-weight: 600;
  border-top: 1px solid #d8dee4;
  padding-top: 8px; margin-top: 4px;
}
.total-num { font-variant-numeric: tabular-nums; }
.empty-block {
  padding: 40px; text-align: center;
  color: #8c959f; font-size: 13px;
  background: #fff; border: 1px solid #eaeef2; border-radius: 4px;
}
</style>
