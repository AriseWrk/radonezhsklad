<template>
  <div class="products-layout">
    <!-- Левая панель: категории -->
    <aside class="categories-panel">
      <div class="cat-section">Справочники</div>
      <div class="cat-item" :class="{ active: selectedCategory === '' }" @click="selectCategory('')">
        <span>Товары и услуги</span>
        <span class="cat-count">{{ totalCount }}</span>
      </div>
      <div class="cat-section">Категории</div>
      <div
        v-for="c in categories"
        :key="c.id"
        class="cat-item"
        :class="{ active: selectedCategory === c.id }"
        @click="selectCategory(c.id)"
      >
        <span class="cat-name">{{ c.name }}</span>
        <span class="cat-count">{{ countByCategory(c.id) }}</span>
      </div>
      <div v-if="categories.length === 0" class="cat-empty muted">Нет категорий</div>
    </aside>

    <!-- Основная область -->
    <div class="products-main">
      <!-- Toolbar -->
      <div class="page-title-bar">
        <div class="page-title">
          <span>Товары и услуги</span>
          <span class="refresh" @click="load" title="Обновить">↻</span>
        </div>
        <div class="page-actions">
          <button class="btn primary icon-btn" @click="openCreate">
            <span class="plus">+</span> Товар
          </button>
          <button class="btn icon-btn" disabled>
            <span class="plus">¥</span> Услуга
          </button>
          <button class="btn icon-btn" disabled>
            <span class="plus">⚙</span> Комплект
          </button>
          <button class="btn icon-btn" disabled>
            <span class="plus">🗂</span> Группа
          </button>
          <button class="btn" @click="showFilter = !showFilter">Фильтр</button>
          <input v-model="search" class="search-input" placeholder="Наименование, код или артикул" />
          <div class="counter" :class="{ active: selected.size > 0 }">{{ selected.size }}</div>
          <select class="mini-select" :disabled="selected.size === 0">
            <option>Изменить</option>
            <option>Удалить</option>
            <option>Архивировать</option>
          </select>
          <button class="btn icon-only" @click="printList" title="Печать">🖨</button>
          <button class="btn" @click="importCsv" disabled>Импорт</button>
          <button class="btn" @click="exportCsv">Экспорт</button>
          <button class="btn icon-only" title="Настройки">⚙</button>
        </div>
      </div>

      <!-- Фильтры -->
      <div v-if="showFilter" class="filter-panel">
        <div class="filter-row">
          <div class="filter-actions">
            <button class="btn-find" @click="page = 1">Найти</button>
            <button class="btn-clear" @click="clearFilters">Очистить</button>
          </div>
          <div class="filter-field">
            <label>Архивные</label>
            <select v-model="includeArchived" @change="load">
              <option :value="false">Скрыть</option>
              <option :value="true">Показать</option>
            </select>
          </div>
          <div class="filter-field">
            <label>Артикул</label>
            <input v-model="filterSku" placeholder="SKU" />
          </div>
          <div class="filter-field">
            <label>Цена от</label>
            <input v-model.number="filterPriceFrom" type="number" min="0" />
          </div>
          <div class="filter-field">
            <label>Цена до</label>
            <input v-model.number="filterPriceTo" type="number" min="0" />
          </div>
        </div>
      </div>

      <div v-if="error" class="error-box">{{ error }}</div>

      <!-- Таблица -->
      <table class="ms-table">
        <thead>
          <tr>
            <th class="chk-col"><input type="checkbox" :checked="allChecked" @change="toggleAll" /></th>
            <th @click="sortBy('name')">
              Наименование
              <span v-if="sortKey === 'name'" class="sort-arrow">{{ sortDir === 'asc' ? '↑' : '↓' }}</span>
            </th>
            <th style="width:100px" @click="sortBy('sku')">
              Код
              <span v-if="sortKey === 'sku'" class="sort-arrow">{{ sortDir === 'asc' ? '↑' : '↓' }}</span>
            </th>
            <th style="width:120px">Артикул</th>
            <th style="width:80px">Ед. изм.</th>
            <th class="num" style="width:130px" @click="sortBy('price')">
              Цена продажи
              <span v-if="sortKey === 'price'" class="sort-arrow">{{ sortDir === 'asc' ? '↑' : '↓' }}</span>
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="loading">
            <td colspan="6" class="muted" style="text-align:center;padding:24px">Загрузка...</td>
          </tr>
          <tr v-else-if="filtered.length === 0">
            <td colspan="6" class="muted" style="text-align:center;padding:24px">Нет товаров</td>
          </tr>
          <tr
            v-else
            v-for="p in paginated"
            :key="p.id"
            class="clickable"
            :class="{ selected: selected.has(p.id) }"
            @click="openEdit(p)"
          >
            <td class="chk-col" @click.stop>
              <input type="checkbox" :checked="selected.has(p.id)" @change="toggleSelect(p.id)" />
            </td>
            <td>
              {{ p.name }}
              <span v-if="p.is_archived" class="pill archived" style="margin-left:8px">архив</span>
            </td>
            <td class="muted mono">{{ shortId(p.id) }}</td>
            <td class="muted mono">{{ p.sku || '—' }}</td>
            <td>{{ unitShort(p.unit_id) }}</td>
            <td class="num">{{ formatMoney(p.price) }}</td>
          </tr>
        </tbody>
      </table>

      <!-- Футер -->
      <div class="ms-footer">
        <div class="ms-pager">
          <button :disabled="page === 1" @click="page--">◀</button>
          <button :disabled="page === 1" @click="page = 1">↤</button>
          <span>{{ rangeFrom }}–{{ rangeTo }} из {{ filtered.length }}</span>
          <button :disabled="rangeTo >= filtered.length" @click="page = Math.ceil(filtered.length / perPage)">↦</button>
          <button :disabled="rangeTo >= filtered.length" @click="page++">▶</button>
        </div>
        <div class="ms-totals">
          <span>Позиций: {{ filtered.length }}</span>
          <span>Сумма: {{ formatMoney(totalSum) }}</span>
        </div>
      </div>
    </div>

    <!-- Модалка создания/редактирования -->
    <div v-if="modal.open" class="modal-backdrop" @click.self="closeModal">
      <form class="card modal big" @submit.prevent="onSubmit">
        <h2>{{ modal.isEdit ? 'Карточка товара' : 'Новый товар' }}</h2>
        <div v-if="modal.error" class="error-box">{{ modal.error }}</div>

        <label>Наименование*
          <input v-model="form.name" required />
        </label>

        <div class="grid2">
          <label>Артикул
            <input v-model="form.sku" placeholder="SKU" />
          </label>
          <label>Штрихкод
            <input v-model="form.barcode" />
          </label>
        </div>

        <div class="grid2">
          <label>Категория
            <select v-model="form.category_id">
              <option value="">— не выбрана —</option>
              <option v-for="c in categories" :key="c.id" :value="c.id">{{ c.name }}</option>
            </select>
          </label>
          <label>Единица измерения
            <select v-model="form.unit_id">
              <option value="">— не выбрана —</option>
              <option v-for="u in units" :key="u.id" :value="u.id">
                {{ u.name }} ({{ u.short_name }})
              </option>
            </select>
          </label>
        </div>

        <div class="grid3">
          <label>Цена продажи
            <input v-model.number="form.price" type="number" step="0.01" min="0" />
          </label>
          <label>Себестоимость
            <input v-model.number="form.cost_price" type="number" step="0.01" min="0" />
          </label>
          <label>Несниж. остаток
            <input v-model.number="form.min_stock" type="number" step="0.001" min="0" />
          </label>
        </div>

        <label>Описание
          <textarea v-model="form.description" rows="2"></textarea>
        </label>

        <div class="modal-actions">
          <button
            v-if="modal.isEdit && !form.is_archived"
            type="button"
            class="danger"
            @click="onArchive"
          >Архивировать</button>
          <div style="flex:1"></div>
          <button type="button" @click="closeModal">Отмена</button>
          <button class="primary" type="submit" :disabled="modal.saving">
            {{ modal.saving ? 'Сохранение...' : 'Сохранить' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import {
  listProducts, createProduct, updateProduct, archiveProduct,
  type Product,
} from '../api/products'
import { listCategories, type Category } from '../api/categories'
import { listUnits, type Unit } from '../api/units'
import { apiErrorMessage } from '../api/client'

const products = ref<Product[]>([])
const categories = ref<Category[]>([])
const units = ref<Unit[]>([])

const loading = ref(false)
const error = ref<string | null>(null)
const showFilter = ref(false)
const page = ref(1)
const perPage = 100

const search = ref('')
const includeArchived = ref(false)
const filterSku = ref('')
const filterPriceFrom = ref<number | null>(null)
const filterPriceTo = ref<number | null>(null)
const selectedCategory = ref('')

const sortKey = ref<'name' | 'sku' | 'price'>('name')
const sortDir = ref<'asc' | 'desc'>('asc')

const selected = ref<Set<string>>(new Set())

const modal = reactive({
  open: false,
  isEdit: false,
  saving: false,
  error: null as string | null,
  editingId: '' as string,
})

const form = reactive({
  name: '',
  sku: '',
  barcode: '',
  category_id: '',
  unit_id: '',
  description: '',
  price: 0,
  cost_price: 0,
  min_stock: 0,
  currency: 'RUB',
  is_archived: false,
})

function shortId(id: string) { return id.slice(0, 8) }
function unitShort(id?: string) {
  if (!id) return '—'
  return units.value.find((u) => u.id === id)?.short_name ?? '—'
}
function formatMoney(n: number) {
  return n.toLocaleString('ru-RU', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function countByCategory(catId: string) {
  return products.value.filter((p) => p.category_id === catId).length
}

const totalCount = computed(() => products.value.length)

function selectCategory(id: string) {
  selectedCategory.value = id
  page.value = 1
}

function sortBy(key: typeof sortKey.value) {
  if (sortKey.value === key) sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  else { sortKey.value = key; sortDir.value = 'asc' }
  page.value = 1
}

const filtered = computed(() => {
  let rows = products.value
  if (selectedCategory.value) rows = rows.filter((p) => p.category_id === selectedCategory.value)
  if (search.value) {
    const q = search.value.toLowerCase()
    rows = rows.filter((p) =>
      p.name.toLowerCase().includes(q) ||
      (p.sku ?? '').toLowerCase().includes(q) ||
      (p.barcode ?? '').toLowerCase().includes(q)
    )
  }
  if (filterSku.value) {
    const q = filterSku.value.toLowerCase()
    rows = rows.filter((p) => (p.sku ?? '').toLowerCase().includes(q))
  }
  if (filterPriceFrom.value != null) rows = rows.filter((p) => p.price >= filterPriceFrom.value!)
  if (filterPriceTo.value != null) rows = rows.filter((p) => p.price <= filterPriceTo.value!)

  return [...rows].sort((a, b) => {
    const av = a[sortKey.value]; const bv = b[sortKey.value]
    if (typeof av === 'number' && typeof bv === 'number') {
      return sortDir.value === 'asc' ? av - bv : bv - av
    }
    const as = String(av ?? '').toLowerCase()
    const bs = String(bv ?? '').toLowerCase()
    return sortDir.value === 'asc' ? as.localeCompare(bs, 'ru') : bs.localeCompare(as, 'ru')
  })
})

const paginated = computed(() => {
  const from = (page.value - 1) * perPage
  return filtered.value.slice(from, from + perPage)
})
const rangeFrom = computed(() => filtered.value.length === 0 ? 0 : (page.value - 1) * perPage + 1)
const rangeTo = computed(() => Math.min(page.value * perPage, filtered.value.length))

const totalSum = computed(() => filtered.value.reduce((s, p) => s + p.price, 0))

const allChecked = computed(() =>
  paginated.value.length > 0 && paginated.value.every((p) => selected.value.has(p.id))
)
function toggleAll(e: Event) {
  const on = (e.target as HTMLInputElement).checked
  selected.value = on ? new Set(paginated.value.map((p) => p.id)) : new Set()
}
function toggleSelect(id: string) {
  const s = new Set(selected.value)
  s.has(id) ? s.delete(id) : s.add(id)
  selected.value = s
}

async function load() {
  loading.value = true
  error.value = null
  try {
    const [p, c, u] = await Promise.all([
      listProducts(includeArchived.value),
      listCategories(),
      listUnits(),
    ])
    products.value = p
    categories.value = c
    units.value = u
  } catch (e) {
    error.value = apiErrorMessage(e)
  } finally {
    loading.value = false
  }
}

function openCreate() {
  Object.assign(form, {
    name: '', sku: '', barcode: '',
    category_id: selectedCategory.value || '',
    unit_id: '', description: '',
    price: 0, cost_price: 0, min_stock: 0,
    currency: 'RUB', is_archived: false,
  })
  modal.open = true; modal.isEdit = false; modal.error = null; modal.editingId = ''
}

function openEdit(p: Product) {
  Object.assign(form, {
    name: p.name,
    sku: p.sku ?? '',
    barcode: p.barcode ?? '',
    category_id: p.category_id ?? '',
    unit_id: p.unit_id ?? '',
    description: p.description ?? '',
    price: p.price,
    cost_price: p.cost_price ?? 0,
    min_stock: p.min_stock ?? 0,
    currency: p.currency || 'RUB',
    is_archived: p.is_archived,
  })
  modal.open = true; modal.isEdit = true; modal.error = null; modal.editingId = p.id
}

function closeModal() { modal.open = false }

async function onSubmit() {
  modal.error = null
  modal.saving = true
  try {
    const payload = {
      name: form.name,
      sku: form.sku || undefined,
      barcode: form.barcode || undefined,
      category_id: form.category_id || undefined,
      unit_id: form.unit_id || undefined,
      description: form.description || undefined,
      price: form.price,
      cost_price: form.cost_price,
      min_stock: form.min_stock,
      currency: form.currency || 'RUB',
    }
    if (modal.isEdit) {
      await updateProduct(modal.editingId, payload)
    } else {
      await createProduct(payload)
    }
    closeModal()
    await load()
  } catch (e) {
    modal.error = apiErrorMessage(e)
  } finally {
    modal.saving = false
  }
}

async function onArchive() {
  if (!confirm(`Заархивировать «${form.name}»?`)) return
  try {
    await archiveProduct(modal.editingId)
    closeModal()
    await load()
  } catch (e) { modal.error = apiErrorMessage(e) }
}

function clearFilters() {
  search.value = ''
  filterSku.value = ''
  filterPriceFrom.value = null
  filterPriceTo.value = null
  page.value = 1
}

function printList() { window.print() }

function exportCsv() {
  const header = ['Наименование', 'Код', 'Артикул', 'Штрихкод', 'Ед. изм.', 'Цена продажи', 'Себестоимость']
  const lines = [header.join(';')]
  for (const p of filtered.value) {
    lines.push([
      csvEscape(p.name),
      shortId(p.id),
      csvEscape(p.sku ?? ''),
      csvEscape(p.barcode ?? ''),
      unitShort(p.unit_id),
      p.price.toFixed(2),
      (p.cost_price ?? 0).toFixed(2),
    ].join(';'))
  }
  const blob = new Blob(['\uFEFF' + lines.join('\r\n')], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `products-${new Date().toISOString().slice(0,10)}.csv`
  a.click()
  URL.revokeObjectURL(url)
}
function csvEscape(s: string) {
  if (s.includes(';') || s.includes('"') || s.includes('\n')) return '"' + s.replace(/"/g, '""') + '"'
  return s
}

function importCsv() { alert('Импорт: в разработке') }

onMounted(load)
</script>

<style scoped>
.products-layout {
  display: grid;
  grid-template-columns: 220px 1fr;
  gap: 16px;
  min-height: calc(100vh - 120px);
}

.categories-panel {
  background: #fff;
  border: 1px solid #d8dee4;
  border-radius: 4px;
  padding: 8px 0;
  height: fit-content;
  position: sticky;
  top: 16px;
}
.cat-section {
  padding: 10px 16px 4px;
  font-size: 11px;
  text-transform: uppercase;
  color: #8c959f;
  letter-spacing: 0.5px;
  font-weight: 600;
}
.cat-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 6px 16px;
  font-size: 13px;
  color: #1f2328;
  cursor: pointer;
  gap: 8px;
}
.cat-item:hover { background: #f6f8fa; }
.cat-item.active {
  background: #eef4ff;
  color: #2c5d9c;
  font-weight: 600;
}
.cat-name { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.cat-count {
  font-size: 11px;
  color: #8c959f;
  background: #f0f2f5;
  padding: 1px 6px;
  border-radius: 8px;
  flex-shrink: 0;
}
.cat-item.active .cat-count {
  background: #d8e4f0;
  color: #2c5d9c;
}
.cat-empty { padding: 8px 16px; font-size: 12px; }

.products-main { min-width: 0; }

/* Toolbar */
.icon-btn { display: inline-flex; align-items: center; gap: 4px; }
.icon-btn .plus { font-size: 13px; font-weight: 700; line-height: 1; }

.search-input {
  padding: 6px 10px;
  font-size: 13px;
  border: 1px solid #d0d7de;
  border-radius: 4px;
  background: #fff;
  color: #1f2328;
  min-width: 220px;
}

.counter {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 40px;
  height: 30px;
  padding: 0 10px;
  background: #fff;
  border: 1px solid #d0d7de;
  border-radius: 4px;
  font-size: 13px;
  color: #8c959f;
}
.counter.active { color: #2c5d9c; border-color: #2c5d9c; font-weight: 600; }

.mini-select {
  padding: 6px 10px;
  font-size: 13px;
  border: 1px solid #d0d7de;
  border-radius: 4px;
  background: #fff;
  color: #1f2328;
  max-width: 130px;
}
.mini-select:disabled { color: #8c959f; background: #f6f8fa; }

.btn.icon-only { padding: 6px 10px; font-size: 15px; }

tr.clickable { cursor: pointer; }
tr.clickable.selected { background: #eef4ff; }
.sort-arrow { color: #2c5d9c; font-size: 11px; margin-left: 4px; }
</style>