<template>
  <div class="col-settings" ref="rootRef">
    <MsButton variant="icon" icon="gear" title="Настройки колонок" @click="open = !open" />
    <div v-if="open" class="col-popover" @click.stop>
      <div class="col-head">Колонки</div>
      <label v-for="d in defs" :key="d.key" class="col-item" :class="{ locked: d.always }">
        <input type="checkbox" :checked="isVisible(d.key)" :disabled="d.always" @change="toggle(d.key)" />
        <span>{{ d.label }}</span>
      </label>
      <button type="button" class="col-reset" @click="reset">По умолчанию</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import MsButton from './MsButton.vue'
import type { ColumnDef } from '../composables/useColumnSettings'

defineProps<{
  defs: ColumnDef[]
  isVisible: (key: string) => boolean
  toggle: (key: string) => void
  reset: () => void
}>()

const open = ref(false)
const rootRef = ref<HTMLElement | null>(null)

function onDocClick(e: MouseEvent): void {
  if (!open.value) return
  const el = rootRef.value
  if (el && !el.contains(e.target as Node)) open.value = false
}

onMounted(() => document.addEventListener('mousedown', onDocClick))
onBeforeUnmount(() => document.removeEventListener('mousedown', onDocClick))
</script>

<style scoped>
.col-settings { position: relative; display: inline-block; }
.col-popover {
  position: absolute; top: calc(100% + 4px); right: 0;
  background: #fff; border: 1px solid #d0d7de; border-radius: 6px;
  box-shadow: 0 6px 20px rgba(0,0,0,0.12); padding: 8px 0;
  min-width: 200px; z-index: 50;
}
.col-head { font-size: 11px; text-transform: uppercase; letter-spacing: .4px; color: #6e7781; padding: 2px 12px 6px; }
.col-item { display: flex; align-items: center; gap: 8px; padding: 5px 12px; cursor: pointer; font-size: 13px; color: #1f2328; user-select: none; }
.col-item:hover { background: #f3f6fb; }
.col-item.locked { color: #6e7781; cursor: default; }
.col-item.locked:hover { background: transparent; }
.col-item input { margin: 0; }
.col-reset { margin: 6px 12px 2px; width: calc(100% - 24px); padding: 5px 8px; border: 1px solid #d0d7de; background: #fff; border-radius: 4px; font-size: 12px; cursor: pointer; color: #1f2328; }
.col-reset:hover { background: #f6f8fa; }
</style>