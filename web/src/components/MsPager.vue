<template>
  <div class="pager" v-if="total > perPage">
    <button :disabled="page === 1" @click="$emit('update:page', 1)" title="Первая">«</button>
    <button :disabled="page === 1" @click="$emit('update:page', page - 1)" title="Назад">‹</button>
    <span class="range">{{ rangeFrom }}-{{ rangeTo }} из {{ total }}</span>
    <button :disabled="rangeTo >= total" @click="$emit('update:page', page + 1)" title="Вперёд">›</button>
    <button :disabled="rangeTo >= total" @click="$emit('update:page', lastPage)" title="Последняя">»</button>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  page: number
  perPage: number
  total: number
}>()

defineEmits<{
  (e: 'update:page', value: number): void
}>()

const lastPage  = computed(() => Math.max(1, Math.ceil(props.total / props.perPage)))
const rangeFrom = computed(() => props.total === 0 ? 0 : (props.page - 1) * props.perPage + 1)
const rangeTo   = computed(() => Math.min(props.page * props.perPage, props.total))
</script>

<style scoped>
.pager { display: flex; align-items: center; gap: 4px; padding: 8px 0; font-size: 12px; }
.pager button { width: 26px; height: 26px; border: 1px solid #d0d7de; background: #fff; border-radius: 4px; cursor: pointer; color: #1f2328; font-size: 13px; line-height: 1; }
.pager button:hover:not(:disabled) { background: #f6f8fa; }
.pager button:disabled { opacity: .45; cursor: default; }
.pager .range { color: #57606a; margin: 0 8px; font-variant-numeric: tabular-nums; }
</style>