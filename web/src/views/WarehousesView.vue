<template>
  <div class="page">
    <div class="ms-title">
      <button class="ms-help" title="Справка"><MsIcon name="help" :size="14" /></button>
      <span>Склады</span>
      <button class="ms-refresh" @click="load" title="Обновить"><MsIcon name="refresh" :size="14" /></button>
    </div>

    <div class="ms-toolbar">
      <MsButton variant="primary" icon="plus" @click="showCreate = true">Склад</MsButton>
      <MsButton icon="filter">Фильтр</MsButton>
      <input v-model="search" class="ms-input" placeholder="Название или адрес" />
      <div class="ms-counter">{{ filtered.length }}</div>
      <MsButton icon="print">Печать</MsButton>
      <MsButton variant="icon" icon="gear" title="Настройки" />
    </div>

    <div v-if="error" class="error-box">{{ error }}</div>

    <table class="ms-table2">
      <thead>
        <tr>
          <th>Название</th>
          <th>Адрес</th>
          <th class="col-status">Статус</th>
          <th class="col-actions"></th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="loading"><td colspan="4" class="empty">Загрузка...</td></tr>
        <tr v-else-if="filtered.length === 0"><td colspan="4" class="empty">Нет складов</td></tr>
        <tr v-else v-for="w in filtered" :key="w.id" class="row">
          <td class="link">{{ w.name }}</td>
          <td class="muted">{{ w.address || '—' }}</td>
          <td class="col-status">
            <span v-if="w.is_active" class="badge bg-posted">активен</span>
            <span v-else class="badge bg-draft">выключен</span>
          </td>
          <td class="col-actions" @click.stop>
            <button class="btn-link-ms danger" @click="onDelete(w)">Удалить</button>
          </td>
        </tr>
      </tbody>
    </table>

    <div class="ms-footer2">
      <div class="pager">
        <span class="range">Всего: {{ filtered.length }}</span>
      </div>
    </div>

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
import { computed, onMounted, reactive, ref } from 'vue'
import { listWarehouses, createWarehouse, deleteWarehouse, type Warehouse } from '../api/warehouses'
import { apiErrorMessage } from '../api/client'
import MsButton from '../components/MsButton.vue'
import MsIcon from '../components/MsIcon.vue'

const items = ref<Warehouse[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const showCreate = ref(false)
const saving = ref(false)
const createError = ref<string | null>(null)
const search = ref('')
const form = reactive({ name: '', address: '' })

const filtered = computed(() => {
  const q = search.value.trim().toLowerCase()
  if (!q) return items.value
  return items.value.filter((w) =>
    (w.name ?? '').toLowerCase().includes(q) ||
    (w.address ?? '').toLowerCase().includes(q)
  )
})

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
    form.name = ''; form.address = ''
    await load()
  } catch (e) {
    createError.value = apiErrorMessage(e)
  } finally {
    saving.value = false
  }
}

async function onDelete(w: Warehouse) {
  if (!confirm(`Удалить склад «${w.name}»?`)) return
  try { await deleteWarehouse(w.id); await load() }
  catch (e) { error.value = apiErrorMessage(e) }
}

onMounted(load)
</script>

<style scoped>
.page { font-size: 13px; }

.ms-title {
  display: flex; align-items: center; gap: 8px;
  font-size: 20px; font-weight: 600; color: #1f2328;
  margin-bottom: 12px;
}
.ms-help {
  width: 20px; height: 20px; border-radius: 50%;
  border: 1px solid #b8c0c8; background: transparent;
  color: #57606a; cursor: pointer;
  display: inline-flex; align-items: center; justify-content: center; padding: 0;
}
.ms-help:hover { background: #f0f2f5; }
.ms-refresh {
  width: 24px; height: 24px; padding: 0;
  display: inline-flex; align-items: center; justify-content: center;
  border: none; background: transparent; color: #57606a; cursor: pointer;
}
.ms-refresh:hover { color: #2c5d9c; }
.ms-toolbar {
  display: flex; align-items: center; gap: 6px;
  margin-bottom: 12px; flex-wrap: wrap;
}
.ms-input {
  flex: 1; max-width: 320px; height: 30px; padding: 0 10px;
  font-size: 13px; border: 1px solid #d0d7de; border-radius: 4px;
  background: #fff; color: #1f2328;
}
.ms-input::placeholder { color: #8c959f; }
.ms-counter {
  min-width: 40px; height: 30px; padding: 0 10px;
  display: inline-flex; align-items: center; justify-content: center;
  border: 1px solid #d0d7de; border-radius: 4px;
  background: #fff; color: #8c959f;
  font-variant-numeric: tabular-nums; font-size: 13px;
}

.ms-table2 { width: 100%; border-collapse: collapse; background: #fff; font-size: 13px; }
.ms-table2 thead th {
  background: #fff; color: #2c5d9c; font-weight: 500;
  padding: 8px 10px; text-align: left;
  border-bottom: 1px solid #d8dee4; white-space: nowrap; font-size: 12px;
}
.ms-table2 tbody td {
  padding: 7px 10px;
  border-bottom: 1px solid #eaeef2;
  vertical-align: middle;
}
.ms-table2 tbody tr.row:hover { background: #f6f8fa; }
.ms-table2 .col-status { width: 130px; }
.ms-table2 .col-actions { width: 120px; text-align: right; }
.ms-table2 .link { color: #2c5d9c; font-weight: 500; }
.ms-table2 .empty { text-align: center; padding: 24px; color: #8c959f; }
.ms-table2 .muted { color: #57606a; }

.badge {
  display: inline-block; padding: 2px 10px; border-radius: 3px;
  font-size: 11px; font-weight: 600; color: #fff;
}
.bg-posted { background: #2196f3; }
.bg-draft  { background: #8c959f; }

.btn-link-ms {
  border: none; background: transparent;
  color: #2c5d9c; cursor: pointer;
  font-size: 12px; padding: 4px 8px; border-radius: 3px;
}
.btn-link-ms:hover { background: #f0f2f5; }
.btn-link-ms.danger { color: #cf222e; }
.btn-link-ms.danger:hover { background: #ffecec; }

.ms-footer2 {
  display: flex; align-items: center; justify-content: space-between;
  padding: 6px 10px; font-size: 12px; color: #57606a;
  background: #fff; border-top: 1px solid #eaeef2;
}
.pager .range { font-variant-numeric: tabular-nums; }

.modal-backdrop {
  position: fixed; inset: 0; background: rgba(0,0,0,0.4);
  display: flex; align-items: center; justify-content: center;
  padding: 20px; z-index: 100;
}
.modal {
  background: #fff; border-radius: 4px;
  width: 420px; max-width: 100%; padding: 20px;
  display: flex; flex-direction: column; gap: 10px;
}
.modal h2 { margin: 0 0 4px; font-size: 18px; }
.modal label { font-size: 13px; color: #444; display: flex; flex-direction: column; gap: 4px; }
.modal-actions { display: flex; justify-content: flex-end; gap: 10px; margin-top: 6px; }
</style>