<template>
  <div class="pp">
    <input
      v-model="query"
      class="pp-input"
      :placeholder="placeholder"
      @focus="open = true"
      @input="search"
      @blur="onBlur"
    />
    <div v-if="open && results.length > 0" class="pp-list">
      <button
        v-for="p in results"
        :key="p.id"
        class="pp-item"
        @mousedown.prevent="pick(p)"
      >
        <span class="pp-item-name">{{ p.name }}</span>
        <span class="pp-item-sku">{{ p.sku || '' }}</span>
      </button>
    </div>
    <div v-if="open && query && results.length === 0" class="pp-empty">Ничего не найдено</div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { smartFilter } from '../utils/smartSearch'
import type { Product } from '../api/products'

const props = defineProps<{
  modelValue: string
  products: Product[]
  placeholder?: string
  disabled?: boolean
}>()
const emit = defineEmits<{
  'update:modelValue': [id: string]
  'picked': [p: Product]
}>()

const query = ref('')
const open = ref(false)
const results = ref<Product[]>([])

const placeholderText = computed(() => props.placeholder || 'Найти товар...')

function search() {
  if (!query.value) { results.value = []; return }
  results.value = smartFilter(props.products, query.value, {
    getFields: (p) => [p.name, p.sku ?? null, p.barcode ?? null, p.id.slice(0, 8)],
    limit: 20,
  })
  open.value = true
}

function pick(p: Product) {
  emit('update:modelValue', p.id)
  emit('picked', p)
  query.value = p.name
  open.value = false
}

function onBlur() {
  setTimeout(() => { open.value = false }, 150)
}

// При изменении внешнего modelValue (например, при загрузке документа) — обновить query
import { watch } from 'vue'
watch(() => props.modelValue, (id) => {
  if (!id) { query.value = ''; return }
  const p = props.products.find((x) => x.id === id)
  if (p) query.value = p.name
})
</script>

<style scoped>
.pp { position: relative; width: 100%; }
.pp-input {
  width: 100%; height: 26px; padding: 0 6px; font-size: 13px;
  border: 1px solid #d0d7de; border-radius: 3px;
  background: #fff; color: #1f2328;
}
.pp-input::placeholder { color: #8c959f; }
.pp-list {
  position: absolute; top: 28px; left: 0; right: 0; z-index: 100;
  background: #fff; border: 1px solid #d0d7de; border-radius: 4px;
  box-shadow: 0 4px 16px rgba(0,0,0,0.12);
  max-height: 260px; overflow-y: auto;
  padding: 4px 0;
}
.pp-item {
  display: flex; justify-content: space-between; align-items: center;
  gap: 8px; width: 100%; text-align: left;
  padding: 6px 10px; border: none; background: transparent;
  font-size: 13px; color: #1f2328; cursor: pointer;
}
.pp-item:hover { background: #f6f8fa; }
.pp-item-name { flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.pp-item-sku { color: #8c959f; font-family: monospace; font-size: 12px; }
.pp-empty {
  position: absolute; top: 28px; left: 0; right: 0; z-index: 100;
  background: #fff; border: 1px solid #d0d7de; border-radius: 4px;
  box-shadow: 0 4px 16px rgba(0,0,0,0.12);
  padding: 8px 12px; font-size: 12px; color: #8c959f;
}
</style>