<script setup lang="ts">
withDefaults(defineProps<{ name: string; size?: number }>(), { size: 16 })

interface IconDef { paths?: string[]; circles?: Array<[number, number, number]>; rects?: Array<[number, number, number, number, number]> }

const ICONS: Record<string, IconDef> = {
  // Действия
  plus:        { paths: ['M8 3.5v9', 'M3.5 8h9'] },
  minus:       { paths: ['M3.5 8h9'] },
  check:       { paths: ['M3.5 8.5l3 3 6-6.5'] },
  close:       { paths: ['M4 4l8 8', 'M12 4l-8 8'] },
  edit:        { paths: ['M11 2.5l2.5 2.5-8 8H3v-2.5z'] },
  copy:        { paths: ['M5 5h7v8H5z', 'M4 3h7v1.5', 'M4 3v9h1'] },
  trash:       { paths: ['M3.5 5h9', 'M5.5 5V3.5h5V5', 'M4.5 5l.5 8h6l.5-8'] },
  save:        { paths: ['M3.5 3.5h7l2 2v7h-9z', 'M5.5 3.5v4h5v-4', 'M5 11h6'] },
  post:        { paths: ['M8 13.5a5.5 5.5 0 1 0 0-11 5.5 5.5 0 0 0 0 11z', 'M5.5 8.2l1.8 1.8 3.5-3.7'] },
  send:        { paths: ['M2.5 8l11-5-3 11-3-4z', 'M10.5 3l-4 6'] },
  print:       { paths: ['M4.5 6V3h7v3', 'M3.5 6h9v5h-2v2h-5v-2h-2z', 'M5 9h6'] },
  refresh:     { paths: ['M13 8a5 5 0 0 1-9 3', 'M3 8a5 5 0 0 1 9-3', 'M12 3v3h-3', 'M4 13v-3h3'] },

  // Навигация
  search:      { paths: ['M6.5 11a4.5 4.5 0 1 0 0-9 4.5 4.5 0 0 0 0 9z', 'M10 10l3 3'] },
  filter:      { paths: ['M2.5 3.5h11', 'M4.5 8h7', 'M6.5 12.5h3'] },
  settings:    { circles: [[8, 8, 2]], paths: ['M8 1.5v2', 'M8 12.5v2', 'M1.5 8h2', 'M12.5 8h2', 'M3.3 3.3l1.4 1.4', 'M11.3 11.3l1.4 1.4', 'M12.7 3.3l-1.4 1.4', 'M4.7 11.3l-1.4 1.4'] },
  help:        { circles: [[8, 8, 6]], paths: ['M6 6.2a2 2 0 1 1 2.7 1.9c-.5.2-.7.6-.7 1.1v.4', 'M8 11.6v.01'] },
  bookmark:    { paths: ['M4 2.5h8v11l-4-3-4 3z'] },

  // Стрелки/шевроны
  'chevron-down':  { paths: ['M3.5 6l4.5 4.5L12.5 6'] },
  'chevron-up':    { paths: ['M3.5 10l4.5-4.5L12.5 10'] },
  'chevron-left':  { paths: ['M10 3.5L5.5 8l4.5 4.5'] },
  'chevron-right': { paths: ['M6 3.5L10.5 8 6 12.5'] },
  'arrow-down':    { paths: ['M8 3v9', 'M4.5 8.5L8 12l3.5-3.5'] },
  'arrow-up':      { paths: ['M8 13V4', 'M4.5 7.5L8 4l3.5 3.5'] },

  // Разное
  dots:        { circles: [[8, 3.5, 0.9], [8, 8, 0.9], [8, 12.5, 0.9]] },
  menu:        { paths: ['M2.5 4.5h11', 'M2.5 8h11', 'M2.5 11.5h11'] },
  bell:        { paths: ['M4 10.5V7a4 4 0 0 1 8 0v3.5l1 1.5H3z', 'M6.5 13.5a1.5 1.5 0 0 0 3 0'] },
  chat:        { paths: ['M3 4h10v6H8l-3 2.5V10H3z'] },
  cart:        { paths: ['M2 3h2l1.5 7h7l1.5-5h-9', 'M6 12.5a.75.75 0 1 0 1.5 0 .75.75 0 0 0-1.5 0', 'M10 12.5a.75.75 0 1 0 1.5 0 .75.75 0 0 0-1.5 0'] },
  excel:       { paths: ['M3.5 2.5h6l3 3v8h-9z', 'M6 7l4 4', 'M10 7l-4 4', 'M9.5 2.5v3h3'] },
  user:        { circles: [[8, 6, 2.5]], paths: ['M3.5 13.5a4.5 4.5 0 0 1 9 0'] },
  package:     { paths: ['M2.5 5l5.5-2.5L13.5 5 8 7.5z', 'M2.5 5v6L8 13.5V7.5', 'M13.5 5v6L8 13.5'] },
  building:    { paths: ['M3.5 13.5v-10h6v10', 'M9.5 6h3v7.5', 'M5.5 6h2', 'M5.5 8.5h2', 'M5.5 11h2'] },
  gear:        { circles: [[8, 8, 2.2]], paths: ['M8 1.5v2', 'M8 12.5v2', 'M1.5 8h2', 'M12.5 8h2', 'M3.3 3.3l1.4 1.4', 'M11.3 11.3l1.4 1.4', 'M12.7 3.3l-1.4 1.4', 'M4.7 11.3l-1.4 1.4'] },
  star:        { paths: ['M8 2.5l1.7 3.5 3.8.5-2.8 2.7.7 3.8L8 11.3 4.6 13l.7-3.8L2.5 6.5l3.8-.5z'] },
  clock:       { circles: [[8, 8, 5.5]], paths: ['M8 4.5V8l2.5 1.5'] },
  eye:         { paths: ['M2 8s2.5-4 6-4 6 4 6 4-2.5 4-6 4-6-4-6-4z'], circles: [[8, 8, 2]] },
  lock:        { paths: ['M4.5 7v-2a3.5 3.5 0 0 1 7 0v2', 'M3.5 7h9v6.5h-9z'], circles: [[8, 10, 0.8]] },
}
</script>

<template>
  <svg
    :width="size" :height="size"
    viewBox="0 0 16 16"
    fill="none"
    stroke="currentColor"
    stroke-width="1.4"
    stroke-linecap="round"
    stroke-linejoin="round"
    aria-hidden="true"
    class="ms-icon"
  >
    <template v-if="ICONS[name]">
      <circle
        v-for="(c, i) in ICONS[name].circles ?? []"
        :key="'c' + i"
        :cx="c[0]" :cy="c[1]" :r="c[2]"
        :fill="c[2] < 1 ? 'currentColor' : 'none'"
        :stroke="c[2] < 1 ? 'none' : 'currentColor'"
      />
      <rect
        v-for="(r, i) in ICONS[name].rects ?? []"
        :key="'r' + i"
        :x="r[0]" :y="r[1]" :width="r[2]" :height="r[3]" :rx="r[4]"
      />
      <path v-for="(d, i) in (ICONS[name].paths ?? [])" :key="'p' + i" :d="d" />
    </template>
  </svg>
</template>

<style scoped>
.ms-icon { display: inline-block; vertical-align: -2px; flex-shrink: 0; }
</style>