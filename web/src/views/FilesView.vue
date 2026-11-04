<template>
  <div class="page">
    <div class="ms-title">
      <span class="title-text">Файлы</span>
      <button class="ms-refresh" @click="load" title="Обновить"><MsIcon name="refresh" :size="14" /></button>
    </div>

    <div class="ms-toolbar">
      <button class="btn-primary" @click="openUpload">
        <MsIcon name="plus" :size="14" /> Загрузить
      </button>
      <button class="btn-default" :class="{ active: showFilter }" @click="showFilter = !showFilter">Фильтр</button>
      <input v-model="search" class="ms-input search" placeholder="Наименование" />
      <div class="ms-counter">{{ filtered.length }}</div>
    </div>

    <div v-if="showFilter" class="ms-filter-panel">
      <div class="filter-actions">
        <button class="btn-green" @click="page = 1">Найти</button>
        <button class="btn-default" @click="clearFilters">Очистить</button>
      </div>
      <div class="filter-grid">
        <div class="filter-field">
          <label><span class="dot"></span>Расположение</label>
          <select v-model="filters.category" class="ms-select">
            <option value="">Все</option>
            <option v-for="c in CATEGORIES" :key="c" :value="c">{{ c }}</option>
          </select>
        </div>
      </div>
    </div>

    <div v-if="error" class="error-box">{{ error }}</div>

    <div class="table-wrap">
      <table class="ms-table">
        <thead>
          <tr>
            <th>Наименование</th>
            <th class="col-num">Размер, МБ</th>
            <th>Расположение</th>
            <th>Дата добавления</th>
            <th>Сотрудник</th>
            <th class="col-actions"></th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="loading"><td colspan="6" class="empty">Загрузка...</td></tr>
          <tr v-else-if="filtered.length === 0"><td colspan="6" class="empty">Нет файлов</td></tr>
          <tr v-else v-for="f in paginated" :key="f.name" class="row">
            <td><a class="link" :href="downloadUrl(f.name)" target="_blank">{{ f.name }}</a></td>
            <td class="col-num">{{ (f.size_bytes / 1024 / 1024).toFixed(2) }}</td>
            <td>{{ f.category || '—' }}</td>
            <td>{{ formatDate(f.created_at) }}</td>
            <td>{{ f.uploaded_by_name || '—' }}</td>
            <td class="col-actions">
              <button class="row-menu" @click.stop="openCategory(f)" title="Изменить расположение"><MsIcon name="edit" :size="14" /></button>
              <button class="row-menu" @click.stop="onDelete(f)" title="Удалить"><MsIcon name="trash" :size="14" /></button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="ms-footer2">
      <div class="pager">
        <button :disabled="page === 1" @click="page = 1">«</button>
        <button :disabled="page === 1" @click="page--">‹</button>
        <span class="range">{{ rangeFrom }}-{{ rangeTo }} из {{ filtered.length }}</span>
        <button :disabled="rangeTo >= filtered.length" @click="page++">›</button>
        <button :disabled="rangeTo >= filtered.length" @click="page = Math.ceil(filtered.length / perPage)">»</button>
      </div>
    </div>

    <div v-if="showUpload" class="ms-modal-overlay" @click.self="showUpload = false">
      <div class="ms-modal">
        <div class="ms-modal-header">
          <h2>Загрузить файл</h2>
          <button class="close-btn" @click="showUpload = false">×</button>
        </div>
        <div class="ms-modal-body">
          <div v-if="uploadError" class="error-box">{{ uploadError }}</div>
          <div class="form-group">
            <label>Файл (.xlsx)</label>
            <input ref="fileInput" type="file" accept=".xlsx,.xls" @change="onFileChange" />
          </div>
          <div class="form-group">
            <label>Расположение</label>
            <select v-model="uploadCategory" class="ms-select">
              <option v-for="c in CATEGORIES" :key="c" :value="c">{{ c }}</option>
            </select>
          </div>
        </div>
        <div class="ms-modal-footer">
          <button class="btn-cancel" @click="showUpload = false">Отмена</button>
          <button class="btn-save" :disabled="uploading" @click="doUpload">Загрузить</button>
        </div>
      </div>
    </div>

    <div v-if="showCategory" class="ms-modal-overlay" @click.self="showCategory = false">
      <div class="ms-modal">
        <div class="ms-modal-header">
          <h2>Расположение файла</h2>
          <button class="close-btn" @click="showCategory = false">×</button>
        </div>
        <div class="ms-modal-body">
          <div class="form-group">
            <label>Файл</label>
            <div>{{ categoryTarget?.name }}</div>
          </div>
          <div class="form-group">
            <label>Расположение</label>
            <select v-model="categoryValue" class="ms-select">
              <option v-for="c in CATEGORIES" :key="c" :value="c">{{ c }}</option>
            </select>
          </div>
        </div>
        <div class="ms-modal-footer">
          <button class="btn-cancel" @click="showCategory = false">Отмена</button>
          <button class="btn-save" :disabled="categorySaving" @click="saveCategory">Сохранить</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { listFiles, uploadFile, deleteFile, updateFileCategory, downloadFileUrl, type FileMeta } from '../api/files'
import { apiErrorMessage } from '../api/client'
import MsIcon from '../components/MsIcon.vue'

const CATEGORIES = ['Отгрузки', 'Перемещения', 'Инвентаризации', 'Внутренние заказы', 'Приёмки', 'Списания', 'Другое']

const items = ref<FileMeta[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const search = ref('')
const showFilter = ref(false)
const filters = reactive({ category: '' })
const page = ref(1)
const perPage = 100

const showUpload = ref(false)
const uploadCategory = ref<string>('Другое')
const uploadFileObj = ref<File | null>(null)
const uploading = ref(false)
const uploadError = ref<string | null>(null)
const fileInput = ref<HTMLInputElement | null>(null)

const showCategory = ref(false)
const categoryTarget = ref<FileMeta | null>(null)
const categoryValue = ref<string>('Другое')
const categorySaving = ref(false)

const filtered = computed(() => {
  let rows = [...items.value]
  if (search.value) {
    const q = search.value.toLowerCase()
    rows = rows.filter((f) => f.name.toLowerCase().includes(q))
  }
  if (filters.category) rows = rows.filter((f) => f.category === filters.category)
  return rows.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
})

const paginated = computed(() => {
  const from = (page.value - 1) * perPage
  return filtered.value.slice(from, from + perPage)
})

const rangeFrom = computed(() => filtered.value.length === 0 ? 0 : (page.value - 1) * perPage + 1)
const rangeTo = computed(() => Math.min(page.value * perPage, filtered.value.length))

function clearFilters() {
  filters.category = ''
  search.value = ''
  page.value = 1
}

async function load() {
  loading.value = true
  error.value = null
  try {
    items.value = await listFiles()
  } catch (e) {
    error.value = apiErrorMessage(e)
  } finally {
    loading.value = false
  }
}

function openUpload() {
  showUpload.value = true
  uploadError.value = null
  uploadFileObj.value = null
  uploadCategory.value = 'Другое'
  if (fileInput.value) fileInput.value.value = ''
}

function onFileChange(e: Event) {
  const f = (e.target as HTMLInputElement).files?.[0] ?? null
  uploadFileObj.value = f
}

async function doUpload() {
  uploadError.value = null
  if (!uploadFileObj.value) {
    uploadError.value = 'Выберите файл'
    return
  }
  uploading.value = true
  try {
    await uploadFile(uploadFileObj.value, uploadCategory.value)
    showUpload.value = false
    await load()
  } catch (e: any) {
    uploadError.value = e?.response?.data?.message || e.message || 'Ошибка загрузки'
  } finally {
    uploading.value = false
  }
}

function openCategory(f: FileMeta) {
  categoryTarget.value = f
  categoryValue.value = f.category || 'Другое'
  showCategory.value = true
}

async function saveCategory() {
  if (!categoryTarget.value) return
  categorySaving.value = true
  try {
    await updateFileCategory(categoryTarget.value.name, categoryValue.value)
    showCategory.value = false
    await load()
  } catch (e: any) {
    alert('Не удалось сохранить: ' + (e?.response?.data?.message || e.message || 'ошибка'))
  } finally {
    categorySaving.value = false
  }
}

async function onDelete(f: FileMeta) {
  if (!confirm(`Удалить файл ${f.name}?`)) return
  try {
    await deleteFile(f.name)
    await load()
  } catch (e: any) {
    alert('Не удалось удалить: ' + (e?.response?.data?.message || e.message || 'ошибка'))
  }
}

function downloadUrl(name: string) { return downloadFileUrl(name) }

function formatDate(s: string) {
  const d = new Date(s)
  return d.toLocaleDateString('ru-RU') + ' ' + d.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })
}

onMounted(load)
</script>

<style scoped>
.page { font-size: 13px; color: #1f2328; padding: 8px 0 32px; }
.ms-title { display: flex; align-items: center; gap: 8px; font-size: 20px; font-weight: 600; margin-bottom: 16px; }
.title-text { line-height: 1; }
.ms-refresh { width: 24px; height: 24px; padding: 0; display: inline-flex; align-items: center; justify-content: center; border: none; background: transparent; color: #57606a; cursor: pointer; }
.ms-refresh:hover { color: #2c5d9c; }
.ms-toolbar { display: flex; align-items: center; gap: 8px; margin-bottom: 12px; flex-wrap: wrap; }
.btn-primary { display: inline-flex; align-items: center; gap: 6px; background: #2da44e; color: #fff; border: 1px solid #2a8e44; border-radius: 4px; padding: 6px 12px; font-size: 13px; cursor: pointer; }
.btn-primary:hover { background: #2c974b; }
.btn-default { background: #f6f8fa; border: 1px solid #d0d7de; border-radius: 4px; padding: 6px 12px; font-size: 13px; cursor: pointer; color: #1f2328; }
.btn-default:hover { background: #eaeef2; }
.btn-default.active { background: #eaf3ff; border-color: #b6ccec; color: #0969da; }
.btn-green { background: #b3d261; border: 1px solid #a5c655; border-radius: 4px; padding: 6px 12px; font-size: 13px; cursor: pointer; color: #fff; font-weight: 500; }
.btn-green:hover { background: #a3c351; }
.ms-input { height: 30px; padding: 0 10px; border: 1px solid #d0d7de; border-radius: 4px; font-size: 13px; }
.ms-input.search { width: 240px; }
.ms-select { height: 30px; padding: 0 10px; border: 1px solid #d0d7de; border-radius: 4px; font-size: 13px; }
.ms-counter { background: #f6f8fa; border: 1px solid #d0d7de; border-radius: 4px; padding: 0 10px; height: 30px; display: inline-flex; align-items: center; justify-content: center; font-size: 13px; color: #8c959f; min-width: 40px; }
.ms-filter-panel { background: #eef1f5; border: 1px solid #d8dee4; border-radius: 4px; padding: 12px 16px; margin-bottom: 12px; }
.filter-actions { display: flex; align-items: center; gap: 8px; margin-bottom: 12px; }
.filter-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 12px 16px; }
.filter-field { display: flex; flex-direction: column; gap: 4px; }
.filter-field label { font-size: 12px; color: #57606a; display: flex; align-items: center; gap: 4px; }
.filter-field .dot { width: 6px; height: 6px; border-radius: 50%; background: #8c959f; }
.table-wrap { background: #fff; border: 1px solid #eaeef2; border-radius: 4px; overflow-x: auto; }
.ms-table { width: 100%; border-collapse: collapse; font-size: 13px; }
.ms-table th { background: #fff; color: #2c5d9c; font-weight: 500; padding: 8px 10px; text-align: left; border-bottom: 1px solid #d8dee4; white-space: nowrap; font-size: 12px; }
.ms-table td { padding: 8px 10px; border-bottom: 1px solid #eaeef2; vertical-align: middle; }
.ms-table tr.row:hover { background: #f6f8fa; }
.ms-table .link { color: #2c5d9c; text-decoration: none; }
.ms-table .link:hover { text-decoration: underline; }
.ms-table .col-num { text-align: right; font-variant-numeric: tabular-nums; }
.ms-table .col-actions { width: 40px; }
.ms-table .empty { text-align: center; padding: 24px; color: #8c959f; }
.row-menu { width: 22px; height: 22px; padding: 0; border: none; background: transparent; color: #57606a; cursor: pointer; border-radius: 3px; }
.row-menu:hover { background: #eaeef2; color: #cf222e; }
.error-box { background: #ffeef0; color: #cf222e; padding: 10px; border-radius: 4px; margin-bottom: 12px; }
.ms-footer2 { display: flex; align-items: center; padding: 8px 10px; font-size: 12px; color: #57606a; background: #fff; border-top: 1px solid #eaeef2; margin-top: 8px; }
.pager { display: flex; align-items: center; gap: 4px; }
.pager button { width: 24px; height: 24px; padding: 0; border: 1px solid #d0d7de; background: #fff; border-radius: 3px; cursor: pointer; font-size: 12px; color: #1f2328; display: inline-flex; align-items: center; justify-content: center; }
.pager button:disabled { opacity: 0.4; cursor: default; }
.pager button:hover:not(:disabled) { background: #f6f8fa; }
.pager .range { margin: 0 8px; font-variant-numeric: tabular-nums; }
.ms-modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.4); display: flex; align-items: center; justify-content: center; z-index: 1000; }
.ms-modal { background: #fff; border-radius: 6px; width: 480px; max-width: 100%; box-shadow: 0 10px 30px rgba(0,0,0,0.15); display: flex; flex-direction: column; }
.ms-modal-header { padding: 16px 20px; border-bottom: 1px solid #eaeef2; display: flex; justify-content: space-between; align-items: center; }
.ms-modal-header h2 { margin: 0; font-size: 16px; }
.close-btn { background: none; border: none; font-size: 22px; color: #8c959f; cursor: pointer; }
.ms-modal-body { padding: 20px; display: flex; flex-direction: column; gap: 14px; }
.form-group { display: flex; flex-direction: column; gap: 6px; }
.form-group label { font-size: 12px; color: #57606a; font-weight: 500; }
.ms-modal-footer { padding: 12px 20px; border-top: 1px solid #eaeef2; display: flex; justify-content: flex-end; gap: 8px; background: #f6f8fa; }
.btn-cancel { background: #fff; border: 1px solid #d0d7de; border-radius: 4px; padding: 6px 14px; font-size: 13px; cursor: pointer; }
.btn-save { background: #2da44e; color: #fff; border: 1px solid #2a8e44; border-radius: 4px; padding: 6px 14px; font-size: 13px; cursor: pointer; }
.btn-save:disabled { opacity: 0.6; cursor: default; }
</style>