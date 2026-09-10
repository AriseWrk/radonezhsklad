<template>
  <div>
    <div class="head">
      <h1>Остатки на складах</h1>
      <button @click="load">Обновить</button>
    </div>

    <div class="filters card">
      <label>Склад
        <select v-model="filterWarehouse">
          <option value="">— все —</option>
          <option v-for="w in warehouses" :key="w.id" :value="w.id">{{ w.name }}</option>
        </select>
      </label>
      <label>Товар
        <select v-model="filterProduct">
          <option value="">— все —</option>
          <option v-for="p in products" :key="p.id" :value="p.id">{{ p.name }}</option>
        </select>
      </label>
    </div>

    <div v-if="error" class="error-box">{{ error }}</div>

    <table>
      <thead>
        <tr>
          <th>Склад</th>
          <th>Товар</th>
          <th>Количество</th>
          <th>Обновлено</th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="loading"><td colspan="4" class="muted">Загрузка...</td></tr>
        <tr v-else-if="items.length === 0"><td colspan="4" class="muted">Остатков нет</td></tr>
        <tr v-for="s in items" :key="s.id">
          <td>{{ warehouseName(s.warehouse_id) }}</td>
          <td>{{ productName(s.product_id) }}</td>
          <td><strong>{{ s.quantity }}</strong></td>
          <td class="muted">{{ formatDate(s.updated_at) }}</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { listStock, type StockBalance } from '../api/stock'
import { listWarehouses, type Warehouse } from '../api/warehouses'
import { listProducts, type Product } from '../api/products'
import { apiErrorMessage } from '../api/client'

const items = ref<StockBalance[]>([])
const warehouses = ref<Warehouse[]>([])
const products = ref<Product[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const filterWarehouse = ref('')
const filterProduct = ref('')

function warehouseName(id: string) {
  return warehouses.value.find((w) => w.id === id)?.name ?? id.slice(0, 8)
}
function productName(id: string) {
  return products.value.find((p) => p.id === id)?.name ?? id.slice(0, 8)
}
function formatDate(s: string) {
  return new Date(s).toLocaleString('ru-RU')
}

async function load() {
  loading.value = true
  error.value = null
  try {
    items.value = await listStock(filterWarehouse.value || undefined, filterProduct.value || undefined)
  } catch (e) {
    error.value = apiErrorMessage(e)
  } finally {
    loading.value = false
  }
}

watch([filterWarehouse, filterProduct], load)

onMounted(async () => {
  try {
    const [w, p] = await Promise.all([listWarehouses(), listProducts(true)])
    warehouses.value = w
    products.value = p
  } catch {
    /* ignore */
  }
  await load()
})
</script>

<style scoped>
.head { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
h1 { margin: 0; }
.muted { color: var(--muted); }
.filters { display: flex; gap: 16px; margin-bottom: 16px; }
.filters label { flex: 1; display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: var(--muted); }
.filters select { color: var(--text); }
</style>