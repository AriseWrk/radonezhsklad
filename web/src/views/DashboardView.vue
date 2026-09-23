<template>
  <div class="page">
    <img src="/logo.png" class="page-watermark" alt="" aria-hidden="true" />
    <div class="page-content">
    <div class="ms-title">
      <button class="ms-help" title="Справка"><MsIcon name="help" :size="14" /></button>
      <span>Показатели</span>
      <button class="ms-refresh" @click="reload" title="Обновить"><MsIcon name="refresh" :size="14" /></button>
    </div>

    <div class="kpi-grid">
      <div class="kpi-card">
        <div class="kpi-icon"><MsIcon name="package" :size="28" /></div>
        <div class="kpi-body">
          <div class="kpi-label">Товаров</div>
          <div class="kpi-value">{{ productsCount }}</div>
        </div>
      </div>
      <div class="kpi-card">
        <div class="kpi-icon"><MsIcon name="package" :size="28" /></div>
        <div class="kpi-body">
          <div class="kpi-label">Категорий</div>
          <div class="kpi-value">{{ categoriesCount }}</div>
        </div>
      </div>
      <div class="kpi-card">
        <div class="kpi-icon"><MsIcon name="building" :size="28" /></div>
        <div class="kpi-body">
          <div class="kpi-label">Единиц измерения</div>
          <div class="kpi-value">{{ unitsCount }}</div>
        </div>
      </div>
    </div>
    <div class="dash-section">
      <div class="dash-section-title">Состояние сервисов</div>
      <div class="svc-grid">
        <div v-for="s in services" :key="s.name" class="svc-card" :class="'svc-' + s.status">
          <span class="svc-dot"></span>
          <span class="svc-name">{{ s.name }}</span>
          <span class="svc-latency">{{ s.latency_ms }} ms</span>
          <span class="svc-status">{{ s.status }}</span>
        </div>
      </div>
    </div>

    <div class="dash-section">
      <div class="dash-section-title">Задачи на будущее</div>
      <div class="todo-grid">
        <div v-for="t in todos" :key="t.title" class="todo-card">
          <span class="todo-tag" :class="'tag-' + t.tag">{{ t.tagLabel }}</span>
          <div class="todo-title">{{ t.title }}</div>
          <div class="todo-desc">{{ t.desc }}</div>
        </div>
      </div>
    </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { listProducts } from '../api/products'
import { listCategories } from '../api/categories'
import { listUnits } from '../api/units'
import MsIcon from '../components/MsIcon.vue'
import { http } from '../api/client'

const productsCount = ref<number | string>('—')
const categoriesCount = ref<number | string>('—')
const unitsCount = ref<number | string>('—')
const services = ref<Array<{ name: string; status: string; latency_ms: number }>>([])

const todos = ref([
  { tag: 'fix',  tagLabel: 'fix',  title: 'Владелец — ФИО', desc: 'Сейчас в карточке заказа показывается ID (e931e0a3). Дотянуть /auth/me до full_name и вывести в карточке.' },
  { tag: 'ui',   tagLabel: 'ui',   title: 'Адрес под проектом', desc: 'Поле Проект как в МС: адрес серой строкой под названием проекта.' },
  { tag: 'back', tagLabel: 'back', title: 'Синк address проектов', desc: 'Проверить, отдаёт ли /entity/project поле address, и синкать его в модель Project.' },
  { tag: 'ui',   tagLabel: 'ui',   title: 'Водяной знак: настройка', desc: 'Проверить opacity на разных мониторах, при необходимости — настройка яркости.' },
  { tag: 'idea', tagLabel: 'idea', title: 'Роль viewer', desc: 'Read-only роль для внешних пользователей (без права править заказы и документы).' },
])

async function loadHealth() {
  try {
    const { data } = await http.get<{ services: Array<{ name: string; status: string; latency_ms: number }> }>('/health/all')
    services.value = data.services
  } catch { /* ignore */ }
}

async function reload() {
  try {
    const [p, c, u] = await Promise.all([listProducts(true), listCategories(), listUnits()])
    productsCount.value = p.length
    categoriesCount.value = c.length
    unitsCount.value = u.length
  } catch { /* ignore */ }
  await loadHealth()
}

onMounted(reload)
</script>

<style scoped>
.page { font-size: 13px; position: relative; min-height: calc(100vh - 120px); }
.page-watermark {
  position: fixed;
  top: 0; right: 0; bottom: 0; left: 0;
  width: 100vw; height: 100vh;
  object-fit: contain;
  opacity: 0.05;
  pointer-events: none;
  user-select: none;
  z-index: 0;
}
.page-content { position: relative; z-index: 1; }
.ms-title { display: flex; align-items: center; gap: 8px; font-size: 20px; font-weight: 600; color: #1f2328; margin-bottom: 16px; }
.ms-help { width: 20px; height: 20px; border-radius: 50%; border: 1px solid #b8c0c8; background: transparent; color: #57606a; cursor: pointer; display: inline-flex; align-items: center; justify-content: center; padding: 0; }
.ms-help:hover { background: #f0f2f5; }
.ms-refresh { width: 24px; height: 24px; padding: 0; display: inline-flex; align-items: center; justify-content: center; border: none; background: transparent; color: #57606a; cursor: pointer; }
.ms-refresh:hover { color: #2c5d9c; }

.kpi-grid { display: grid; grid-template-columns: repeat(3, minmax(0, 260px)); gap: 14px; }
.kpi-card {
  background: #fff; border: 1px solid #eaeef2; border-radius: 4px;
  padding: 16px 20px; display: flex; align-items: center; gap: 16px;
}
.kpi-icon {
  width: 48px; height: 48px; border-radius: 50%;
  background: #eaf3ff; color: #2c5d9c;
  display: inline-flex; align-items: center; justify-content: center;
  flex-shrink: 0;
}
.kpi-body { display: flex; flex-direction: column; gap: 2px; }
.kpi-label { font-size: 12px; color: #8c959f; text-transform: uppercase; letter-spacing: 0.3px; }
.kpi-value { font-size: 26px; font-weight: 600; color: #1f2328; font-variant-numeric: tabular-nums; line-height: 1.1; }

.dash-section { margin-top: 24px; }
.dash-section-title { font-size: 12px; font-weight: 600; color: #57606a; margin-bottom: 10px; text-transform: uppercase; letter-spacing: 0.4px; }
.svc-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(220px, 1fr)); gap: 10px; }
.svc-card { display: flex; align-items: center; gap: 10px; padding: 10px 14px; background: #fff; border: 1px solid #eaeef2; border-radius: 4px; font-size: 13px; }
.svc-dot { width: 8px; height: 8px; border-radius: 50%; background: #8c959f; flex-shrink: 0; }
.svc-ok .svc-dot { background: #1a7f37; box-shadow: 0 0 0 3px rgba(26,127,55,0.15); }
.svc-degraded .svc-dot { background: #d4a017; box-shadow: 0 0 0 3px rgba(212,160,23,0.15); }
.svc-down .svc-dot { background: #cf222e; box-shadow: 0 0 0 3px rgba(207,34,46,0.15); }
.svc-name { flex: 1; color: #1f2328; text-transform: capitalize; }
.svc-latency { color: #8c959f; font-variant-numeric: tabular-nums; font-size: 12px; }
.svc-status { font-size: 11px; text-transform: uppercase; letter-spacing: 0.3px; color: #57606a; }
.svc-ok .svc-status { color: #1a7f37; }
.svc-degraded .svc-status { color: #d4a017; }
.svc-down .svc-status { color: #cf222e; }

.todo-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 10px; }
.todo-card { background: #fff; border: 1px solid #eaeef2; border-radius: 4px; padding: 12px 14px; font-size: 13px; }
.todo-tag { display: inline-block; font-size: 10px; text-transform: uppercase; letter-spacing: 0.4px; padding: 2px 6px; border-radius: 3px; background: #eaf3ff; color: #2c5d9c; margin-bottom: 6px; font-weight: 600; }
.todo-tag.tag-ui { background: #e6f4ea; color: #1a7f37; }
.todo-tag.tag-back { background: #fff4e5; color: #a85400; }
.todo-tag.tag-idea { background: #f3e8ff; color: #6e40c9; }
.todo-title { font-weight: 600; color: #1f2328; margin-bottom: 4px; }
.todo-desc { font-size: 12px; color: #57606a; line-height: 1.4; }
</style>