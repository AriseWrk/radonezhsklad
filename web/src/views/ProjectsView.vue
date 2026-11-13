<template>
  <div class="projects-page">
    <aside class="prj-side">
      <div class="prj-side-head">
        <input
          v-model="search"
          class="prj-search"
          type="text"
          placeholder="Поиск проектов..."
          autocomplete="off"
        />
        <label class="prj-arch">
          <input v-model="showArchived" type="checkbox" />
          <span>Архив</span>
        </label>
      </div>
      <div class="prj-list" v-if="filteredProjects.length">
        <button
          v-for="p in filteredProjects"
          :key="p.id"
          type="button"
          class="prj-item"
          :class="{ active: p.id === selectedId, archived: p.archived }"
          @click="selectProject(p.id)"
          :title="p.name"
        >
          <span class="prj-name">{{ p.name }}</span>
          <span v-if="p.archived" class="prj-badge">архив</span>
        </button>
      </div>
      <div v-else class="prj-empty">
        {{ search ? 'Ничего не найдено' : 'Проектов пока нет' }}
      </div>
    </aside>

    <section class="prj-main">
      <div v-if="!selectedId" class="prj-placeholder">
        Выберите проект слева
      </div>
      <template v-else>
        <div class="prj-header">
          <div class="prj-title">{{ selectedProject?.name || '—' }}</div>
          <div class="prj-tabs">
            <button
              type="button"
              class="prj-tab"
              :class="{ active: tab === 'orders' }"
              @click="tab = 'orders'"
            >Заказы</button>
            <button
              type="button"
              class="prj-tab"
              :class="{ active: tab === 'systems' }"
              @click="tab = 'systems'"
            >По системам</button>
          </div>
        </div>

        <div v-if="loading" class="prj-loading">Загрузка...</div>
        <div v-else-if="error" class="prj-error">{{ error }}</div>

        <template v-else>
          <!-- TAB: Заказы -->
          <div v-if="tab === 'orders'" class="prj-orders">
            <table v-if="orders.length" class="prj-table">
              <thead>
                <tr>
                  <th>Номер</th>
                  <th>Дата</th>
                  <th>Статус</th>
                  <th class="num">Позиций</th>
                  <th class="num">Сумма</th>
                  <th class="num">Отгружено</th>
                  <th>Комментарий</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="o in orders" :key="o.id" class="prj-order-row" @click="openOrder(o.id)">
                  <td class="mono">{{ o.number }}</td>
                  <td>{{ formatDate(o.doc_date) }}</td>
                  <td><span class="status-pill" :class="'st-' + o.status">{{ statusLabel(o.status) }}</span></td>
                  <td class="num">{{ o.item_count }}</td>
                  <td class="num">{{ formatMoney(o.total) }}</td>
                  <td class="num">{{ formatMoney(o.shipped_amount) }}</td>
                  <td class="muted ellipsis" :title="o.comment || ''">{{ o.comment || '' }}</td>
                </tr>
              </tbody>
            </table>
            <div v-else class="prj-empty">По этому проекту ещё нет внутренних заказов</div>
          </div>

          <!-- TAB: По системам -->
          <div v-if="tab === 'systems'" class="prj-systems">
            <div v-for="sys in systemGroups" :key="sys.code || '_none'" class="sys-block">
              <button type="button" class="sys-head" @click="toggle(sys.code)">
                <MsIcon :name="openSys[sys.code] ? 'chevron-down' : 'chevron-right'" :size="14" />
                <span class="sys-name">{{ sys.name }}</span>
                <span class="sys-meta">{{ sys.items.length }} поз. · {{ totalQty(sys.items) }} ед.</span>
              </button>
              <div v-if="openSys[sys.code]" class="sys-body">
                <table class="prj-table">
                  <thead>
                    <tr>
                      <th class="col-sku">Артикул</th>
                      <th>Наименование</th>
                      <th class="num">Кол-во</th>
                      <th>Ед.</th>
                      <th class="num">ВЗ</th>
                      <th>Последняя выдача</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="it in sys.items" :key="it.system_code + ':' + it.product_id">
                      <td class="col-sku mono muted">{{ it.sku || '—' }}</td>
                      <td>{{ it.name || '—' }}</td>
                      <td class="num">{{ formatQty(it.quantity) }}</td>
                      <td class="muted">{{ unitShort(it.unit_id) }}</td>
                      <td class="num">{{ it.order_count }}</td>
                      <td class="muted">{{ it.last_doc_date ? formatDate(it.last_doc_date) : '—' }}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
            <div v-if="!systemGroups.length" class="prj-empty">Нет позиций с указанными системами</div>
          </div>
        </template>
      </template>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import MsIcon from '../components/MsIcon.vue'
import { listProjects, listProjectOrders, listProjectSystems, type Project, type ProjectOrderBrief, type ProjectSystemItem } from '../api/projects'
import { listUnits, type Unit } from '../api/units'

const router = useRouter()

const projects = ref<Project[]>([])
const units = ref<Unit[]>([])
const search = ref('')
const showArchived = ref(false)
const selectedId = ref<string>('')
const tab = ref<'orders' | 'systems'>('orders')

const orders = ref<ProjectOrderBrief[]>([])
const systems = ref<ProjectSystemItem[]>([])
const loading = ref(false)
const error = ref('')

const openSys = reactive<Record<string, boolean>>({})

const filteredProjects = computed(() => {
  const q = search.value.trim().toLowerCase()
  return projects.value
    .filter((p) => showArchived.value || !p.archived)
    .filter((p) => !q || p.name.toLowerCase().includes(q))
    .sort((a, b) => a.name.localeCompare(b.name, 'ru'))
})

const selectedProject = computed(() => projects.value.find((p) => p.id === selectedId.value) || null)

const KNOWN_SYSTEMS: Array<{ code: string; name: string }> = [
  { code: 'SOTS', name: 'СОТС' },
  { code: 'APS',  name: 'АПС' },
  { code: 'SOUE', name: 'СОУЭ' },
  { code: 'SKUD', name: 'СКУД' },
  { code: 'SOT',  name: 'СОТ' },
  { code: 'SKS',  name: 'СКС' },
  { code: '',     name: 'Без системы' },
]

const systemGroups = computed(() => {
  const map = new Map<string, ProjectSystemItem[]>()
  for (const it of systems.value) {
    const key = it.system_code || ''
    if (!map.has(key)) map.set(key, [])
    map.get(key)!.push(it)
  }
  const out: Array<{ code: string; name: string; items: ProjectSystemItem[] }> = []
  for (const s of KNOWN_SYSTEMS) {
    const items = map.get(s.code)
    if (items && items.length) out.push({ code: s.code, name: s.name, items })
  }
  return out
})

function toggle(code: string) {
  openSys[code] = !openSys[code]
}

function totalQty(items: ProjectSystemItem[]) {
  return items.reduce((a, b) => a + b.quantity, 0)
}

function unitShort(id?: string | null) {
  if (!id) return ''
  return units.value.find((u) => u.id === id)?.short_name || ''
}

function formatDate(s: string | null | undefined) {
  if (!s) return ''
  const d = new Date(s)
  return d.toLocaleDateString('ru-RU', { day: '2-digit', month: '2-digit', year: 'numeric' })
}

function formatMoney(n: number) {
  return new Intl.NumberFormat('ru-RU', { maximumFractionDigits: 2 }).format(n || 0)
}

function formatQty(n: number) {
  return new Intl.NumberFormat('ru-RU', { maximumFractionDigits: 3 }).format(n || 0)
}

function statusLabel(s: string) {
  switch (s) {
    case 'draft': return 'Черновик'
    case 'posted': return 'Проведён'
    case 'shipped': return 'Отгружен'
    case 'closed': return 'Закрыт'
    case 'cancelled': return 'Отменён'
    default: return s
  }
}

function openOrder(id: string) {
  router.push(`/internal-orders/${id}`)
}

async function loadStats() {
  if (!selectedId.value) return
  loading.value = true
  error.value = ''
  try {
    const [os, sy] = await Promise.all([
      listProjectOrders(selectedId.value),
      listProjectSystems(selectedId.value),
    ])
    orders.value = os
    systems.value = sy
    // авто-раскрыть первую непустую систему
    openSys[''] = false
    for (const s of KNOWN_SYSTEMS) { if (s.code) openSys[s.code] = false }
  } catch (e: any) {
    error.value = e?.response?.data?.error || e?.message || 'Не удалось загрузить данные проекта'
  } finally {
    loading.value = false
  }
}

function selectProject(id: string) {
  selectedId.value = id
}

watch(selectedId, () => { loadStats() })

onMounted(async () => {
  try {
    const [ps, us] = await Promise.all([
      listProjects(true),
      listUnits().catch(() => [] as Unit[]),
    ])
    projects.value = ps
    units.value = us
    if (!selectedId.value && filteredProjects.value.length) {
      selectedId.value = filteredProjects.value[0].id
    }
  } catch (e: any) {
    error.value = e?.response?.data?.error || e?.message || 'Не удалось загрузить проекты'
  }
})
</script>

<style scoped>
.projects-page {
  display: grid;
  grid-template-columns: 320px 1fr;
  gap: 12px;
  height: calc(100vh - 120px);
  min-height: 400px;
}
.prj-side {
  display: flex;
  flex-direction: column;
  border: 1px solid var(--ms-border, #e5e5e5);
  border-radius: 6px;
  overflow: hidden;
  background: #fff;
}
.prj-side-head {
  padding: 10px;
  border-bottom: 1px solid var(--ms-border, #e5e5e5);
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.prj-search {
  width: 100%;
  padding: 6px 10px;
  border: 1px solid var(--ms-border, #e5e5e5);
  border-radius: 4px;
  font-size: 13px;
  outline: none;
}
.prj-search:focus { border-color: var(--ms-accent, #2b78e4); }
.prj-arch {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #666;
  cursor: pointer;
  user-select: none;
}
.prj-list {
  flex: 1;
  overflow-y: auto;
  padding: 4px;
}
.prj-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 6px;
  width: 100%;
  text-align: left;
  padding: 8px 10px;
  border: none;
  background: transparent;
  border-radius: 4px;
  cursor: pointer;
  font-size: 13px;
  color: #222;
}
.prj-item:hover { background: #f3f6fb; }
.prj-item.active { background: #e7f0fc; color: var(--ms-accent, #2b78e4); font-weight: 500; }
.prj-item.archived .prj-name { color: #999; }
.prj-name { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; flex: 1; }
.prj-badge { font-size: 10px; color: #888; border: 1px solid #ccc; border-radius: 3px; padding: 1px 4px; }
.prj-empty { padding: 20px; text-align: center; color: #999; font-size: 13px; }
.prj-main {
  border: 1px solid var(--ms-border, #e5e5e5);
  border-radius: 6px;
  background: #fff;
  overflow-y: auto;
  padding: 14px 16px;
}
.prj-placeholder { padding: 40px; text-align: center; color: #999; }
.prj-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  margin-bottom: 14px;
  flex-wrap: wrap;
}
.prj-title { font-size: 16px; font-weight: 600; color: #222; }
.prj-tabs { display: flex; gap: 2px; border-bottom: 1px solid var(--ms-border, #e5e5e5); }
.prj-tab {
  padding: 6px 14px;
  border: none;
  background: transparent;
  cursor: pointer;
  font-size: 13px;
  color: #666;
  border-bottom: 2px solid transparent;
}
.prj-tab.active { color: var(--ms-accent, #2b78e4); border-bottom-color: var(--ms-accent, #2b78e4); font-weight: 500; }
.prj-loading, .prj-error { padding: 20px; text-align: center; color: #999; }
.prj-error { color: #c00; }

.prj-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}
.prj-table th {
  text-align: left;
  padding: 6px 8px;
  border-bottom: 1px solid var(--ms-border, #e5e5e5);
  font-weight: 500;
  color: #666;
  font-size: 12px;
}
.prj-table td { padding: 6px 8px; border-bottom: 1px solid #f2f2f2; }
.prj-table .num { text-align: right; white-space: nowrap; }
.prj-table .muted { color: #888; }
.prj-table .mono { font-family: ui-monospace, Menlo, Consolas, monospace; }
.prj-table .ellipsis { max-width: 220px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.prj-table .col-sku { white-space: nowrap; width: 1%; }
.prj-order-row { cursor: pointer; }
.prj-order-row:hover { background: #f3f6fb; }
.status-pill {
  display: inline-block;
  padding: 1px 8px;
  border-radius: 10px;
  font-size: 11px;
  background: #eee;
  color: #555;
}
.status-pill.st-draft { background: #f4f4f4; color: #666; }
.status-pill.st-posted { background: #e7f0fc; color: #2b78e4; }
.status-pill.st-shipped { background: #e6f7ec; color: #1f8a4c; }
.status-pill.st-closed { background: #ecf0f5; color: #444; }
.status-pill.st-cancelled { background: #fbe9e9; color: #c00; }

.prj-systems { display: flex; flex-direction: column; gap: 8px; }
.sys-block { border: 1px solid var(--ms-border, #e5e5e5); border-radius: 6px; overflow: hidden; }
.sys-head {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  padding: 10px 12px;
  background: #fafbfc;
  border: none;
  cursor: pointer;
  text-align: left;
  font-size: 14px;
  font-weight: 500;
}
.sys-head:hover { background: #f3f6fb; }
.sys-name { flex: 1; }
.sys-meta { color: #888; font-weight: 400; font-size: 12px; }
.sys-body { padding: 8px 12px; }
</style>