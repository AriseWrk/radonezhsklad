<template>
  <div>
    <div class="head">
      <h1>Склады</h1>
      <button class="primary" @click="showCreate = true">+ Новый склад</button>
    </div>

    <div v-if="error" class="error-box">{{ error }}</div>

    <table>
      <thead>
        <tr>
          <th>Название</th>
          <th>Адрес</th>
          <th>Статус</th>
          <th style="width: 100px"></th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="loading"><td colspan="4" class="muted">Загрузка...</td></tr>
        <tr v-else-if="items.length === 0"><td colspan="4" class="muted">Нет складов</td></tr>
        <tr v-for="w in items" :key="w.id">
          <td>{{ w.name }}</td>
          <td class="muted">{{ w.address || '—' }}</td>
          <td>
            <span v-if="w.is_active" class="pill active">активен</span>
            <span v-else class="pill archived">выключен</span>
          </td>
          <td>
            <button class="danger" @click="onDelete(w)">Удалить</button>
          </td>
        </tr>
      </tbody>
    </table>

    <div v-if="showCreate" class="modal-backdrop" @click.self="showCreate = false">
      <form class="card modal" @submit.prevent="onCreate">
        <h2>Новый склад</h2>
        <div v-if="createError" class="error-box">{{ createError }}</div>

        <label>Название*
          <input v-model="form.name" required />
        </label>
        <label>Адрес
          <input v-model="form.address" />
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
import { listWarehouses, createWarehouse, deleteWarehouse, type Warehouse } from '../api/warehouses'
import { apiErrorMessage } from '../api/client'

const items = ref<Warehouse[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const showCreate = ref(false)
const saving = ref(false)
const createError = ref<string | null>(null)
const form = reactive({ name: '', address: '' })

async function load() {
  loading.value = true
  error.value = null
  try {
    items.value = await listWarehouses()
  } catch (e) {
    error.value = apiErrorMessage(e)
  } finally {
    loading.value = false
  }
}

async function onCreate() {
  createError.value = null
  saving.value = true
  try {
    await createWarehouse(form.name, form.address || undefined)
    showCreate.value = false
    form.name = ''
    form.address = ''
    await load()
  } catch (e) {
    createError.value = apiErrorMessage(e)
  } finally {
    saving.value = false
  }
}

async function onDelete(w: Warehouse) {
  if (!confirm(`Удалить склад «${w.name}»?`)) return
  try {
    await deleteWarehouse(w.id)
    await load()
  } catch (e) {
    error.value = apiErrorMessage(e)
  }
}

onMounted(load)
</script>

<style scoped>
.head { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
h1 { margin: 0; }
.muted { color: var(--muted); }
.pill { padding: 2px 8px; border-radius: 10px; font-size: 12px; font-weight: 600; }
.pill.active { background: #e8f5e9; color: var(--success); }
.pill.archived { background: #f3f4f6; color: var(--muted); }
.modal-backdrop { position: fixed; inset: 0; background: rgba(0,0,0,0.4); display: flex; align-items: center; justify-content: center; padding: 20px; z-index: 100; }
.modal { width: 480px; max-width: 100%; display: flex; flex-direction: column; gap: 12px; }
.modal h2 { margin: 0 0 4px; font-size: 18px; }
label { display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: var(--muted); }
label input { color: var(--text); }
.modal-actions { display: flex; justify-content: flex-end; gap: 10px; margin-top: 6px; }
</style>