<template>
  <div ref="root" class="qty-stepper" :class="{ 'is-disabled': disabled }">
    <button type="button" class="qs-btn" :disabled="disabled" title="Уменьшить" @click.stop="toggle('minus')">−</button>
    <input
      class="qs-input"
      type="number"
      :value="display"
      :min="min"
      :step="step"
      :disabled="disabled"
      @input="onInput"
      @focus="open = null"
    />
    <button type="button" class="qs-btn" :disabled="disabled" title="Увеличить" @click.stop="toggle('plus')">+</button>
    <Teleport to="body">
      <div
        v-if="open && menuStyle"
        class="qs-menu"
        :style="menuStyle"
        @mousedown.stop
      >
        <button
          v-for="p in presets"
          :key="p"
          type="button"
          class="qs-menu-item"
          @mousedown.prevent="apply(open === 'plus' ? p : -p)"
        >{{ open === 'plus' ? '+' : '−' }}{{ p }}</button>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'

const props = withDefaults(
  defineProps<{
    modelValue: number | null | undefined
    disabled?: boolean
    min?: number
    step?: number
    presets?: number[]
  }>(),
  {
    disabled: false,
    min: 0,
    step: 0.001,
    presets: () => [1, 10, 100],
  },
)

const emit = defineEmits<{ (e: 'update:modelValue', v: number): void }>()

const root = ref<HTMLElement | null>(null)
const open = ref<'plus' | 'minus' | null>(null)
const menuStyle = ref<Record<string, string> | null>(null)

const display = computed(() => {
  const v = props.modelValue
  if (v === null || v === undefined || (v as unknown) === '') return ''
  return String(v)
})

function round3(n: number) {
  return Math.round(n * 1000) / 1000
}

function clamp(n: number) {
  const lo = props.min ?? 0
  if (!Number.isFinite(n)) return lo
  if (n < lo) return lo
  return round3(n)
}

function onInput(e: Event) {
  const raw = (e.target as HTMLInputElement).value
  const num = raw === '' ? 0 : Number(raw)
  emit('update:modelValue', clamp(num))
}

function apply(delta: number) {
  const cur = Number(props.modelValue) || 0
  emit('update:modelValue', clamp(cur + delta))
  open.value = null
  menuStyle.value = null
}

function updateMenuPos() {
  const el = root.value
  if (!el || !open.value) return
  const r = el.getBoundingClientRect()
  const estH = (props.presets?.length ?? 0) * 26 + 8
  const spaceBelow = window.innerHeight - r.bottom
  const top = spaceBelow < estH + 8 ? Math.max(4, r.top - estH - 4) : r.bottom + 2
  menuStyle.value = {
    position: 'fixed',
    top: `${top}px`,
    left: `${r.left}px`,
    zIndex: '9999',
  }
}

async function toggle(which: 'plus' | 'minus') {
  if (props.disabled) return
  if (open.value === which) {
    open.value = null
    menuStyle.value = null
    return
  }
  open.value = which
  menuStyle.value = null
  await nextTick()
  updateMenuPos()
}

function onDocClick(e: MouseEvent) {
  if (!open.value) return
  const el = root.value
  const target = e.target as Node
  if (el && el.contains(target)) return
  const menu = document.querySelector('.qs-menu') as HTMLElement | null
  if (menu && menu.contains(target)) return
  open.value = null
  menuStyle.value = null
}

function onKey(e: KeyboardEvent) {
  if (e.key === 'Escape') {
    open.value = null
    menuStyle.value = null
  }
}

function onScroll() {
  if (open.value) updateMenuPos()
}

onMounted(() => {
  document.addEventListener('mousedown', onDocClick)
  document.addEventListener('keydown', onKey)
  window.addEventListener('scroll', onScroll, true)
  window.addEventListener('resize', onScroll)
})
onBeforeUnmount(() => {
  document.removeEventListener('mousedown', onDocClick)
  document.removeEventListener('keydown', onKey)
  window.removeEventListener('scroll', onScroll, true)
  window.removeEventListener('resize', onScroll)
})
</script>

<style scoped>
.qty-stepper { position: relative; display: inline-flex; align-items: center; gap: 2px; }
.qs-btn {
  width: 20px; height: 26px; padding: 0;
  border: 1px solid #d0d7de; background: #fff; border-radius: 3px;
  cursor: pointer; font-size: 14px; line-height: 1; color: #57606a;
  display: inline-flex; align-items: center; justify-content: center; flex-shrink: 0;
}
.qs-btn:hover:not(:disabled) { background: #eaf3ff; color: #2c5d9c; border-color: #2c5d9c; }
.qs-btn:disabled { opacity: 0.5; cursor: default; }
.qs-input {
  width: 72px; height: 26px; padding: 0 6px;
  font-size: 13px; text-align: right;
  border: 1px solid #d0d7de; border-radius: 3px;
  background: #fff; color: #1f2328;
  font-variant-numeric: tabular-nums;
}
.qs-input:focus { outline: 2px solid rgba(44,93,156,0.3); border-color: #2c5d9c; }
.qs-input:disabled { background: #f6f8fa; color: #8c959f; }
</style>

<style>
.qs-menu {
  background: #fff; border: 1px solid #d0d7de; border-radius: 4px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.12);
  padding: 3px; min-width: 78px;
  display: flex; flex-direction: column; gap: 1px;
}
.qs-menu-item {
  padding: 4px 10px; text-align: left;
  background: transparent; border: none; border-radius: 3px;
  font-size: 12px; cursor: pointer; color: #1f2328; white-space: nowrap;
}
.qs-menu-item:hover { background: #eaf3ff; color: #2c5d9c; }
</style>