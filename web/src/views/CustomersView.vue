<template>
  <div>
    <div class="head">
      <h1>Покупатели</h1>
      <button class="primary" @click="showCreate = true">+ Новый покупатель</button>
    </div>

    <div v-if="error" class="error-box">{{ error }}</div>

    <table>
      <thead>
        <tr>
          <th>Название</th>
          <th>Телефон</th>
          <th>Email</th>
          <th>Адрес</th>
          <th style="width: 100px"></th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="loading"><td colspan="5" class="muted">Загрузка...</td></tr>
        <tr v-else-if="items.length === 0"><td colspan="5" class="muted">Нет покупателей</td></tr>
        <tr v-for="c in items" :key="c.id">
          <td>{{ c.name }}</td>
          <td>{{ c.phone || '—' }}</td>
          <td>{{ c.email || '—' }}</td>
          <td class="muted">{{ c.address || '—' }}</td>
          <td>
            <button class="danger" @click="onDelete(c)">Удалить</button>
          </td>
        </tr>
      </tbody>
    </table>

    <div v-if="showCreate" class="modal-backdrop" @click.self="showCreate = false">
      <form class="card modal" @submit.prevent="onCreate">
        <h2>Новый покупатель</h2>
        <div v-if="createError" class="error-box">{{ createError }}</div>

        <label>Название*
          <input v-model="form.name" required />
        </label>
        <label>Телефон
          <input v-model="form.phone" />
        </label>
        <label>Email
          <input v-model="form.email" type="email" />
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
import { listCustomers, createCustomer, deleteCustomer, type Customer } from '../api/customers'
import { apiErrorMessage } from '../api/client'

const items = ref<Customer[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const showCreate = ref(false)
const saving = ref(false)
const createError = ref<string | null>(null)
const form = reactive({ name: '', phone: '', email: '', address: '' })

async function load() {
  loading.value = true
  error.value = null
  try {
    items.value = await listCustomers()
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
    await createCustomer({
      name: form.name,
      phone: form.phone || undefined,
      email: form.email || undefined,
      address: form.address || undefined,
    })
    showCreate.value = false
    form.name = ''; form.phone = ''; form.email = ''; form.address = ''
    await load()
  } catch (e) {
    createError.value = apiErrorMessage(e)
  } finally {
    saving.value = false
  }
}

async function onDelete(c: Customer) {
  if (!confirm(`Удалить «${c.name}»?`)) return
  try { await deleteCustomer(c.id); await load() } catch (e) { error.value = apiErrorMessage(e) }
}

onMounted(load)
</script>

<style scoped>
.head { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
h1 { margin: 0; }
.muted { color: var(--muted); }
.modal-backdrop { position: fixed; inset: 0; background: rgba(0,0,0,0.4); display: flex; align-items: center; justify-content: center; padding: 20px; z-index: 100; }
.modal { width: 480px; max-width: 100%; display: flex; flex-direction: column; gap: 12px; }
.modal h2 { margin: 0; font-size: 18px; }
label { display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: var(--muted); }
label input { color: var(--text); }
.modal-actions { display: flex; justify-content: flex-end; gap: 10px; margin-top: 6px; }
</style>