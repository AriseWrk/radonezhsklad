<template>
  <div class="card-page">
    <div v-if="error" class="error-box">{{ error }}</div>

    <!-- Шапка -->
    <div class="doc-header">
      <div class="doc-header-left">
        <div class="header-title-row">
          <span class="doc-type-label">{{ typeLabel(docType) }} №</span>
          <input v-model="form.number" class="num-input" :disabled="!isNew" placeholder="авто" />
          <span class="from-sep">от</span>
          <input v-model="form.doc_date" type="datetime-local" class="date-input" :disabled="!canEdit" />
          <span class="from-sep">Статус</span>
          <select v-model="form.status" class="status-select" :disabled="!isNew">
            <option value="draft">Черновик</option>
            <option value="posted">Проведён</option>
          </select>
        </div>
      </div>
      <div class="doc-header-right">
        <div class="user-box">
          <div class="user-name">{{ auth.userId ? auth.userId.slice(0,8) : '—' }}</div>
          <div class="user-role">Основной</div>
        </div>
      </div>
    </div>

    <!-- Тулбар -->
    <div class="toolbar">
      <button class="btn-save" :disabled="saving" @click="save">{{ saving ? 'Сохранение...' : 'Сохранить' }}</button>
      <button class="btn-plain" v-if="doc && doc.status === 'draft'" :disabled="saving" @click="doPost">Провести</button>
      <button class="btn-plain" v-if="doc && doc.status === 'posted'" :disabled="saving" @click="doCancel">Отменить</button>
      <button class="btn-plain" @click="close">Закрыть</button>
      <button class="btn-plain" :disabled="!doc?.id">Изменить ▾</button>
      <button class="btn-plain" :disabled="!doc?.id">Печать ▾</button>
      <button class="btn-plain" :disabled="!doc?.id">Отправить ▾</button>
    </div>

    <div v-if="loading" class="empty-block">Загрузка…</div>

    <template v-if="!loading">
      <!-- Поля -->
      <div class="fields-grid">
        <div class="field">
          <label>* Организация</label>
          <div class="with-plus">
            <select v-model="form.organization_id">
              <option value="">—</option>
              <option v-for="o in organizations" :key="o.id" :value="o.id">{{ o.name }}</option>
            </select>
            <button class="plus-btn" type="button">+</button>
          </div>
        </div>
        <div class="field">
          <label>Проект</label>
          <div class="with-plus">
            <select disabled><option>—</option></select>
            <button class="plus-btn" type="button">+</button>
          </div>
        </div>
        <div class="field">
          <label>* Склад</label>
          <div class="with-plus">
            <select v-model="form.warehouse_id">
              <option value="">—</option>
              <option v-for="w in warehouses" :key="w.id" :value="w.id">{{ w.name }}</option>
            </select>
            <button class="plus-btn" type="button">+</button>
          </div>
        </div>
        <div class="field" v-if="docType === 'transfer'">
          <label>* Склад-получатель</label>
          <div class="with-plus">
            <select v-model="form.target_warehouse_id">
              <option value="">—</option>
              <option v-for="w in warehouses" :key="w.id" :value="w.id" :disabled="w.id === form.warehouse_id">{{ w.name }}</option>
            </select>
            <button class="plus-btn" type="button">+</button>
          </div>
        </div>
        <div class="field" v-if="docType === 'enter'">
          <label>Договор</label>
          <div class="with-plus">
            <select v-model="form.contract_id">
              <option value="">—</option>
              <option v-for="ct in contracts" :key="ct.id" :value="ct.id">{{ ct.number }}</option>
            </select>
            <button class="plus-btn" type="button">+</button>
          </div>
        </div>
        <div class="field" v-if="docType === 'writeoff'">
          <label>* Статья расходов</label>
          <div class="with-plus">
            <select v-model="form.expense_item">
              <option value="">—</option>
              <option value="Возврат">Возврат</option>
              <option value="Закупка товаров">Закупка товаров</option>
              <option value="Налоги и сборы">Налоги и сборы</option>
              <option value="Перемещение">Перемещение</option>
              <option value="Списания">Списания</option>
              <option value="Аренда">Аренда</option>
              <option value="Зарплата">Зарплата</option>
              <option value="Маркетинг и реклама">Маркетинг и реклама</option>
            </select>
            <button class="plus-btn" type="button">+</button>
          </div>
        </div>
      </div>

      <!-- Вкладки -->
      <div class="tabs">
        <button class="tab" :class="{ active: activeTab === 'main' }" @click="activeTab = 'main'">Главная</button>
        <button class="tab" :class="{ active: activeTab === 'linked' }" @click="activeTab = 'linked'">Связанные документы</button>
      </div>

      <template v-if="activeTab === 'main'">
        <!-- Позиции -->
        <table class="items-table">
          <thead>
            <tr>
              <th class="col-chk"></th>
              <th>Наименование</th>
              <th class="col-num-right">Кол-во</th>
              <th class="col-num-right">Остаток</th>
              <th class="col-num-right">Цена</th>
              <th class="col-num-right">Сумма</th>
              <th>{{ docType === 'writeoff' ? 'Причина списания' : 'Причина оприходования' }}</th>
              <th>ГТД</th>
              <th>РНПТ</th>
              <th>Страна</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(it, idx) in form.items" :key="idx">
              <td class="col-chk">
                <button class="row-del" @click="removeItem(idx)" title="Удалить">×</button>
              </td>
              <td>
                <ProductPicker v-model="it.product_id" :products="products" placeholder="— товар —" @picked="(p) => onProductPicked(it, p)" />
              </td>
              <td class="col-num-right"><input v-model.number="it.quantity" type="number" step="0.001" class="qty-input" /></td>
              <td class="col-num-right muted">{{ stockOf(it.product_id) }}</td>
              <td class="col-num-right"><input v-model.number="it.price" type="number" step="0.01" class="price-input" /></td>
              <td class="col-num-right">{{ formatMoney((it.quantity || 0) * (it.price || 0)) }}</td>
              <td><input v-model="it.reason" class="reason-input" :disabled="docType !== 'writeoff'" placeholder="—" /></td>
              <td class="muted">—</td>
              <td class="muted">—</td>
              <td class="muted">—</td>
            </tr>
            <tr v-if="form.items.length === 0">
              <td colspan="10" class="empty">Добавьте позиции</td>
            </tr>
          </tbody>
        </table>

        <div class="add-row">
          <input
            class="add-input"
            placeholder="Добавить позицию — введите наименование, код, штрихкод или артикул"
            :value="addQuery"
            @input="onAddInput"
            @focus="addOpen = true"
            @blur="onAddBlur"
          />
          <button class="btn-from-dict" @click="addItem">Добавить из справочника</button>
          <button class="btn-import" disabled>Импорт</button>
          <div v-if="addOpen && addResults.length > 0" class="ac-list">
            <button
              v-for="pr in addResults"
              :key="pr.id"
              class="ac-item"
              @mousedown.prevent="addPicked(pr)"
            >
              <span class="ac-name">{{ pr.name }}</span>
              <span class="ac-sku">{{ pr.sku || '' }}</span>
            </button>
          </div>
          <div v-if="addOpen && addQuery && addResults.length === 0" class="ac-empty">Ничего не найдено</div>
        </div>

        <!-- Комментарий + Итого -->
        <div class="bottom-row">
          <div class="comment-box">
            <label class="lbl">Комментарий</label>
            <textarea v-model="form.comment" rows="4" class="comment-textarea" :disabled="!canEdit"></textarea>
          </div>
          <div class="totals-box">
            <div class="total-row big">
              <span>Итого:</span>
              <span class="total-num">{{ formatMoney(totalSum) }}</span>
            </div>
            <div class="total-row small">
              <span>Накладные расходы</span>
              <input class="overhead-input" disabled placeholder="0" />
              <span class="overhead-label">Распределить</span>
              <a href="#" class="overhead-link">по цене</a>
            </div>
          </div>
        </div>

        <!-- Задачи / Файлы -->
        <div class="collapsed-section">
          <div class="collapsed-head">▼ Задачи</div>
          <div class="collapsed-body">
            <button class="btn-add">+ Задача</button>
            <span class="muted">Нет задач</span>
          </div>
        </div>
        <div class="collapsed-section">
          <div class="collapsed-head">▼ Файлы</div>
          <div class="collapsed-body">
            <table class="files-table">
              <thead><tr><th>Наименование</th><th>Размер, МБ</th><th>Дата добавления</th><th>Сотрудник</th></tr></thead>
              <tbody><tr><td colspan="4" class="muted">Нет файлов</td></tr></tbody>
            </table>
          </div>
        </div>
      </template>

      <template v-else>
        <div class="empty-block">Связанных документов нет</div>
      </template>
    </template>
  </div>
</template>
<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getDocument, postDocument, cancelDocument, createDocument, updateDocument, type Document } from '../api/documents'
import { listWarehouses, type Warehouse } from '../api/warehouses'
import { listOrganizations, type Organization } from '../api/suppliers'
import { listProducts, type Product } from '../api/products'
import { listContracts, type Contract } from '../api/contracts'
import { apiErrorMessage } from '../api/client'
import { useAuthStore } from '../stores/auth'
import ProductPicker from '../components/ProductPicker.vue'
import { smartFilter } from '../utils/smartSearch'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const doc = ref<Document | null>(null)
const warehouses = ref<Warehouse[]>([])
const organizations = ref<Organization[]>([])
const products = ref<Product[]>([])
const contracts = ref<Contract[]>([])
const loading = ref(false)
const saving = ref(false)
const error = ref<string | null>(null)
const activeTab = ref<'main' | 'linked'>('main')
const addQuery = ref('')
const addOpen = ref(false)
const addResults = computed(() => {
  if (!addQuery.value) return []
  return smartFilter(products.value, addQuery.value, {
    getFields: (p) => [p.name, p.sku ?? null, p.barcode ?? null, p.id.slice(0, 8)],
    limit: 15,
  })
})
const stockMap = ref<Record<string, number>>({})

const isNew = computed(() => route.path === '/documents/new')
const docType = computed(() => {
  if (isNew.value) return String(route.query.type || 'enter')
  return doc.value?.type || 'enter'
})
const canEdit = computed(() => isNew.value || (doc.value?.status === 'draft' && !doc.value?.external_id))

function nowLocal(): string {
  const d = new Date()
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth()+1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

const form = reactive({
  number: '',
  doc_date: nowLocal(),
  status: 'draft',
  warehouse_id: '',
  target_warehouse_id: '',
  organization_id: '',
  contract_id: '',
  expense_item: '',
  comment: '',
  items: [] as Array<{ product_id: string; quantity: number; price: number; reason: string }>,
})

const totalSum = computed(() =>
  form.items.reduce((s, it) => s + (it.quantity || 0) * (it.price || 0), 0)
)

function typeLabel(t?: string): string {
  switch (t) {
    case 'receipt': return 'Оприходование'
    case 'enter': return 'Оприходование'
    case 'shipment': return 'Отгрузка'
    case 'transfer': return 'Перемещение'
    case 'writeoff': return 'Списание'
    case 'inventory': return 'Инвентаризация'
    default: return t ?? '—'
  }
}
function formatMoney(n: number): string {
  return (n || 0).toLocaleString('ru-RU', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}
function stockOf(productId: string): string {
  if (!productId) return '—'
  const v = stockMap.value[productId]
  return v != null ? v.toLocaleString('ru-RU', { maximumFractionDigits: 3 }) : '—'
}

function onAddInput(e: Event) {
  addQuery.value = (e.target as HTMLInputElement).value
}
function onAddBlur() {
  setTimeout(() => { addOpen.value = false }, 150)
}
function addPicked(p: any) {
  form.items.push({ product_id: p.id, quantity: 1, price: p.price || 0, reason: '' })
  addQuery.value = ''
  addOpen.value = false
}

function onProductPicked(it: any, p: any) {
  it.product_id = p.id
  if (!it.price) it.price = p.price || 0
}

function addItem() {
  form.items.push({ product_id: '', quantity: 1, price: 0, reason: '' })
}
function removeItem(i: number) { form.items.splice(i, 1) }

async function load() {
  loading.value = true
  error.value = null
  try {
    const [w, orgs, prods, ct] = await Promise.all([
      listWarehouses(), listOrganizations(), listProducts(true), listContracts(),
    ])
    warehouses.value = w
    organizations.value = orgs
    products.value = prods
    contracts.value = ct

    if (isNew.value) {
      form.organization_id = orgs.find((o) => o.is_default)?.id ?? ''
      form.warehouse_id = w[0]?.id ?? ''
      form.expense_item = 'Списания'
      form.items = []
    } else {
      const id = route.params.id as string
      const d = await getDocument(id)
      doc.value = d
      form.number = d.number
      form.doc_date = d.created_at ? new Date(d.created_at).toISOString().slice(0, 16) : nowLocal()
      form.status = d.status
      form.warehouse_id = d.warehouse_id ?? ''
      form.target_warehouse_id = d.target_warehouse_id ?? ''
      form.organization_id = d.organization_id ?? ''
      form.contract_id = d.contract_id ?? ''
      form.expense_item = (d as any).expense_item ?? ''
      form.comment = d.comment ?? ''
      form.items = (d.items ?? []).map((it: any) => ({
        product_id: it.product_id,
        quantity: it.quantity,
        price: it.price,
        reason: it.reason ?? '',
      }))
    }
  } catch (e) {
    error.value = apiErrorMessage(e)
  } finally {
    loading.value = false
  }
}

async function save() {
  error.value = null
  if (!form.warehouse_id) { error.value = 'Выберите склад'; return }
  if (form.items.some((it) => !it.product_id || !it.quantity)) { error.value = 'Заполните позиции'; return }
  saving.value = true
  try {
    if (isNew.value) {
      const d = await createDocument({
        type: docType.value as any,
        warehouse_id: form.warehouse_id,
        target_warehouse_id: form.target_warehouse_id || undefined,
        organization_id: form.organization_id || undefined,
        contract_id: form.contract_id || undefined,
        expense_item: form.expense_item || undefined,
        comment: form.comment || undefined,
        items: form.items.map((it) => ({ product_id: it.product_id, quantity: it.quantity, price: it.price })),
      } as any)
      router.replace('/documents/' + d.id)
    } else if (doc.value) {
      const upd = await updateDocument(doc.value.id, {
        warehouse_id: form.warehouse_id,
        target_warehouse_id: form.target_warehouse_id || undefined,
        organization_id: form.organization_id || undefined,
        contract_id: form.contract_id || undefined,
        expense_item: form.expense_item || undefined,
        comment: form.comment || undefined,
        items: form.items.map((it) => ({ product_id: it.product_id, quantity: it.quantity, price: it.price })),
      } as any)
      doc.value = upd
      form.comment = upd.comment ?? ''
      form.expense_item = (upd as any).expense_item ?? ''
    }
  } catch (e) {
    error.value = apiErrorMessage(e)
  } finally {
    saving.value = false
  }
}

async function doPost() {
  if (!doc.value) return
  saving.value = true
  try {
    doc.value = await postDocument(doc.value.id)
    form.status = doc.value.status
  } catch (e) { error.value = apiErrorMessage(e) }
  finally { saving.value = false }
}
async function doCancel() {
  if (!doc.value || !confirm('Отменить документ?')) return
  saving.value = true
  try {
    doc.value = await cancelDocument(doc.value.id)
    form.status = doc.value.status
  } catch (e) { error.value = apiErrorMessage(e) }
  finally { saving.value = false }
}

function close() {
  const t = docType.value
  router.push(
    t === 'enter' ? '/stock/receipts' :
    t === 'writeoff' ? '/stock/writeoffs' :
    t === 'transfer' ? '/stock/transfers' :
    '/documents'
  )
}

onMounted(load)
watch(() => route.params.id, load)
</script>
<style scoped>
.card-page { font-size: 13px; }
.error-box { background: #ffebe9; border: 1px solid #ff8182; color: #82071e; padding: 8px 12px; border-radius: 4px; margin-bottom: 10px; }
.empty-block { padding: 40px; text-align: center; color: #8c959f; background: #fff; border: 1px solid #eaeef2; border-radius: 4px; }

/* Шапка */
.doc-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 6px; padding-bottom: 6px; border-bottom: 1px solid #eaeef2; }
.doc-header-left { display: flex; flex-direction: column; gap: 4px; flex: 1; }
.header-title-row { display: flex; align-items: center; gap: 6px; font-size: 16px; font-weight: 600; color: #1f2328; flex-wrap: wrap; }
.doc-type-label { font-weight: 600; }
.num-input { width: 90px; height: 26px; padding: 0 6px; font-size: 15px; font-weight: 600; border: 1px solid #d0d7de; border-radius: 3px; background: #fff; color: #1f2328; }
.num-input:disabled { background: #f6f8fa; color: #57606a; }
.from-sep { font-weight: 400; color: #57606a; font-size: 13px; }
.date-input { height: 26px; padding: 0 6px; font-size: 13px; border: 1px solid #d0d7de; border-radius: 3px; background: #fff; color: #1f2328; }
.status-select { height: 26px; padding: 0 6px; font-size: 13px; border: 1px solid #d0d7de; border-radius: 3px; }
.doc-header-right .user-box { text-align: right; }
.doc-header-right .user-name { font-size: 13px; font-weight: 500; color: #1f2328; }
.doc-header-right .user-role { font-size: 11px; color: #8c959f; }

/* Тулбар */
.toolbar { display: flex; gap: 4px; margin-bottom: 10px; flex-wrap: wrap; padding-bottom: 8px; border-bottom: 1px solid #eaeef2; }
.btn-save { height: 26px; padding: 0 14px; background: #a8ce5e; border: 1px solid #a8ce5e; border-radius: 3px; color: #1f2328; font-weight: 600; cursor: pointer; font-size: 12px; }
.btn-plain { height: 26px; padding: 0 10px; background: #fff; border: 1px solid #d0d7de; border-radius: 3px; color: #1f2328; cursor: pointer; font-size: 12px; }
.btn-save { height: 30px; padding: 0 14px; background: #a8ce5e; border: 1px solid #a8ce5e; border-radius: 4px; color: #1f2328; font-weight: 600; cursor: pointer; font-size: 13px; }
.btn-save:hover:not(:disabled) { background: #97bd4b; }
.btn-save:disabled { opacity: .6; cursor: default; }
.btn-plain { height: 30px; padding: 0 12px; background: #fff; border: 1px solid #d0d7de; border-radius: 4px; color: #1f2328; cursor: pointer; font-size: 13px; }
.btn-plain:hover:not(:disabled) { background: #f6f8fa; }
.btn-plain:disabled { opacity: .5; cursor: default; }

/* Поля */
.fields-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 6px 24px; margin-bottom: 8px; max-width: 1100px; }
.field { display: flex; flex-direction: column; gap: 2px; }
.field > label { font-size: 12px; color: #57606a; }
.with-plus { display: flex; gap: 4px; }
.with-plus select { flex: 1; height: 28px; padding: 0 8px; font-size: 13px; border: 1px solid #d0d7de; border-radius: 3px; background: #fff; color: #1f2328; }
.plus-btn { width: 28px; height: 28px; border: 1px solid #d0d7de; background: #fff; border-radius: 3px; cursor: pointer; color: #2c5d9c; }
.plus-btn:hover { background: #f6f8fa; }

/* Вкладки */
.tabs { display: flex; gap: 0; border-bottom: 1px solid #d8dee4; margin-bottom: 8px; margin-top: 4px; }
.tab { padding: 6px 12px; background: transparent; border: none; border-bottom: 2px solid transparent; color: #57606a; cursor: pointer; font-size: 12px; }
.tab.active { color: #2c5d9c; border-bottom-color: #2c5d9c; font-weight: 600; }

/* Таблица позиций */
.items-table { width: 100%; border-collapse: collapse; background: #fff; font-size: 13px; margin-bottom: 8px; }
.items-table thead th { background: #fff; color: #2c5d9c; font-weight: 500; padding: 6px 8px; text-align: left; border-bottom: 1px solid #d8dee4; font-size: 12px; white-space: nowrap; }
.items-table tbody td { padding: 4px 8px; border-bottom: 1px solid #eaeef2; vertical-align: middle; }
.col-chk { width: 30px; }
.col-num-right { text-align: right; font-variant-numeric: tabular-nums; }
.row-del { width: 22px; height: 22px; border: none; background: transparent; color: #cf222e; cursor: pointer; font-size: 16px; }
.product-select { width: 100%; height: 26px; padding: 0 6px; font-size: 13px; border: 1px solid #d0d7de; border-radius: 3px; background: #fff; }
.qty-input, .price-input { width: 80px; height: 26px; padding: 0 6px; font-size: 13px; border: 1px solid #d0d7de; border-radius: 3px; text-align: right; background: #fff; }
.reason-input { width: 100%; height: 26px; padding: 0 6px; font-size: 13px; border: 1px solid #d0d7de; border-radius: 3px; background: #fff; }
.reason-input:disabled { background: #f6f8fa; }
.empty { text-align: center; padding: 24px; color: #8c959f; }
.add-row { position: relative; display: flex; gap: 6px; margin-bottom: 10px; align-items: center; }
.add-input {
  flex: 1; height: 28px; padding: 0 10px; font-size: 13px;
  border: 1px solid #2c5d9c; border-radius: 3px; background: #fff;
}
.add-input::placeholder { color: #8c959f; }
.btn-from-dict, .btn-import {
  height: 28px; padding: 0 12px; font-size: 12px;
  border: 1px solid #d0d7de; background: #fff; color: #1f2328;
  border-radius: 3px; cursor: pointer;
}
.btn-from-dict:hover, .btn-import:hover:not(:disabled) { background: #f6f8fa; }
.btn-import { margin-left: auto; }
.btn-import:disabled { opacity: .5; cursor: default; }
.ac-list {
  position: absolute; top: 32px; left: 0; right: 0; z-index: 100;
  background: #fff; border: 1px solid #d0d7de; border-radius: 4px;
  box-shadow: 0 4px 16px rgba(0,0,0,0.12);
  max-height: 260px; overflow-y: auto; padding: 4px 0;
}
.ac-item {
  display: flex; justify-content: space-between; gap: 8px; width: 100%;
  padding: 6px 10px; border: none; background: transparent;
  font-size: 13px; color: #1f2328; cursor: pointer; text-align: left;
}
.ac-item:hover { background: #f6f8fa; }
.ac-name { flex: 1; }
.ac-sku { color: #8c959f; font-family: monospace; font-size: 12px; }
.ac-empty {
  position: absolute; top: 32px; left: 0; right: 0; z-index: 100;
  background: #fff; border: 1px solid #d0d7de; border-radius: 4px;
  box-shadow: 0 4px 16px rgba(0,0,0,0.12);
  padding: 8px 12px; font-size: 12px; color: #8c959f;
}

/* Низ */
.bottom-row { display: grid; grid-template-columns: 1fr 380px; gap: 24px; margin-bottom: 16px; }
.comment-box { display: flex; flex-direction: column; gap: 4px; }
.comment-box .lbl { font-size: 12px; color: #57606a; }
.comment-textarea { width: 100%; padding: 8px; font-size: 13px; border: 1px solid #d0d7de; border-radius: 3px; resize: vertical; }
.comment-textarea:disabled { background: #f6f8fa; }
.totals-box { background: #fff; padding: 6px 0; }
.total-row { display: flex; justify-content: space-between; align-items: center; }
.total-row.big { font-size: 15px; font-weight: 600; margin-bottom: 8px; }
.total-num { font-variant-numeric: tabular-nums; }
.total-row.small { gap: 8px; font-size: 12px; color: #57606a; }
.overhead-input { width: 80px; height: 26px; padding: 0 6px; font-size: 12px; border: 1px solid #d0d7de; border-radius: 3px; text-align: right; }
.overhead-label { color: #57606a; }
.overhead-link { color: #2c5d9c; text-decoration: underline; font-size: 12px; }

/* Задачи/Файлы */
.collapsed-section { background: #fff; border: 1px solid #eaeef2; border-radius: 4px; padding: 8px 12px; margin-bottom: 8px; }
.collapsed-head { font-weight: 600; font-size: 13px; color: #1f2328; padding: 4px 0; cursor: default; }
.collapsed-body { padding: 8px 0; display: flex; gap: 10px; align-items: center; }
.muted { color: #8c959f; }
.files-table { width: 100%; border-collapse: collapse; font-size: 12px; }
.files-table th, .files-table td { padding: 6px 8px; border-bottom: 1px solid #eaeef2; text-align: left; }
</style>
