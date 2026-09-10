<template>
  <div>
    <h1>Дашборд</h1>
    <div class="cards">
      <div class="card stat">
        <div class="label">Товаров</div>
        <div class="value">{{ productsCount }}</div>
      </div>
      <div class="card stat">
        <div class="label">Категорий</div>
        <div class="value">{{ categoriesCount }}</div>
      </div>
      <div class="card stat">
        <div class="label">Единиц измерения</div>
        <div class="value">{{ unitsCount }}</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { listProducts } from '../api/products'
import { listCategories } from '../api/categories'
import { listUnits } from '../api/units'

const productsCount = ref<number | string>('—')
const categoriesCount = ref<number | string>('—')
const unitsCount = ref<number | string>('—')

onMounted(async () => {
  try {
    const [p, c, u] = await Promise.all([listProducts(true), listCategories(), listUnits()])
    productsCount.value = p.length
    categoriesCount.value = c.length
    unitsCount.value = u.length
  } catch {
    // 401 обработает интерцептор, остальное тихо игнорируем
  }
})
</script>

<style scoped>
h1 { margin-top: 0; }
.cards { display: flex; gap: 16px; flex-wrap: wrap; }
.stat { width: 200px; }
.stat .label { color: var(--muted); font-size: 13px; margin-bottom: 6px; }
.stat .value { font-size: 28px; font-weight: 700; }
</style>