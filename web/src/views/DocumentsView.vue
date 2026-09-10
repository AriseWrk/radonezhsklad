<template>
  <div>
    <div class="head">
      <h1>Документы</h1>
      <button class="primary" @click="openCreate">+ Новый документ</button>
    </div>

    <div class="filters card">
      <label>Тип
        <select v-model="filterType" @change="load">
          <option value="">— все —</option>
          <option value="receipt">Приёмка</option>
          <option value="shipment">Отгрузка</option>
          <option value="transfer">Перемещение</option>
        </select>
      </label>
      <label>Статус
        <select v-model="filterStatus" @change="load">
          <option value="">— все —</option>
          <option value="draft">Черновик</option>
          <option value="posted">Проведён</option>
          <option value="cancelled">Отменён</option>
        </select>
      </label>
      <label>Склад
        <select v-model="filterWarehouse" @change="load">
          <option value="">— все —</option>
          <option v-for="w in warehouses" :key="w.id" :value="w.id">{{ w.name }}</option>
        </select>
      </label>
    </div>

    <div v-if="error" class="error-box">{{ error }}</div>

    <table>
      <thead>
        <tr>
          <th>Номер</th>
          <th>Тип</th>
          <th>Склад</th>
          <th>Статус</th>
          <th>Позиций</th>
          <th>Дата</th>
          <th style="width: 220px"></th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="loading"><td colspan="7" class="muted">Загрузка...</td></tr>
        <tr v-else-if="items.length === 0"><td colspan="7" class="muted">Нет документов</td></tr>
        <tr v-for="d in items" :key="d.id">
          <td class="mono">{{ d.number }}</td>
          <td>{{ typeLabel(d.type) }}</td>
          <td>{{ warehouseName(d.warehouse_id) }}</td>
          <td><span :class="['pill', d.status]">{{ statusLabel(d.status) }}</span></td>
          <td>{{ d.items?.length ?? '—' }}</td>
          <td class="muted">{{ formatDate(d.created_at) }}</td>
          <td>
            <button v-if="d.status === 'draft'" class="primary" @click="onPost(d)">Провести</button>
            <button v-if="d.status === 'posted'" class="danger" @click="onCancel(d)">Отменить</button>
            <button @click="showDetails(d)">Детали</button>
          </td>
        </tr>
      </tbody>
    </table>

    <!-- модалка создания -->
    <div v-if="showCreate" class="modal-backdrop" @click.self="closeCreate">
      <form class="card modal" @submit.prevent="onCreate">
        <h2>Новый документ</h2>
        <div v-if="createError" class="error-box">{{ createError }}</div>

        <label>Тип*
          <select v-model="form.type" required>
            <option value="receipt">Приёмка</option>
            <option value="shipment">Отгрузка</option>
            <option value="transfer">Перемещение</option>
          </select>
        </label>

        <label>Склад*
          <select v-model="form.warehouse_id" required>
            <option value="">— выберите —</option>
            <option v-for="w in warehouses" :key="w.id" :value="w.id">{{ w.name }}</option>
          </select>
        </label>

        <label v-if="form.type === 'transfer'">Целевой склад*
          <select v-model="form.target_warehouse_id" required>
            <option value="">— выберите —</option>
            <option v-for="w in warehouses" :key="w.id" :value="w.id">{{ w.name }}</option>
          </select>
        </label>

        <label>Комментарий
          <input v-model="form.comment" />
        </label>

        <div class="items-section">
          <div class="items-head">
            <span>Позиции</span>
            <button type="button" @click="addItem">+ Добавить</button>
          </div>
          <div v-for="(it, idx) in form.items" :key="idx" class="item-row">
            <select v-model="it.product_id" required>
              <option value="">— товар —</option>
              <option v-for="p in products" :key="p.id" :value="p.id">{{ p.name }}</option>
            </select>
            <input v-model.number="it.quantity" type="number" step="0.001" min="0.001" placeholder="Кол-во" required />
            <input v-model.number="it.price" type="number" step="0.01" min="0" placeholder="Цена" />
            <button type="button" class="danger" @click="removeItem(idx)">×</button>
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

    <!-- модалка деталей -->
    <div v-if="detailDoc" class="modal-backdrop" @click.self="detailDoc = null">
      <div class="card modal">
        <h2>Документ {{ detailDoc.number }}</h2>
        <p>
          <strong>Тип:</strong> {{ typeLabel(detailDoc.type) }}<br />
          <strong>Статус:</strong> {{ statusLabel(detailDoc.status) }}<br />
          <strong>Склад:</strong> {{ warehouseName(detailDoc.warehouse_id) }}
          <template v-if="detailDoc.target_warehouse_id">
            → {{ warehouseName(detailDoc.target_warehouse_id) }}
          </template>
        </p>
        <table>
          <thead>
            <tr><th>Товар</th><th>Кол-во</th><th>Цена</th></tr>
          </thead>
          <tbody>
            <tr v-for="it in detailDoc.items" :key="it.id">
              <td>{{ productName(it.product_id) }}</td>
              <td>{{ it.quantity }}</td>
              <td>{{ it.price.toFixed(2) }}</td>
            </tr>
          </tbody>
        </table>
        <div class="modal-actions">
          <button @click="detailDoc = null">Закрыть</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import {
  listDocuments, createDocument, postDocument, cancelDocument, getDocument,
  type Document, type DocType,
} from '../api/documents'
import { listWarehouses, type Warehouse } from '../api/warehouses'
import { listProducts, type Product } from '../api/products'
import { apiErrorMessage } from '../api/client'

const items = ref<Document[]>([])
const warehouses = ref<Warehouse[]>([])
const products = ref<Product[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const filterType = ref('')
const filterStatus = ref('')
const filterWarehouse = ref('')

const showCreate = ref(false)
const saving = ref(false)
const createError = ref<string | null>(null)
const form = reactive({
  type: 'receipt' as DocType,
  warehouse_id: '',
  target_warehouse_id: '',
  comment: '',
  items: [] as Array<{ product_id: string; quantity: number; price: number }>,
})

const detailDoc = ref<Document | null>(null)

function typeLabel(t: string) {
  return { receipt: 'Приёмка', shipment: 'Отгрузка', transfer: 'Перемещение', inventory: 'Инвентаризация' }[t] ?? t
}
function statusLabel(s: string) {
  return { draft: 'Черновик', posted: 'Проведён', cancelled: 'Отменён' }[s] ?? s
}
function warehouseName(id: string) {
  return warehouses.value.find((w) => w.id === id)?.name ?? id.slice(0, 8)
}
function productName(id: string) {
  return products.value.find((p) => p.id === id)?.name ?? id.slice(0, 8)
}
function formatDate(s: string) { return new Date(s).toLocaleString('ru-RU') }

async function load() {
  loading.value = true
  error.value = null
  try {
    items.value = await listDocuments({
      type: filterType.value || undefined,
      status: filterStatus.value || undefined,
      warehouse_id: filterWarehouse.value || undefined,
    })
  } catch (e) {
    error.value = apiErrorMessage(e)
  } finally {
    loading.value = false
  }
}

function openCreate() {
  form.type = 'receipt'
  form.warehouse_id = ''
  form.target_warehouse_id = ''
  form.comment = ''
  form.items = [{ product_id: '', quantity: 1, price: 0 }]
  createError.value = null
  showCreate.value = true
}
function closeCreate() { showCreate.value = false }

function addItem() { form.items.push({ product_id: '', quantity: 1, price: 0 }) }
function removeItem(i: number) { form.items.splice(i, 1) }

async function onCreate() {
  createError.value = null
  saving.value = true
  try {
    if (form.items.length === 0) throw new Error('Добавьте хотя бы одну позицию')
    if (form.items.some((it) => !it.product_id || it.quantity <= 0)) throw new Error('Заполните все позиции')

    await createDocument({
      type: form.type,
      warehouse_id: form.warehouse_id,
      target_warehouse_id: form.type === 'transfer' ? form.target_warehouse_id : undefined,
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
  if (!confirm(`Провести документ ${d.number}?`)) return
  try { await postDocument(d.id); await load() } catch (e) { error.value = apiErrorMessage(e) }
}

async function onCancel(d: Document) {
  if (!confirm(`Отменить документ ${d.number}?`)) return
  try { await cancelDocument(d.id); await load() } catch (e) { error.value = apiErrorMessage(e) }
}

async function showDetails(d: Document) {
  try {
    detailDoc.value = await getDocument(d.id)
  } catch (e) {
    error.value = apiErrorMessage(e)
  }
}

onMounted(async () => {
  try {
    const [w, p] = await Promise.all([listWarehouses(), listProducts(true)])
    warehouses.value = w
    products.value = p
  } catch { /* ignore */ }
  await load()
})
</script>

<style scoped>
.head { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
h1 { margin: 0; }
.muted { color: var(--muted); }
.mono { font-family: monospace; font-size: 13px; }
.filters { display: flex; gap: 16px; margin-bottom: 16px; }
.filters label { flex: 1; display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: var(--muted); }
.filters select { color: var(--text); }
.pill { padding: 2px 8px; border-radius: 10px; font-size: 12px; font-weight: 600; }
.pill.draft { background: #fff8e1; color: #9a6a00; }
.pill.posted { background: #e8f5e9; color: var(--success); }
.pill.cancelled { background: #f3f4f6; color: var(--muted); }

.modal-backdrop { position: fixed; inset: 0; background: rgba(0,0,0,0.4); display: flex; align-items: center; justify-content: center; padding: 20px; z-index: 100; }
.modal { width: 640px; max-width: 100%; max-height: 90vh; overflow-y: auto; display: flex; flex-direction: column; gap: 12px; }
.modal h2 { margin: 0; font-size: 18px; }
label { display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: var(--muted); }
label input, label select { color: var(--text); }
.items-section { border: 1px solid var(--border); border-radius: 6px; padding: 12px; }
.items-head { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; font-size: 13px; color: var(--muted); }
.item-row { display: grid; grid-template-columns: 2fr 1fr 1fr 40px; gap: 8px; margin-bottom: 6px; }
.modal-actions { display: flex; justify-content: flex-end; gap: 10px; margin-top: 6px; }
</style>