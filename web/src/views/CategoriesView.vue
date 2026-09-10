<template>
  <div>
    <div class="head">
      <h1>Категории</h1>
      <button class="primary" @click="showCreate = true">+ Новая категория</button>
    </div>

    <div v-if="error" class="error-box">{{ error }}</div>

    <table>
      <thead>
        <tr>
          <th>Название</th>
          <th>Родитель</th>
          <th style="width: 120px"></th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="loading"><td colspan="3" class="muted">Загрузка...</td></tr>
        <tr v-else-if="categories.length === 0"><td colspan="3" class="muted">Нет категорий</td></tr>
        <tr v-for="c in categories" :key="c.id">
          <td>{{ c.name }}</td>
          <td class="muted">{{ parentName(c.parent_id) }}</td>
          <td>
            <button class="danger" @click="onDelete(c)">Удалить</button>
          </td>
        </tr>
      </tbody>
    </table>

    <div v-if="showCreate" class="modal-backdrop" @click.self="showCreate = false">
      <form class="card modal" @submit.prevent="onCreate">
        <h2>Новая категория</h2>
        <div v-if="createError" class="error-box">{{ createError }}</div>

        <label>Название*
          <input v-model="form.name" required />
        </label>

        <label>Родительская категория
          <select v-model="form.parent_id">
            <option value="">— нет (корневая) —</option>
            <option v-for="c in categories" :key="c.id" :value="c.id">{{ c.name }}</option>
          </select>
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
import { listCategories, createCategory, deleteCategory, type Category } from '../api/categories'
import { apiErrorMessage } from '../api/client'

const categories = ref<Category[]>([])
const loading = ref(false)
const error = ref<string | null>(null)

const showCreate = ref(false)
const saving = ref(false)
const createError = ref<string | null>(null)

const form = reactive({ name: '', parent_id: '' })

async function load() {
  loading.value = true
  error.value = null
  try {
    categories.value = await listCategories()
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
    await createCategory(form.name, form.parent_id || null)
    showCreate.value = false
    form.name = ''
    form.parent_id = ''
    await load()
  } catch (e) {
    createError.value = apiErrorMessage(e)
  } finally {
    saving.value = false
  }
}

async function onDelete(c: Category) {
  if (!confirm(`Удалить «${c.name}»?`)) return
  try {
    await deleteCategory(c.id)
    await load()
  } catch (e) {
    error.value = apiErrorMessage(e)
  }
}

function parentName(id?: string) {
  if (!id) return '—'
  return categories.value.find((c) => c.id === id)?.name ?? '—'
}

onMounted(load)
</script>

<style scoped>
.head { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
h1 { margin: 0; }
.muted { color: var(--muted); }
.modal-backdrop {
  position: fixed; inset: 0;
  background: rgba(0,0,0,0.4);
  display: flex; align-items: center; justify-content: center;
  padding: 20px; z-index: 100;
}
.modal { width: 480px; max-width: 100%; display: flex; flex-direction: column; gap: 12px; }
.modal h2 { margin: 0 0 4px; font-size: 18px; }
label { display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: var(--muted); }
label input, label select { color: var(--text); }
.modal-actions { display: flex; justify-content: flex-end; gap: 10px; margin-top: 6px; }
</style>