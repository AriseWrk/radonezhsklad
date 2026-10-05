<template>
  <div class="page">
    <div class="ms-title">
      <span>Синхронизация с МойСклад</span>
      <button class="ms-refresh" @click="loadAll" title="Обновить">
        <MsIcon name="refresh" :size="14" />
      </button>
    </div>

    <div v-if="error" class="error-box">{{ error }}</div>

    <!-- Автосинк -->
    <div class="sync-card">
      <div class="sync-card-head">Автосинк (retry-pending)</div>
      <div v-if="autoSync" class="sync-card-body">
        <div class="kv">
          <span class="k">Статус:</span>
          <span class="v" :class="autoSync.enabled ? 'ok' : 'muted'">
            {{ autoSync.enabled ? 'включён' : 'выключен' }}
          </span>
          <span class="hint">(MS_AUTOSYNC_ENABLED в warehouse/.env)</span>
        </div>
        <div class="kv">
          <span class="k">Интервал:</span>
          <span class="v">{{ autoSync.interval_s }} сек</span>
        </div>
        <div class="kv">
          <span class="k">Последний тик:</span>
          <span class="v">{{ autoSync.last_tick ? formatDt(autoSync.last_tick) : '—' }}</span>
        </div>
        <div class="kv">
          <span class="k">Тиков всего:</span>
          <span class="v">{{ autoSync.tick_count }}</span>
        </div>
        <div v-if="autoSync.last_error" class="kv">
          <span class="k">Последняя ошибка:</span>
          <span class="v err">{{ autoSync.last_error }}</span>
        </div>
      </div>
      <div v-else class="sync-card-body muted">Загрузка...</div>
    </div>

    <!-- Pull вручную -->
    <div class="sync-card">
      <div class="sync-card-head">Синхронизировать из МС</div>
      <div class="sync-card-body">
        <div class="pull-row">
          <select v-model="selectedKind" class="ms-input">
            <option v-for="k in PULL_KINDS" :key="k.kind" :value="k.kind">{{ k.label }}</option>
          </select>
          <MsButton variant="primary" icon="refresh" :disabled="busy" @click="doPull">
            Тянуть
          </MsButton>
          <MsButton icon="check" :disabled="busy" @click="doRetry">
            Retry pending ({{ openErrorsCount }})
          </MsButton>
        </div>
        <div v-if="activeJob" class="active-job">
          <span class="tag running">{{ activeJob.kind }}</span>
          <span>{{ activeJob.status }}</span>
          <span v-if="activeJob.total">{{ activeJob.fetched }} / {{ activeJob.total }}</span>
          <span v-else>{{ activeJob.fetched }}</span>
          <span v-if="activeJob.last_line" class="last-line">{{ activeJob.last_line }}</span>
        </div>
      </div>
    </div>

    <!-- Ошибки -->
    <div class="sync-card">
      <div class="sync-card-head">
        Ошибки синхронизации
        <span class="badge" v-if="openErrorsCount > 0">{{ openErrorsCount }}</span>
      </div>
      <div class="sync-card-body">
        <table class="ms-table2" v-if="errors.length > 0">
          <thead>
            <tr>
              <th>Когда</th>
              <th>Сущность</th>
              <th>Оп</th>
              <th>local_id</th>
              <th>Ошибка</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="e in errors" :key="e.id">
              <td>{{ formatDt(e.created_at) }}</td>
              <td>{{ e.entity }}</td>
              <td>{{ e.op }}</td>
              <td><span v-if="e.local_id" class="mono">{{ short(e.local_id) }}</span></td>
              <td class="err">{{ e.error }}</td>
              <td>
                <MsButton
                  v-if="e.entity === 'document' && e.local_id"
                  variant="icon"
                  icon="refresh"
                  title="Повторить"
                  :disabled="busy"
                  @click="doRetryOne(e.local_id!)"
                />
              </td>
            </tr>
          </tbody>
        </table>
        <div v-else class="muted">Открытых ошибок нет.</div>
      </div>
    </div>

    <!-- История job'ов -->
    <div class="sync-card">
      <div class="sync-card-head">История задач</div>
      <div class="sync-card-body">
        <table class="ms-table2" v-if="jobs.length > 0">
          <thead>
            <tr>
              <th>Kind</th>
              <th>Статус</th>
              <th>Начато</th>
              <th>Закончено</th>
              <th>Fetched</th>
              <th>Total</th>
              <th>Last line / error</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="j in jobs" :key="j.id">
              <td><span class="mono">{{ j.kind }}</span></td>
              <td>
                <span class="st" :class="'st-' + j.status">{{ j.status }}</span>
              </td>
              <td>{{ formatDt(j.started_at) }}</td>
              <td>{{ j.ended_at ? formatDt(j.ended_at) : '—' }}</td>
              <td class="num">{{ j.fetched }}</td>
              <td class="num">{{ j.total ?? '—' }}</td>
              <td class="ellip">{{ j.error || j.last_line || '—' }}</td>
            </tr>
          </tbody>
        </table>
        <div v-else class="muted">Пусто.</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import MsIcon from '../components/MsIcon.vue'
import MsButton from '../components/MsButton.vue'
import { apiErrorMessage } from '../api/client'
import {
  listSyncJobs,
  listSyncErrors,
  startPull,
  startRetryPending,
  retryDocument,
  getAutoSyncStatus,
  PULL_KINDS,
  type SyncJob,
  type SyncErrorItem,
  type AutoSyncStatus,
} from '../api/sync'

const jobs = ref<SyncJob[]>([])
const errors = ref<SyncErrorItem[]>([])
const openErrorsCount = ref(0)
const autoSync = ref<AutoSyncStatus | null>(null)
const error = ref<string | null>(null)
const busy = ref(false)
const selectedKind = ref(PULL_KINDS[0].kind)

const activeJob = computed<SyncJob | null>(() =>
  jobs.value.find((j) => j.status === 'running' || j.status === 'queued') ?? null,
)

let timer: number | null = null

function formatDt(s?: string | null): string {
  if (!s) return '—'
  const d = new Date(s)
  if (Number.isNaN(d.getTime())) return s
  return d.toLocaleString('ru-RU', { hour12: false })
}

function short(uuid?: string): string {
  if (!uuid) return ''
  return uuid.slice(0, 8) + '…'
}

async function loadAll() {
  error.value = null
  try {
    const [js, es, as] = await Promise.all([
      listSyncJobs(30),
      listSyncErrors(50),
      getAutoSyncStatus().catch(() => null),
    ])
    jobs.value = js
    errors.value = es.items
    openErrorsCount.value = es.count_open
    autoSync.value = as
  } catch (e) {
    error.value = apiErrorMessage(e)
  }
}

async function doPull() {
  busy.value = true
  error.value = null
  try {
    const j = await startPull(selectedKind.value)
    await waitJob(j.id)
  } catch (e) {
    error.value = apiErrorMessage(e)
  } finally {
    busy.value = false
    await loadAll()
  }
}

async function doRetry() {
  busy.value = true
  error.value = null
  try {
    const j = await startRetryPending()
    await waitJob(j.id)
  } catch (e) {
    error.value = apiErrorMessage(e)
  } finally {
    busy.value = false
    await loadAll()
  }
}

async function doRetryOne(docId: string) {
  busy.value = true
  error.value = null
  try {
    await retryDocument(docId)
  } catch (e) {
    error.value = apiErrorMessage(e)
  } finally {
    busy.value = false
    await loadAll()
  }
}

async function waitJob(id: string) {
  for (let i = 0; i < 600; i++) {
    await new Promise((r) => setTimeout(r, 1000))
    try {
      const js = await listSyncJobs(5)
      const j = js.find((x) => x.id === id)
      if (!j) return
      jobs.value = js
      if (j.status === 'done' || j.status === 'error' || j.status === 'cancelled') return
    } catch {
      return
    }
  }
}

onMounted(async () => {
  await loadAll()
  timer = window.setInterval(loadAll, 5000)
})

onUnmounted(() => {
  if (timer !== null) window.clearInterval(timer)
})
</script>

<style scoped>
.page { padding: 16px; }
.ms-title {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 15px;
  font-weight: 600;
  margin-bottom: 12px;
}
.ms-refresh {
  background: none;
  border: 1px solid #d0d7de;
  border-radius: 6px;
  width: 26px;
  height: 26px;
  cursor: pointer;
}
.error-box {
  background: #ffeef0;
  color: #cf222e;
  border: 1px solid #f5b3ba;
  border-radius: 6px;
  padding: 8px 12px;
  margin-bottom: 12px;
}
.sync-card {
  border: 1px solid #d0d7de;
  border-radius: 8px;
  background: #fff;
  margin-bottom: 14px;
}
.sync-card-head {
  padding: 10px 14px;
  border-bottom: 1px solid #d0d7de;
  font-weight: 600;
  font-size: 13px;
  background: #f6f8fa;
  border-radius: 8px 8px 0 0;
  display: flex;
  align-items: center;
  gap: 8px;
}
.sync-card-body { padding: 12px 14px; }
.badge {
  background: #cf222e;
  color: #fff;
  border-radius: 10px;
  padding: 1px 8px;
  font-size: 11px;
}
.kv {
  display: flex;
  gap: 8px;
  padding: 3px 0;
  font-size: 13px;
}
.kv .k { color: #57606a; min-width: 160px; }
.kv .v { color: #24292f; }
.kv .v.ok { color: #1a7f37; font-weight: 600; }
.kv .v.muted { color: #8c959f; }
.kv .hint { color: #8c959f; font-size: 12px; }
.pull-row {
  display: flex;
  gap: 8px;
  align-items: center;
}
.ms-input {
  padding: 5px 8px;
  border: 1px solid #d0d7de;
  border-radius: 6px;
  font-size: 13px;
  background: #fff;
  min-width: 180px;
}
.active-job {
  margin-top: 10px;
  display: flex;
  gap: 10px;
  align-items: center;
  font-size: 13px;
  flex-wrap: wrap;
}
.tag.running {
  background: #ddf4ff;
  color: #0969da;
  padding: 2px 8px;
  border-radius: 6px;
  font-size: 12px;
}
.last-line {
  color: #57606a;
  font-size: 12px;
  font-family: ui-monospace, monospace;
}
.muted { color: #8c959f; }
.err { color: #cf222e; word-break: break-word; }
.mono { font-family: ui-monospace, monospace; font-size: 12px; }

.ms-table2 {
  width: 100%;
  border-collapse: collapse;
  font-size: 12.5px;
}
.ms-table2 th,
.ms-table2 td {
  border-bottom: 1px solid #eaeef2;
  padding: 6px 8px;
  text-align: left;
  vertical-align: top;
}
.ms-table2 th {
  background: #f6f8fa;
  font-weight: 600;
  color: #57606a;
  white-space: nowrap;
}
.ms-table2 .num { text-align: right; font-variant-numeric: tabular-nums; }
.ms-table2 .ellip {
  max-width: 360px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: #57606a;
}
.st {
  padding: 1px 6px;
  border-radius: 4px;
  font-size: 11px;
  text-transform: uppercase;
  font-weight: 600;
}
.st-done { background: #dcffe4; color: #1a7f37; }
.st-running { background: #ddf4ff; color: #0969da; }
.st-queued { background: #f6f8fa; color: #57606a; }
.st-error { background: #ffebe9; color: #cf222e; }
.st-cancelled { background: #f6f8fa; color: #8c959f; }
</style>