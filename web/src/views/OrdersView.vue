<template>
  <div class="page">
    <div class="ms-title">
      <button class="ms-help" title="Справка"><MsIcon name="help" :size="14" /></button>
      <span>Заказы</span>
      <button class="ms-refresh" @click="load" title="Обновить"><MsIcon name="refresh" :size="14" /></button>
    </div>

    <div class="ms-toolbar">
      <MsButton variant="primary" icon="plus" @click="openCreate">Заказ</MsButton>
      <MsButton icon="filter">Фильтр</MsButton>
      <input v-model="search" class="ms-input" placeholder="Номер или комментарий" />
      <div class="ms-counter">{{ filtered.length }}</div>
      <select class="ms-select" v-model="filterStatus" @change="load">
        <option value="">Статус</option>
        <option value="draft">Черновик</option>
        <option value="confirmed">Подтверждён</option>
        <option value="shipped">Отгружен</option>
        <option value="cancelled">Отменён</option>
      </select>
      <MsButton icon="print">Печать</MsButton>
      <MsButton variant="icon" icon="gear" title="Настройки" />
    </div>

    <div v-if="error" class="error-box">{{ error }}</div>

    <table class="ms-table2">
      <thead>
        <tr>
          <th class="col-num">Номер</th>
          <th>Покупатель</th>
          <th>Склад</th>
          <th class="col-num-right">Сумма</th>
          <th class="col-status">Статус</th>
          <th class="col-date">Дата</th>
          <th class="col-actions"></th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="loading"><td colspan="7" class="empty">Загрузка...</td></tr>
        <tr v-else-if="filtered.length === 0"><td colspan="7" class="empty">Нет заказов</td></tr>
        <tr v-else v-for="o in filtered" :key="o.id" class="row">
          <td class="col-num"><span class="link">{{ o.number }}</span></td>
          <td>{{ customerName(o.customer_id) }}</td>
          <td>{{ warehouseName(o.warehouse_id) }}</td>
          <td class="col-num-right"><b>{{ o.total.toFixed(2) }}</b> {{ o.currency }}</td>
          <td class="col-status"><span class="badge" :class="'bg-' + o.status">{{ statusLabel(o.status) }}</span></td>
          <td class="col-date">{{ formatDate(o.created_at) }}</td>
          <td class="col-actions" @click.stop>
            <button v-if="o.status === 'draft'" class="btn-link-ms" @click="onConfirm(o)">Подтвердить</button>
            <button v-if="o.status === 'confirmed'" class="btn-link-ms" @click="onShip(o)">Отгрузить</button>
            <button v-if="o.status === 'draft' || o.status === 'confirmed'" class="btn-link-ms danger" @click="onCancel(o)">Отменить</button>
            <button class="row-menu" @click="showDetails(o)"><MsIcon name="dots" :size="14" /></button>
          </td>
        </tr>
      </tbody>
    </table>

    <div class="ms-footer2">
      <div class="pager"><span class="range">Всего: {{ filtered.length }}</span></div>
      <div class="totals-inline"><span>Сумма: <b>{{ totalSum.toFixed(2) }}</b></span></div>
    </div>

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
            <button type="button" class="btn-link" @click="addItem">+ Добавить</button>
          </div>
          <div v-for="(it, idx) in form.items" :key="idx" class="item-row">
            <select v-model="it.product_id" required>
              <option value="">— товар —</option>
              <option v-for="p in products" :key="p.id" :value="p.id">{{ p.name }}</option>
            </select>
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

    <div v-if="detailOrder" class="modal-backdrop" @click.self="detailOrder = null">
      <div class="card modal big">
        <h2>Заказ {{ detailOrder.number }}</h2>
        <div class="doc-info">
          <div><span class="lbl">Покупатель:</span> {{ customerName(detailOrder.customer_id) }}</div>
          <div><span class="lbl">Склад:</span> {{ warehouseName(detailOrder.warehouse_id) }}</div>
          <div><span class="lbl">Статус:</span> {{ statusLabel(detailOrder.status) }}</div>
          <div><span class="lbl">Сумма:</span> {{ detailOrder.total.toFixed(2) }} {{ detailOrder.currency }}</div>
          <div v-if="detailOrder.warehouse_doc_id" style="grid-column: 1/-1">
            <span class="lbl">Документ отгрузки:</span> <span class="mono">{{ detailOrder.warehouse_doc_id }}</span>
          </div>
        </div>
        <table class="ms-table2">
          <thead><tr><th>Товар</th><th class="col-num-right">Кол-во</th><th class="col-num-right">Цена</th></tr></thead>
          <tbody>
            <tr v-for="it in detailOrder.items" :key="it.id">
              <td>{{ productName(it.product_id) }}</td>
              <td class="col-num-right">{{ it.quantity }}</td>
              <td class="col-num-right">{{ it.price.toFixed(2) }}</td>
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
import { computed, onMounted, reactive, ref } from 'vue'
import {
  listOrders, createOrder, confirmOrder, shipOrder, cancelOrder, getOrder,
  type Order,
} from '../api/orders'
import { listCustomers, type Customer } from '../api/customers'
import { listWarehouses, type Warehouse } from '../api/warehouses'
import { listProducts, type Product } from '../api/products'
import { apiErrorMessage } from '../api/client'
import MsButton from '../components/MsButton.vue'
import MsIcon from '../components/MsIcon.vue'

const items = ref<Order[]>([])
const customers = ref<Customer[]>([])
const warehouses = ref<Warehouse[]>([])
const products = ref<Product[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const filterStatus = ref('')
const search = ref('')

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

const filtered = computed(() => {
  const q = search.value.trim().toLowerCase()
  if (!q) return items.value
  return items.value.filter((o) =>
    (o.number ?? '').toLowerCase().includes(q) ||
    (o.comment ?? '').toLowerCase().includes(q)
  )
})
const totalSum = computed(() => filtered.value.reduce((s, o) => s + (o.total ?? 0), 0))

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
  if (!confirm(`Отгрузить заказ ${o.number}?`)) return
  try { await shipOrder(o.id); await load() } catch (e) { error.value = apiErrorMessage(e) }
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
.page { font-size: 13px; }
.ms-title { display: flex; align-items: center; gap: 8px; font-size: 20px; font-weight: 600; color: #1f2328; margin-bottom: 12px; }
.ms-help { width: 20px; height: 20px; border-radius: 50%; border: 1px solid #b8c0c8; background: transparent; color: #57606a; cursor: pointer; display: inline-flex; align-items: center; justify-content: center; padding: 0; }
.ms-help:hover { background: #f0f2f5; }
.ms-refresh { width: 24px; height: 24px; padding: 0; display: inline-flex; align-items: center; justify-content: center; border: none; background: transparent; color: #57606a; cursor: pointer; }
.ms-refresh:hover { color: #2c5d9c; }
.ms-toolbar { display: flex; align-items: center; gap: 6px; margin-bottom: 12px; flex-wrap: wrap; }
.ms-input { flex: 1; max-width: 320px; height: 30px; padding: 0 10px; font-size: 13px; border: 1px solid #d0d7de; border-radius: 4px; background: #fff; color: #1f2328; }
.ms-input::placeholder { color: #8c959f; }
.ms-counter { min-width: 40px; height: 30px; padding: 0 10px; display: inline-flex; align-items: center; justify-content: center; border: 1px solid #d0d7de; border-radius: 4px; background: #fff; color: #8c959f; font-variant-numeric: tabular-nums; font-size: 13px; }
.ms-select { height: 30px; padding: 0 8px; font-size: 13px; border: 1px solid #d0d7de; border-radius: 4px; background: #fff; color: #1f2328; max-width: 180px; }

.ms-table2 { width: 100%; border-collapse: collapse; background: #fff; font-size: 13px; }
.ms-table2 thead th { background: #fff; color: #2c5d9c; font-weight: 500; padding: 8px 10px; text-align: left; border-bottom: 1px solid #d8dee4; white-space: nowrap; font-size: 12px; }
.ms-table2 tbody td { padding: 7px 10px; border-bottom: 1px solid #eaeef2; vertical-align: middle; }
.ms-table2 tbody tr.row:hover { background: #f6f8fa; }
.ms-table2 .col-num { width: 100px; }
.ms-table2 .col-date { width: 160px; color: #57606a; }
.ms-table2 .col-num-right { text-align: right; font-variant-numeric: tabular-nums; }
.ms-table2 .col-status { width: 130px; }
.ms-table2 .col-actions { width: 240px; text-align: right; white-space: nowrap; }
.ms-table2 .link { color: #2c5d9c; font-weight: 500; }
.ms-table2 .empty { text-align: center; padding: 24px; color: #8c959f; }

.badge { display: inline-block; padding: 2px 10px; border-radius: 3px; font-size: 11px; font-weight: 600; color: #fff; }
.bg-draft     { background: #8c959f; }
.bg-confirmed { background: #2196f3; }
.bg-shipped   { background: #4caf50; }
.bg-cancelled { background: #cf222e; }

.btn-link-ms { border: none; background: transparent; color: #2c5d9c; cursor: pointer; font-size: 12px; padding: 4px 8px; border-radius: 3px; }
.btn-link-ms:hover { background: #f0f2f5; }
.btn-link-ms.danger { color: #cf222e; }
.btn-link-ms.danger:hover { background: #ffecec; }
.row-menu { width: 22px; height: 22px; padding: 0; border: none; background: transparent; color: #57606a; cursor: pointer; border-radius: 3px; vertical-align: middle; }
.row-menu:hover { background: #eaeef2; color: #1f2328; }

.ms-footer2 { display: flex; align-items: center; justify-content: space-between; padding: 6px 10px; font-size: 12px; color: #57606a; background: #fff; border-top: 1px solid #eaeef2; }
.pager .range { font-variant-numeric: tabular-nums; }
.totals-inline { font-variant-numeric: tabular-nums; }

.modal-backdrop { position: fixed; inset: 0; background: rgba(0,0,0,0.4); display: flex; align-items: center; justify-content: center; padding: 20px; z-index: 100; }
.modal { background: #fff; border-radius: 4px; width: 480px; max-width: 100%; padding: 20px; display: flex; flex-direction: column; gap: 10px; max-height: 90vh; overflow-y: auto; }
.modal.big { width: 720px; }
.modal h2 { margin: 0 0 4px; font-size: 18px; }
.modal label { font-size: 13px; color: #444; display: flex; flex-direction: column; gap: 4px; }
.modal-actions { display: flex; justify-content: flex-end; gap: 10px; margin-top: 6px; }
.items-section { display: flex; flex-direction: column; gap: 8px; margin-top: 6px; }
.items-head { display: flex; justify-content: space-between; align-items: center; font-size: 13px; font-weight: 500; }
.item-row { display: grid; grid-template-columns: 1fr 90px 90px 30px; gap: 6px; align-items: center; }
.item-row select, .item-row input { height: 28px; padding: 0 8px; font-size: 13px; border: 1px solid #d0d7de; border-radius: 3px; background: #fff; }
.doc-info { display: grid; grid-template-columns: 1fr 1fr; gap: 8px 24px; font-size: 13px; margin-bottom: 12px; }
.doc-info .lbl { color: #57606a; }
.mono { font-family: monospace; font-size: 12px; }
</style>