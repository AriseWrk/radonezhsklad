<template>
  <div class="card-page">
    <div class="ms-doc-head">
      <div class="ms-title">
        <button class="ms-help" title="Справка"><MsIcon name="help" :size="14" /></button>
        <span>{{ typeLabel(doc?.type) }}</span>
        <button class="ms-refresh" @click="load" title="Обновить"><MsIcon name="refresh" :size="14" /></button>
        <span v-if="isImported" class="ext-badge" title="Импортирован из МойСклад">МойСклад</span>
      </div>
      <div class="ms-status">
        <span v-if="doc" class="st" :class="'st-' + doc.status">{{ statusLabel(doc.status) }}</span>
      </div>
    </div>

    <div class="ms-toolbar">
      <MsButton variant="primary" icon="check" v-if="canEdit" :disabled="saving" @click="doPost">
        {{ saving ? 'Сохранение...' : 'Провести' }}
      </MsButton>
      <MsButton icon="print" @click="print">Печать</MsButton>
      <MsButton icon="close" v-if="isPosted && !isImported" :disabled="saving" @click="doCancel">Отменить</MsButton>
      <div class="toolbar-spacer"></div>
      <MsButton variant="icon" icon="close" @click="close" title="Закрыть" />
    </div>

    <div v-if="error" class="error-box">{{ error }}</div>
    <div v-if="loading" class="empty-block">Загрузка…</div>

    <template v-if="doc">
      <div class="ms-doc-fields">
        <div class="doc-left">
          <div class="row">
            <label>Номер</label>
            <input :value="doc.number" disabled />
          </div>
          <div class="row">
            <label>Дата</label>
            <input :value="formatDate(doc.created_at)" disabled />
          </div>
          <div class="row">
            <label>Склад</label>
            <span class="ro">{{ warehouseName(doc.warehouse_id) }}</span>
          </div>
          <div class="row" v-if="doc.target_warehouse_id">
            <label>Склад-получатель</label>
            <span class="ro">{{ warehouseName(doc.target_warehouse_id) }}</span>
          </div>
        </div>
        <div class="doc-right">
          <div class="row" v-if="doc.organization_id">
            <label>Организация</label>
            <span class="ro">{{ organizationName(doc.organization_id) }}</span>
          </div>
          <div class="row" v-if="doc.incoming_number">
            <label>Входящий №</label>
            <span class="ro">{{ doc.incoming_number }}</span>
          </div>
          <div class="row" v-if="doc.incoming_date">
            <label>Входящая дата</label>
            <span class="ro">{{ formatDate(doc.incoming_date) }}</span>
          </div>
          <div class="row" v-if="doc.paid_amount">
            <label>Оплачено</label>
            <span class="ro">{{ formatMoney(doc.paid_amount) }}</span>
          </div>
          <div class="row" v-if="doc.posted_at">
            <label>Проведён</label>
            <span class="ro">{{ formatDateTime(doc.posted_at) }}</span>
          </div>
          <div class="row" v-if="doc.cancelled_at">
            <label>Отменён</label>
            <span class="ro">{{ formatDateTime(doc.cancelled_at) }}</span>
          </div>
        </div>
      </div>

      <table class="ms-items">
        <thead>
          <tr>
            <th class="col-num">#</th>
            <th>Наименование</th>
            <th>Код</th>
            <th class="col-num-right">Кол-во</th>
            <th class="col-num-right">Цена</th>
            <th class="col-num-right">Сумма</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(it, idx) in (doc.items ?? [])" :key="it.id">
            <td class="col-num">{{ idx + 1 }}</td>
            <td>{{ it.product_name || '—' }}</td>
            <td class="muted mono">{{ it.product_sku || '—' }}</td>
            <td class="col-num-right">{{ formatQty(it.quantity) }}</td>
            <td class="col-num-right">{{ formatMoney(it.price) }}</td>
            <td class="col-num-right">{{ formatMoney(it.quantity * it.price) }}</td>
          </tr>
          <tr v-if="!doc.items || doc.items.length === 0">
            <td colspan="6" class="empty">Нет позиций</td>
          </tr>
        </tbody>
      </table>

      <div class="bottom-row">
        <div class="comment-box" v-if="doc.comment">
          <label class="lbl">Комментарий</label>
          <div class="comment-body">{{ doc.comment }}</div>
        </div>
        <div class="comment-box" v-else></div>
        <div class="totals-box">
          <div class="total-row big">
            <span>Итого:</span>
            <span class="total-num">{{ formatMoney(doc.total ?? 0) }}</span>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getDocument, postDocument, cancelDocument, type Document } from '../api/documents'
import { listWarehouses, type Warehouse } from '../api/warehouses'
import { listOrganizations, type Organization } from '../api/suppliers'
import { apiErrorMessage } from '../api/client'
import MsButton from '../components/MsButton.vue'
import MsIcon from '../components/MsIcon.vue'

const route = useRoute()
const router = useRouter()

const doc = ref<Document | null>(null)
const warehouses = ref<Warehouse[]>([])
const organizations = ref<Organization[]>([])
const loading = ref(false)
const saving = ref(false)
const error = ref<string | null>(null)

const isImported = computed(() => !!doc.value?.external_id)
const isPosted = computed(() => doc.value?.status === 'posted')
const canEdit = computed(() => doc.value?.status === 'draft' && !isImported.value)

function typeLabel(t?: string): string {
  switch (t) {
    case 'receipt':   return 'Оприходование'
    case 'shipment':  return 'Отгрузка'
    case 'transfer':  return 'Перемещение'
    case 'writeoff':  return 'Списание'
    case 'inventory': return 'Инвентаризация'
    default: return t ?? '—'
  }
}
function statusLabel(s?: string): string {
  switch (s) {
    case 'draft':     return 'Черновик'
    case 'posted':    return 'Проведён'
    case 'cancelled': return 'Отменён'
    default: return s ?? '—'
  }
}
function warehouseName(id?: string): string {
  if (!id) return '—'
  return warehouses.value.find((w) => w.id === id)?.name ?? '—'
}
function organizationName(id?: string): string {
  if (!id) return '—'
  return organizations.value.find((o) => o.id === id)?.name ?? '—'
}
function formatDate(s?: string): string {
  if (!s) return '—'
  return new Date(s).toLocaleDateString('ru-RU')
}
function formatDateTime(s?: string): string {
  if (!s) return '—'
  return new Date(s).toLocaleString('ru-RU')
}
function formatQty(n: number): string {
  return n.toLocaleString('ru-RU', { maximumFractionDigits: 3 })
}
function formatMoney(n: number): string {
  return n.toLocaleString('ru-RU', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

async function load() {
  const id = route.params.id as string
  if (!id) return
  loading.value = true
  error.value = null
  try {
    const [d, w, orgs] = await Promise.all([
      getDocument(id),
      listWarehouses(),
      listOrganizations(),
    ])
    doc.value = d
    warehouses.value = w
    organizations.value = orgs
  } catch (e) {
    error.value = apiErrorMessage(e)
  } finally {
    loading.value = false
  }
}

function close() { router.push('/documents') }
function print() { window.print() }

async function doPost() {
  if (!doc.value) return
  saving.value = true
  error.value = null
  try {
    const updated = await postDocument(doc.value.id)
    doc.value = updated
  } catch (e) {
    error.value = apiErrorMessage(e)
  } finally {
    saving.value = false
  }
}
async function doCancel() {
  if (!doc.value) return
  if (!window.confirm('Отменить документ?')) return
  saving.value = true
  error.value = null
  try {
    const updated = await cancelDocument(doc.value.id)
    doc.value = updated
  } catch (e) {
    error.value = apiErrorMessage(e)
  } finally {
    saving.value = false
  }
}

onMounted(load)
watch(() => route.params.id, load)
</script>

<style scoped>
.card-page { padding: 0; }
.toolbar {
  display: flex; gap: 8px; align-items: center;
  padding: 8px 0 12px; border-bottom: 1px solid #e1e4e8; margin-bottom: 12px;
}
.toolbar-info { margin-left: auto; font-size: 13px; color: #8c959f; }
.doc-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 14px; }
.doc-title { display: flex; align-items: center; gap: 8px; font-size: 16px; }
.doc-number { font-weight: 600; }
.num-input {
  width: 100px; padding: 4px 8px;
  border: 1px solid transparent; border-radius: 4px;
  font-size: 15px; font-weight: 600; background: transparent;
}
.date-input {
  width: 180px; padding: 4px 8px;
  border: 1px solid transparent; border-radius: 4px;
  font-size: 13px; background: transparent;
}
.status-label { margin-left: 12px; color: #57606a; font-size: 13px; }
.status-pill {
  padding: 2px 8px; font-size: 12px; border-radius: 10px; font-weight: 500;
}
.status-draft     { background: #eaeef2; color: #57606a; }
.status-posted    { background: #dafbe1; color: #1a7f37; }
.status-cancelled { background: #ffebe9; color: #cf222e; }
.ext-badge {
  padding: 2px 6px; font-size: 11px; font-weight: 500;
  background: #eaeef2; color: #57606a; border-radius: 3px;
  text-transform: uppercase; letter-spacing: 0.3px;
}
.fields-grid {
  display: grid; grid-template-columns: repeat(3, 1fr);
  gap: 12px 20px; margin-bottom: 16px;
}
.field { display: flex; flex-direction: column; gap: 4px; font-size: 12px; color: #57606a; }
.field label { font-weight: 500; }
.ro { font-size: 13px; color: #1f2328; padding: 5px 0; }
.items-section { margin: 16px 0; }
.items-table { width: 100%; border-collapse: collapse; font-size: 13px; }
.items-table th, .items-table td { padding: 6px 8px; border-bottom: 1px solid #eaeef2; text-align: left; }
.items-table th { font-weight: 500; color: #57606a; background: #f6f8fa; }
.items-table .num { text-align: right; }
.items-table tfoot td { border-top: 2px solid #d0d7de; border-bottom: none; padding-top: 8px; }
.comment-section { margin-top: 16px; }
.comment-section label { font-size: 12px; color: #57606a; font-weight: 500; display: block; margin-bottom: 4px; }
.comment-body {
  padding: 8px 10px; background: #f6f8fa; border-radius: 4px;
  font-size: 13px; white-space: pre-wrap;
}
.error-box {
  padding: 8px 12px; background: #ffebe9; color: #cf222e;
  border-radius: 4px; font-size: 13px; margin-bottom: 8px;
}
.muted { color: #8c959f; }

/* === МойСклад-стиль карточки === */
.card-page { font-size: 13px; }
.ms-doc-head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 10px; }
.ms-title { display: flex; align-items: center; gap: 8px; font-size: 20px; font-weight: 600; color: #1f2328; }
.ms-help { width: 20px; height: 20px; border-radius: 50%; border: 1px solid #b8c0c8; background: transparent; color: #57606a; cursor: pointer; display: inline-flex; align-items: center; justify-content: center; padding: 0; }
.ms-help:hover { background: #f0f2f5; }
.ms-refresh { width: 24px; height: 24px; padding: 0; display: inline-flex; align-items: center; justify-content: center; border: none; background: transparent; color: #57606a; cursor: pointer; }
.ms-refresh:hover { color: #2c5d9c; }
.ext-badge { background: #eaf3ff; color: #2c5d9c; border: 1px solid #c8dcf0; font-size: 11px; padding: 1px 8px; border-radius: 3px; font-weight: 500; }
.ms-status { display: flex; align-items: center; gap: 12px; }
.st { display: inline-block; padding: 3px 12px; border-radius: 3px; font-size: 12px; font-weight: 500; }
.st-draft     { background: #eef1f5; color: #57606a; }
.st-posted    { background: #d8e8c8; color: #2d6a1e; }
.st-cancelled { background: #f5d5d5; color: #a01c1c; }

.ms-toolbar { display: flex; align-items: center; gap: 6px; margin-bottom: 12px; flex-wrap: wrap; padding-bottom: 10px; border-bottom: 1px solid #eaeef2; }
.toolbar-spacer { flex: 1; }

.ms-doc-fields { display: grid; grid-template-columns: 1fr 1fr; gap: 6px 40px; margin-bottom: 14px; max-width: 1100px; }
.doc-left, .doc-right { display: flex; flex-direction: column; gap: 6px; }
.row { display: grid; grid-template-columns: 160px 1fr; align-items: center; gap: 10px; }
.row > label { font-size: 13px; color: #57606a; }
.row > input, .row > .ro { height: 28px; padding: 0 8px; font-size: 13px; line-height: 28px; border: 1px solid #d0d7de; border-radius: 3px; background: #fff; color: #1f2328; width: 100%; }
.row > input:disabled { background: #f6f8fa; color: #57606a; }
.row > .ro { border-color: transparent; background: transparent; }

.ms-items { width: 100%; border-collapse: collapse; background: #fff; font-size: 13px; margin-bottom: 14px; }
.ms-items thead th { background: #fff; color: #2c5d9c; font-weight: 500; padding: 8px 10px; text-align: left; border-bottom: 1px solid #d8dee4; white-space: nowrap; font-size: 12px; }
.ms-items tbody td { padding: 6px 10px; border-bottom: 1px solid #eaeef2; vertical-align: middle; }
.ms-items .col-num { width: 40px; color: #57606a; text-align: center; }
.ms-items .col-num-right { text-align: right; font-variant-numeric: tabular-nums; }
.ms-items .empty { text-align: center; padding: 24px; color: #8c959f; }
.mono { font-family: monospace; font-size: 12px; }
.muted { color: #8c959f; }

.bottom-row { display: grid; grid-template-columns: 1fr 380px; gap: 24px; align-items: start; margin-top: 12px; }
.comment-box { display: flex; flex-direction: column; gap: 4px; }
.comment-box .lbl { font-size: 12px; color: #57606a; }
.comment-body { background: #f6f8fa; border: 1px solid #eaeef2; border-radius: 3px; padding: 8px 10px; font-size: 13px; color: #1f2328; }
.totals-box { background: #f6f8fa; border: 1px solid #eaeef2; border-radius: 4px; padding: 12px 16px; }
.total-row { display: flex; justify-content: space-between; align-items: center; font-size: 15px; font-weight: 600; color: #1f2328; }
.total-num { font-variant-numeric: tabular-nums; }
.empty-block { padding: 40px; text-align: center; color: #8c959f; font-size: 13px; background: #fff; border: 1px solid #eaeef2; border-radius: 4px; }
</style>
