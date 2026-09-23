<template>
  <div class="page">
    <div class="ms-title">
      <button class="ms-help" title="Справка"><MsIcon name="help" :size="14" /></button>
      <span>Инвентаризация</span>
      <button v-if="warehouseId" class="ms-refresh" @click="load" title="Обновить"><MsIcon name="refresh" :size="14" /></button>
    </div>

    <div class="ms-toolbar">
      <select class="ms-select" v-model="warehouseId" @change="onWarehouseChange">
        <option value="">— выберите склад —</option>
        <option v-for="w in warehouses" :key="w.id" :value="w.id">{{ w.name }}</option>
      </select>
      <MsButton icon="refresh" @click="load" :disabled="!warehouseId">Загрузить остатки</MsButton>
      <MsButton variant="primary" icon="save" :disabled="!canSave" @click="save">
        {{ saving ? 'Сохранение...' : 'Создать документ' }}
      </MsButton>
      <div class="toolbar-spacer"></div>
      <MsButton variant="icon" icon="gear" title="Настройки" />
    </div>

    <div v-if="error" class="error-box">{{ error }}</div>

    <div v-if="!warehouseId" class="empty-block">
      <h2>Выберите склад для инвентаризации</h2>
      <p class="muted">Будут загружены все товары, числящиеся на складе. Введите фактическое количество — система рассчитает расхождения и создаст документ.</p>
    </div>

    <div v-else-if="loading" class="empty-block">Загрузка остатков...</div>

    <div v-else-if="rows.length === 0" class="empty-block">
      На этом складе нет товаров с ненулевым остатком.
    </div>

    <table v-else class="ms-table2">
      <thead>
        <tr>
          <th>Наименование</th>
          <th>Артикул</th>
          <th class="col-num-right">Учётный остаток</th>
          <th class="col-num-right" style="width:140px">Факт</th>
          <th>Ед.</th>
          <th class="col-num-right">Расхождение</th>
          <th class="col-num-right">Себестоимость</th>
          <th class="col-num-right">Сумма расхождения</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="r in rows" :key="r.product_id" class="row">
          <td>{{ r.product_name }}</td>
          <td class="muted mono">{{ r.sku || '—' }}</td>
          <td class="col-num-right">{{ formatQty(r.book_quantity) }}</td>
          <td class="col-num-right">
            <input
              type="number"
              step="0.001"
              min="0"
              v-model.number="r.fact_quantity"
              class="fact-input"
              :class="{ changed: r.fact_quantity !== r.book_quantity }"
            />
          </td>
          <td>{{ r.unit_short }}</td>
          <td class="col-num-right" :class="diffClass(r)">
            {{ diffLabel(r) }}
          </td>
          <td class="col-num-right">{{ formatMoney(r.cost_price) }}</td>
          <td class="col-num-right" :class="diffClass(r)">
            {{ formatMoney(diffAmount(r)) }}
          </td>
        </tr>
      </tbody>
    </table>

    <div v-if="rows.length" class="ms-footer2">
      <div class="pager">
        <span class="range">Позиций: {{ rows.length }}</span>
        <span class="range">Изменено: {{ changedCount }}</span>
      </div>
      <div class="totals-inline">
        <span :class="totalDiffAmount > 0 ? 'danger-text' : ''">
          Расхождение: <b>{{ formatMoney(totalDiffAmount) }}</b>
        </span>
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { prepareInventory, type InventoryPrepareRow } from '../api/inventory'
import { listWarehouses, type Warehouse } from '../api/warehouses'
import { createDocument } from '../api/documents'
import { apiErrorMessage } from '../api/client'
import MsButton from '../components/MsButton.vue'
import MsIcon from '../components/MsIcon.vue'
import { useRouter } from 'vue-router'

interface Row extends InventoryPrepareRow {
  fact_quantity: number
}

const warehouses = ref<Warehouse[]>([])
const warehouseId = ref('')
const rows = ref<Row[]>([])
const loading = ref(false)
const saving = ref(false)
const error = ref<string | null>(null)

const router = useRouter()

function formatQty(n: number) { return n.toLocaleString('ru-RU', { maximumFractionDigits: 3 }) }
function formatMoney(n: number) {
  const sign = n < 0 ? '−' : ''
  return sign + Math.abs(n).toLocaleString('ru-RU', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function diffAmount(r: Row): number {
  return (r.fact_quantity - r.book_quantity) * r.cost_price
}
function diffClass(r: Row): string {
  const d = r.fact_quantity - r.book_quantity
  if (d > 0) return 'diff-plus'
  if (d < 0) return 'diff-minus'
  return 'muted'
}
function diffLabel(r: Row): string {
  const d = r.fact_quantity - r.book_quantity
  if (d === 0) return '0'
  const sign = d > 0 ? '+' : ''
  return sign + formatQty(d)
}

const changedCount = computed(() => rows.value.filter((r) => r.fact_quantity !== r.book_quantity).length)
const canSave = computed(() => rows.value.length > 0 && changedCount.value > 0)
const totalDiffAmount = computed(() => rows.value.reduce((s, r) => s + diffAmount(r), 0))

async function load() {
  if (!warehouseId.value) return
  loading.value = true
  error.value = null
  try {
    const resp = await prepareInventory(warehouseId.value)
    rows.value = resp.items.map((r) => ({ ...r, fact_quantity: r.book_quantity }))
  } catch (e) {
    error.value = apiErrorMessage(e)
    rows.value = []
  } finally {
    loading.value = false
  }
}

function onWarehouseChange() {
  rows.value = []
  if (warehouseId.value) load()
}

async function save() {
  if (!canSave.value) return
  saving.value = true
  error.value = null
  try {
    const items = rows.value
      .filter((r) => r.fact_quantity !== r.book_quantity)
      .map((r) => ({
        product_id: r.product_id,
        quantity: r.fact_quantity,
        price: r.cost_price,
      }))
    const doc = await createDocument({
      type: 'inventory',
      warehouse_id: warehouseId.value,
      comment: `Инвентаризация от ${new Date().toLocaleString('ru-RU')}`,
      items,
    })
    router.push({ name: 'documents', query: { type: 'inventory', open: doc.id } })
  } catch (e) {
    error.value = apiErrorMessage(e)
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  try { warehouses.value = await listWarehouses() } catch { /* ignore */ }
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
.toolbar-spacer { flex: 1; }
.ms-select { height: 30px; padding: 0 8px; font-size: 13px; border: 1px solid #d0d7de; border-radius: 4px; background: #fff; color: #1f2328; min-width: 220px; }

.ms-table2 { width: 100%; border-collapse: collapse; background: #fff; font-size: 13px; }
.ms-table2 thead th { background: #fff; color: #2c5d9c; font-weight: 500; padding: 8px 10px; text-align: left; border-bottom: 1px solid #d8dee4; white-space: nowrap; font-size: 12px; }
.ms-table2 tbody td { padding: 7px 10px; border-bottom: 1px solid #eaeef2; vertical-align: middle; }
.ms-table2 tbody tr.row:hover { background: #f6f8fa; }
.ms-table2 .col-num-right { text-align: right; font-variant-numeric: tabular-nums; }
.mono { font-family: monospace; font-size: 12px; }
.muted { color: #8c959f; }

.fact-input {
  width: 110px;
  padding: 4px 8px;
  font-size: 13px;
  border: 1px solid #d0d7de;
  border-radius: 3px;
  text-align: right;
  font-variant-numeric: tabular-nums;
}
.fact-input:focus { outline: 2px solid rgba(44,93,156,0.3); border-color: #2c5d9c; }
.fact-input.changed { background: #fff8e1; border-color: #f0c040; font-weight: 600; }

.diff-plus  { color: #1a7f37; font-weight: 600; }
.diff-minus { color: #cf222e; font-weight: 600; }
.danger-text { color: #cf222e; }

.ms-footer2 { display: flex; align-items: center; justify-content: space-between; padding: 6px 10px; font-size: 12px; color: #57606a; background: #fff; border-top: 1px solid #eaeef2; }
.pager { display: flex; gap: 16px; align-items: center; }
.pager .range { font-variant-numeric: tabular-nums; }
.totals-inline { font-variant-numeric: tabular-nums; }

.empty-block { padding: 48px 24px; text-align: center; background: #fff; border: 1px solid #eaeef2; border-radius: 6px; color: #57606a; font-size: 13px; }
.empty-block h2 { margin: 12px 0 8px; font-size: 18px; color: #1f2328; }
.empty-block p { max-width: 480px; margin: 0 auto; }
</style>