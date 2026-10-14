<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'

export interface SmartItem {
  id: string
  name: string
  sublabel?: string
}

const props = withDefaults(defineProps<{
  modelValue: string | null | undefined
  items: SmartItem[]
  placeholder?: string
  clearable?: boolean
  disabled?: boolean
  maxItems?: number
}>(), {
  placeholder: '',
  clearable: true,
  disabled: false,
  maxItems: 50,
})

const emit = defineEmits<{
  (e: 'update:modelValue', v: string | null): void
  (e: 'change', v: SmartItem | null): void
}>()

const open = ref(false)
const query = ref('')
const activeIdx = ref(0)
const rootEl = ref<HTMLElement | null>(null)
const inputEl = ref<HTMLInputElement | null>(null)

const selected = computed(() => props.items.find((i) => i.id === props.modelValue) ?? null)

const filtered = computed(() => {
  const q = query.value.trim().toLowerCase()
  if (!q) return props.items.slice(0, props.maxItems)
  return props.items
    .filter((i) =>
      i.name.toLowerCase().includes(q) || (i.sublabel ?? '').toLowerCase().includes(q)
    )
    .slice(0, props.maxItems)
})

function openDropdown() {
  if (props.disabled) return
  open.value = true
  query.value = ''
  activeIdx.value = 0
  nextTick(() => inputEl.value?.focus())
}

function closeDropdown() {
  open.value = false
  query.value = ''
}

function selectItem(it: SmartItem) {
  emit('update:modelValue', it.id)
  emit('change', it)
  closeDropdown()
}

function clearSel(e: Event) {
  e.stopPropagation()
  emit('update:modelValue', null)
  emit('change', null)
  closeDropdown()
}

function onKey(e: KeyboardEvent) {
  if (!open.value) {
    if (e.key === 'Enter' || e.key === ' ' || e.key === 'ArrowDown') {
      e.preventDefault()
      openDropdown()
    }
    return
  }
  if (e.key === 'ArrowDown') {
    e.preventDefault()
    activeIdx.value = Math.min(activeIdx.value + 1, filtered.value.length - 1)
  } else if (e.key === 'ArrowUp') {
    e.preventDefault()
    activeIdx.value = Math.max(activeIdx.value - 1, 0)
  } else if (e.key === 'Enter') {
    e.preventDefault()
    const it = filtered.value[activeIdx.value]
    if (it) selectItem(it)
  } else if (e.key === 'Escape') {
    e.preventDefault()
    closeDropdown()
  }
}

function onDocClick(e: MouseEvent) {
  if (!rootEl.value) return
  if (!rootEl.value.contains(e.target as Node)) closeDropdown()
}

watch(open, (v) => {
  if (v) document.addEventListener('mousedown', onDocClick)
  else document.removeEventListener('mousedown', onDocClick)
})
</script>

<template>
  <div ref="rootEl" class="ms-smart" :class="{ 'is-disabled': disabled }">
    <div
      v-if="!open"
      class="ss-display"
      :tabindex="disabled ? -1 : 0"
      @click="openDropdown"
      @keydown="onKey"
    >
      <span v-if="selected" class="ss-val" :title="selected.name">
        {{ selected.name }}
        <span v-if="selected.sublabel" class="ss-sub">{{ selected.sublabel }}</span>
      </span>
      <span v-else class="ss-ph">{{ placeholder || '—' }}</span>
      <button
        v-if="clearable && selected"
        class="ss-clear"
        type="button"
        title="Очистить"
        @click="clearSel"
      >✕</button>
      <span v-else class="ss-caret">▾</span>
    </div>

    <div v-else class="ss-input-wrap">
      <input
        ref="inputEl"
        v-model="query"
        class="ss-input"
        :placeholder="placeholder"
        @keydown="onKey"
      />
    </div>

    <div v-if="open" class="ss-dd">
      <div v-if="filtered.length === 0" class="ss-empty">Ничего не найдено</div>
      <div
        v-for="(it, idx) in filtered"
        :key="it.id"
        class="ss-opt"
        :class="{ 'is-active': idx === activeIdx }"
        @mouseenter="activeIdx = idx"
        @mousedown.prevent="selectItem(it)"
      >
        <span class="ss-opt-name">{{ it.name }}</span>
        <span v-if="it.sublabel" class="ss-opt-sub">{{ it.sublabel }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.ms-smart {
  position: relative;
  width: 100%;
  font-size: 12px;
}
.ss-display {
  display: flex;
  align-items: center;
  gap: 6px;
  height: 28px;
  padding: 0 8px;
  border: 1px solid #d0d7de;
  border-radius: 3px;
  background: #fff;
  color: #1f2328;
  cursor: pointer;
  outline: none;
  overflow: hidden;
}
.ss-display:focus { border-color: #2c5d9c; box-shadow: 0 0 0 2px rgba(44,93,156,.15); }
.ss-val {
  flex: 1;
  min-width: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.ss-sub { color: #8c959f; margin-left: 6px; font-size: 11px; }
.ss-ph { flex: 1; color: #8c959f; }
.ss-caret { color: #8c959f; font-size: 10px; }
.ss-clear {
  border: none;
  background: transparent;
  color: #8c959f;
  cursor: pointer;
  padding: 0 2px;
  font-size: 12px;
  line-height: 1;
}
.ss-clear:hover { color: #cf222e; }
.ss-input-wrap { height: 28px; }
.ss-input {
  width: 100%;
  height: 28px;
  padding: 0 8px;
  border: 1px solid #2c5d9c;
  border-radius: 3px;
  font-size: 12px;
  outline: none;
  box-shadow: 0 0 0 2px rgba(44,93,156,.15);
}
.ss-dd {
  position: absolute;
  top: calc(100% + 2px);
  left: 0;
  right: 0;
  z-index: 50;
  max-height: 260px;
  overflow-y: auto;
  background: #fff;
  border: 1px solid #d0d7de;
  border-radius: 4px;
  box-shadow: 0 6px 18px rgba(0,0,0,.08);
  padding: 3px 0;
}
.ss-opt {
  padding: 5px 10px;
  cursor: pointer;
  display: flex;
  align-items: baseline;
  gap: 8px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.ss-opt.is-active { background: #eef4ff; }
.ss-opt-name { color: #1f2328; }
.ss-opt-sub { color: #8c959f; font-size: 11px; }
.ss-empty { padding: 8px 10px; color: #8c959f; font-size: 12px; text-align: center; }
.is-disabled { opacity: .55; pointer-events: none; }
</style>
