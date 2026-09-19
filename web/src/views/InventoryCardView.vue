<template>
  <div class="card-page">
    <div class="toolbar">
      <button class="btn primary" disabled>Сохранить</button>
      <button class="btn" @click="close">Закрыть</button>
      <button class="btn" @click="print">Печать</button>
      <button class="btn" @click="prev">‹</button>
      <span class="muted" style="font-size:12px">1 из 749</span>
      <button class="btn" @click="next">›</button>
      <button class="btn">Изменить</button>
      <button class="btn">Создать документ</button>
      <div class="toolbar-info">
        <span>{{ userLabel }}</span>
      </div>
    </div>

    <div v-if="error" class="error-box">{{ error }}</div>

    <div class="doc-header">
      <div class="doc-title">
        <span class="muted">Инвентаризация</span>
        <span class="doc-number">№</span>
        <input v-model="doc.number" class="num-input" disabled />
        <span class="muted">от</span>
        <input :value="formatDateTime(doc.doc_date)" class="date-input" disabled />
        <span class="status-label">Статус:</span>
        <select class="status-select" disabled>
          <option>Проведён</option>
        </select>
      </div>
      <div class="toolbar-info">
        <span>Автор: {{ userLabel }}</span>
        <span style="margin-left:12px">Изменён: {{ formatDateTime(doc.updated_at) }}</span>
      </div>
    </div>

    <div class="fields-grid">
      <div class="field">
        <label>Организация</label>
        <input :value="organizationName(doc.organization_id)" disabled />
      </div>
      <div class="field">
        <label>Склад</label>
        <input :value="doc.warehouse_name" disabled />
      </div>
    </div>

    <div class="doc-tabs">
      <button class="doc-tab" :class="{ active: tab === 'main' }" @click="tab = 'main'">Главная</button>
      <button class="doc-tab" :class="{ active: tab === 'related' }" @click="tab = 'related'">Связанные документы</button>
    </div>

    <template v-if="tab === 'main'">
      <div class="add-row">
        <input disabled placeholder="Добавьте позицию — введите наименование, код, штрихкод или артикул" class="add-input" />
        <button class="btn" disabled>Добавить из справочника</button>
        <button class="btn" disabled>Дополнить из остатков</button>
        <button class="btn" disabled>Дополнить из номенклатуры</button>
        <button class="btn" disabled>Импорт</button>
      </div>

      <table class="items-table">
        <thead>
          <tr>
            <th style="width:36px">№</th>
            <th>Наименование</th>
            <th class="num" style="width:120px">Расчётный остаток</th>
            <th class="num" style="width:120px">Фактический остаток</th>
            <th class="num" style="width:90px">Разница</th>
            <th class="num" style="width:110px">Цена</th>
            <th class="num" style="width:130px">Избыток/недостача</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="loading"><td colspan="7" class="muted" style="text-align:center;padding:24px">Загрузка…</td></tr>
          <tr v-else-if="items.length === 0"><td colspan="7" class="muted" style="text-align:center;padding:24px">Нет позиций</td></tr>
          <tr v-else v-for="(it, idx) in items" :key="it.id">
            <td class="muted num">{{ idx + 1 }}</td>
            <td>{{ it.product_name || it.product_id }}</td>
            <td class="num">{{ formatQty(it.calculated_quantity) }}</td>
            <td class="num">{{ formatQty(it.quantity) }}</td>
            <td class="num" :class="{ neg: it.correction_amount < 0 }">{{ formatQty(it.correction_amount) }}</td>
            <td class="num">{{ formatMoney(it.price) }}</td>
            <td class="num" :class="{ neg: it.correction_sum < 0 }">{{ formatMoney(it.correction_sum) }}</td>
          </tr>
        </tbody>
      </table>

      <div class="bottom-row">
        <div class="comment-box">
          <label class="lbl">Комментарий</label>
          <textarea :value="doc.comment" rows="6" disabled></textarea>
        </div>
        <div class="totals-box">
          <div class="total-row big">
            <span>Итого:</span>
            <span class="total-num">{{ formatMoney(doc.total) }}</span>
          </div>
          <div class="total-row small">
            <span>Кол-во:</span>
            <span class="total-num">{{ items.length }}</span>
          </div>
        </div>
      </div>
    </template>

    <template v-else>
      <div class="card muted" style="padding:24px;text-align:center">
        Связанных документов пока нет
      </div>
    </template>

    <div class="section-block">
      <div class="section-head"><span>Задачи</span><button class="btn-link" disabled>+ Задача</button></div>
      <div class="muted" style="font-size:12px">Нет задач</div>
    </div>

    <div class="section-block">
      <div class="section-head"><span>Файлы</span><button class="btn-link" disabled>+ Файл</button></div>
      <div class="muted" style="font-size:12px">Нет файлов</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getInventory, type Inventory, type InventoryItem } from '../api/inventories'
import { listOrganizations, type Organization } from '../api/suppliers'
import { apiErrorMessage } from '../api/client'
import { useAuthStore } from '../stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const doc = ref<Inventory>({
  id: '', number: '', doc_date: '', total: 0, items_count: 0, created_at: '', updated_at: '',
} as Inventory)
const items = ref<InventoryItem[]>([])
const organizations = ref<Organization[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const tab = ref<'main' | 'related'>('main')

const userLabel = computed(() => auth.userId ? auth.userId.slice(0, 8) : '')

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
    organizations.value = orgs
  } catch (e) {
    error.value = apiErrorMessage(e)
  } finally {
    loading.value = false
  }
}

function close() { router.push('/inventories') }
function print() { window.print() }
function prev() { router.push('/inventories') }
function next() { router.push('/inventories') }

onMounted(load)
</script>

<style scoped>
.card-page { padding: 0; }
.toolbar {
  display: flex; gap: 8px; align-items: center;
  padding: 8px 0 12px;
  border-bottom: 1px solid #e1e4e8;
  margin-bottom: 12px;
}
.toolbar-info { margin-left: auto; font-size: 12px; color: #8c959f; display: flex; gap: 4px; }
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
.date-input {
  width: 180px; padding: 4px 8px;
  border: 1px solid transparent; border-radius: 4px;
  font-size: 13px; background: transparent;
}
.status-label { margin-left: 12px; color: #57606a; font-size: 13px; }
.status-select {
  padding: 4px 8px; font-size: 13px;
  border: 1px solid #d0d7de; border-radius: 4px;
  background: #fff;
}
.fields-grid {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr 1fr;
  gap: 12px 20px;
  margin-bottom: 16px;
}
.field { display: flex; flex-direction: column; gap: 4px; font-size: 12px; color: #57606a; }
.field label { font-weight: 500; }
.field input {
  padding: 5px 8px; font-size: 13px;
  border: 1px solid #d0d7de; border-radius: 3px;
  background: #f6f8fa;
}
.doc-tabs { display: flex; gap: 4px; border-bottom: 1px solid #e1e4e8; margin-bottom: 12px; }
.doc-tab {
  padding: 8px 16px; background: none; border: none;
  font-size: 13px; cursor: pointer; color: #57606a;
  border-bottom: 2px solid transparent;
}
.doc-tab.active { color: #1f2328; font-weight: 600; border-bottom-color: #2c5d9c; }
.add-row { display: flex; gap: 8px; margin-bottom: 12px; }
.add-input { flex: 1; padding: 6px 10px; border: 1px solid #d0d7de; border-radius: 4px; background: #f6f8fa; }
.items-table { width: 100%; border-collapse: collapse; font-size: 13px; }
.items-table th {
  text-align: left; padding: 8px; border-bottom: 1px solid #d0d7de;
  font-weight: 500; color: #57606a; font-size: 12px;
}
.items-table td { padding: 8px; border-bottom: 1px solid #eaeef2; }
.items-table th.num, .items-table td.num { text-align: right; }
.neg { color: #c00; }
.bottom-row { display: grid; grid-template-columns: 1fr 320px; gap: 20px; margin-top: 16px; }
.comment-box { display: flex; flex-direction: column; gap: 4px; }
.lbl { font-size: 12px; color: #57606a; }
.comment-box textarea {
  padding: 8px; border: 1px solid #d0d7de; border-radius: 4px;
  font-family: inherit; font-size: 13px; resize: vertical;
  background: #f6f8fa;
}
.totals-box { display: flex; flex-direction: column; gap: 6px; }
.total-row { display: flex; justify-content: space-between; font-size: 13px; }
.total-row.big { font-size: 16px; font-weight: 600; padding-top: 6px; border-top: 1px solid #d0d7de; }
.total-row.small { font-size: 12px; color: #57606a; }
.total-num { font-variant-numeric: tabular-nums; }
.section-block { margin-top: 24px; border-top: 1px solid #e1e4e8; padding-top: 12px; }
.section-head { display: flex; justify-content: space-between; align-items: center; margin-bottom: 6px; font-weight: 600; font-size: 13px; }
.btn-link { background: none; border: none; color: #2c5d9c; cursor: pointer; font-size: 13px; }
.btn-link:disabled { color: #8c959f; cursor: default; }
</style>
