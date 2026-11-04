<template>
  <div class="page">
    <div class="header-row">
      <button class="save-btn" :disabled="loading" @click="save">Сохранить</button>
      <div class="meta">
        <div class="meta-line">Изменения: {{ editedBy }}</div>
        <div class="meta-line">{{ editedAt }}</div>
      </div>
    </div>

    <h1 class="page-title">Настройки компании</h1>

    <section class="block">
      <h2 class="block-title">Правила нумерации документов <span class="q">?</span></h2>
      <label class="opt">
        <input type="radio" name="numbering" value="year" v-model="state.numbering" />
        <span>Внутри календарного года</span>
      </label>
      <label class="opt">
        <input type="radio" name="numbering" value="history" v-model="state.numbering" />
        <span>По порядку за всю историю</span>
      </label>
    </section>

    <section class="block">
      <h2 class="block-title">Обратный адрес в письмах</h2>
      <label class="opt">
        <input type="radio" name="reply" value="employee" v-model="state.reply_address" />
        <span>Адрес сотрудника, который отправляет документ</span>
      </label>
      <label class="opt">
        <input type="radio" name="reply" value="common" v-model="state.reply_address" />
        <span>Общий адрес</span>
      </label>
    </section>

    <section class="block">
      <h2 class="block-title">Другие политики</h2>
      <label class="opt">
        <input type="checkbox" v-model="state.forbid_ship_no_stock" />
        <span>Запретить отгрузку товаров, которых нет на складе</span>
      </label>
      <label class="opt">
        <input type="checkbox" v-model="state.auto_min_price" />
        <span>Автоматически устанавливать минимальную цену <span class="q">?</span></span>
      </label>
      <label class="opt">
        <input type="checkbox" v-model="state.move_deleted_to_bin" />
        <span>Перемещать удаленные документы на 7 дней в корзину</span>
      </label>
      <label class="opt">
        <input type="checkbox" v-model="state.use_lots" />
        <span>Использовать партии товаров</span>
      </label>
      <label class="opt">
        <input type="checkbox" v-model="state.show_extra_fields" />
        <span>Включить отображение дополнительных полей товаров и услуг в позициях документов <span class="q">?</span></span>
      </label>
    </section>

    <section class="block">
      <h2 class="block-title">Страна для базовых настроек</h2>
      <p class="block-desc">
        Влияет на настройки шаблонов печатных форм документов,<br/>
        предзаданный набор ставок НДС и предустановленную валюту учёта.
      </p>
      <div class="country">Россия</div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, onMounted } from 'vue'
import { http } from '../api/client'

interface Settings {
  numbering: string
  reply_address: string
  forbid_ship_no_stock: boolean
  auto_min_price: boolean
  move_deleted_to_bin: boolean
  use_lots: boolean
  show_extra_fields: boolean
}

const defaults: Settings = {
  numbering: 'year',
  reply_address: 'employee',
  forbid_ship_no_stock: false,
  auto_min_price: false,
  move_deleted_to_bin: true,
  use_lots: false,
  show_extra_fields: false,
}

const state = reactive<Settings>({ ...defaults })
const editedBy = ref('Козляков А. С.')
const editedAt = ref('18.07.2019 16:44')
const loading = ref(false)

async function load() {
  loading.value = true
  try {
    const { data } = await http.get<Settings & { updated_by_name: string; updated_at: string }>('/company/settings')
    Object.assign(state, data)
    if (data.updated_by_name) editedBy.value = data.updated_by_name
    if (data.updated_at) {
      const d = new Date(data.updated_at)
      editedAt.value = d.toLocaleDateString('ru-RU') + ' ' + d.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })
    }
  } catch (e) {
    console.error('Не удалось загрузить настройки', e)
  } finally {
    loading.value = false
  }
}

async function save() {
  loading.value = true
  try {
    await http.put('/company/settings', state)
    await load()
  } catch (e) {
    console.error('Не удалось сохранить настройки', e)
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.page {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
  font-size: 13px;
  color: #1f2328;
  max-width: 800px; /* Узкая колонка, как в МС */
  padding: 8px 0 32px 16px;
}

.header-row {
  display: flex;
  align-items: center;
  gap: 20px;
  margin-bottom: 12px;
}

.save-btn {
  background-color: #b3d261; /* Светло-зеленый из МС */
  color: #fff;
  border: 1px solid #a5c655;
  border-radius: 3px;
  padding: 5px 16px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  outline: none;
}
.save-btn:hover:not(:disabled) { background-color: #a3c351; }
.save-btn:disabled { opacity: 0.6; cursor: default; }

.meta {
  display: flex;
  flex-direction: column;
  justify-content: center;
  color: #57606a;
  font-size: 11px;
  line-height: 1.4;
}

.page-title {
  font-size: 26px;
  font-weight: 700;
  margin: 16px 0 24px;
}

.block {
  margin-bottom: 28px;
}

.block-title {
  font-size: 15px;
  font-weight: 600;
  margin: 0 0 14px;
  display: flex;
  align-items: center;
  gap: 6px;
}

.block-desc {
  color: #57606a;
  font-size: 12px;
  line-height: 1.4;
  margin: 0 0 10px;
}

.opt {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 4px 0;
  cursor: pointer;
  font-size: 13px;
  line-height: 1.4;
}

/* Выравниваем инпуты по базовой линии текста и запрещаем им сжиматься */
.opt input[type="radio"],
.opt input[type="checkbox"] {
  margin: 1px 0 0 0;
  flex-shrink: 0;
  width: 14px;
  height: 14px;
  cursor: pointer;
}

/* Текст внутри метки не должен сжиматься и переноситься некрасиво */
.opt span {
  flex: 1;
}

.q {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 14px;
  height: 14px;
  border-radius: 50%;
  border: 1px solid #b8c0c8;
  color: #8c959f;
  font-size: 10px;
  font-weight: 700;
  margin-left: 2px;
  vertical-align: text-top;
}

.country {
  font-size: 13px;
  font-weight: 600;
  margin-top: 4px;
}
</style>