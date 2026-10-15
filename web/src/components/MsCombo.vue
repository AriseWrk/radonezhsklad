<script setup lang="ts">
import { computed, ref } from 'vue'
import MsIcon from './MsIcon.vue'
import { smartFilter } from '../utils/smartSearch'

interface Item { id: string; name: string; sku?: string }
const props = withDefaults(defineProps<{
  modelValue: string
  items: Item[]
  placeholder?: string
  disabled?: boolean
}>(), { placeholder: '—' })
const emit = defineEmits<{ (e: 'update:modelValue', v: string): void }>()

const open = ref(false)
const query = ref('')
const inputEl = ref<HTMLInputElement | null>(null)

const selectedName = computed(() => {
  const it = props.items.find((x) => x.id === props.modelValue)
  return it ? it.name : ''
})

const filtered = computed(() => {
  const q = query.value.trim()
  if (!q) return props.items.slice(0, 50)
  return smartFilter(props.items, q, {
    getFields: (it) => [it.name, it.sku ?? null, null, null],
    limit: 50,
  })
})

function onFocus() {
  open.value = true
  query.value = ''
}

function onInput(e: Event) {
  query.value = (e.target as HTMLInputElement).value
  open.value = true
}

function pick(it: Item | null) {
  emit('update:modelValue', it ? it.id : '')
  query.value = ''
  open.value = false
}

function onBlur() {
  // небольшая задержка, чтобы клик по элементу успел сработать
  setTimeout(() => { open.value = false }, 150)
}

// Разбивает name на [до, совпадение, после] для подсветки по query
function parts(name: string): [string, string, string] {
  const q = query.value.trim()
  if (!q) return [name, '', '']
  const i = name.toLowerCase().indexOf(q.toLowerCase())
  if (i < 0) return [name, '', '']
  return [name.slice(0, i), name.slice(i, i + q.length), name.slice(i + q.length)]
}
</script>

<template>
  <div class="ms-combo" :class="{ open, disabled }">
    <input
      ref="inputEl"
      class="ms-combo-input"
      :value="open ? query : selectedName"
      :placeholder="selectedName ? '' : placeholder"
      :disabled="disabled"
      @focus="onFocus"
      @input="onInput"
      @blur="onBlur"
    />
    <button type="button" class="ms-combo-clear" v-if="selectedName && !disabled" @mousedown.prevent="pick(null)" title="Очистить">×</button>
    <MsIcon name="chevron-down" :size="12" class="ms-combo-caret" />
    <div v-if="open" class="ms-combo-drop">
      <div class="ms-combo-item muted" @mousedown.prevent="pick(null)">— не выбрано —</div>
      <div
        v-for="it in filtered"
        :key="it.id"
        class="ms-combo-item"
        :class="{ active: it.id === modelValue }"
        @mousedown.prevent="pick(it)"
      ><template v-for="(p, i) in [parts(it.name)]" :key="i"><span>{{ p[0] }}</span><mark v-if="p[1]">{{ p[1] }}</mark><span>{{ p[2] }}</span></template></div>
      <div v-if="filtered.length === 0" class="ms-combo-item muted">Ничего не найдено</div>
    </div>
  </div>
</template>

<style scoped>
.ms-combo { position: relative; width: 100%; }
.ms-combo-input {
  width: 100%; height: 28px; padding: 0 24px 0 8px;
  font-size: 12px; border: 1px solid #d0d7de; border-radius: 3px;
  background: #fff; color: #1f2328; box-sizing: border-box;
}
.ms-combo-input:focus { outline: none; border-color: #2c5d9c; }
.ms-combo.open .ms-combo-input { border-color: #2c5d9c; }
.ms-combo-caret {
  position: absolute; right: 6px; top: 50%; transform: translateY(-50%);
  color: #8c959f; pointer-events: none;
}
.ms-combo-clear {
  position: absolute; right: 20px; top: 50%; transform: translateY(-50%);
  width: 16px; height: 16px; padding: 0; border: none; background: transparent;
  color: #8c959f; cursor: pointer; font-size: 14px; line-height: 1;
}
.ms-combo-clear:hover { color: #cf222e; }
.ms-combo-drop {
  position: absolute; z-index: 100; left: 0; right: 0; top: calc(100% + 2px);
  max-height: 240px; overflow-y: auto;
  background: #fff; border: 1px solid #d0d7de; border-radius: 3px;
  box-shadow: 0 4px 12px rgba(0,0,0,.12);
}
.ms-combo-item {
  padding: 5px 10px; font-size: 12px; color: #1f2328; cursor: pointer;
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
}
.ms-combo-item:hover, .ms-combo-item.active { background: #eef4fb; }
.ms-combo-item.muted { color: #8c959f; font-style: italic; }
.ms-combo-item mark { background: #fff3a3; color: inherit; padding: 0; }
</style>