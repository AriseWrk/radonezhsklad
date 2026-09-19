<template>
  <div>
    <div class="page-title-bar">
      <div class="page-title">
        <span>Инвентаризации</span>
        <span class="refresh" @click="load" title="Обновить">↻</span>
      </div>
      <div class="page-actions">
        <button class="btn" @click="showFilter = !showFilter">Фильтр</button>
        <router-link to="/inventory" class="btn">Создать инвентаризацию</router-link>
      </div>
    </div>

    <div v-if="showFilter" class="filter-panel">
      <div class="filter-row">
        <div class="filter-actions">
          <button class="btn-find" @click="load">Найти</button>
          <button class="btn-clear" @click="clearFilters">Очистить</button>
        </div>
        <div class="filter-field">
          <label>Склад</label>
          <select v-model="filters.warehouse_id" @change="load">
            <option value="">Все</option>
            <option v-for="w in warehouses" :key="w.id" :value="w.id">{{ w.name }}</option>
          </select>
        </div>
        <div class="filter-field">
          <label>Период с</label>
          <input v-model="filters.from" type="date" @change="load" />
        </div>
        <div class="filter-field">
          <label>Период по</label>
          <input v-model="filters.to" type="date" @change="load" />
        </div>
      </div>
    </div>

    <div v-if="error" class="error-box">{{ error }}</div>

    <table class="ms-table">
      <thead>
        <tr>
          <th style="width:60px">№</th>
          <th style="width:120px">Номер</th>
          <th style="width:170px">Дата</th>
          <th>Склад</th>
          <th class="num" style="width:90px">Позиций</th>
          <th class="num" style="width:130px">Сумма</th>
          <th>Комментарий</th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="loading"><td colspan="7" class="muted" style="text-align:center;padding:24px">Загрузка...</td></tr>
        <tr v-else-if="rows.length === 0"><td colspan="7" class="muted" style="text-align:center;padding:24px">Нет данных</td></tr>
        <tr v-else v-for="(r, idx) in rows" :key="r.id" class="clickable" @click="open(r)">
          <td class="muted">{{ idx + 1 }}</td>
          <td class="link">{{ r.number }}</td>
          <td class="muted">{{ formatDate(r.doc_date) }}</td>
          <td>{{ r.warehouse_name || '—' }}</td>
          <td class="num">{{ r.items_count }}</td>
          <td class="num">{{ formatNum(r.total) }}</td>
          <td class="muted">{{ r.comment || '' }}</td>
        </tr>
      </tbody>
    </table>

    <div v-if="rows.length > pageSize" class="pager">
      <button class="btn" :disabled="page <= 1" @click="page--">← Назад</button>
      <span class="muted">Стр. {{ page }} / {{ totalPages }}</span>
      <button class="btn" :disabled="page >= totalPages" @click="page++">Вперёд →</button>
    </div>

    <!-- Модалка деталей -->
    <div v-if="selected" class="modal-overlay" @click.self="selected = null">
      <div class="modal-card" style="max-width:900px">
        <div class="modal-head">
          <div>Инвентаризация {{ selected.number }}</div>
          <button class="btn-close" @click="selected = null">×</button>
        </div>
        <div class="modal-body">
          <p><b>Дата:</b> {{ formatDate(selected.doc_date) }}</p>
          <p><b>Склад:</b> {{ selected.warehouse_name || '—' }}</p>
          <p><b>Комментарий:</b> {{ selected.comment || '—' }}</p>
          <table class="ms-table" style="margin-top:12px">
            <thead>
              <tr>
                <th>Товар</th>
                <th style="width:80px">Артикул</th>
                <th class="num">Учёт</th>
                <th class="num">Факт</th>
                <th class="num">Расхождение</th>
                <th class="num">Цена</th>
                <th class="num">Сумма расхождения</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="it in selected.items || []" :key="it.id">
                <td>{{ it.product_name || it.product_id.slice(0, 8) + '…' }}</td>
                <td class="muted mono">{{ it.product_sku || '—' }}</td>
                <td class="num">{{ formatNum(it.calculated_quantity) }}</td>
                <td class="num">{{ formatNum(it.quantity) }}</td>
                <td class="num" :class="{ neg: it.correction_amount < 0 }">{{ formatNum(it.correction_amount) }}</td>
                <td class="num">{{ formatNum(it.price) }}</td>
                <td class="num">{{ formatNum(it.correction_sum) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { listInventories, type Inventory } from '../api/inventories'
import { listWarehouses, type Warehouse } from '../api/warehouses'
import { apiErrorMessage } from '../api/client'

const router = useRouter()
const rows = ref<Inventory[]>([])
const warehouses = ref<Warehouse[]>([])
const loading = ref(false)
const error = ref('')
const showFilter = ref(false)
const selected = ref<Inventory | null>(null)
const page = ref(1)
const pageSize = 50

const filters = reactive({
  warehouse_id: '',
  from: '',
  to: '',
})

const totalPages = computed(() => Math.max(1, Math.ceil(rows.value.length / pageSize)))

function formatDate(s: string) {
  if (!s) return '—'
  const d = new Date(s)
  return d.toLocaleDateString('ru-RU')
}
function formatNum(n: number | undefined) {
  if (n == null) return '—'
  return n.toLocaleString('ru-RU', { maximumFractionDigits: 3 })
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    rows.value = await listInventories({
      warehouse_id: filters.warehouse_id || undefined,
      from: filters.from || undefined,
      to: filters.to || undefined,
    })
    page.value = 1
  } catch (e) {
    error.value = apiErrorMessage(e)
  } finally {
    loading.value = false
  }
}

function clearFilters() {
  filters.warehouse_id = ''
  filters.from = ''
  filters.to = ''
  load()
}

function open(r: Inventory) {
  router.push('/inventories/' + r.id)
}

onMounted(async () => {
  try { warehouses.value = await listWarehouses() } catch {}
  load()
})
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(0,0,0,0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}
.modal-card {
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 10px 40px rgba(0,0,0,0.25);
  max-height: 90vh;
  width: 100%;
  max-width: 900px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}
.modal-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 18px;
  border-bottom: 1px solid #e5e7eb;
  font-weight: 600;
  font-size: 16px;
}
.modal-body {
  padding: 16px 18px;
  overflow-y: auto;
}
.modal-body p { margin: 6px 0; }
.btn-close {
  background: none;
  border: none;
  font-size: 24px;
  line-height: 1;
  cursor: pointer;
  color: #666;
  padding: 0 6px;
}
.btn-close:hover { color: #000; }
.neg { color: #c00; }
</style>
