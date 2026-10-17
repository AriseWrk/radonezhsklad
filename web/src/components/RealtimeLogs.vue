<template>
  <div v-if="isAdmin" class="rt">
    <div class="rt-tabs">
      <button v-for="t in tabs" :key="t.key" class="rt-tab" :class="{ active: tab === t.key }" @click="tab = t.key">
        {{ t.label }}
        <span v-if="t.key === 'errors' && errorLines.length" class="rt-badge">{{ errorLines.length }}</span>
      </button>
      <div class="rt-spacer"></div>
      <select v-if="tab !== 'actions'" v-model="serviceFilter" class="rt-select">
        <option value="all">все сервисы</option>
        <option v-for="s in serviceNames" :key="s" :value="s">{{ s }}</option>
      </select>
      <span class="rt-status" :class="{ live: !paused }"><span class="rt-dot"></span>{{ paused ? 'paused' : 'live' }}</span>
      <button class="rt-btn" @click="paused = !paused" :title="paused ? 'Возобновить' : 'Пауза'">{{ paused ? '▶' : '❚❚' }}</button>
      <button class="rt-btn" @click="refresh" title="Обновить">⟳</button>
    </div>
    <div class="rt-body" :class="{ dimmed: paused }">
      <template v-if="tab === 'system'">
        <div v-if="!systemLines.length" class="rt-empty">Нет данных</div>
        <div v-for="(l, i) in systemLines" :key="'s' + i" class="rt-line">
          <span class="rt-svc">{{ l.service }}</span>
          <span class="rt-text">{{ l.line }}</span>
        </div>
      </template>
      <template v-else-if="tab === 'errors'">
        <div v-if="!errorLines.length" class="rt-empty">Ошибок нет</div>
        <div v-for="(l, i) in errorLines" :key="'e' + i" class="rt-entry">
          <div class="rt-line err">
            <span class="rt-svc">{{ l.service }}</span>
            <span class="rt-text">{{ l.line }}</span>
          </div>
          <div v-if="hint(l.line)" class="rt-hint">Возможно: {{ hint(l.line) }}</div>
        </div>
      </template>
      <template v-else>
        <div v-if="!actions.length" class="rt-empty">Нет действий</div>
        <div v-for="a in actions" :key="a.id" class="rt-action">
          <span class="rt-time">{{ formatDateTimeRu(a.created_at) }}</span>
          <span class="rt-user">{{ userLabel(a.user_id) }}</span>
          <span class="rt-verb">{{ formatEventRu(a) }}</span>
        </div>
      </template>
    </div>
  </div>
  <div v-else class="rt-denied">Логи и действия доступны только администратору</div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { listAudit, listAuditLogs, type AuditLog, type AuditLogLine } from '../api/audit'
import { listUsers, type User } from '../api/users'
import { formatDateTimeRu, formatEventRu, formatUserShort } from '../utils/auditFormat'
import { useAuthStore } from '../stores/auth'

type Tab = 'system' | 'errors' | 'actions'
const tabs: Array<{ key: Tab; label: string }> = [
  { key: 'system', label: 'Системный лог' },
  { key: 'errors', label: 'Ошибки' },
  { key: 'actions', label: 'Действия пользователей' },
]
const serviceNames = ['gateway', 'auth', 'product', 'warehouse', 'order', 'audit', 'front']

const auth = useAuthStore()
const isAdmin = computed(() => auth.role === 'admin')

const tab = ref<Tab>('system')
const serviceFilter = ref('all')
const paused = ref(false)

const systemLines = ref<AuditLogLine[]>([])
const errorLines = ref<AuditLogLine[]>([])
const actions = ref<AuditLog[]>([])
const users = ref<User[]>([])

const userMap = computed(() => {
  const m = new Map<string, User>()
  for (const u of users.value) m.set(u.id, u)
  return m
})

function userLabel(id?: string): string {
  if (!id) return 'Система'
  return formatUserShort(userMap.value.get(id) ?? null)
}

const HINTS: Array<{ re: RegExp; hint: string }> = [
  { re: /connection refused|dial tcp/i, hint: 'Сервис не запущен или недоступен.' },
  { re: /context deadline exceeded|timeout/i, hint: 'Таймаут запроса (БД или соседний сервис).' },
  { re: /address already in use|Only one usage of each socket/i, hint: 'Порт занят: останови старый процесс.' },
  { re: /password authentication failed|role .* does not exist/i, hint: 'Postgres: неверный логин/пароль или нет роли.' },
  { re: /no such host|lookup .* on/i, hint: 'DNS/URL сервиса неверный — проверь .env.' },
  { re: /panic:|goroutine \d+/i, hint: 'Паника в Go. Смотри стектрейс выше.' },
  { re: /duplicate key|unique constraint/i, hint: 'Дубликат уникального поля.' },
  { re: /no rows in result set/i, hint: 'Запись не найдена в БД.' },
  { re: /401|unauthorized/i, hint: 'Проблема с JWT.' },
  { re: /403|forbidden/i, hint: 'Недостаточно прав.' },
  { re: /invalid input syntax|invalid uuid/i, hint: 'Невалидный UUID/тип в запросе.' },
]

function hint(line: string): string | null {
  for (const h of HINTS) if (h.re.test(line)) return h.hint
  return null
}

async function refresh() {
  try {
    if (tab.value === 'system') {
      const r = await listAuditLogs({ kind: 'out', service: serviceFilter.value, lines: 200 })
      systemLines.value = r.items.slice().reverse()
    } else if (tab.value === 'errors') {
      const r = await listAuditLogs({ kind: 'err', service: serviceFilter.value, lines: 200 })
      errorLines.value = r.items.slice().reverse()
    } else {
      const r = await listAudit({ limit: 50 })
      actions.value = r.items
    }
  } catch { /* dev dashboard: silent */ }
}

let timer: number | null = null
function startTimer() {
  if (timer !== null) window.clearInterval(timer)
  timer = window.setInterval(() => { if (!paused.value) refresh() }, 3000)
}

watch(tab, () => refresh())
watch(serviceFilter, () => refresh())

onMounted(async () => {
  try { users.value = await listUsers() } catch { /* ignore */ }
  await refresh()
  startTimer()
})
onBeforeUnmount(() => {
  if (timer !== null) window.clearInterval(timer)
})
</script>

<style scoped>
.rt { margin-top: 24px; background: #0d1117; border: 1px solid #30363d; border-radius: 6px; overflow: hidden; font-family: ui-monospace, 'Cascadia Code', Menlo, Consolas, monospace; font-size: 12px; }
.rt-tabs { display: flex; align-items: center; gap: 4px; padding: 6px 8px; background: #161b22; border-bottom: 1px solid #30363d; flex-wrap: wrap; }
.rt-tab { background: transparent; border: none; color: #8b949e; padding: 6px 12px; font-size: 12px; cursor: pointer; border-radius: 4px; font-family: inherit; }
.rt-tab:hover { background: #21262d; color: #c9d1d9; }
.rt-tab.active { background: #0d419d; color: #fff; }
.rt-badge { display: inline-block; margin-left: 6px; padding: 0 6px; background: #cf222e; color: #fff; border-radius: 10px; font-size: 10px; line-height: 16px; }
.rt-spacer { flex: 1; }
.rt-select { background: #0d1117; color: #c9d1d9; border: 1px solid #30363d; padding: 3px 6px; font-size: 11px; border-radius: 4px; font-family: inherit; }
.rt-status { display: inline-flex; align-items: center; gap: 5px; color: #8b949e; font-size: 11px; padding: 0 6px; }
.rt-dot { width: 7px; height: 7px; border-radius: 50%; background: #6e7681; }
.rt-status.live .rt-dot { background: #3fb950; box-shadow: 0 0 0 3px rgba(63,185,80,0.2); }
.rt-btn { background: #21262d; color: #c9d1d9; border: 1px solid #30363d; padding: 3px 8px; font-size: 11px; border-radius: 4px; cursor: pointer; font-family: inherit; }
.rt-btn:hover { background: #30363d; }
.rt-body { max-height: 360px; overflow-y: auto; padding: 8px 10px; }
.rt-body.dimmed { opacity: 0.55; }
.rt-empty { color: #6e7681; padding: 20px; text-align: center; font-style: italic; }
.rt-entry { padding: 2px 0; border-bottom: 1px dashed #21262d; }
.rt-line { padding: 2px 0; display: flex; gap: 8px; align-items: flex-start; border-bottom: 1px dashed #21262d; }
.rt-entry .rt-line { border-bottom: none; padding: 0; }
.rt-line.err { border-left: 3px solid #cf222e; padding-left: 6px; }
.rt-svc { color: #58a6ff; font-size: 10px; text-transform: uppercase; flex-shrink: 0; width: 72px; padding-top: 2px; }
.rt-text { color: #c9d1d9; white-space: pre-wrap; word-break: break-all; flex: 1; }
.rt-hint { color: #d29922; font-size: 11px; padding: 4px 0 4px 80px; }
.rt-action { display: flex; gap: 10px; align-items: baseline; padding: 4px 0; border-bottom: 1px dashed #21262d; color: #c9d1d9; }
.rt-action .rt-time { color: #8b949e; font-size: 11px; width: 130px; flex-shrink: 0; }
.rt-action .rt-user { color: #7ee787; font-weight: 600; }
.rt-action .rt-verb { color: #c9d1d9; }
.rt-denied { margin-top: 24px; padding: 20px; background: #f6f8fa; border: 1px dashed #d0d7de; border-radius: 6px; color: #57606a; text-align: center; font-size: 12px; }
</style>
