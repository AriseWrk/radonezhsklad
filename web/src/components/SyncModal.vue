<template>
  <div v-if="visible" class="sync-overlay" @click.self="onClose">
    <div class="sync-dialog">
      <div class="sync-head">
        <span class="sync-icon">🔄</span>
        <span class="sync-title">Синхронизация с МойСклад</span>
      </div>

      <div class="sync-body">
        <div class="sync-status-row">
          <span class="sync-status">{{ statusText }}</span>
          <span v-if="job && job.total > 0" class="sync-counter">{{ job.fetched }} / {{ job.total }}</span>
        </div>

        <div class="sync-bar">
          <div class="sync-bar-fill" :class="barClass" :style="{ width: percent + '%' }"></div>
        </div>

        <div class="sync-details">
          <div v-if="job && job.last_line" class="sync-line">{{ job.last_line }}</div>
          <div v-if="job && job.error" class="sync-error">{{ job.error }}</div>
        </div>
      </div>

      <div class="sync-foot">
        <button
          class="sync-btn"
          :class="{ primary: finished }"
          :disabled="!finished"
          @click="onClose"
        >
          {{ finished ? 'Закрыть' : 'Идёт синхронизация...' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { startPullOrders, getSyncJob, type SyncJob } from '../api/sync'
import { apiErrorMessage } from '../api/client'

const visible = ref(false)
const job = ref<SyncJob | null>(null)
const error = ref<string | null>(null)

const finished = computed(() => !job.value || job.value.status === 'done' || job.value.status === 'error')

const percent = computed(() => {
  if (error.value) return 100
  if (!job.value) return 0
  if (job.value.status === 'done') return 100
  if (job.value.status === 'error') return 100
  if (job.value.total <= 0) {
    // индикация «живости»: если total ещё не пришёл — крутим до 30%
    return job.value.status === 'running' ? 30 : 5
  }
  return Math.min(100, Math.round((job.value.fetched / job.value.total) * 100))
})

const barClass = computed(() => {
  if (error.value) return 'error'
  if (!job.value) return ''
  if (job.value.status === 'done') return 'done'
  if (job.value.status === 'error') return 'error'
  return 'running'
})

const statusText = computed(() => {
  if (error.value) return 'Ошибка'
  if (!job.value) return 'Подготовка...'
  switch (job.value.status) {
    case 'queued': return 'В очереди...'
    case 'running': return 'Загрузка из МойСклад...'
    case 'done': return 'Синхронизация завершена'
    case 'error': return 'Ошибка синхронизации'
  }
  return ''
})

async function poll(id: string) {
  let attempts = 0
  while (attempts < 2000) {
    attempts++
    await new Promise((r) => setTimeout(r, 500))
    try {
      const j = await getSyncJob(id)
      job.value = j
      if (j.status === 'done' || j.status === 'error') {
        if (j.status === 'error') error.value = j.error || 'Ошибка'
        return
      }
    } catch (e) {
      error.value = apiErrorMessage(e)
      return
    }
  }
  error.value = 'Превышено время ожидания'
}

async function open() {
  visible.value = true
  job.value = null
  error.value = null
  try {
    const j = await startPullOrders()
    job.value = j
    await poll(j.id)
  } catch (e) {
    error.value = apiErrorMessage(e)
  }
}

function onClose() {
  if (!finished.value) return
  visible.value = false
}

defineExpose({ open })
</script>

<style scoped>
.sync-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.35);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}
.sync-dialog {
  width: 440px;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.2);
  padding: 20px;
  font-family: inherit;
}
.sync-head {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
}
.sync-icon { font-size: 20px; }
.sync-title { font-weight: 600; font-size: 15px; }
.sync-body { margin-bottom: 16px; }
.sync-status-row {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  margin-bottom: 8px;
  font-size: 13px;
}
.sync-status { color: #333; }
.sync-counter { color: #666; font-variant-numeric: tabular-nums; }
.sync-bar {
  height: 8px;
  background: #eee;
  border-radius: 4px;
  overflow: hidden;
  margin-bottom: 10px;
}
.sync-bar-fill {
  height: 100%;
  background: #2ea44f;
  transition: width 0.3s ease;
}
.sync-bar-fill.running { background: #2ea44f; }
.sync-bar-fill.done    { background: #2ea44f; }
.sync-bar-fill.error   { background: #cf222e; }
.sync-details {
  font-size: 12px;
  color: #555;
  min-height: 32px;
  max-height: 80px;
  overflow-y: auto;
}
.sync-line { white-space: pre-wrap; word-break: break-word; }
.sync-error { color: #cf222e; margin-top: 4px; font-weight: 500; }
.sync-foot { display: flex; justify-content: flex-end; }
.sync-btn {
  padding: 6px 14px;
  border: 1px solid #d0d7de;
  background: #f6f8fa;
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
}
.sync-btn:disabled {
  opacity: 0.6;
  cursor: default;
}
.sync-btn.primary {
  background: #1f883d;
  color: #fff;
  border-color: #1f883d;
}
</style>