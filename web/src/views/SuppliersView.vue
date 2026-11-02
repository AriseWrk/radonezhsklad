<template>
  <div class="page">
    <div class="ms-title">
      <button class="ms-help" title="Справка"><MsIcon name="help" :size="14" /></button>
      <span>Поставщики</span>
      <button class="ms-refresh" @click="load" title="Обновить"><MsIcon name="refresh" :size="14" /></button>
    </div>

    <div class="ms-toolbar">
      <MsButton variant="primary" icon="plus" @click="openCreate">Поставщик</MsButton>
      <input v-model="search" class="ms-input" placeholder="Наименование, ИНН, телефон, email" />
      <div class="ms-counter">{{ filtered.length }}</div>
    </div>

    <div v-if="error" class="error-box">{{ error }}</div>

    <table class="ms-table2">
      <thead>
        <tr>
          <th>Название</th>
          <th class="col-phone">ИНН</th>
          <th class="col-phone">Телефон</th>
          <th class="col-email">Email</th>
          <th>Адрес</th>
          <th class="col-actions"></th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="loading"><td colspan="6" class="empty">Загрузка...</td></tr>
        <tr v-else-if="filtered.length === 0"><td colspan="6" class="empty">Нет поставщиков</td></tr>
        <tr v-else v-for="s in filtered" :key="s.id" class="row">
          <td class="link" @click="openEdit(s)">{{ s.name }}</td>
          <td class="mono">{{ s.inn || "—" }}</td>
          <td>{{ s.phone || "—" }}</td>
          <td>{{ s.email || "—" }}</td>
          <td class="muted">{{ s.address || "—" }}</td>
          <td class="col-actions" @click.stop>
            <button class="btn-link-ms danger" @click="onDelete(s)">Удалить</button>
          </td>
        </tr>
      </tbody>
    </table>

    <div class="ms-footer2">
      <div class="pager"><span class="range">Всего: {{ filtered.length }}</span></div>
    </div>

    <div v-if="showForm" class="modal-backdrop" @click.self="showForm = false">
      <form class="card modal" @submit.prevent="onSave">
        <h2>{{ editing ? "Поставщик" : "Новый поставщик" }}</h2>
        <div v-if="formError" class="error-box">{{ formError }}</div>
        <label>Название*
          <input v-model="form.name" required />
        </label>
        <label>ИНН
          <input v-model="form.inn" />
        </label>
        <label>Телефон
          <input v-model="form.phone" />
        </label>
        <label>Email
          <input v-model="form.email" type="email" />
        </label>
        <label>Адрес
          <input v-model="form.address" />
        </label>
        <div class="modal-actions">
          <button type="button" @click="showForm = false">Отмена</button>
          <button class="primary" type="submit" :disabled="saving">{{ saving ? "Сохранение..." : "Сохранить" }}</button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue"
import { listSuppliers, createSupplier, updateSupplier, deleteSupplier, type Supplier } from "../api/suppliers"
import { apiErrorMessage } from "../api/client"
import MsButton from "../components/MsButton.vue"
import MsIcon from "../components/MsIcon.vue"

const items = ref<Supplier[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const search = ref("")
const showForm = ref(false)
const editing = ref<Supplier | null>(null)
const saving = ref(false)
const formError = ref<string | null>(null)
const form = reactive({ name: "", inn: "", phone: "", email: "", address: "" })

const filtered = computed(() => {
  const q = search.value.trim().toLowerCase()
  if (!q) return items.value
  return items.value.filter((s) =>
    (s.name ?? "").toLowerCase().includes(q) ||
    (s.inn ?? "").toLowerCase().includes(q) ||
    (s.phone ?? "").toLowerCase().includes(q) ||
    (s.email ?? "").toLowerCase().includes(q)
  )
})

async function load() {
  loading.value = true; error.value = null
  try { items.value = await listSuppliers() }
  catch (e) { error.value = apiErrorMessage(e) }
  finally { loading.value = false }
}

function openCreate() {
  editing.value = null
  form.name = ""; form.inn = ""; form.phone = ""; form.email = ""; form.address = ""
  formError.value = null
  showForm.value = true
}

function openEdit(s: Supplier) {
  editing.value = s
  form.name = s.name; form.inn = s.inn ?? ""; form.phone = s.phone ?? ""; form.email = s.email ?? ""; form.address = s.address ?? ""
  formError.value = null
  showForm.value = true
}

async function onSave() {
  formError.value = null; saving.value = true
  try {
    if (editing.value) await updateSupplier(editing.value.id, form)
    else await createSupplier(form)
    showForm.value = false
    await load()
  } catch (e) { formError.value = apiErrorMessage(e) }
  finally { saving.value = false }
}

async function onDelete(s: Supplier) {
  if (!confirm("Удалить поставщика " + s.name + "?")) return
  try { await deleteSupplier(s.id); await load() }
  catch (e) { error.value = apiErrorMessage(e) }
}

onMounted(load)
</script>

<style scoped>
.col-phone { width: 140px; }
.col-email { width: 200px; }
.col-actions { width: 100px; text-align: right; }
.mono { font-family: ui-monospace, monospace; }
.btn-link-ms { background: none; border: 0; cursor: pointer; color: #2c5d9c; padding: 0; }
.btn-link-ms.danger { color: #b02a37; }
</style>
