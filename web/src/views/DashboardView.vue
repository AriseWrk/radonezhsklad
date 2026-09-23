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
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { listProducts } from '../api/products'
import { listCategories } from '../api/categories'
import { listUnits } from '../api/units'
import MsIcon from '../components/MsIcon.vue'

const productsCount = ref<number | string>('—')
const categoriesCount = ref<number | string>('—')
const unitsCount = ref<number | string>('—')

async function reload() {
  try {
    const [p, c, u] = await Promise.all([listProducts(true), listCategories(), listUnits()])
    productsCount.value = p.length
    categoriesCount.value = c.length
    unitsCount.value = u.length
  } catch { /* ignore */ }
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
</style>