<template>
  <div>
    <!-- Заголовок -->
    <div class="ms-title">
      <button class="ms-help" title="Справка"><MsIcon name="help" :size="14" /></button>
      <span>Сотрудники</span>
      <button class="ms-refresh" @click="load" title="Обновить"><MsIcon name="refresh" :size="14" /></button>
    </div>

    <!-- Тулбар -->
    <div class="ms-toolbar">
      <MsButton variant="primary" icon="user" @click="openCreate">Сотрудник</MsButton>
      <MsButton icon="filter" @click="showFilter = !showFilter">Фильтр</MsButton>
      <select class="ms-select" v-model="filterRole" @change="page = 1">
        <option value="">Все роли</option>
        <option value="admin">Администратор</option>
        <option value="manager">Менеджер</option>
        <option value="warehouse">Кладовщик</option>
        <option value="user">Пользователь</option>
      </select>
      <div class="ms-counter">{{ filtered.length }}</div>
      <MsButton icon="print">Печать</MsButton>
      <MsButton variant="icon" icon="gear" title="Настройки" />
    </div>

    <!-- Фильтр-панель -->
    <div v-if="showFilter" class="ms-filter">
      <div class="filter-actions">
        <MsButton variant="green" @click="page = 1">Найти</MsButton>
        <MsButton @click="clearFilters">Очистить</MsButton>
      </div>
      <div class="filter-grid">
        <div class="filter-field">
          <label class="filter-label"><span class="dot"></span>ФИО, e-mail, телефон или логин</label>
          <input v-model="filterSearch" placeholder="Начните вводить" @input="page = 1" />
        </div>
        <div class="filter-field">
          <label class="filter-label"><span class="dot"></span>Активен</label>
          <select v-model="filterActive" @change="page = 1">
            <option value="">—</option>
            <option value="yes">Да</option>
            <option value="no">Нет</option>
          </select>
        </div>
      </div>
    </div>

    <div v-if="error" class="error-box">{{ error }}</div>

    <!-- Таблица -->
    <table class="ms-table users-table">
      <thead>
        <tr>
          <th class="chk-col"><input type="checkbox" :checked="allChecked" @change="toggleAll" /></th>
          <th class="check-col" title="Вход в систему"></th>
          <th class="avatar-col"></th>
          <th @click="sortBy('last_name')">Фамилия <span v-if="sortKey === 'last_name'">{{ sortDir === 'asc' ? '↑' : '↓' }}</span></th>
          <th>Имя</th>
          <th>Отчество</th>
          <th>E-mail</th>
          <th>Телефон</th>
          <th>Логин</th>
          <th>Описание</th>
          <th>Роль</th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="loading"><td colspan="11" class="muted" style="text-align:center;padding:24px">Загрузка...</td></tr>
        <tr v-else-if="filtered.length === 0"><td colspan="11" class="muted" style="text-align:center;padding:24px">Нет сотрудников</td></tr>
        <tr v-for="u in paginated" :key="u.id" class="clickable" @click="openEdit(u)">
          <td class="chk-col" @click.stop>
            <input type="checkbox" :checked="selected.has(u.id)" @change="toggleSelect(u.id)" />
          </td>
          <td class="check-col">
            <span v-if="u.is_active" class="check-on" title="Активен">✓</span>
            <span v-else class="check-off" title="Отключён">—</span>
          </td>
          <td class="avatar-col">
            <div class="avatar">{{ initials(u) }}</div>
          </td>
          <td>{{ u.last_name || '—' }}</td>
          <td>{{ u.first_name || '—' }}</td>
          <td>{{ u.middle_name || '—' }}</td>
          <td>{{ u.email }}</td>
          <td>{{ u.phone || '—' }}</td>
          <td>{{ u.login || '—' }}</td>
          <td class="muted">{{ u.description || '—' }}</td>
          <td>{{ roleLong(u.role) }}</td>
        </tr>
      </tbody>
    </table>

    <!-- Футер -->
    <div class="ms-footer">
      <div class="ms-pager">
        <button :disabled="page === 1" @click="page--">◀</button>
        <button :disabled="page === 1" @click="page--">↤</button>
        <span>{{ rangeFrom }}–{{ rangeTo }} из {{ filtered.length }}</span>
        <button :disabled="rangeTo >= filtered.length" @click="page++">↦</button>
        <button :disabled="rangeTo >= filtered.length" @click="page++">▶</button>
      </div>
    </div>

    <!-- Модалка создания/редактирования -->
    <div v-if="modal.open" class="modal-backdrop" @click.self="closeModal">
      <form class="card modal big" @submit.prevent="onSubmit">
        <h2>{{ modal.isEdit ? 'Карточка сотрудника' : 'Новый сотрудник' }}</h2>
        <div v-if="modal.error" class="error-box">{{ modal.error }}</div>

        <div class="grid2">
          <label>Фамилия*
            <input v-model="form.last_name" required />
          </label>
          <label>Имя*
            <input v-model="form.first_name" required />
          </label>
          <label>Отчество
            <input v-model="form.middle_name" />
          </label>
          <label>Телефон
            <input v-model="form.phone" placeholder="+7 999 123-45-67" />
          </label>
        </div>

        <label>E-mail*
          <input v-model="form.email" type="email" required :disabled="modal.isEdit" />
        </label>

        <div class="grid2">
          <label>Логин
            <input v-model="form.login" placeholder="(по умолчанию = email до @)" />
          </label>
          <label>Пароль {{ modal.isEdit ? '(оставьте пустым, чтобы не менять)' : '*' }}
            <input v-model="form.password" type="password" :required="!modal.isEdit" :minlength="8" />
          </label>
        </div>

        <label>Роль*
          <select v-model="form.role" required>
            <option value="admin">Администратор</option>
            <option value="manager">Менеджер</option>
            <option value="warehouse">Кладовщик</option>
            <option value="user">Пользователь (индивидуальные настройки)</option>
          </select>
        </label>

        <label>Описание
          <textarea v-model="form.description" rows="2"></textarea>
        </label>

        <label v-if="modal.isEdit" class="chk">
          <input type="checkbox" v-model="form.is_active" /> Вход разрешён
        </label>

        <div class="modal-actions">
          <button type="button" class="danger" v-if="modal.isEdit && form.email !== 'admin@radonezh.local'" @click="onArchive">Отключить</button>
          <div style="flex:1"></div>
          <button type="button" @click="closeModal">Отмена</button>
          <button class="primary" type="submit" :disabled="modal.saving">
            {{ modal.saving ? 'Сохранение...' : 'Сохранить' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { listUsers, createUser, updateUser, updateActive, type User, type UserInput } from '../api/users'
import { apiErrorMessage } from '../api/client'
import MsButton from '../components/MsButton.vue'
import MsIcon from '../components/MsIcon.vue'

const items = ref<User[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const page = ref(1)
const perPage = 25

const showFilter = ref(false)
const filterSearch = ref('')
const filterRole = ref('')
const filterActive = ref('')

const selected = ref<Set<string>>(new Set())

const sortKey = ref<'last_name' | 'first_name'>('last_name')
const sortDir = ref<'asc' | 'desc'>('asc')

const modal = reactive({
  open: false,
  isEdit: false,
  saving: false,
  error: null as string | null,
  editingId: '' as string,
})

const form = reactive({
  last_name: '',
  first_name: '',
  middle_name: '',
  phone: '',
  email: '',
  login: '',
  password: '',
  description: '',
  role: 'user' as string,
  is_active: true,
})

function roleLong(r: string) {
  return {
    admin: 'Администратор',
    manager: 'Менеджер',
    warehouse: 'Кладовщик',
    user: 'Пользователь (индивидуальные настройки)',
  }[r] ?? r
}

function initials(u: User) {
  return ((u.last_name[0] ?? '') + (u.first_name[0] ?? '')).toUpperCase() || '👤'
}

function sortBy(k: typeof sortKey.value) {
  if (sortKey.value === k) sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  else { sortKey.value = k; sortDir.value = 'asc' }
}

const filtered = computed(() => {
  let rows = items.value
  if (filterSearch.value) {
    const q = filterSearch.value.toLowerCase()
    rows = rows.filter((u) =>
      [u.last_name, u.first_name, u.middle_name, u.email, u.phone ?? '', u.login ?? '']
        .some((s) => s.toLowerCase().includes(q))
    )
  }
  if (filterRole.value) rows = rows.filter((u) => u.role === filterRole.value)
  if (filterActive.value === 'yes') rows = rows.filter((u) => u.is_active)
  if (filterActive.value === 'no') rows = rows.filter((u) => !u.is_active)

  return [...rows].sort((a, b) => {
    const av = a[sortKey.value].toLowerCase()
    const bv = b[sortKey.value].toLowerCase()
    return sortDir.value === 'asc' ? av.localeCompare(bv, 'ru') : bv.localeCompare(av, 'ru')
  })
})

const paginated = computed(() => {
  const from = (page.value - 1) * perPage
  return filtered.value.slice(from, from + perPage)
})

const rangeFrom = computed(() => filtered.value.length === 0 ? 0 : (page.value - 1) * perPage + 1)
const rangeTo = computed(() => Math.min(page.value * perPage, filtered.value.length))

const allChecked = computed(() => paginated.value.length > 0 && paginated.value.every((u) => selected.value.has(u.id)))

function toggleAll(e: Event) {
  const on = (e.target as HTMLInputElement).checked
  selected.value = on ? new Set(paginated.value.map((u) => u.id)) : new Set()
}
function toggleSelect(id: string) {
  const s = new Set(selected.value)
  s.has(id) ? s.delete(id) : s.add(id)
  selected.value = s
}

function clearFilters() {
  filterSearch.value = ''
  filterRole.value = ''
  filterActive.value = ''
  page.value = 1
}

async function load() {
  loading.value = true
  error.value = null
  try { items.value = await listUsers() }
  catch (e) { error.value = apiErrorMessage(e) }
  finally { loading.value = false }
}

function openCreate() {
  Object.assign(form, {
    last_name: '', first_name: '', middle_name: '', phone: '',
    email: '', login: '', password: '', description: '',
    role: 'user', is_active: true,
  })
  modal.open = true; modal.isEdit = false; modal.error = null; modal.editingId = ''
}

function openEdit(u: User) {
  Object.assign(form, {
    last_name: u.last_name, first_name: u.first_name, middle_name: u.middle_name,
    phone: u.phone ?? '', email: u.email, login: u.login ?? '',
    password: '', description: u.description ?? '',
    role: u.role, is_active: u.is_active,
  })
  modal.open = true; modal.isEdit = true; modal.error = null; modal.editingId = u.id
}

function closeModal() { modal.open = false }

async function onSubmit() {
  modal.error = null
  modal.saving = true
  try {
    const payload: UserInput = {
      email: form.email,
      last_name: form.last_name,
      first_name: form.first_name,
      middle_name: form.middle_name || undefined,
      phone: form.phone || undefined,
      login: form.login || undefined,
      description: form.description || undefined,
      role: form.role as any,
    }
    if (modal.isEdit) {
      await updateUser(modal.editingId, payload)
      const u = items.value.find((x) => x.id === modal.editingId)
      if (u && u.is_active !== form.is_active) {
        await updateActive(modal.editingId, form.is_active)
      }
    } else {
      await createUser({ ...payload, password: form.password })
    }
    closeModal()
    await load()
  } catch (e) {
    modal.error = apiErrorMessage(e)
  } finally {
    modal.saving = false
  }
}

async function onArchive() {
  if (!confirm('Отключить вход сотруднику?')) return
  try {
    await updateActive(modal.editingId, false)
    closeModal()
    await load()
  } catch (e) { modal.error = apiErrorMessage(e) }
}

onMounted(load)
</script>

<style scoped>
.users-table .chk-col  { width: 32px; }
.users-table .check-col { width: 32px; text-align: center; }
.users-table .avatar-col { width: 42px; }
.users-table tr.clickable { cursor: pointer; }

.check-on  { color: #1a7f37; font-weight: 700; }
.check-off { color: #8c959f; }

.avatar {
  width: 28px; height: 28px; border-radius: 50%;
  background: #d8e4f0; color: #2c5d9c;
  display: flex; align-items: center; justify-content: center;
  font-size: 11px; font-weight: 700;
}

.role-filter {
  padding: 6px 10px; font-size: 13px;
  border: 1px solid #d0d7de; border-radius: 4px;
  background: #fff; color: #1f2328;
  max-width: 200px;
}

.link-help {
  color: #2c5d9c; font-size: 13px;
  margin-left: 8px;
}
.link-help:hover { text-decoration: underline; }

.modal.big { width: 720px; max-width: 100%; }
.grid2 { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }
label.chk { flex-direction: row; align-items: center; gap: 8px; color: #1f2328; }
label.chk input { width: auto; }

/* === МойСклад-стиль === */
.ms-title { display: flex; align-items: center; gap: 8px; font-size: 20px; font-weight: 600; color: #1f2328; margin-bottom: 12px; }
.ms-help { width: 20px; height: 20px; border-radius: 50%; border: 1px solid #b8c0c8; background: transparent; color: #57606a; cursor: pointer; display: inline-flex; align-items: center; justify-content: center; padding: 0; }
.ms-help:hover { background: #f0f2f5; }
.ms-refresh { width: 24px; height: 24px; padding: 0; display: inline-flex; align-items: center; justify-content: center; border: none; background: transparent; color: #57606a; cursor: pointer; }
.ms-refresh:hover { color: #2c5d9c; }
.ms-toolbar { display: flex; align-items: center; gap: 6px; margin-bottom: 12px; flex-wrap: wrap; }
.ms-select { height: 30px; padding: 0 8px; font-size: 13px; border: 1px solid #d0d7de; border-radius: 4px; background: #fff; color: #1f2328; max-width: 180px; }
.ms-counter { min-width: 40px; height: 30px; padding: 0 10px; display: inline-flex; align-items: center; justify-content: center; border: 1px solid #d0d7de; border-radius: 4px; background: #fff; color: #8c959f; font-variant-numeric: tabular-nums; font-size: 13px; }
.ms-filter { background: #eef1f5; border: 1px solid #d8dee4; border-radius: 4px; padding: 10px 14px 12px; margin-bottom: 12px; }
.filter-actions { display: flex; align-items: center; gap: 6px; margin-bottom: 10px; }
.filter-grid { display: grid; grid-template-columns: repeat(6, minmax(0, 1fr)); gap: 10px 14px; }
.filter-field { display: flex; flex-direction: column; gap: 3px; }
.filter-label { font-size: 12px; color: #57606a; display: flex; align-items: center; gap: 5px; }
.filter-label .dot { width: 8px; height: 8px; border-radius: 50%; background: #2c5d9c; flex-shrink: 0; }
.filter-field input, .filter-field select { height: 28px; padding: 0 8px; font-size: 12px; border: 1px solid #d0d7de; border-radius: 3px; background: #fff; color: #1f2328; }
</style>
