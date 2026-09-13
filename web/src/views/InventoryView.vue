<template>
  <div>
    <!-- Заголовок -->
    <div class="page-title-bar">
      <div class="page-title">
        <span>Инвентаризация</span>
        <span v-if="rows.length" class="refresh" @click="load" title="Обновить">↻</span>
      </div>
      <div class="page-actions">
        <select v-model="warehouseId" class="role-filter" @change="onWarehouseChange">
          <option value="">— выберите склад —</option>
          <option v-for="w in warehouses" :key="w.id" :value="w.id">{{ w.name }}</option>
        </select>
        <button class="btn" @click="load" :disabled="!warehouseId">Загрузить остатки</button>
        <button
          class="btn primary"
          :disabled="!canSave"
          @click="save"
        >
          {{ saving ? 'Сохранение...' : 'Создать документ' }}
        </button>
      </div>
    </div>

    <div v-if="error" class="error-box">{{ error }}</div>

    <!-- Пояснение -->
    <div v-if="!warehouseId" class="empty-state">
      <div style="font-size:48px">📋</div>
      <h2>Выберите склад для инвентаризации</h2>
      <p class="muted">Будут загружены все товары, числящиеся на складе. Введите фактическое количество — система рассчитает расхождения и создаст документ.</p>
    </div>

    <div v-else-if="loading" class="empty-state">
      <p class="muted">Загрузка остатков...</p>
    </div>

    <div v-else-if="rows.length === 0" class="empty-state">
      <p class="muted">На этом складе нет товаров с ненулевым остатком.</p>
    </div>

    <!-- Таблица для ввода фактов -->
    <table v-else class="ms-table">
      <thead>
        <tr>
          <th>Наименование</th>
          <th>Артикул</th>
          <th class="num">Учётный остаток</th>
          <th class="num" style="width:140px">Факт</th>
          <th>Ед.</th>
          <th class="num">Расхождение</th>
          <th class="num">Себестоимость</th>
          <th class="num">Сумма расхождения</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="r in rows" :key="r.product_id">
          <td>{{ r.product_name }}</td>
          <td class="muted mono">{{ r.sku || '—' }}</td>
          <td class="num">{{ formatQty(r.book_quantity) }}</td>
          <td class="num">
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
          <td class="num" :class="diffClass(r)">
            {{ diffLabel(r) }}
          </td>
          <td class="num">{{ formatMoney(r.cost_price) }}</td>
          <td class="num" :class="diffClass(r)">
            {{ formatMoney(diffAmount(r)) }}
          </td>
        </tr>
      </tbody>
    </table>

    <!-- Итоги -->
    <div v-if="rows.length" class="ms-footer">
      <div class="ms-pager">
        <span>Позиций: {{ rows.length }}</span>
        <span style="margin-left:16px">Изменено: {{ changedCount }}</span>
      </div>
      <div class="ms-totals">
        <span :class="totalDiffAmount > 0 ? 'danger-text' : ''">
          Расхождение: {{ formatMoney(totalDiffAmount) }}
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
.empty-state {
  background: #fff;
  border: 1px solid #d8dee4;
  border-radius: 6px;
  padding: 48px 24px;
  text-align: center;
}
.empty-state h2 { margin: 12px 0 8px; font-size: 18px; }
.empty-state p { max-width: 480px; margin: 0 auto; }

.fact-input {
  width: 100%;
  padding: 5px 8px;
  font-size: 13px;
  border: 1px solid #d0d7de;
  border-radius: 3px;
  text-align: right;
  font-variant-numeric: tabular-nums;
}
.fact-input:focus {
  outline: 2px solid rgba(9, 105, 218, 0.3);
  border-color: #2c5d9c;
}
.fact-input.changed {
  background: #fff8e1;
  border-color: #f0c040;
  font-weight: 600;
}

.diff-plus  { color: #1a7f37; font-weight: 600; }
.diff-minus { color: #cf222e; font-weight: 600; }
</style>