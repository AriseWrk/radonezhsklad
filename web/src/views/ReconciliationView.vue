<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import MsButton from '../components/MsButton.vue'
import {
  apiErrorMessage,
} from '../api/client'
import {
  exportReconciliation,
  getActiveReconciliationRun,
  getReconciliationRun,
  listReconciliationItems,
  listReconciliationRuns,
  startReconciliation,
  type ReconciliationItem,
  type ReconciliationRun,
} from '../api/reconciliation'

const runs = ref<ReconciliationRun[]>([])
const selectedRunId = ref('')
const items = ref<ReconciliationItem[]>([])
const totalItems = ref(0)
const page = ref(1)
const pageSize = 100

const loading = ref(false)
const loadingItems = ref(false)
const starting = ref(false)
const exporting = ref(false)
const error = ref('')
const notice = ref('')

const filterType = ref('')
const filterStatus = ref('')
const searchInput = ref('')
const appliedSearch = ref('')

let pollTimer: number | undefined

const selectedRun = computed(() =>
  runs.value.find((run) => run.id === selectedRunId.value) ?? null,
)

const totalPages = computed(() =>
  Math.max(1, Math.ceil(totalItems.value / pageSize)),
)

const isRunning = computed(() =>
  runs.value.some((run) => run.status === 'running'),
)

function formatDate(value?: string | null): string {
  if (!value) return '—'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '—'
  return new Intl.DateTimeFormat('ru-RU', {
    dateStyle: 'short',
    timeStyle: 'medium',
  }).format(date)
}

function formatNumber(value?: number | null): string {
  if (value === null || value === undefined) return '—'
  return new Intl.NumberFormat('ru-RU', {
    maximumFractionDigits: 4,
  }).format(value)
}

function formatMoney(value?: number | null): string {
  if (value === null || value === undefined) return '—'
  return new Intl.NumberFormat('ru-RU', {
    style: 'currency',
    currency: 'RUB',
    maximumFractionDigits: 2,
  }).format(value)
}

function statusLabel(status: string): string {
  const labels: Record<string, string> = {
    running: 'Выполняется',
    completed: 'Завершена',
    failed: 'Ошибка',
    matched: 'Совпадает',
    difference: 'Расхождение',
    missing_local: 'Нет локально',
    missing_moysklad: 'Нет в МойСклад',
    ambiguous: 'Неоднозначно',
    unmatched: 'Не сопоставлено',
  }
  return labels[status] ?? status
}

function typeLabel(type: string): string {
  const labels: Record<string, string> = {
    product: 'Товар',
    price: 'Цена',
    stock: 'Остаток',
    warehouse: 'Склад',
  }
  return labels[type] ?? type
}

function statusClass(status: string): string {
  if (status === 'completed' || status === 'matched') return 'success'
  if (status === 'running') return 'running'
  if (status === 'failed' || status === 'difference') return 'danger'
  return 'warning'
}

async function loadRuns(preferActive = false): Promise<void> {
  const [history, active] = await Promise.all([
    listReconciliationRuns(50),
    getActiveReconciliationRun(),
  ])

  runs.value = history

  if (active) {
    const index = runs.value.findIndex((run) => run.id === active.id)
    if (index >= 0) runs.value[index] = active
    else runs.value.unshift(active)
  }

  if (preferActive && active) {
    selectedRunId.value = active.id
  } else if (!selectedRunId.value && runs.value.length) {
    selectedRunId.value = runs.value[0].id
  }
}

async function loadItems(): Promise<void> {
  if (!selectedRunId.value) {
    items.value = []
    totalItems.value = 0
    return
  }

  loadingItems.value = true
  error.value = ''

  try {
    const result = await listReconciliationItems(selectedRunId.value, {
      type: filterType.value || undefined,
      status: filterStatus.value || undefined,
      search: appliedSearch.value || undefined,
      limit: pageSize,
      offset: (page.value - 1) * pageSize,
    })
    items.value = result.items ?? []
    totalItems.value = result.total ?? 0
  } catch (err) {
    error.value = apiErrorMessage(err)
  } finally {
    loadingItems.value = false
  }
}

async function selectRun(id: string): Promise<void> {
  selectedRunId.value = id
  page.value = 1
  await loadItems()
}

async function refresh(): Promise<void> {
  loading.value = true
  error.value = ''

  try {
    await loadRuns()
    if (selectedRunId.value) {
      const run = await getReconciliationRun(selectedRunId.value)
      const index = runs.value.findIndex((item) => item.id === run.id)
      if (index >= 0) runs.value[index] = run
      else runs.value.unshift(run)
      if (run.status !== 'running') await loadItems()
    }
  } catch (err) {
    error.value = apiErrorMessage(err)
  } finally {
    loading.value = false
  }
}

async function poll(): Promise<void> {
  try {
    const active = await getActiveReconciliationRun()

    if (active) {
      const index = runs.value.findIndex((run) => run.id === active.id)
      if (index >= 0) runs.value[index] = active
      else runs.value.unshift(active)

      if (selectedRunId.value !== active.id) {
        selectedRunId.value = active.id
        page.value = 1
      }
    } else if (selectedRunId.value) {
      const current = await getReconciliationRun(selectedRunId.value)
      const index = runs.value.findIndex((run) => run.id === current.id)
      if (index >= 0) runs.value[index] = current

      if (current.status !== 'running') {
        await loadItems()
        await loadRuns()
      }
    }

    if (!active) {
      const history = await listReconciliationRuns(50)
      runs.value = history
    }
  } catch {
    // Следующий цикл повторит запрос; не прерываем работу экрана.
  }
}

async function startRun(): Promise<void> {
  if (starting.value || isRunning.value) return

  starting.value = true
  error.value = ''
  notice.value = ''

  try {
    const run = await startReconciliation()
    runs.value.unshift(run)
    selectedRunId.value = run.id
    items.value = []
    totalItems.value = 0
    page.value = 1
    notice.value = 'Сверка запущена. Результаты появятся после завершения.'
  } catch (err) {
    error.value = apiErrorMessage(err)
    await refresh()
  } finally {
    starting.value = false
  }
}

async function applyFilters(): Promise<void> {
  appliedSearch.value = searchInput.value.trim()
  page.value = 1
  await loadItems()
}

async function changePage(nextPage: number): Promise<void> {
  if (nextPage < 1 || nextPage > totalPages.value) return
  page.value = nextPage
  await loadItems()
}

async function downloadExcel(): Promise<void> {
  if (!selectedRunId.value || exporting.value) return

  exporting.value = true
  error.value = ''

  try {
    const blob = await exportReconciliation(selectedRunId.value)
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `reconciliation-${selectedRunId.value}.xlsx`
    document.body.appendChild(link)
    link.click()
    link.remove()
    URL.revokeObjectURL(url)
  } catch (err) {
    error.value = apiErrorMessage(err)
  } finally {
    exporting.value = false
  }
}
onMounted(async () => {
  await refresh()
  pollTimer = window.setInterval(poll, 5000)
})

onUnmounted(() => {
  if (pollTimer !== undefined) window.clearInterval(pollTimer)
})
</script>

<template>
  <div class="reconciliation-page">
    <div class="page-title-bar">
      <div>
        <div class="page-title">Сверка с МойСклад</div>
        <div class="page-subtitle">
          Диагностическая проверка цен продажи и остатков. Данные системы не изменяются.
        </div>
      </div>
      <div class="page-actions">
        <MsButton variant="default" icon="refresh" :disabled="loading" @click="refresh">
          Обновить
        </MsButton>
        <MsButton
          variant="primary"
          icon="play"
          :disabled="starting || isRunning"
          @click="startRun"
        >
          {{ starting ? 'Запуск...' : isRunning ? 'Сверка выполняется' : 'Запустить сверку' }}
        </MsButton>
      </div>
    </div>

    <div v-if="error" class="error-box">{{ error }}</div>
    <div v-if="notice" class="notice-box">{{ notice }}</div>

    <section v-if="selectedRun" class="run-summary">
      <div class="summary-heading">
        <div>
          <strong>Последний выбранный запуск</strong>
          <span class="run-id">{{ selectedRun.id }}</span>
        </div>
        <span class="status-pill" :class="statusClass(selectedRun.status)">
          {{ statusLabel(selectedRun.status) }}
        </span>
      </div>

      <div class="run-meta">
        <span>Начало: {{ formatDate(selectedRun.started_at) }}</span>
        <span>Завершение: {{ formatDate(selectedRun.finished_at) }}</span>
        <span>Запустил: {{ selectedRun.triggered_by_name || '—' }}</span>
        <span>Тип цены: {{ selectedRun.price_type_name }}</span>
      </div>

      <div v-if="selectedRun.status === 'running'" class="running-note">
        Сверка выполняется в фоновом режиме. Статус обновляется автоматически.
      </div>
      <div v-if="selectedRun.error" class="error-box run-error">
        {{ selectedRun.error }}
      </div>

      <div class="summary-grid">
        <div class="summary-card">
          <span>Товаров проверено</span>
          <strong>{{ formatNumber(selectedRun.total_products) }}</strong>
        </div>
        <div class="summary-card">
          <span>Совпадений</span>
          <strong>{{ formatNumber(selectedRun.matched_products) }}</strong>
        </div>
        <div class="summary-card">
          <span>Расхождения цен</span>
          <strong>{{ formatNumber(selectedRun.price_differences) }}</strong>
        </div>
        <div class="summary-card">
          <span>Расхождения остатков</span>
          <strong>{{ formatNumber(selectedRun.stock_differences) }}</strong>
        </div>
        <div class="summary-card">
          <span>Нет локально</span>
          <strong>{{ formatNumber(selectedRun.missing_local_products) }}</strong>
        </div>
        <div class="summary-card">
          <span>Нет в МойСклад</span>
          <strong>{{ formatNumber(selectedRun.missing_ms_products) }}</strong>
        </div>
        <div class="summary-card">
          <span>Неоднозначные товары</span>
          <strong>{{ formatNumber(selectedRun.ambiguous_products) }}</strong>
        </div>
        <div class="summary-card">
          <span>Несопоставленные склады</span>
          <strong>{{ formatNumber(selectedRun.unmatched_warehouses) }}</strong>
        </div>
        <div class="summary-card">
          <span>Нулевые остатки</span>
          <strong>{{ formatNumber(selectedRun.zero_stock_items) }}</strong>
        </div>
      </div>
    </section>

    <section class="history-panel">
      <div class="section-heading">
        <strong>История запусков</strong>
        <span>{{ runs.length }} из последних 50</span>
      </div>

      <div v-if="runs.length" class="history-list">
        <button
          v-for="run in runs"
          :key="run.id"
          type="button"
          class="history-row"
          :class="{ selected: selectedRunId === run.id }"
          @click="selectRun(run.id)"
        >
          <span class="history-date">{{ formatDate(run.started_at) }}</span>
          <span class="history-user">{{ run.triggered_by_name || '—' }}</span>
          <span class="status-pill" :class="statusClass(run.status)">
            {{ statusLabel(run.status) }}
          </span>
          <span class="history-count">{{ run.total_products }} товаров</span>
        </button>
      </div>
      <div v-else class="empty-state">Запусков пока нет. Нажмите «Запустить сверку».</div>
    </section>

    <section class="items-panel">
      <div class="section-heading">
        <strong>Результаты сверки</strong>
        <div class="section-actions">
          <span>{{ totalItems }} строк</span>
          <MsButton
            variant="default"
            icon="download"
            size="sm"
            :disabled="!selectedRun || selectedRun.status === 'running' || exporting"
            @click="downloadExcel"
          >
            {{ exporting ? 'Подготовка...' : 'Экспорт Excel' }}
          </MsButton>
        </div>
      </div>

      <div class="filters">
        <label>
          Тип
          <select v-model="filterType">
            <option value="">Все типы</option>
            <option value="product">Товары</option>
            <option value="price">Цены</option>
            <option value="stock">Остатки</option>
            <option value="warehouse">Склады</option>
          </select>
        </label>

        <label>
          Результат
          <select v-model="filterStatus">
            <option value="">Все результаты</option>
            <option value="matched">Совпадает</option>
            <option value="difference">Расхождение</option>
            <option value="missing_local">Нет локально</option>
            <option value="missing_moysklad">Нет в МойСклад</option>
            <option value="ambiguous">Неоднозначно</option>
            <option value="unmatched">Не сопоставлено</option>
          </select>
        </label>

        <label class="search-field">
          Поиск
          <input
            v-model="searchInput"
            type="search"
            placeholder="Название товара, артикул или склад"
            @keyup.enter="applyFilters"
          />
        </label>

        <MsButton variant="primary" size="md" @click="applyFilters">
          Применить
        </MsButton>
      </div>

      <div class="table-wrap">
        <table class="reconciliation-table">
          <thead>
            <tr>
              <th>Тип</th>
              <th>Результат</th>
              <th>Артикул</th>
              <th>Локальный товар / склад</th>
              <th>Товар / склад МойСклад</th>
              <th>Цена локально</th>
              <th>Цена МойСклад</th>
              <th>Остаток локально</th>
              <th>Остаток МойСклад</th>
              <th>Нулевой остаток</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loadingItems">
              <td colspan="10" class="table-message">Загрузка результатов...</td>
            </tr>
            <tr v-else-if="!items.length">
              <td colspan="10" class="table-message">
                {{ selectedRun?.status === 'running'
                  ? 'Результаты появятся после завершения сверки.'
                  : 'По выбранным фильтрам строк не найдено.' }}
              </td>
            </tr>
            <tr v-for="item in items" :key="item.id">
              <td>{{ typeLabel(item.item_type) }}</td>
              <td>
                <span class="status-pill" :class="statusClass(item.result_status)">
                  {{ statusLabel(item.result_status) }}
                </span>
              </td>
              <td class="mono">{{ item.sku || '—' }}</td>
              <td>
                <div>{{ item.local_product_name || item.local_warehouse_name || '—' }}</div>
                <small v-if="item.local_warehouse_name && item.local_product_name">
                  {{ item.local_warehouse_name }}
                </small>
              </td>
              <td>
                <div>{{ item.ms_product_name || item.ms_store_name || '—' }}</div>
                <small v-if="item.ms_store_name && item.ms_product_name">
                  {{ item.ms_store_name }}
                </small>
              </td>
              <td class="numeric">{{ formatMoney(item.local_price) }}</td>
              <td class="numeric">{{ formatMoney(item.ms_price) }}</td>
              <td class="numeric">{{ formatNumber(item.local_stock) }}</td>
              <td class="numeric">{{ formatNumber(item.ms_stock) }}</td>
              <td>
                <span v-if="item.is_zero_stock" class="zero-mark">Да</span>
                <span v-else>—</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="table-footer">
        <span>
          Показано {{ items.length }} из {{ totalItems }}
        </span>
        <div class="pagination">
          <MsButton
            size="sm"
            :disabled="page <= 1 || loadingItems"
            @click="changePage(page - 1)"
          >
            Назад
          </MsButton>
          <span>Страница {{ page }} из {{ totalPages }}</span>
          <MsButton
            size="sm"
            :disabled="page >= totalPages || loadingItems"
            @click="changePage(page + 1)"
          >
            Далее
          </MsButton>
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped>
.reconciliation-page {
  padding: 4px 0 28px;
  min-width: 0;
}
.page-title-bar {
  align-items: flex-start;
  margin-bottom: 18px;
}
.page-subtitle {
  color: var(--muted);
  font-size: 13px;
  margin-top: 6px;
}
.page-actions, .section-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}
.run-summary, .history-panel, .items-panel {
  background: var(--panel);
  border: 1px solid var(--border);
  border-radius: 8px;
  margin-bottom: 16px;
  padding: 16px;
}
.summary-heading, .section-heading {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  margin-bottom: 14px;
}
.run-id {
  display: block;
  color: var(--muted);
  font: 11px monospace;
  margin-top: 5px;
  overflow-wrap: anywhere;
}
.run-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 22px;
  color: var(--muted);
  font-size: 12px;
  margin-bottom: 14px;
}
.summary-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}
.summary-card {
  background: #f8f9fb;
  border: 1px solid #e8ebef;
  border-radius: 6px;
  padding: 12px;
  min-width: 0;
}
.summary-card span {
  display: block;
  color: var(--muted);
  font-size: 12px;
  margin-bottom: 7px;
}
.summary-card strong {
  font-size: 20px;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
}
.status-pill {
  display: inline-flex;
  align-items: center;
  border-radius: 12px;
  padding: 3px 9px;
  font-size: 11px;
  white-space: nowrap;
  background: #f1f2f4;
  color: #57606a;
}
.status-pill.success { background: #dafbe1; color: #116329; }
.status-pill.running { background: #ddf4ff; color: #0550ae; }
.status-pill.danger { background: #ffebe9; color: #cf222e; }
.status-pill.warning { background: #fff8c5; color: #7d4e00; }
.running-note, .notice-box {
  background: #ddf4ff;
  color: #0550ae;
  border-radius: 5px;
  padding: 9px 12px;
  font-size: 13px;
  margin-bottom: 12px;
}
.notice-box { margin-bottom: 12px; }
.run-error { margin-top: 10px; }
.history-list {
  display: flex;
  flex-direction: column;
}
.history-row {
  display: grid;
  grid-template-columns: minmax(150px, 1fr) minmax(100px, 1fr) auto auto;
  align-items: center;
  gap: 12px;
  text-align: left;
  width: 100%;
  border: 0;
  border-top: 1px solid var(--border);
  border-radius: 0;
  padding: 10px 8px;
  background: transparent;
  font-size: 12px;
}
.history-row:hover, .history-row.selected { background: #f3f6fa; }
.history-date { font-variant-numeric: tabular-nums; }
.history-user, .history-count { color: var(--muted); }
.empty-state, .table-message {
  text-align: center;
  color: var(--muted);
  padding: 28px 12px;
  font-size: 13px;
}
.filters {
  display: flex;
  align-items: flex-end;
  flex-wrap: wrap;
  gap: 10px;
  margin-bottom: 14px;
}
.filters label {
  display: flex;
  flex-direction: column;
  gap: 5px;
  min-width: 150px;
  color: var(--muted);
  font-size: 12px;
}
.filters .search-field { flex: 1; min-width: 220px; }
.filters input, .filters select {
  height: 32px;
  padding: 5px 9px;
  font-size: 12px;
}
.table-wrap {
  width: 100%;
  overflow-x: auto;
}
.reconciliation-table {
  min-width: 1120px;
  border-radius: 5px;
}
.reconciliation-table th, .reconciliation-table td {
  padding: 9px 10px;
  font-size: 12px;
  vertical-align: middle;
}
.reconciliation-table th { white-space: nowrap; }
.reconciliation-table small {
  display: block;
  color: var(--muted);
  margin-top: 3px;
}
.numeric { text-align: right !important; white-space: nowrap; }
.mono { font-family: monospace; }
.zero-mark { color: var(--danger); font-weight: 600; }
.table-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  margin-top: 12px;
  color: var(--muted);
  font-size: 12px;
}
.pagination {
  display: flex;
  align-items: center;
  gap: 10px;
}
@media (max-width: 850px) {
  .page-title-bar { flex-direction: column; }
  .summary-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .history-row {
    grid-template-columns: 1fr auto;
  }
  .history-user, .history-count { display: none; }
  .table-footer { align-items: flex-start; flex-direction: column; }
}
</style>
