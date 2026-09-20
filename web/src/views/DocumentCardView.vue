<template>
  <div class="card-page">
    <div class="toolbar">
      <button class="btn primary" v-if="canEdit" :disabled="saving" @click="doPost">
        {{ saving ? 'Сохранение...' : 'Провести' }}
      </button>
      <button class="btn" @click="close">Закрыть</button>
      <button class="btn" @click="print">Печать</button>
      <button class="btn" v-if="isPosted && !isImported" :disabled="saving" @click="doCancel">Отменить</button>
      <div class="toolbar-info">{{ doc ? typeLabel(doc.type) : '' }}</div>
    </div>

    <div v-if="error" class="error-box">{{ error }}</div>
    <div v-if="loading" class="muted" style="padding:16px">Загрузка…</div>

    <template v-if="doc">
      <div class="doc-header">
        <div class="doc-title">
          <span class="muted">{{ typeLabel(doc.type) }}</span>
          <span v-if="doc.external_id" class="ext-badge" title="Импортирован из МойСклад — только просмотр">МойСклад</span>
          <span class="doc-number">№</span>
          <input :value="doc.number" class="num-input" disabled />
          <span class="muted">от</span>
          <input :value="formatDate(doc.created_at)" class="date-input" disabled />
          <span class="status-label">Статус:</span>
          <span :class="['status-pill', 'status-' + doc.status]">{{ statusLabel(doc.status) }}</span>
        </div>
      </div>

      <div class="fields-grid">
        <div class="field">
          <label>Склад</label>
          <div class="ro">{{ warehouseName(doc.warehouse_id) }}</div>
        </div>
        <div class="field" v-if="doc.target_warehouse_id">
          <label>Склад-получатель</label>
          <div class="ro">{{ warehouseName(doc.target_warehouse_id) }}</div>
        </div>
        <div class="field" v-if="doc.organization_id">
          <label>Организация</label>
          <div class="ro">{{ organizationName(doc.organization_id) }}</div>
        </div>
        <div class="field" v-if="doc.incoming_number">
          <label>Входящий №</label>
          <div class="ro">{{ doc.incoming_number }}</div>
        </div>
        <div class="field" v-if="doc.incoming_date">
          <label>Входящая дата</label>
          <div class="ro">{{ formatDate(doc.incoming_date) }}</div>
        </div>
        <div class="field" v-if="doc.paid_amount">
          <label>Оплачено</label>
          <div class="ro">{{ formatMoney(doc.paid_amount) }}</div>
        </div>
        <div class="field" v-if="doc.posted_at">
          <label>Проведён</label>
          <div class="ro">{{ formatDateTime(doc.posted_at) }}</div>
        </div>
        <div class="field" v-if="doc.cancelled_at">
          <label>Отменён</label>
          <div class="ro">{{ formatDateTime(doc.cancelled_at) }}</div>
        </div>
      </div>

      <div class="items-section">
        <table class="items-table">
          <thead>
            <tr>
              <th style="width:40px">#</th>
              <th>Наименование</th>
              <th>Код</th>
              <th class="num">Кол-во</th>
              <th class="num">Цена</th>
              <th class="num">Сумма</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(it, idx) in (doc.items ?? [])" :key="it.id">
              <td>{{ idx + 1 }}</td>
              <td>{{ it.product_name || '—' }}</td>
              <td class="muted">{{ it.product_sku || '—' }}</td>
              <td class="num">{{ formatQty(it.quantity) }}</td>
              <td class="num">{{ formatMoney(it.price) }}</td>
              <td class="num">{{ formatMoney(it.quantity * it.price) }}</td>
            </tr>
            <tr v-if="!doc.items || doc.items.length === 0">
              <td colspan="6" class="muted" style="text-align:center;padding:16px">Нет позиций</td>
            </tr>
          </tbody>
          <tfoot>
            <tr>
              <td colspan="5" class="num"><strong>Итого:</strong></td>
              <td class="num"><strong>{{ formatMoney(doc.total ?? 0) }}</strong></td>
            </tr>
          </tfoot>
        </table>
      </div>

      <div class="comment-section" v-if="doc.comment">
        <label>Комментарий</label>
        <div class="comment-body">{{ doc.comment }}</div>
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
</style>