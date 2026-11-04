<template>
  <div v-if="visible" class="ms-modal-overlay" @click.self="close">
    <div class="ms-modal">
      <div class="ms-modal-header">
        <h2>Новый документ</h2>
        <button class="close-btn" @click="close">×</button>
      </div>

      <div class="ms-modal-body">
        <div v-if="error" class="error-msg">{{ error }}</div>

        <div class="form-row">
          <div class="form-group">
            <label>Тип документа</label>
            <select v-model="form.type" class="ms-select">
              <option value="receipt">Приёмка</option>
              <option value="shipment">Отгрузка</option>
              <option value="transfer">Перемещение</option>
              <option value="writeoff">Списание</option>
              <option value="inventory">Инвентаризация</option>
            </select>
          </div>
          <div class="form-group">
            <label>Склад</label>
            <select v-model="form.warehouse_id" class="ms-select">
              <option value="">— выберите —</option>
              <option v-for="w in warehouses" :key="w.id" :value="w.id">{{ w.name }}</option>
            </select>
          </div>
        </div>

        <div class="form-row" v-if="form.type === 'transfer'">
          <div class="form-group">
            <label>Целевой склад</label>
            <select v-model="form.target_warehouse_id" class="ms-select">
              <option value="">— выберите —</option>
              <option v-for="w in warehouses" :key="w.id" :value="w.id">{{ w.name }}</option>
            </select>
          </div>
        </div>

        <div class="form-group">
          <label>Комментарий</label>
          <input v-model="form.comment" class="ms-input" placeholder="Комментарий" />
        </div>

        <div class="form-group">
          <label>Позиции</label>
          <div class="items-list">
            <div v-for="(it, idx) in form.items" :key="idx" class="item-row">
              <select v-model="it.product_id" class="ms-select product">
                <option value="">— товар —</option>
                <option v-for="p in products" :key="p.id" :value="p.id">{{ p.name }}</option>
              </select>
              <input v-model.number="it.quantity" type="number" step="0.001" min="0.001" class="ms-input qty" placeholder="Кол-во" />
              <input v-model.number="it.price" type="number" step="0.01" min="0" class="ms-input price" placeholder="Цена" />
              <button class="del-btn" @click="removeItem(idx)">×</button>
            </div>
            <button class="add-item-btn" @click="addItem">+ Добавить строку</button>
          </div>
        </div>
      </div>

      <div class="ms-modal-footer">
        <button class="btn-cancel" @click="close">Отмена</button>
        <button class="btn-save" @click="submit" :disabled="saving">Сохранить</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import { createDocument } from '../api/documents'
import { listWarehouses, type Warehouse } from '../api/warehouses'
import { listProducts, type Product } from '../api/products'

const props = defineProps<{
  visible: boolean
  defaultType?: string
  defaultWarehouseId?: string
}>()

const emit = defineEmits<{
  (e: 'update:visible', v: boolean): void
  (e: 'created'): void
}>()

const warehouses = ref<Warehouse[]>([])
const products = ref<Product[]>([])
const saving = ref(false)
const error = ref<string | null>(null)

const form = reactive({
  type: 'receipt',
  warehouse_id: '',
  target_warehouse_id: '',
  comment: '',
  items: [] as Array<{ product_id: string; quantity: number; price: number }>,
})

watch(() => props.visible, async (v) => {
  if (v) {
    error.value = null
    form.type = props.defaultType || 'receipt'
    form.warehouse_id = props.defaultWarehouseId || ''
    form.target_warehouse_id = ''
    form.comment = ''
    form.items = [{ product_id: '', quantity: 1, price: 0 }]
    try {
      const [w, p] = await Promise.all([listWarehouses(), listProducts(true)])
      warehouses.value = w
      products.value = p
    } catch { /* ignore */ }
  }
})

function close() {
  emit('update:visible', false)
}

function addItem() {
  form.items.push({ product_id: '', quantity: 1, price: 0 })
}

function removeItem(idx: number) {
  form.items.splice(idx, 1)
}

async function submit() {
  error.value = null
  if (!form.warehouse_id) { error.value = 'Выберите склад'; return }
  if (form.type === 'transfer' && !form.target_warehouse_id) { error.value = 'Выберите целевой склад'; return }
  if (form.items.some(i => !i.product_id || i.quantity <= 0)) { error.value = 'Заполните все позиции'; return }

  saving.value = true
  try {
    await createDocument({
      type: form.type,
      warehouse_id: form.warehouse_id,
      target_warehouse_id: form.type === 'transfer' ? form.target_warehouse_id : undefined,
      comment: form.comment || undefined,
      items: form.items,
    })
    emit('created')
    close()
  } catch (e: any) {
    error.value = e.response?.data?.message || e.message || 'Ошибка сохранения'
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.ms-modal-overlay {
  position: fixed; inset: 0; background: rgba(0,0,0,0.4);
  display: flex; align-items: center; justify-content: center; z-index: 1000;
}
.ms-modal {
  background: #fff; border-radius: 6px; width: 700px; max-width: 100%;
  box-shadow: 0 10px 30px rgba(0,0,0,0.15); display: flex; flex-direction: column;
  max-height: 90vh;
}
.ms-modal-header {
  padding: 16px 24px; border-bottom: 1px solid #eaeef2;
  display: flex; justify-content: space-between; align-items: center;
}
.ms-modal-header h2 { margin: 0; font-size: 18px; font-weight: 600; color: #1f2328; }
.close-btn { background: none; border: none; font-size: 24px; color: #8c959f; cursor: pointer; padding: 0; line-height: 1; }
.close-btn:hover { color: #cf222e; }

.ms-modal-body { padding: 24px; overflow-y: auto; flex: 1; }
.error-msg { background: #ffeef0; color: #cf222e; padding: 10px; border-radius: 4px; margin-bottom: 16px; font-size: 13px; }

.form-row { display: flex; gap: 16px; margin-bottom: 16px; }
.form-group { display: flex; flex-direction: column; gap: 6px; flex: 1; margin-bottom: 16px; }
.form-group label { font-size: 12px; color: #57606a; font-weight: 500; }

.ms-select, .ms-input {
  height: 32px; padding: 0 10px; border: 1px solid #d0d7de; border-radius: 4px;
  font-size: 13px; color: #1f2328; background: #fff; outline: none;
}
.ms-select:focus, .ms-input:focus { border-color: #2c5d9c; box-shadow: 0 0 0 2px rgba(44,93,156,0.2); }

.items-list { display: flex; flex-direction: column; gap: 8px; }
.item-row { display: grid; grid-template-columns: 1fr 100px 100px 32px; gap: 8px; }
.del-btn { background: #f6f8fa; color: #cf222e; border: 1px solid #d0d7de; border-radius: 4px; cursor: pointer; font-size: 18px; line-height: 1; display: flex; align-items: center; justify-content: center; }
.del-btn:hover { background: #ffeef0; border-color: #ffb3b8; }
.add-item-btn { background: none; border: none; color: #0969da; font-size: 13px; cursor: pointer; align-self: flex-start; padding: 8px 0; font-weight: 500; }
.add-item-btn:hover { text-decoration: underline; }

.ms-modal-footer {
  padding: 16px 24px; border-top: 1px solid #eaeef2;
  display: flex; justify-content: flex-end; gap: 12px; background: #f6f8fa; border-radius: 0 0 6px 6px;
}
.btn-cancel { background: #fff; border: 1px solid #d0d7de; border-radius: 4px; padding: 6px 16px; font-size: 13px; cursor: pointer; color: #1f2328; }
.btn-cancel:hover { background: #eaeef2; }
.btn-save { background: #2da44e; color: #fff; border: 1px solid #2a8e44; border-radius: 4px; padding: 6px 16px; font-size: 13px; cursor: pointer; font-weight: 500; }
.btn-save:hover { background: #2c974b; }
.btn-save:disabled { opacity: 0.6; cursor: default; }
</style>