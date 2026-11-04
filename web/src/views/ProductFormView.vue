<template>
  <div class="pf-page">
    <!-- Верхняя шапка -->
    <div class="pf-topbar">
      <div class="pf-topbar-left">
        <button class="pf-btn pf-btn-green" @click="onSave" :disabled="saving">
          {{ saving ? 'Сохранение...' : 'Сохранить' }}
        </button>
        <button class="pf-btn" @click="onClose">Закрыть</button>
      </div>
      <div class="pf-topbar-right">
        <button class="pf-btn pf-btn-icon" title="Печать" @click="onPrint"><MsIcon name="print" :size="14" /><span>Печать</span><span class="caret">▾</span></button>
        <button class="pf-btn pf-btn-icon" title="Настройки"><MsIcon name="gear" :size="14" /></button>
        <button class="pf-btn pf-btn-icon" title="Ещё"><MsIcon name="dots" :size="14" /></button>
      </div>
    </div>

    <div v-if="error" class="pf-error">{{ error }}</div>

    <!-- Наименование -->
    <div class="pf-name-row">
      <label class="pf-required">* Наименование товара</label>
      <input v-model="form.name" class="pf-name-input" placeholder="" />
    </div>

    <!-- Основная сетка -->
    <div class="pf-grid">
      <!-- Левая колонка -->
      <div class="pf-col-left">
        <section class="pf-card">
          <header class="pf-card-h">
            <span class="dots-grip">⋮⋮</span>
            <span class="pf-card-title">Изображения</span>
            <span class="chev">▾</span>
          </header>
          <div class="pf-card-body">
            <button class="pf-dashed" type="button"><MsIcon name="plus" :size="12" /> Изображение</button>
          </div>
        </section>

        <section class="pf-card">
          <header class="pf-card-h">
            <span class="dots-grip">⋮⋮</span>
            <span class="pf-card-title">Общие данные</span>
            <span class="chev">▾</span>
          </header>
          <div class="pf-card-body">
            <div class="pf-field">
              <label>Описание</label>
              <textarea v-model="form.description" rows="2"></textarea>
            </div>
            <div class="pf-field">
              <label>Группа</label>
              <select v-model="form.category_id">
                <option value="">—</option>
                <option v-for="c in categories" :key="c.id" :value="c.id">{{ c.name }}</option>
              </select>
            </div>
            <div class="pf-field">
              <label>Страна</label>
              <select><option>—</option></select>
            </div>
            <div class="pf-field">
              <label>Поставщик</label>
              <div class="pf-with-plus">
                <select><option>—</option></select>
                <button class="pf-plus-btn" type="button"><MsIcon name="plus" :size="12" /></button>
              </div>
            </div>
            <div class="pf-field">
              <label>Артикул</label>
              <input v-model="form.sku" />
            </div>
            <div class="pf-field">
              <label>Код</label>
              <input :value="form.code" readonly class="pf-readonly" />
            </div>
            <div class="pf-field">
              <label>Внешний код</label>
              <input v-model="form.external_id" />
            </div>
            <div class="pf-field">
              <label>Единица измерения</label>
              <div class="pf-with-plus">
                <select v-model="form.unit_id">
                  <option value="">—</option>
                  <option v-for="u in units" :key="u.id" :value="u.id">{{ u.short_name }} ({{ u.name }})</option>
                </select>
                <button class="pf-plus-btn" type="button"><MsIcon name="plus" :size="12" /></button>
              </div>
            </div>
            <div class="pf-field">
              <label>Вес</label>
              <input v-model.number="form.weight" type="number" min="0" step="0.001" />
            </div>
            <div class="pf-field">
              <label>Объём</label>
              <input v-model.number="form.volume" type="number" min="0" step="0.001" />
            </div>
            <div class="pf-field">
              <label>НДС</label>
              <select v-model="form.vat">
                <option value="">—</option>
                <option value="0">0%</option>
                <option value="10">10%</option>
                <option value="20">20%</option>
              </select>
            </div>
          </div>
        </section>

        <section class="pf-card">
          <header class="pf-card-h">
            <span class="dots-grip">⋮⋮</span>
            <span class="pf-card-title">Контент для разных торговых площадок <span class="pf-help">?</span></span>
            <span class="chev">▾</span>
          </header>
          <div class="pf-card-body">
            <div class="pf-info">Создайте уникальное название и описание для каждой торговой площадки с помощью искусственного интеллекта — с учетом требований к контенту, особенностей аудитории и других параметров.</div>
          </div>
        </section>
      </div>

      <!-- Правая колонка -->
      <div class="pf-col-right">
        <div class="pf-tabs">
          <button v-for="t in tabs" :key="t.key" class="pf-tab" :class="{active: activeTab===t.key}" @click="activeTab=t.key">
            {{ t.label }}
          </button>
        </div>

        <div v-if="activeTab==='prices'" class="pf-tab-body">
          <div class="pf-field">
            <label>Минимальная цена <span class="pf-help">?</span></label>
            <input v-model.number="form.min_price" type="number" step="0.01" min="0" placeholder="0" />
          </div>
          <div class="pf-field">
            <label>Закупочная цена <span class="pf-help">?</span></label>
            <input v-model.number="form.cost_price" type="number" step="0.01" min="0" placeholder="0" />
          </div>
          <div class="pf-field pf-field-inline">
            <label>Цены продажи <span class="pf-help">?</span></label>
            <button class="pf-btn" type="button">Настроить цены</button>
          </div>
          <div class="pf-field">
            <label>Цена продажи</label>
            <input v-model.number="form.price" type="number" step="0.01" min="0" placeholder="0" />
          </div>
          <label class="pf-checkbox">
            <input type="checkbox" v-model="form.no_discount" />
            <span>Запретить скидки при продаже в розницу</span>
          </label>
        </div>

        <div v-else class="pf-tab-empty">
          Пока нет данных.
        </div>
      </div>
    </div>

    <!-- Нижние блоки (на всю ширину, в левой колонке под "Изображениями" и т.п.) -->
    <div class="pf-bottom-left">
      <section class="pf-card">
        <header class="pf-card-h">
          <span class="dots-grip">⋮⋮</span>
          <span class="pf-card-title">Неснижаемый остаток <span class="pf-help">?</span></span>
          <span class="chev">▾</span>
        </header>
        <div class="pf-card-body">
          <div class="pf-info">Минимальное количество товара, которое всегда должно быть на складе. Когда количество становится меньше заданного, приходит уведомление. В разделе Товары → Остатки можно Пополнить резервы — заказать товары для пополнения до указанного значения.</div>
          <label class="pf-radio"><input type="radio" checked /><span>В сумме на всех складах</span>
            <input v-model.number="form.min_stock" type="number" class="pf-mini-input" placeholder="Не указан" />
          </label>
          <label class="pf-radio"><input type="radio" disabled /><span>Одинаковый на всех складах</span></label>
          <label class="pf-radio"><input type="radio" disabled /><span>Задать для каждого склада</span></label>
        </div>
      </section>

      <section class="pf-card">
        <header class="pf-card-h">
          <span class="dots-grip">⋮⋮</span>
          <span class="pf-card-title">Особенности учета</span>
          <span class="chev">▾</span>
        </header>
        <div class="pf-card-body">
          <div class="pf-field"><label>Фасовка</label><select><option>Штучная</option><option>Весовая</option></select></div>
          <div class="pf-field"><label>Тип учета</label><select><option>Без специализированного учета</option></select></div>
          <div class="pf-field pf-field-inline"><label>Прослеживаемый <span class="pf-help">?</span></label><input type="checkbox" disabled /></div>
          <div class="pf-subtitle">Маркировка</div>
          <a href="#" class="pf-link" @click.prevent>Подключить маркировку</a>
          <div class="pf-field"><label>Владелец карточки <span class="pf-help">?</span></label><input disabled /></div>
          <div class="pf-field"><label>Тип продукции</label><select disabled><option>Не маркируется</option></select></div>
        </div>
      </section>

      <section class="pf-card">
        <header class="pf-card-h">
          <span class="dots-grip">⋮⋮</span>
          <span class="pf-card-title">Штрихкоды товара <span class="pf-help">?</span></span>
          <span class="chev">▾</span>
        </header>
        <div class="pf-card-body">
          <div class="pf-info">Укажите штрихкод товара, чтобы добавлять его в документы при помощи сканера штрихкодов. Код GTIN включает в себя только цифры и может иметь длину 8, 12, 13 или 14 цифр. В учетных системах всегда используется 14 знаков, коды с меньшим количеством цифр дополняют ведущими нулями, например, 95050003 → 00000095050003.</div>
          <div class="pf-barcode-row">
            <select><option>EAN13</option></select>
            <input v-model="form.barcode" placeholder="Штрихкод" />
            <button class="pf-plus-btn" type="button"><MsIcon name="dots" :size="12" /></button>
          </div>
          <button class="pf-dashed" type="button"><MsIcon name="plus" :size="12" /> Штрихкод</button>
        </div>
      </section>

      <section class="pf-card">
        <header class="pf-card-h">
          <span class="dots-grip">⋮⋮</span>
          <span class="pf-card-title">Кассовый чек</span>
          <span class="chev">▾</span>
        </header>
        <div class="pf-card-body">
          <div class="pf-field"><label>Система налогообложения</label><select><option>Совпадает с точкой</option></select></div>
          <div class="pf-field"><label>Признак предмета расчета</label><select><option>Товар</option></select></div>
        </div>
      </section>

      <section class="pf-card">
        <header class="pf-card-h">
          <span class="dots-grip">⋮⋮</span>
          <span class="pf-card-title">Доступ</span>
          <span class="chev">▾</span>
        </header>
        <div class="pf-card-body">
          <div class="pf-field"><label>Сотрудник</label><input disabled value="—" /></div>
          <div class="pf-field"><label>Отдел</label><input disabled value="Основной" /></div>
          <label class="pf-checkbox"><input type="checkbox" checked /><span>Общий доступ</span></label>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  getProduct, createProduct, updateProduct,
  type ProductInput,
} from '../api/products'
import { listCategories, type Category } from '../api/categories'
import { listUnits, type Unit } from '../api/units'
import { apiErrorMessage } from '../api/client'
import MsIcon from '../components/MsIcon.vue'

const route = useRoute()
const router = useRouter()

const isNew = computed(() => route.name === 'product-new' || route.params.id === 'new')
const editingId = computed(() => isNew.value ? '' : String(route.params.id || ''))

const categories = ref<Category[]>([])
const units = ref<Unit[]>([])
const saving = ref(false)
const error = ref<string | null>(null)

const activeTab = ref<'prices' | 'mods' | 'analogs' | 'pack' | 'stock' | 'history' | 'files'>('prices')
const tabs = [
  { key: 'prices',  label: 'Цены' },
  { key: 'mods',    label: 'Модификации (0)' },
  { key: 'analogs', label: 'Аналоги' },
  { key: 'pack',    label: 'Упаковка (0)' },
  { key: 'stock',   label: 'Остатки' },
  { key: 'history', label: 'История' },
  { key: 'files',   label: 'Файлы (0)' },
] as const

const form = reactive({
  name: '',
  sku: '',
  barcode: '',
  external_id: '',
  category_id: '',
  unit_id: '',
  description: '',
  price: 0,
  cost_price: 0,
  min_price: 0,
  min_stock: 0,
  weight: 0,
  volume: 0,
  vat: '',
  code: '',
  no_discount: false,
  currency: 'RUB',
})

async function load() {
  try {
    const [c, u] = await Promise.all([listCategories(), listUnits()])
    categories.value = c
    units.value = u
    if (editingId.value) {
      const p = await getProduct(editingId.value)
      Object.assign(form, {
        name: p.name,
        sku: p.sku ?? '',
        barcode: p.barcode ?? '',
        external_id: p.external_id ?? '',
        category_id: p.category_id ?? '',
        unit_id: p.unit_id ?? '',
        description: p.description ?? '',
        price: p.price,
        cost_price: p.cost_price ?? 0,
        min_stock: p.min_stock ?? 0,
        currency: p.currency || 'RUB',
        code: p.id.slice(0, 8),
      })
    } else {
      form.code = Math.floor(10000 + Math.random() * 90000).toString()
    }
  } catch (e) {
    error.value = apiErrorMessage(e)
  }
}

async function onSave() {
  error.value = null
  if (!form.name.trim()) { error.value = 'Введите наименование'; return }
  saving.value = true
  try {
    const payload: ProductInput = {
      name: form.name.trim(),
      sku: form.sku || undefined,
      barcode: form.barcode || undefined,
      category_id: form.category_id || undefined,
      unit_id: form.unit_id || undefined,
      description: form.description || undefined,
      price: Number(form.price) || 0,
      cost_price: Number(form.cost_price) || 0,
      min_stock: Number(form.min_stock) || 0,
      currency: form.currency || 'RUB',
    }
    if (editingId.value) {
      await updateProduct(editingId.value, payload)
    } else {
      const created = await createProduct(payload)
      router.replace({ name: 'product-edit', params: { id: created.id } })
    }
  } catch (e) {
    error.value = apiErrorMessage(e)
  } finally {
    saving.value = false
  }
}

function onClose() { router.push({ name: 'products' }) }
function onPrint() { window.print() }

onMounted(load)
</script>

<style scoped>
.pf-page { padding: 0; }
.pf-topbar {
  display: flex; align-items: center; justify-content: space-between;
  padding: 8px 0 12px;
  gap: 8px;
}
.pf-topbar-left, .pf-topbar-right { display: flex; gap: 6px; align-items: center; }
.pf-btn {
  height: 30px; padding: 0 12px;
  border-radius: 4px; border: 1px solid #d0d7de;
  background: #fff; color: #1f2328; font-size: 13px; cursor: pointer;
  display: inline-flex; align-items: center; gap: 6px;
}
.pf-btn:hover { background: #f6f8fa; }
.pf-btn:disabled { opacity: .6; cursor: default; }
.pf-btn-green {
  background: #a8ce5e; border-color: #a8ce5e; color: #1f2328; font-weight: 600;
}
.pf-btn-green:hover:not(:disabled) { background: #97bd4b; border-color: #97bd4b; }
.pf-btn-icon { padding: 0 8px; }
.pf-btn-icon .caret { font-size: 10px; color: #8c959f; margin-left: 2px; }

.pf-error {
  background: #fff0f0; border: 1px solid #cf222e; color: #cf222e;
  padding: 8px 12px; border-radius: 4px; margin-bottom: 10px; font-size: 13px;
}

.pf-name-row { display: flex; flex-direction: column; gap: 4px; margin-bottom: 16px; }
.pf-required { color: #cf222e; font-size: 12px; }
.pf-name-input {
  height: 36px; padding: 0 10px; font-size: 14px;
  border: 1px solid #2c5d9c; border-radius: 4px; outline: none;
  background: #fff; color: #1f2328;
}

.pf-grid {
  display: grid;
  grid-template-columns: 320px 1fr;
  gap: 16px;
  align-items: start;
}

.pf-col-left { display: flex; flex-direction: column; gap: 12px; }
.pf-col-right { min-width: 0; }
.pf-tabs { display: flex; gap: 0; border-bottom: 1px solid #d8dee4; padding: 0; margin-bottom: 14px; }
.pf-tab-body { padding: 0; display: flex; flex-direction: column; gap: 12px; max-width: 380px; }
.pf-bottom-left { display: flex; flex-direction: column; gap: 12px; max-width: 320px; margin-top: 12px; }

.pf-card { background: #fff; border: 1px solid #d8dee4; border-radius: 4px; overflow: hidden; }
.pf-card-h {
  display: flex; align-items: center; gap: 8px;
  padding: 8px 12px; background: #f6f8fa; border-bottom: 1px solid #eaeef2;
  font-size: 13px; font-weight: 600; color: #1f2328;
}
.pf-card-title { flex: 1; }
.pf-card-h .dots-grip { color: #8c959f; font-size: 12px; }
.pf-card-h .chev { color: #8c959f; font-size: 11px; }
.pf-help {
  display: inline-flex; align-items: center; justify-content: center;
  width: 14px; height: 14px; border-radius: 50%;
  background: #d8dee4; color: #57606a; font-size: 10px; font-weight: 600;
}
.pf-card-body { padding: 12px; display: flex; flex-direction: column; gap: 10px; }

.pf-field { display: flex; flex-direction: column; gap: 4px; }
.pf-field label { font-size: 12px; color: #57606a; }
.pf-field input, .pf-field select, .pf-field textarea {
  height: 28px; padding: 0 8px; font-size: 13px;
  border: 1px solid #d0d7de; border-radius: 3px;
  background: #fff; color: #1f2328;
}
.pf-field textarea { height: auto; padding: 6px 8px; resize: vertical; }
.pf-field input:disabled, .pf-field select:disabled { background: #f6f8fa; color: #8c959f; }
.pf-readonly { background: #f6f8fa !important; color: #57606a; }
.pf-field-inline { flex-direction: row; align-items: center; gap: 8px; }
.pf-field-inline label { flex: 1; }

.pf-with-plus { display: flex; gap: 4px; }
.pf-with-plus select { flex: 1; }
.pf-plus-btn {
  width: 28px; height: 28px; flex-shrink: 0;
  border: 1px solid #d0d7de; background: #fff; border-radius: 3px;
  cursor: pointer; color: #2c5d9c;
  display: inline-flex; align-items: center; justify-content: center;
}
.pf-plus-btn:hover { background: #f6f8fa; }

.pf-dashed {
  display: inline-flex; align-items: center; gap: 4px;
  height: 28px; padding: 0 10px; font-size: 12px;
  border: 1px dashed #d0d7de; border-radius: 3px;
  background: #fff; color: #2c5d9c; cursor: pointer;
}
.pf-dashed:hover { background: #f6f8fa; border-color: #2c5d9c; }

.pf-info {
  font-size: 12px; color: #57606a; line-height: 1.4;
  background: #f0f8ff; border: 1px solid #b6dcff; border-radius: 3px;
  padding: 8px 10px;
}

.pf-subtitle { font-size: 12px; font-weight: 600; color: #1f2328; margin-top: 4px; }
.pf-link { color: #2c5d9c; font-size: 12px; text-decoration: none; }
.pf-link:hover { text-decoration: underline; }

.pf-radio, .pf-checkbox {
  display: flex; align-items: center; gap: 6px;
  font-size: 13px; color: #1f2328; cursor: pointer;
}
.pf-mini-input { width: 120px; height: 24px; padding: 0 6px; font-size: 12px; border: 1px solid #d0d7de; border-radius: 3px; }
.pf-barcode-row { display: flex; gap: 6px; align-items: center; }
.pf-barcode-row select { width: 90px; }
.pf-barcode-row input { flex: 1; height: 28px; padding: 0 8px; border: 1px solid #d0d7de; border-radius: 3px; font-size: 13px; }

.pf-tabs {
  display: flex; gap: 0; border-bottom: 1px solid #eaeef2; padding: 0 8px;
}
.pf-tab {
  padding: 10px 14px; font-size: 13px; color: #57606a; background: transparent;
  border: none; border-bottom: 2px solid transparent; cursor: pointer;
  white-space: nowrap;
}
.pf-tab:hover { color: #2c5d9c; }
.pf-tab.active { color: #2c5d9c; border-bottom-color: #2c5d9c; font-weight: 600; }


.pf-tab-empty { padding: 24px; text-align: left; color: #8c959f; font-size: 13px; }
</style>