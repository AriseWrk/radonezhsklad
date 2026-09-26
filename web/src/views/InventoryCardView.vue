<template>
  <div class="card-page">
    <div class="ms-doc-head">
      <div class="ms-title">
        <button class="ms-help" title="Справка"><MsIcon name="help" :size="14" /></button>
        <span>Инвентаризация</span>
        <button class="ms-refresh" @click="load" title="Обновить"><MsIcon name="refresh" :size="14" /></button>
      </div>
      <div class="ms-status">
        <span class="st st-posted">Проведён</span>
      </div>
    </div>

    <div class="ms-toolbar">
      <MsButton variant="primary" icon="plus" @click="showCreateDialog = true">Создать документ</MsButton>
      <MsButton icon="print" @click="print">Печать</MsButton>
      <div class="pager-nav">
        <button class="pg" :disabled="!prevId" @click="prev">‹</button>
        <span class="pg-info">{{ position }} из {{ total }}</span>
        <button class="pg" :disabled="!nextId" @click="next">›</button>
      </div>
      <div class="toolbar-spacer"></div>
      <MsButton variant="icon" icon="close" @click="close" title="Закрыть" />
    </div>

    <div v-if="error" class="error-box">{{ error }}</div>

    <div class="print-header">
      <div class="ph-title">Инвентаризация: {{ doc.number }}</div>
      <div class="ph-row"><span class="ph-lbl">Дата проведения:</span> <span>{{ formatDateTime(doc.doc_date) }}</span></div>
      <div class="ph-row"><span class="ph-lbl">Склад:</span> <span>{{ doc.warehouse_name || '—' }}</span></div>
    </div>

    <div class="ms-doc-fields">
      <div class="doc-left">
        <div class="row">
          <label>Номер</label>
          <input :value="doc.number" disabled />
        </div>
        <div class="row">
          <label>Дата</label>
          <input :value="formatDateTime(doc.doc_date)" disabled />
        </div>
      </div>
      <div class="doc-right">
        <div class="row">
          <label>Организация</label>
          <span class="ro">{{ organizationName(doc.organization_id) }}</span>
        </div>
        <div class="row">
          <label>Склад</label>
          <span class="ro">{{ doc.warehouse_name || '—' }}</span>
        </div>
      </div>
    </div>

    <div class="ms-tabs">
      <button class="tab" :class="{ active: tab === 'main' }" @click="tab = 'main'">Главная</button>
      <button class="tab" :class="{ active: tab === 'related' }" @click="tab = 'related'">Связанные документы</button>
    </div>

    <template v-if="tab === 'main'">
      <table class="ms-items">
        <thead>
          <tr>
            <th class="col-num">№</th>
            <th class="col-code">Код</th>
            <th>Наименование</th>
            <th class="col-num-right">Расчётный остаток</th>
            <th class="col-num-right">Фактический остаток</th>
            <th class="col-num-right">Разница</th>
            <th class="col-unit">Ед. изм.</th>
            <th class="col-num-right">Цена</th>
            <th class="col-num-right">Избыток / недостача</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="loading"><td colspan="9" class="empty">Загрузка…</td></tr>
          <tr v-else-if="items.length === 0"><td colspan="9" class="empty">Нет позиций</td></tr>
          <tr v-else v-for="(it, idx) in items" :key="it.id">
            <td class="col-num">{{ idx + 1 }}</td>
            <td class="col-code">{{ it.product_sku || '' }}</td>
            <td>{{ it.product_name || it.product_id }}</td>
            <td class="col-num-right">{{ formatQty(it.calculated_quantity) }}</td>
            <td class="col-num-right">{{ formatQty(it.quantity) }}</td>
            <td class="col-num-right" :class="{ neg: it.correction_amount < 0 }">{{ formatQty(it.correction_amount) }}</td>
            <td class="col-unit">{{ it.product_unit_short || '' }}</td>
            <td class="col-num-right">{{ formatMoney(it.price) }}</td>
            <td class="col-num-right" :class="{ neg: it.correction_sum < 0 }">{{ formatMoney(it.correction_sum) }}</td>
          </tr>
        </tbody>
      </table>

      <div class="bottom-row">
        <div class="comment-box">
          <label class="lbl">Комментарий</label>
          <div class="comment-body">{{ doc.comment || '—' }}</div>
        </div>
        <div class="totals-box">
          <div class="total-row">
            <span>Позиций:</span>
            <span class="total-num">{{ items.length }}</span>
          </div>
          <div class="total-row big">
            <span>Итого:</span>
            <span class="total-num">{{ formatMoney(doc.total) }}</span>
          </div>
        </div>
      </div>
    </template>

    <template v-else>
      <div class="empty-block">Связанных документов пока нет</div>
    </template>

    <div v-if="showCreateDialog" class="modal-backdrop" @click.self="showCreateDialog = false">
      <div class="card modal">
        <h2>Создать корректирующий документ</h2>
        <p class="muted" style="font-size:13px;margin:8px 0 16px">
          Из инвентаризации №{{ doc.number }} от {{ formatDateTime(doc.doc_date) }}
        </p>
        <div class="dialog-actions">
          <MsButton :disabled="creating" @click="createCorrection('shortage')">Списать недостачи</MsButton>
          <MsButton :disabled="creating" @click="createCorrection('surplus')">Оприходовать избытки</MsButton>
        </div>
        <div v-if="createError" class="error-box" style="margin-top:12px">{{ createError }}</div>
        <div class="modal-actions">
          <button type="button" @click="showCreateDialog = false">Отмена</button>
        </div>
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getInventory, getInventoryNeighbors, createInventoryCorrection, type Inventory, type InventoryItem } from '../api/inventories'
import { listOrganizations, type Organization } from '../api/suppliers'
import { apiErrorMessage } from '../api/client'
import MsButton from '../components/MsButton.vue'
import MsIcon from '../components/MsIcon.vue'

const route = useRoute()
const router = useRouter()

const doc = ref<Inventory>({
  id: '', number: '', doc_date: '', total: 0, items_count: 0, created_at: '', updated_at: '',
} as Inventory)
const items = ref<InventoryItem[]>([])
const organizations = ref<Organization[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const tab = ref<'main' | 'related'>('main')
const showCreateDialog = ref(false)
const creating = ref(false)
const createError = ref<string | null>(null)


function formatDateTime(s: string) {
  if (!s) return ''
  const d = new Date(s)
  return d.toLocaleString('ru-RU', { day: '2-digit', month: '2-digit', year: 'numeric', hour: '2-digit', minute: '2-digit' })
}
function formatQty(n: number | undefined) {
  if (n == null) return '0'
  return n.toLocaleString('ru-RU', { maximumFractionDigits: 3 })
}
function formatMoney(n: number | undefined) {
  if (n == null) return '0,00'
  return n.toLocaleString('ru-RU', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}
function organizationName(id?: string) {
  if (!id) return ''
  return organizations.value.find((o) => o.id === id)?.name ?? ''
}

async function load() {
  loading.value = true; error.value = null
  try {
    const id = route.params.id as string
    const [d, orgs] = await Promise.all([
      getInventory(id),
      listOrganizations().catch(() => [] as Organization[]),
    ])
    doc.value = d
    items.value = d.items ?? []
    try {
      const nb = await getInventoryNeighbors(d.id)
      prevId.value = nb.prev_id
      nextId.value = nb.next_id
      position.value = nb.position
      total.value = nb.total
    } catch { /* ignore */ }
    organizations.value = orgs
  } catch (e) {
    error.value = apiErrorMessage(e)
  } finally {
    loading.value = false
  }
}

const prevId = ref<string | null>(null)
const nextId = ref<string | null>(null)
const position = ref(0)
const total = ref(0)

function close() { router.push('/inventories') }
function print() { window.print() }

async function createCorrection(kind: 'shortage' | 'surplus') {
  if (!doc.value.id) return
  creating.value = true
  createError.value = null
  try {
    const d = await createInventoryCorrection(doc.value.id, kind)
    showCreateDialog.value = false
    router.push('/documents/' + d.id)
  } catch (e) {
    createError.value = apiErrorMessage(e)
  } finally {
    creating.value = false
  }
}
function prev() { if (prevId.value) router.push('/inventories/' + prevId.value) }
function next() { if (nextId.value) router.push('/inventories/' + nextId.value) }

onMounted(load)
watch(() => route.params.id, load)
</script>

<style scoped>
/* === МойСклад-стиль === */
.card-page { font-size: 13px; }
.ms-doc-head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 10px; }
.ms-title { display: flex; align-items: center; gap: 8px; font-size: 20px; font-weight: 600; color: #1f2328; }
.ms-help { width: 20px; height: 20px; border-radius: 50%; border: 1px solid #b8c0c8; background: transparent; color: #57606a; cursor: pointer; display: inline-flex; align-items: center; justify-content: center; padding: 0; }
.ms-help:hover { background: #f0f2f5; }
.ms-refresh { width: 24px; height: 24px; padding: 0; display: inline-flex; align-items: center; justify-content: center; border: none; background: transparent; color: #57606a; cursor: pointer; }
.ms-refresh:hover { color: #2c5d9c; }
.ms-status { display: flex; align-items: center; gap: 12px; }
.st { display: inline-block; padding: 3px 12px; border-radius: 3px; font-size: 12px; font-weight: 500; }
.st-posted { background: #d8e8c8; color: #2d6a1e; }

.ms-toolbar { display: flex; align-items: center; gap: 6px; margin-bottom: 12px; flex-wrap: wrap; padding-bottom: 10px; border-bottom: 1px solid #eaeef2; }
.toolbar-spacer { flex: 1; }
.pager-nav { display: inline-flex; align-items: center; gap: 4px; margin-left: 8px; }
.pg { width: 24px; height: 24px; border: 1px solid #d0d7de; background: #fff; border-radius: 3px; cursor: pointer; font-size: 12px; color: #1f2328; }
.pg:disabled { opacity: 0.4; cursor: default; }
.pg-info { font-size: 12px; color: #57606a; margin: 0 6px; }

.ms-doc-fields { display: grid; grid-template-columns: 1fr 1fr; gap: 6px 40px; margin-bottom: 14px; max-width: 1100px; }
.doc-left, .doc-right { display: flex; flex-direction: column; gap: 6px; }
.row { display: grid; grid-template-columns: 160px 1fr; align-items: center; gap: 10px; }
.row > label { font-size: 13px; color: #57606a; }
.row > input, .row > .ro { height: 28px; padding: 0 8px; font-size: 13px; line-height: 28px; border: 1px solid #d0d7de; border-radius: 3px; background: #fff; color: #1f2328; width: 100%; }
.row > input:disabled { background: #f6f8fa; color: #57606a; }
.row > .ro { border-color: transparent; background: transparent; }

.ms-tabs { display: flex; gap: 2px; border-bottom: 1px solid #d8dee4; margin-bottom: 12px; }
.tab { border: none; background: transparent; padding: 10px 16px; font-size: 13px; color: #57606a; cursor: pointer; border-bottom: 2px solid transparent; }
.tab:hover { color: #2c5d9c; }
.tab.active { color: #2c5d9c; border-bottom-color: #2c5d9c; font-weight: 600; }

.ms-items { width: 100%; border-collapse: collapse; background: #fff; font-size: 13px; margin-bottom: 14px; }
.ms-items thead th { background: #fff; color: #2c5d9c; font-weight: 500; padding: 8px 10px; text-align: left; border-bottom: 1px solid #d8dee4; white-space: nowrap; font-size: 12px; }
.ms-items tbody td { padding: 6px 10px; border-bottom: 1px solid #eaeef2; vertical-align: middle; }
.ms-items .col-num { width: 40px; color: #57606a; text-align: center; }
.ms-items .col-num-right { text-align: right; font-variant-numeric: tabular-nums; }
.ms-items .empty { text-align: center; padding: 24px; color: #8c959f; }
.ms-items .col-code { width: 90px; color: #57606a; }
.ms-items .col-unit { width: 70px; color: #57606a; text-align: center; }
.neg { color: #c00; }

.bottom-row { display: grid; grid-template-columns: 1fr 380px; gap: 24px; align-items: start; margin-top: 12px; }
.comment-box { display: flex; flex-direction: column; gap: 4px; }
.comment-box .lbl { font-size: 12px; color: #57606a; }
.comment-body { background: #f6f8fa; border: 1px solid #eaeef2; border-radius: 3px; padding: 8px 10px; font-size: 13px; color: #1f2328; min-height: 60px; }
.totals-box { background: #f6f8fa; border: 1px solid #eaeef2; border-radius: 4px; padding: 12px 16px; display: flex; flex-direction: column; gap: 6px; }
.total-row { display: flex; justify-content: space-between; align-items: center; font-size: 13px; }
.total-row.big { font-size: 15px; font-weight: 600; border-top: 1px solid #d8dee4; padding-top: 8px; margin-top: 4px; }
.total-num { font-variant-numeric: tabular-nums; }
.empty-block { padding: 40px; text-align: center; color: #8c959f; font-size: 13px; background: #fff; border: 1px solid #eaeef2; border-radius: 4px; }

.modal-backdrop { position: fixed; inset: 0; background: rgba(0,0,0,0.4); display: flex; align-items: center; justify-content: center; padding: 20px; z-index: 100; }
.modal { background: #fff; border-radius: 4px; width: 480px; max-width: 100%; padding: 20px; display: flex; flex-direction: column; gap: 10px; }
.modal h2 { margin: 0 0 4px; font-size: 18px; }
.modal-actions { display: flex; justify-content: flex-end; gap: 10px; margin-top: 6px; }
.dialog-actions { display: flex; gap: 8px; flex-wrap: wrap; }
.muted { color: #8c959f; }
/* === Печатная форма инвентаризации (по образцу МойСклад) === */
.print-header { display: none; }

@media print {
  .ms-toolbar, .ms-tabs, .ms-doc-head, .ms-doc-fields,
  .comment-box, .pager-nav, .modal-backdrop, .empty-block,
  .ms-help, .ms-refresh { display: none !important; }

  .card-page { font-size: 12px; padding: 0; }
  .print-header { display: block !important; margin-bottom: 12px; }
  .ph-title { font-size: 16px; font-weight: 600; margin-bottom: 4px; }
  .ph-row { font-size: 13px; margin-bottom: 2px; }
  .ph-lbl { color: #57606a; display: inline-block; min-width: 140px; }

  .ms-items { font-size: 11px; margin-bottom: 0; }
  .ms-items thead th { border-bottom: 1px solid #000; padding: 4px 6px; color: #000; }
  .ms-items tbody td { border-bottom: 1px solid #ccc; padding: 3px 6px; }

  .bottom-row { grid-template-columns: 1fr; gap: 0; }
  .totals-box { background: transparent; border: none; padding: 8px 0 0; }
}
</style>
