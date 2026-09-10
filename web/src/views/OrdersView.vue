<template>
  <div>
    <div class="head">
      <h1>Заказы</h1>
      <button class="primary" @click="openCreate">+ Новый заказ</button>
    </div>

    <div class="filters card">
      <label>Статус
        <select v-model="filterStatus" @change="load">
          <option value="">— все —</option>
          <option value="draft">Черновик</option>
          <option value="confirmed">Подтверждён</option>
          <option value="shipped">Отгружен</option>
          <option value="cancelled">Отменён</option>
        </select>
      </label>
    </div>

    <div v-if="error" class="error-box">{{ error }}</div>

    <table>
      <thead>
        <tr>
          <th>Номер</th>
          <th>Покупатель</th>
          <th>Склад</th>
          <th>Сумма</th>
          <th>Статус</th>
          <th>Дата</th>
          <th style="width: 300px"></th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="loading"><td colspan="7" class="muted">Загрузка...</td></tr>
        <tr v-else-if="items.length === 0"><td colspan="7" class="muted">Нет заказов</td></tr>
        <tr v-for="o in items" :key="o.id">
          <td class="mono">{{ o.number }}</td>
          <td>{{ customerName(o.customer_id) }}</td>
          <td>{{ warehouseName(o.warehouse_id) }}</td>
          <td><strong>{{ o.total.toFixed(2) }}</strong> {{ o.currency }}</td>
          <td><span :class="['pill', o.status]">{{ statusLabel(o.status) }}</span></td>
          <td class="muted">{{ formatDate(o.created_at) }}</td>
          <td>
            <button v-if="o.status === 'draft'" class="primary" @click="onConfirm(o)">Подтвердить</button>
            <button v-if="o.status === 'confirmed'" class="primary" @click="onShip(o)">Отгрузить</button>
            <button v-if="o.status === 'draft' || o.status === 'confirmed'" class="danger" @click="onCancel(o)">Отменить</button>
            <button @click="showDetails(o)">Детали</button>
          </td>
        </tr>
      </tbody>
    </table>

    <!-- модалка создания -->
    <div v-if="showCreate" class="modal-backdrop" @click.self="closeCreate">
      <form class="card modal" @submit.prevent="onCreate">
        <h2>Новый заказ</h2>
        <div v-if="createError" class="error-box">{{ createError }}</div>

        <label>Покупатель*
          <select v-model="form.customer_id" required>
            <option value="">— выберите —</option>
            <option v-for="c in customers" :key="c.id" :value="c.id">{{ c.name }}</option>
          </select>
        </label>

        <label>Склад*
          <select v-model="form.warehouse_id" required>
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
    <div v-if="detailOrder" class="modal-backdrop" @click.self="detailOrder = null">
      <div class="card modal">
        <h2>Заказ {{ detailOrder.number }}</h2>
        <p>
          <strong>Покупатель:</strong> {{ customerName(detailOrder.customer_id) }}<br />
          <strong>Склад:</strong> {{ warehouseName(detailOrder.warehouse_id) }}<br />
          <strong>Статус:</strong> {{ statusLabel(detailOrder.status) }}<br />
          <strong>Сумма:</strong> {{ detailOrder.total.toFixed(2) }} {{ detailOrder.currency }}
        </p>
        <p v-if="detailOrder.warehouse_doc_id" class="muted mono">
          Документ отгрузки: {{ detailOrder.warehouse_doc_id }}
        </p>
        <table>
          <thead><tr><th>Товар</th><th>Кол-во</th><th>Цена</th></tr></thead>
          <tbody>
            <tr v-for="it in detailOrder.items" :key="it.id">
              <td>{{ productName(it.product_id) }}</td>
              <td>{{ it.quantity }}</td>
              <td>{{ it.price.toFixed(2) }}</td>
            </tr>
          </tbody>
        </table>
        <div class="modal-actions">
          <button @click="detailOrder = null">Закрыть</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import {
  listOrders, createOrder, confirmOrder, shipOrder, cancelOrder, getOrder,
  type Order,
} from '../api/orders'
import { listCustomers, type Customer } from '../api/customers'
import { listWarehouses, type Warehouse } from '../api/warehouses'
import { listProducts, type Product } from '../api/products'
import { apiErrorMessage } from '../api/client'

const items = ref<Order[]>([])
const customers = ref<Customer[]>([])
const warehouses = ref<Warehouse[]>([])
const products = ref<Product[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const filterStatus = ref('')

const showCreate = ref(false)
const saving = ref(false)
const createError = ref<string | null>(null)
const form = reactive({
  customer_id: '',
  warehouse_id: '',
  comment: '',
  items: [] as Array<{ product_id: string; quantity: number; price: number }>,
})

const detailOrder = ref<Order | null>(null)

function customerName(id: string) { return customers.value.find((c) => c.id === id)?.name ?? id.slice(0, 8) }
function warehouseName(id: string) { return warehouses.value.find((w) => w.id === id)?.name ?? id.slice(0, 8) }
function productName(id: string) { return products.value.find((p) => p.id === id)?.name ?? id.slice(0, 8) }
function statusLabel(s: string) {
  return { draft: 'Черновик', confirmed: 'Подтверждён', shipped: 'Отгружен', cancelled: 'Отменён' }[s] ?? s
}
function formatDate(s: string) { return new Date(s).toLocaleString('ru-RU') }

async function load() {
  loading.value = true
  error.value = null
  try {
    items.value = await listOrders({ status: filterStatus.value || undefined })
  } catch (e) {
    error.value = apiErrorMessage(e)
  } finally {
    loading.value = false
  }
}

function openCreate() {
  form.customer_id = ''
  form.warehouse_id = ''
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
    await createOrder({
      customer_id: form.customer_id,
      warehouse_id: form.warehouse_id,
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

async function onConfirm(o: Order) {
  if (!confirm(`Подтвердить заказ ${o.number}?`)) return
  try { await confirmOrder(o.id); await load() } catch (e) { error.value = apiErrorMessage(e) }
}

async function onShip(o: Order) {
  if (!confirm(`Отгрузить заказ ${o.number}? Будет создан документ отгрузки на складе.`)) return
  try {
    await shipOrder(o.id)
    await load()
  } catch (e) {
    alert('Не удалось отгрузить: ' + apiErrorMessage(e))
  }
}

async function onCancel(o: Order) {
  if (!confirm(`Отменить заказ ${o.number}?`)) return
  try { await cancelOrder(o.id); await load() } catch (e) { error.value = apiErrorMessage(e) }
}

async function showDetails(o: Order) {
  try { detailOrder.value = await getOrder(o.id) } catch (e) { error.value = apiErrorMessage(e) }
}

onMounted(async () => {
  try {
    const [c, w, p] = await Promise.all([listCustomers(), listWarehouses(), listProducts(true)])
    customers.value = c
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
.pill.confirmed { background: #e3f2fd; color: #0d47a1; }
.pill.shipped { background: #e8f5e9; color: var(--success); }
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