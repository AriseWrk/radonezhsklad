<template>
  <div>
    <div class="page-title-bar">
      <div class="page-title">
        <span>Аудит</span>
        <span class="refresh" @click="load" title="Обновить">↻</span>
      </div>
      <div class="page-actions">
        <button class="btn" @click="showFilter = !showFilter">Фильтр</button>
      </div>
    </div>

    <div v-if="showFilter" class="filter-panel">
      <div class="filter-row">
        <div class="filter-actions">
          <button class="btn-find" @click="load">Найти</button>
          <button class="btn-clear" @click="clearFilters">Очистить</button>
        </div>
        <div class="filter-field">
          <label>Метод</label>
          <select v-model="filters.method">
            <option value="">Все</option>
            <option value="POST">POST</option>
            <option value="PUT">PUT</option>
            <option value="PATCH">PATCH</option>
            <option value="DELETE">DELETE</option>
          </select>
        </div>
        <div class="filter-field">
          <label>Ресурс</label>
          <input v-model="filters.resource" placeholder="products, orders, warehouses..." />
        </div>
        <div class="filter-field">
          <label>Период с</label>
          <input v-model="filters.dateFrom" type="date" />
        </div>
        <div class="filter-field">
          <label>Период по</label>
          <input v-model="filters.dateTo" type="date" />
        </div>
      </div>
    </div>

    <div v-if="error" class="error-box">{{ error }}</div>

    <table class="ms-table">
      <thead>
        <tr>
          <th style="width:160px">Время</th>
          <th style="width:80px">Метод</th>
          <th>Путь</th>
          <th style="width:120px">Пользователь</th>
          <th style="width:80px">Статус</th>
          <th style="width:140px">IP</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="loading">
          <td colspan="7" class="muted" style="text-align:center;padding:24px">Загрузка...</td>
        </tr>
        <tr v-else-if="items.length === 0">
          <td colspan="7" class="muted" style="text-align:center;padding:24px">Нет событий</td>
        </tr>
        <tr v-else v-for="l in items" :key="l.id">
          <td class="muted">{{ formatDate(l.created_at) }}</td>
          <td><span :class="['method-badge', 'method-' + l.method.toLowerCase()]">{{ l.method }}</span></td>
          <td class="mono">{{ l.path }}</td>
          <td class="mono muted">{{ (l.user_id || '—').slice(0, 8) }}</td>
          <td>
            <span :class="['status-badge', l.status < 300 ? 'ok' : 'err']">{{ l.status }}</span>
          </td>
          <td class="muted mono">{{ l.client_ip || '—' }}</td>
          <td class="actions-col">
            <button class="btn-link" @click="showDetail(l)">Детали</button>
          </td>
        </tr>
      </tbody>
    </table>

    <div class="ms-footer">
      <div class="ms-pager">
        <button :disabled="filters.offset === 0" @click="prevPage">◀</button>
        <span>{{ rangeFrom }}–{{ rangeTo }} из {{ total }}</span>
        <button :disabled="rangeTo >= total" @click="nextPage">▶</button>
      </div>
    </div>

    <!-- Модалка деталей -->
    <div v-if="detail" class="modal-backdrop" @click.self="detail = null">
      <div class="card modal big">
        <h2>Событие аудита</h2>
        <div class="doc-info">
          <div><span class="lbl">Время:</span> {{ formatDate(detail.created_at) }}</div>
          <div><span class="lbl">Метод:</span> {{ detail.method }}</div>
          <div><span class="lbl">Путь:</span> {{ detail.path }}</div>
          <div><span class="lbl">Статус:</span> {{ detail.status }}</div>
          <div><span class="lbl">Пользователь:</span> {{ detail.user_id || '—' }}</div>
          <div><span class="lbl">IP:</span> {{ detail.client_ip || '—' }}</div>
          <div><span class="lbl">Request ID:</span> {{ detail.request_id || '—' }}</div>
        </div>
        <div v-if="detail.request_body">
          <div class="lbl" style="margin-bottom:6px">Тело запроса:</div>
          <pre class="code-block">{{ prettyBody(detail.request_body) }}</pre>
        </div>
        <div class="modal-actions">
          <button @click="detail = null">Закрыть</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { listAudit, type AuditLog } from '../api/audit'
import { apiErrorMessage } from '../api/client'

const items = ref<AuditLog[]>([])
const total = ref(0)
const loading = ref(false)
const error = ref<string | null>(null)
const showFilter = ref(false)
const detail = ref<AuditLog | null>(null)

const filters = reactive({
  method: '',
  resource: '',
  dateFrom: '',
  dateTo: '',
  offset: 0,
  limit: 100,
})

function formatDate(s: string) {
  const d = new Date(s)
  return d.toLocaleDateString('ru-RU') + ' ' + d.toLocaleTimeString('ru-RU')
}

function prettyBody(s: string) {
  try { return JSON.stringify(JSON.parse(s), null, 2) } catch { return s }
}

const rangeFrom = computed(() => total.value === 0 ? 0 : filters.offset + 1)
const rangeTo = computed(() => Math.min(filters.offset + filters.limit, total.value))

async function load() {
  loading.value = true
  error.value = null
  try {
    const resp = await listAudit({
      method: filters.method || undefined,
      resource: filters.resource || undefined,
      date_from: filters.dateFrom || undefined,
      date_to: filters.dateTo || undefined,
      limit: filters.limit,
      offset: filters.offset,
    })
    items.value = resp.items
    total.value = resp.total
  } catch (e) {
    error.value = apiErrorMessage(e)
  } finally {
    loading.value = false
  }
}

function clearFilters() {
  filters.method = ''
  filters.resource = ''
  filters.dateFrom = ''
  filters.dateTo = ''
  filters.offset = 0
  load()
}

function prevPage() {
  if (filters.offset === 0) return
  filters.offset = Math.max(0, filters.offset - filters.limit)
  load()
}

function nextPage() {
  if (rangeTo.value >= total.value) return
  filters.offset += filters.limit
  load()
}

function showDetail(l: AuditLog) { detail.value = l }

onMounted(load)
</script>

<style scoped>
.mono { font-family: monospace; font-size: 12px; }

.method-badge {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 3px;
  font-size: 11px;
  font-weight: 600;
  font-family: monospace;
}
.method-post   { background: #e8f5e9; color: #1a7f37; }
.method-put    { background: #fff8e1; color: #9a6a00; }
.method-patch  { background: #f3e5f5; color: #6a1b9a; }
.method-delete { background: #ffebee; color: #cf222e; }

.status-badge {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 10px;
  font-size: 11px;
  font-weight: 600;
}
.status-badge.ok  { background: #e8f5e9; color: #1a7f37; }
.status-badge.err { background: #ffebee; color: #cf222e; }

.code-block {
  background: #f6f8fa;
  border: 1px solid #d8dee4;
  border-radius: 4px;
  padding: 10px;
  font-size: 12px;
  font-family: monospace;
  max-height: 300px;
  overflow: auto;
  margin: 0;
}
</style>