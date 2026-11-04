<template>
  <div class="page">
    <div class="ms-title">
      <span class="title-text">Документы</span>
      <button class="ms-help" title="Справка"><MsIcon name="help" :size="14" /></button>
      <button class="ms-refresh" @click="load" title="Обновить"><MsIcon name="refresh" :size="14" /></button>
    </div>

    <!-- Верхняя панель (Создать, Фильтр, Поиск, Печать) -->
    <div class="ms-toolbar">
      <div class="create-dropdown">
        <button class="btn-primary" @click="showCreateMenu = !showCreateMenu">
          <MsIcon name="plus" :size="14" /> Документ
        </button>
        
        <!-- Меню создания как в МС (в 2 колонки) -->
        <div v-if="showCreateMenu" class="dropdown-menu-ms" @mouseleave="showCreateMenu = false">
          <div class="dd-col">
            <div class="dd-group">Продажи</div>
            <button @click="openCreate('shipment')">Заказы покупателей</button>
            <button @click="openCreate('invoice_out')">Счета покупателям</button>
            <button @click="openCreate('shipment')">Отгрузка</button>
            
            <div class="dd-group">Закупки</div>
            <button @click="openCreate('purchase_order')">Заказы поставщикам</button>
            <button @click="openCreate('invoice_in')">Счета поставщиков</button>
            <button @click="openCreate('receipt')">Приёмка</button>
            
            <div class="dd-group">Деньги</div>
            <button @click="openCreate('payment_in')">Входящие платежи</button>
            <button @click="openCreate('payment_out')">Исходящие платежи</button>
          </div>
          <div class="dd-col">
            <div class="dd-group">Склад</div>
            <button @click="openCreate('internal_order')">Внутренние заказы</button>
            <button @click="openCreate('transfer')">Перемещения</button>
            <button @click="openCreate('inventory')">Инвентаризация</button>
            <button @click="openCreate('enter')">Оприходование</button>
            <button @click="openCreate('writeoff')">Списание</button>
            
            <div class="dd-group">Производство</div>
            <button @click="openCreate('tech_op')">Технологическая операция</button>
            <button @click="openCreate('prod_order')">Заказ на производство</button>
          </div>
        </div>
      </div>

      <button class="btn-default" :class="{ active: showFilter }" @click="showFilter = !showFilter">Фильтр</button>
      <input v-model="filters.search" class="ms-input search" placeholder="Номер или комментарий" />
      <div class="ms-counter">{{ filtered.length }}</div>
      <button class="btn-default" @click="print">Печать</button>
    </div>

    <!-- Блок фильтров (серый, как в МС) -->
    <div v-if="showFilter" class="ms-filter-panel">
      <div class="filter-actions">
        <button class="btn-green" @click="page = 1">Найти</button>
        <button class="btn-default" @click="clearFilters">Очистить</button>
      </div>
      
      <div class="filter-grid">
        <div class="filter-field">
          <label><span class="dot"></span>Период: <span class="link">вч·сг·нед·мес</span></label>
          <div class="date-range">
            <input v-model="filters.dateFrom" type="date" class="ms-input" />
            <span class="dash">—</span>
            <input v-model="filters.dateTo" type="date" class="ms-input" />
          </div>
        </div>
        <div class="filter-field">
          <label><span class="dot"></span>Товар или группа</label>
          <select class="ms-select"><option>Все</option></select>
        </div>
        <div class="filter-field">
          <label><span class="dot"></span>Склад</label>
          <select v-model="filters.warehouseId" class="ms-select">
            <option value="">Все</option>
            <option v-for="w in warehouses" :key="w.id" :value="w.id">{{ w.name }}</option>
          </select>
        </div>
        <div class="filter-field">
          <label><span class="dot"></span>Проект</label>
          <select class="ms-select"><option>Все</option></select>
        </div>
        <div class="filter-field">
          <label><span class="dot"></span>Канал продаж</label>
          <select class="ms-select"><option>Все</option></select>
        </div>
        <div class="filter-field">
          <label><span class="dot"></span>Контрагент или организация</label>
          <select class="ms-select"><option>Все</option></select>
        </div>
        
        <div class="filter-field">
          <label><span class="dot"></span>Группа контрагента</label>
          <select class="ms-select"><option>Все</option></select>
        </div>
        <div class="filter-field">
          <label><span class="dot"></span>Договор</label>
          <select class="ms-select"><option>Все</option></select>
        </div>
        <div class="filter-field">
          <label><span class="dot"></span>Тип документа</label>
          <select v-model="filters.type" class="ms-select">
            <option value="">Все</option>
            <option value="receipt">Приёмка</option>
            <option value="shipment">Отгрузка</option>
            <option value="transfer">Перемещение</option>
            <option value="internal_order">Внутренний заказ</option>
          </select>
        </div>
        <div class="filter-field">
          <label><span class="dot"></span>Статус</label>
          <select v-model="filters.status" class="ms-select">
            <option value="">Все</option>
            <option value="draft">Черновик</option>
            <option value="posted">Проведён</option>
          </select>
        </div>
        <div class="filter-field">
          <label><span class="dot"></span>Проведено</label>
          <select class="ms-select"><option>Все</option></select>
        </div>
        <div class="filter-field">
          <label><span class="dot"></span>Кто изменил</label>
          <select class="ms-select"><option>Все</option></select>
        </div>
      </div>
    </div>

    <!-- Таблица -->
    <div class="table-wrap">
      <table class="ms-table">
        <thead>
          <tr>
            <th class="col-check"><input type="checkbox" /></th>
            <th>Тип документа</th>
            <th>№</th>
            <th>Время</th>
            <th class="col-num">Сумма</th>
            <th>Со склада</th>
            <th>На склад</th>
            <th>Организация</th>
            <th>Контрагент</th>
            <th>Статус</th>
            <th>Отправлено</th>
            <th>Напечатано</th>
            <th>Комментарий</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="loading"><td colspan="14" class="empty">Загрузка...</td></tr>
          <tr v-else-if="filtered.length === 0"><td colspan="14" class="empty">Нет документов</td></tr>
          <tr v-else v-for="d in paginated" :key="d.id" class="row" @click="openCard(d)">
            <td class="col-check" @click.stop><input type="checkbox" /></td>
            <td>{{ typeLabel(d.type) }}</td>
            <td><span class="link">{{ d.number }}</span></td>
            <td>{{ formatDate(d.created_at) }}</td>
            <td class="col-num"><b>{{ formatMoney(d.total ?? 0) }}</b></td>
            <td>{{ warehouseName(d.warehouse_id) }}</td>
            <td>{{ warehouseName(d.target_warehouse_id) }}</td>
            <td>{{ orgName(d.organization_id) }}</td>
            <td>{{ supplierName(d.supplier_id) }}</td>
            <td>
              <span class="badge" :class="'bg-' + d.status">{{ statusLabel(d.status) }}</span>
            </td>
            <td><span v-if="d.sent_at" class="pill pill-blue">Отправлен</span></td>
            <td><span v-if="d.printed_at" class="pill pill-blue">Напечатан</span></td>
            <td class="comment">{{ d.comment || '' }}</td>
            <td class="col-actions">
              <button class="row-menu" @click.stop="onSoftDelete(d)" title="В корзину"><MsIcon name="trash" :size="14" /></button>
              <button class="row-menu" @click.stop="openCard(d)"><MsIcon name="dots" :size="14" /></button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Пагинация -->
    <div class="ms-footer2">
      <div class="pager">
        <button :disabled="page === 1" @click="page = 1">«</button>
        <button :disabled="page === 1" @click="page--">‹</button>
        <span class="range">{{ rangeFrom }}-{{ rangeTo }} из {{ filtered.length }}</span>
        <button :disabled="rangeTo >= filtered.length" @click="page++">›</button>
        <button :disabled="rangeTo >= filtered.length" @click="page = Math.ceil(filtered.length / perPage)">»</button>
      </div>
    </div>

    <DocumentCreateModal 
      v-model:visible="showCreateModal" 
      :default-type="createType"
      :default-warehouse-id="filters.warehouseId"
      @created="onCreated" 
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { listDocuments, softDeleteDocument, type Document } from '../api/documents'
import { listInternalOrders, softDeleteInternalOrder, type InternalOrder } from '../api/internalOrders'
import { listWarehouses, type Warehouse } from '../api/warehouses'
import { apiErrorMessage } from '../api/client'
import MsIcon from '../components/MsIcon.vue'
import DocumentCreateModal from '../components/DocumentCreateModal.vue'

const router = useRouter()

const items = ref<Document[]>([])
const internalOrders = ref<InternalOrder[]>([])
const warehouses = ref<Warehouse[]>([])
const loading = ref(false)
const error = ref<string | null>(null)

const showFilter = ref(false)
const page = ref(1)
const perPage = 100

const showCreateMenu = ref(false)
const showCreateModal = ref(false)
const createType = ref('receipt')

const filters = reactive({
  dateFrom: '',
  dateTo: '',
  warehouseId: '',
  type: '',
  status: '',
  search: '',
})

function openCreate(type: string) {
  createType.value = type
  showCreateModal.value = true
  showCreateMenu.value = false
}

async function onCreated() {
  await load()
}

const filtered = computed(() => {
  let rows = [...items.value]
  for (const io of internalOrders.value) {
    rows.push({
      id: io.id,
      type: 'internal_order' as any,
      number: io.number,
      status: io.status === 'draft' ? 'draft' : 'posted',
      warehouse_id: io.warehouse_id,
      target_warehouse_id: (io as any).target_warehouse_id,
      organization_id: io.organization_id,
      supplier_id: undefined,
      total: io.total,
      comment: io.comment,
      created_at: io.created_at,
      sent_at: io.sent_at,
      printed_at: io.printed_at,
    } as Document)
  }

  if (filters.dateFrom) {
    const t = new Date(filters.dateFrom).getTime()
    rows = rows.filter((d) => new Date(d.created_at).getTime() >= t)
  }
  if (filters.dateTo) {
    const t = new Date(filters.dateTo).getTime() + 86400000
    rows = rows.filter((d) => new Date(d.created_at).getTime() <= t)
  }
  if (filters.warehouseId) rows = rows.filter((d) => d.warehouse_id === filters.warehouseId)
  if (filters.type) rows = rows.filter((d) => d.type === filters.type)
  if (filters.status) rows = rows.filter((d) => d.status === filters.status)
  if (filters.search) {
    const q = filters.search.toLowerCase()
    rows = rows.filter((d) => (d.number ?? '').toLowerCase().includes(q) || (d.comment ?? '').toLowerCase().includes(q))
  }

  return rows.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
})

const paginated = computed(() => {
  const from = (page.value - 1) * perPage
  return filtered.value.slice(from, from + perPage)
})

const rangeFrom = computed(() => filtered.value.length === 0 ? 0 : (page.value - 1) * perPage + 1)
const rangeTo = computed(() => Math.min(page.value * perPage, filtered.value.length))

function clearFilters() {
  filters.dateFrom = ''
  filters.dateTo = ''
  filters.warehouseId = ''
  filters.type = ''
  filters.status = ''
  filters.search = ''
  page.value = 1
}

async function load() {
  loading.value = true
  error.value = null
  try {
    const [docs, ios, wh] = await Promise.all([listDocuments(), listInternalOrders(), listWarehouses()])
    items.value = docs
    internalOrders.value = ios
    warehouses.value = wh
  } catch (e) {
    error.value = apiErrorMessage(e)
  } finally {
    loading.value = false
  }
}

function openCard(d: Document) {
  if (d.type === 'internal_order') {
    router.push('/internal-orders/' + d.id)
  } else {
    router.push('/documents/' + d.id)
  }
}

async function onSoftDelete(d: Document) {
  if (!confirm(`Удалить документ ${d.number} в корзину?`)) return
  try {
    if (d.type === 'internal_order') {
      await softDeleteInternalOrder(d.id)
    } else {
      await softDeleteDocument(d.id)
    }
    await load()
  } catch (e: any) {
    alert('Не удалось удалить: ' + (e?.response?.data?.message || e.message || 'ошибка'))
  }
}

function typeLabel(t: string) {
  return { receipt: 'Приёмка', shipment: 'Отгрузка', transfer: 'Перемещение', writeoff: 'Списание', inventory: 'Инвентаризация', internal_order: 'Внутренний заказ' }[t] ?? t
}
function statusLabel(s: string) {
  return { draft: 'Черновик', posted: 'Проведён', cancelled: 'Отменён' }[s] ?? s
}
function warehouseName(id?: string) {
  return id ? warehouses.value.find((w) => w.id === id)?.name ?? id.slice(0, 8) : ''
}
function orgName(id?: string) { return id ? id.slice(0, 8) : '' }
function supplierName(id?: string) { return id ? id.slice(0, 8) : '' }
function formatDate(s: string) {
  const d = new Date(s)
  return d.toLocaleDateString('ru-RU') + ' ' + d.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })
}
function formatMoney(n: number) {
  return n.toLocaleString('ru-RU', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}
function print() { window.print() }

onMounted(load)
</script>

<style scoped>
.page { font-size: 13px; color: #1f2328; padding: 8px 0 32px; }

.ms-title { display: flex; align-items: center; gap: 8px; font-size: 20px; font-weight: 600; margin-bottom: 16px; }
.title-text { line-height: 1; }
.ms-help { width: 20px; height: 20px; border-radius: 50%; border: 1px solid #b8c0c8; background: transparent; color: #57606a; cursor: pointer; display: inline-flex; align-items: center; justify-content: center; padding: 0; }
.ms-help:hover { background: #f0f2f5; }
.ms-refresh { width: 24px; height: 24px; padding: 0; display: inline-flex; align-items: center; justify-content: center; border: none; background: transparent; color: #57606a; cursor: pointer; }
.ms-refresh:hover { color: #2c5d9c; }

.ms-toolbar { display: flex; align-items: center; gap: 8px; margin-bottom: 12px; flex-wrap: wrap; }
.btn-primary { display: inline-flex; align-items: center; gap: 6px; background: #2da44e; color: #fff; border: 1px solid #2a8e44; border-radius: 4px; padding: 6px 12px; font-size: 13px; cursor: pointer; }
.btn-primary:hover { background: #2c974b; }
.create-dropdown { position: relative; }
.btn-default { background: #f6f8fa; border: 1px solid #d0d7de; border-radius: 4px; padding: 6px 12px; font-size: 13px; cursor: pointer; color: #1f2328; }
.btn-default:hover { background: #eaeef2; }
.btn-default.active { background: #eaf3ff; border-color: #b6ccec; color: #0969da; }
.btn-green { background: #b3d261; border: 1px solid #a5c655; border-radius: 4px; padding: 6px 12px; font-size: 13px; cursor: pointer; color: #fff; font-weight: 500; }
.btn-green:hover { background: #a3c351; }

.ms-input, .ms-select {
  height: 30px; padding: 0 10px; border: 1px solid #d0d7de; border-radius: 4px;
  font-size: 13px; color: #1f2328; background: #fff; outline: none;
}
.ms-input.search { width: 220px; }
.ms-counter { background: #f6f8fa; border: 1px solid #d0d7de; border-radius: 4px; padding: 0 10px; height: 30px; display: inline-flex; align-items: center; justify-content: center; font-size: 13px; color: #8c959f; min-width: 40px; }

/* === СЕРЫЙ БЛОК ФИЛЬТРОВ === */
.ms-filter-panel { background: #eef1f5; border: 1px solid #d8dee4; border-radius: 4px; padding: 12px 16px; margin-bottom: 12px; }
.filter-actions { display: flex; align-items: center; gap: 8px; margin-bottom: 12px; }
.filter-grid { display: grid; grid-template-columns: repeat(6, 1fr); gap: 12px 16px; }
.filter-field { display: flex; flex-direction: column; gap: 4px; }
.filter-field label { font-size: 12px; color: #57606a; display: flex; align-items: center; gap: 4px; }
.filter-field .dot { width: 6px; height: 6px; border-radius: 50%; background: #8c959f; }
.filter-field .link { color: #0969da; cursor: pointer; margin-left: 2px; }
.filter-field .link:hover { text-decoration: underline; }
.filter-field .ms-select, .filter-field .ms-input { width: 100%; height: 28px; font-size: 12px; }
.date-range { display: flex; align-items: center; gap: 6px; }
.date-range .dash { color: #8c959f; }

/* === ВЫПАДАЮЩЕЕ МЕНЮ ДОКУМЕНТОВ === */
.dropdown-menu-ms {
  position: absolute; top: 100%; left: 0; z-index: 100;
  background: #fff; border: 1px solid #d0d7de; border-radius: 4px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.1); display: flex; gap: 20px;
  padding: 12px; margin-top: 4px; width: 500px;
}
.dd-col { display: flex; flex-direction: column; gap: 2px; flex: 1; }
.dd-group { font-size: 11px; color: #8c959f; text-transform: uppercase; font-weight: 600; margin: 8px 0 4px 0; padding-left: 8px; background: #f6f8fa; padding: 4px 8px; border-radius: 2px; }
.dd-group:first-child { margin-top: 0; }
.dd-col button { width: 100%; text-align: left; padding: 6px 8px; background: none; border: none; font-size: 13px; cursor: pointer; color: #1f2328; border-radius: 3px; }
.dd-col button:hover { background: #eaf3ff; color: #0969da; }

/* === ТАБЛИЦА === */
.table-wrap { background: #fff; border: 1px solid #eaeef2; border-radius: 4px; overflow-x: auto; }
.ms-table { width: 100%; border-collapse: collapse; font-size: 13px; }
.ms-table th { background: #fff; color: #2c5d9c; font-weight: 500; padding: 8px 10px; text-align: left; border-bottom: 1px solid #d8dee4; white-space: nowrap; font-size: 12px; }
.ms-table td { padding: 8px 10px; border-bottom: 1px solid #eaeef2; vertical-align: middle; }
.ms-table tr.row { cursor: pointer; }
.ms-table tr.row:hover { background: #f6f8fa; }
.ms-table .link { color: #2c5d9c; font-weight: 500; }
.ms-table .link:hover { text-decoration: underline; }
.ms-table .col-num { text-align: right; font-variant-numeric: tabular-nums; }
.ms-table .col-check { width: 32px; padding-left: 12px; }
.ms-table .comment { color: #8c959f; max-width: 200px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.ms-table .empty { text-align: center; padding: 24px; color: #8c959f; }

.badge { display: inline-block; padding: 2px 8px; border-radius: 3px; font-size: 11px; font-weight: 600; color: #fff; }
.bg-draft { background: #8c959f; }
.bg-posted { background: #2196f3; }
.bg-cancelled { background: #cf222e; }

.pill { display: inline-block; padding: 2px 8px; border-radius: 10px; font-size: 11px; font-weight: 600; }
.pill-blue { background: #dbeafe; color: #1e40af; }

.row-menu { width: 22px; height: 22px; padding: 0; border: none; background: transparent; color: #57606a; cursor: pointer; border-radius: 3px; }
.row-menu:hover { background: #eaeef2; color: #1f2328; }

.ms-footer2 { display: flex; align-items: center; padding: 8px 10px; font-size: 12px; color: #57606a; background: #fff; border-top: 1px solid #eaeef2; margin-top: 8px; }
.pager { display: flex; align-items: center; gap: 4px; }
.pager button { width: 24px; height: 24px; padding: 0; border: 1px solid #d0d7de; background: #fff; border-radius: 3px; cursor: pointer; font-size: 12px; color: #1f2328; display: inline-flex; align-items: center; justify-content: center; }
.pager button:disabled { opacity: 0.4; cursor: default; }
.pager button:hover:not(:disabled) { background: #f6f8fa; }
.pager .range { margin: 0 8px; font-variant-numeric: tabular-nums; }
</style>