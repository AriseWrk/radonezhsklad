<script setup lang="ts">
import MsIcon from './MsIcon.vue'

withDefaults(defineProps<{
  variant?: 'default' | 'primary' | 'green' | 'ghost' | 'danger' | 'icon'
  icon?: string
  iconRight?: string
  size?: 'sm' | 'md'
  disabled?: boolean
  title?: string
}>(), { variant: 'default', size: 'md' })
</script>

<template>
  <button
    type="button"
    class="ms-btn"
    :class="[`v-${variant}`, `s-${size}`]"
    :disabled="disabled"
    :title="title"
  >
    <MsIcon v-if="icon" :name="icon" :size="size === 'sm' ? 14 : 16" />
    <span v-if="$slots.default" class="lbl"><slot /></span>
    <MsIcon v-if="iconRight" :name="iconRight" :size="size === 'sm' ? 14 : 16" class="icon-right" />
  </button>
</template>

<style scoped>
.ms-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  font-family: inherit;
  font-size: 13px;
  line-height: 1;
  padding: 0 12px;
  height: 30px;
  border-radius: 4px;
  border: 1px solid #d0d7de;
  background: #fff;
  color: #1f2328;
  cursor: pointer;
  white-space: nowrap;
  transition: background .08s, border-color .08s;
}
.ms-btn:hover:not(:disabled) { background: #f6f8fa; }
.ms-btn:disabled { opacity: .55; cursor: default; }
.ms-btn.s-sm { height: 26px; padding: 0 9px; font-size: 12px; gap: 4px; }

.ms-btn.v-primary {
  background: #2c5d9c;
  border-color: #2c5d9c;
  color: #fff;
}
.ms-btn.v-primary:hover:not(:disabled) { background: #234a7d; border-color: #234a7d; }

.ms-btn.v-green {
  background: #a8ce5e;
  border-color: #a8ce5e;
  color: #1f2328;
  font-weight: 600;
}
.ms-btn.v-green:hover:not(:disabled) { background: #97bd4b; border-color: #97bd4b; }

.ms-btn.v-ghost {
  background: transparent;
  border-color: transparent;
  color: #1f2328;
}
.ms-btn.v-ghost:hover:not(:disabled) { background: #f0f2f5; }

.ms-btn.v-danger {
  background: #fff;
  border-color: #cf222e;
  color: #cf222e;
}
.ms-btn.v-danger:hover:not(:disabled) { background: #fff0f0; }

.ms-btn.v-icon {
  width: 30px;
  padding: 0;
  background: transparent;
  border-color: transparent;
  color: #57606a;
}
.ms-btn.v-icon:hover:not(:disabled) { background: #f0f2f5; color: #1f2328; }

.icon-right { margin-left: -1px; }
.lbl { line-height: 1; }
</style>