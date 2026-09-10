<template>
  <div>
    <div class="head">
      <h1>Товары</h1>
      <div class="actions">
        <label class="chk">
          <input type="checkbox" v-model="includeArchived" @change="load" />
          Показать архивные
        </label>
        <button class="primary" @click="showCreate = true">+ Новый товар</button>
      </div>
    </div>

    <div v-if="error" class="error-box">{{ error }}</div>

    <table>
      <thead>
        <tr>
          <th>Название</th>
          <th>SKU</th>
          <th>Цена</th>
          <th>Статус</th>
          <th style="width: 120px"></th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="loading"><td colspan="5" class="muted">Загрузка...</td></tr>
        <tr v-else-if="products.length === 0"><td colspan="5" class="muted">Нет товаров</td></tr>
        <tr v-for="p in products" :key="p.id">
          <td>{{ p.name }}</td>
          <td class="mono">{{ p.sku || '—' }}</td>
          <td>{{ p.price.toFixed(2) }} {{ p.currency }}</td>
          <td>
            <span v-if="p.is_archived" class="pill archived">архив</span>
            <span v-else class="pill active">активен</span>
          </td>
          <td>
            <button v-if="!p.is_archived" class="danger" @click="onArchive(p)">Архив</button>
          </td>
        </tr>
      </tbody>
    </table>

    <!-- модалка создания -->
    <div v-if="showCreate" class="modal-backdrop" @click.self="showCreate = false">
      <form class="card modal" @submit.prevent="onCreate">
        <h2>Новый товар</h2>
        <div v-if="createError" class="error-box">{{ createError }}</div>

        <label>Название*
          <input v-model="form.name" required />
        </label>

        <div class="row">
          <label>SKU
            <input v-model="form.sku" />
          </label>
          <label>Штрихкод
            <input v-model="form.barcode" />
          </label>
        </div>

        <div class="row">
          <label>Категория
            <select v-model="form.category_id">
              <option value="">— не выбрана —</option>
              <option v-for="c in categories" :key="c.id" :value="c.id">{{ c.name }}</option>
            </select>
          </label>
          <label>Единица
            <select v-model="form.unit_id">
              <option value="">— не выбрана —</option>
              <option v-for="u in units" :key="u.id" :value="u.id">
                {{ u.name }} ({{ u.short_name }})
              </option>
            </select>
          </label>
        </div>

        <div class="row">
          <label>Цена
            <input v-model.number="form.price" type="number" step="0.01" min="0" />
          </label>
          <label>Валюта
            <input v-model="form.currency" maxlength="3" />
          </label>
        </div>

        <label>Описание
          <textarea v-model="form.description" rows="3"></textarea>
        </label>

        <div class="modal-actions">
          <button type="button" @click="showCreate = false">Отмена</button>
          <button class="primary" type="submit" :disabled="saving">
            {{ saving ? 'Сохранение...' : 'Создать' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { listProducts, createProduct, archiveProduct, type Product } from '../api/products'
import { listCategories, type Category } from '../api/categories'
import { listUnits, type Unit } from '../api/units'
import { apiErrorMessage } from '../api/client'

const products = ref<Product[]>([])
const categories = ref<Category[]>([])
const units = ref<Unit[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const includeArchived = ref(false)

const showCreate = ref(false)
const saving = ref(false)
const createError = ref<string | null>(null)

const form = reactive({
  name: '',
  sku: '',
  barcode: '',
  category_id: '',
  unit_id: '',
  description: '',
  price: 0,
  currency: 'RUB',
})

async function load() {
  loading.value = true
  error.value = null
  try {
    products.value = await listProducts(includeArchived.value)
  } catch (e) {
    error.value = apiErrorMessage(e)
  } finally {
    loading.value = false
  }
}

async function loadRefs() {
  try {
    const [c, u] = await Promise.all([listCategories(), listUnits()])
    categories.value = c
    units.value = u
  } catch {
    // тихо
  }
}

async function onCreate() {
  createError.value = null
  saving.value = true
  try {
    await createProduct({
      name: form.name,
      sku: form.sku || undefined,
      barcode: form.barcode || undefined,
      category_id: form.category_id || undefined,
      unit_id: form.unit_id || undefined,
      description: form.description || undefined,
      price: form.price,
      currency: form.currency || 'RUB',
    })
    showCreate.value = false
    form.name = ''; form.sku = ''; form.barcode = ''
    form.category_id = ''; form.unit_id = ''
    form.description = ''; form.price = 0; form.currency = 'RUB'
    await load()
  } catch (e) {
    createError.value = apiErrorMessage(e)
  } finally {
    saving.value = false
  }
}

async function onArchive(p: Product) {
  if (!confirm(`Заархивировать «${p.name}»?`)) return
  try {
    await archiveProduct(p.id)
    await load()
  } catch (e) {
    error.value = apiErrorMessage(e)
  }
}

onMounted(async () => {
  await loadRefs()
  await load()
})
</script>

<style scoped>
.head { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
h1 { margin: 0; }
.actions { display: flex; gap: 12px; align-items: center; }
.chk { display: flex; gap: 6px; align-items: center; font-size: 13px; color: var(--muted); }
.chk input { width: auto; }
.mono { font-family: monospace; font-size: 13px; color: var(--muted); }
.pill { padding: 2px 8px; border-radius: 10px; font-size: 12px; font-weight: 600; }
.pill.active { background: #e8f5e9; color: var(--success); }
.pill.archived { background: #f3f4f6; color: var(--muted); }

.modal-backdrop {
  position: fixed; inset: 0;
  background: rgba(0,0,0,0.4);
  display: flex; align-items: center; justify-content: center;
  padding: 20px;
  z-index: 100;
}
.modal { width: 560px; max-width: 100%; display: flex; flex-direction: column; gap: 12px; }
.modal h2 { margin: 0 0 4px; font-size: 18px; }
label { display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: var(--muted); }
label input, label select, label textarea { color: var(--text); }
.row { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }
.modal-actions { display: flex; justify-content: flex-end; gap: 10px; margin-top: 6px; }
</style>