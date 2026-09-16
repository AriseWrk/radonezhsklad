<template>
  <div class="contract-card">
    <!-- Верхний toolbar -->
    <div class="toolbar">
      <button class="btn success" @click="save" :disabled="saving">
        {{ saving ? 'Сохранение...' : 'Сохранить' }}
      </button>
      <button class="btn" @click="close">Закрыть</button>

      <div class="pager" v-if="neighbors">
        <span class="pager-count">{{ neighbors.index }} из {{ neighbors.total }}</span>
        <button class="pager-btn" :disabled="!neighbors.prev_id" @click="goTo(neighbors.prev_id)">
          <span>◀</span>
        </button>
        <button class="pager-btn" :disabled="!neighbors.next_id" @click="goTo(neighbors.next_id)">
          <span>▶</span>
        </button>
      </div>

      <div class="toolbar-right">
        <button class="btn" @click="onArchive">
          {{ form.archived ? 'Вернуть из архива' : 'Поместить в архив' }}
        </button>
        <select class="mini-select">
          <option>Изменить</option>
        </select>
        <button class="btn">Печать</button>
        <select class="mini-select">
          <option>Отправить</option>
        </select>
      </div>

      <div class="owner-block">
        <div class="owner-name">{{ ownerLabel }}</div>
        <div class="owner-sub muted">Основной ▾</div>
      </div>
      <div class="changed-block">
        <div class="owner-name">{{ changerLabel }}</div>
        <div class="owner-sub muted">{{ formatDate(contract?.updated_at) }}</div>
      </div>
    </div>

    <div v-if="error" class="error-box">{{ error }}</div>

    <!-- Заголовок -->
    <div class="doc-header">
      <div class="doc-title">
        <span class="doc-label">Договор №</span>
        <input v-model="form.number" class="num-input" />
        <span class="muted">от</span>
        <input v-model="form.doc_date" type="datetime-local" class="date-input" />
        <select class="status-select" v-model="form.archived">
          <option :value="false">Статус</option>
          <option :value="true">Архивный</option>
        </select>
      </div>
    </div>

    <!-- Поля -->
    <div class="fields">
      <div class="field-left">
        <div class="field">
          <label>Организация *</label>
          <div class="combo">
            <select v-model="form.organization_id">
              <option value="">—</option>
              <option v-for="o in organizations" :key="o.id" :value="o.id">{{ o.name }}</option>
            </select>
            <span class="edit-icon" title="Редактировать">✎</span>
          </div>
        </div>

        <div class="field">
          <label>Тип договора</label>
          <select v-model="form.contract_type">
            <option value="Договор купли-продажи">Договор купли-продажи</option>
            <option value="Договор оказания услуг">Договор оказания услуг</option>
            <option value="Договор поставки">Договор поставки</option>
            <option value="Договор подряда">Договор подряда</option>
            <option value="Договор аренды">Договор аренды</option>
            <option value="Счёт-оферта">Счёт-оферта</option>
          </select>
        </div>

        <div class="field">
          <label>Код</label>
          <input v-model="form.code" />
        </div>

        <div class="field">
          <label>Сумма договора</label>
          <input v-model.number="form.amount" type="number" step="0.01" min="0" />
        </div>

        <div class="field">
          <label>Комментарий</label>
          <textarea v-model="form.comment" rows="5"></textarea>
        </div>
      </div>

      <div class="field-right">
        <div class="field">
          <label>Контрагент *</label>
          <div class="combo">
            <select v-model="form.customer_id">
              <option value="">—</option>
              <option v-for="c in customers" :key="c.id" :value="c.id">{{ c.name }}</option>
            </select>
            <span class="edit-icon" title="Редактировать">✎</span>
          </div>
        </div>

        <div class="field" v-if="form.customer_id">
          <label class="balance-label">Баланс (Мы должны):</label>
          <div class="balance-value">{{ formatMoney(customerBalance) }} руб</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  getContract, createContract, updateContract, getContractNeighbors,
  type Contract, type ContractNeighbors,
} from '../api/contracts'
import { listCustomers, type Customer } from '../api/customers'
import { listOrganizations, type Organization } from '../api/suppliers'
import { apiErrorMessage } from '../api/client'
import { useAuthStore } from '../stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const isNew = computed(() => route.params.id === 'new')

const contract = ref<Contract | null>(null)
const neighbors = ref<ContractNeighbors | null>(null)
const customers = ref<Customer[]>([])
const organizations = ref<Organization[]>([])

const saving = ref(false)
const error = ref<string | null>(null)

const form = reactive({
  number: '',
  contract_type: 'Договор купли-продажи',
  code: '',
  doc_date: '',
  customer_id: '',
  organization_id: '',
  amount: 0,
  currency: 'RUB',
  paid: 0,
  fulfilled: 0,
  comment: '',
  archived: false,
})

const currentId = ref<string | null>(null)

const ownerLabel = computed(() => auth.userId ? auth.userId.slice(0, 8) : '')
const changerLabel = computed(() => auth.userId ? auth.userId.slice(0, 8) : '')
const customerBalance = computed(() => {
  // Простая эвристика: сумма всех договоров этого контрагента минус оплачено.
  // В реальном проекте — запрос к API.
  return form.amount - form.paid
})

function formatMoney(n: number) {
  return n.toLocaleString('ru-RU', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}
function formatDate(s?: string) {
  if (!s) return ''
  const d = new Date(s)
  return d.toLocaleDateString('ru-RU') + ' ' + d.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })
}

async function loadNeighbors(id: string) {
  try {
    neighbors.value = await getContractNeighbors(id)
  } catch { neighbors.value = null }
}

async function load() {
  error.value = null
  try {
    const [custs, orgs] = await Promise.all([
      listCustomers(false).catch(() => []),
      listOrganizations().catch(() => []),
    ])
    customers.value = custs
    organizations.value = orgs

    if (isNew.value) {
      form.number = 'ДГ-' + new Date().getTime().toString().slice(-6)
      form.doc_date = new Date().toISOString().slice(0, 16)
      form.contract_type = 'Договор купли-продажи'
      currentId.value = null
      neighbors.value = null
    } else {
      const id = route.params.id as string
      currentId.value = id
      const c = await getContract(id)
      contract.value = c
      form.number = c.number
      form.contract_type = c.contract_type || 'Договор купли-продажи'
      form.code = c.code ?? ''
      form.doc_date = c.doc_date.slice(0, 16)
      form.customer_id = c.customer_id ?? ''
      form.organization_id = c.organization_id ?? ''
      form.amount = c.amount
      form.currency = c.currency
      form.paid = c.paid
      form.fulfilled = c.fulfilled
      form.comment = c.comment ?? ''
      form.archived = c.archived
      await loadNeighbors(id)
    }
  } catch (e) {
    error.value = apiErrorMessage(e)
  }
}

async function save() {
  error.value = null
  saving.value = true
  try {
    const payload = {
      number: form.number,
      contract_type: form.contract_type,
      code: form.code || undefined,
      doc_date: form.doc_date ? new Date(form.doc_date).toISOString() : undefined,
      customer_id: form.customer_id || undefined,
      organization_id: form.organization_id || undefined,
      amount: form.amount,
      currency: form.currency,
      paid: form.paid,
      fulfilled: form.fulfilled,
      comment: form.comment || undefined,
      archived: form.archived,
    }
    if (isNew.value || !currentId.value) {
      const c = await createContract(payload)
      currentId.value = c.id
      contract.value = c
      router.replace(`/contracts/${c.id}`)
      await loadNeighbors(c.id)
    } else {
      const c = await updateContract(currentId.value, payload)
      contract.value = c
    }
  } catch (e) {
    error.value = apiErrorMessage(e)
  } finally {
    saving.value = false
  }
}

async function onArchive() {
  form.archived = !form.archived
  await save()
}

function close() { router.push('/contracts') }
function goTo(id: string) { router.push(`/contracts/${id}`) }

onMounted(load)
</script>

<style scoped>
.contract-card { padding: 0; }

.toolbar {
  display: flex;
  gap: 8px;
  align-items: center;
  padding: 8px 0 12px;
  border-bottom: 1px solid #e1e4e8;
  margin-bottom: 16px;
}
.btn {
  display: inline-flex; align-items: center; gap: 6px;
  padding: 6px 12px; font-size: 13px;
  border: 1px solid #d0d7de; background: #fff;
  border-radius: 4px; cursor: pointer; color: #1f2328;
}
.btn:hover { background: #f6f8fa; }
.btn.success {
  background: #8bc34a;
  border-color: #7cb342;
  color: #fff;
  font-weight: 600;
}
.btn.success:hover { background: #7cb342; }
.btn:disabled { opacity: 0.6; cursor: default; }

.pager {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  margin-left: 12px;
}
.pager-count { font-size: 13px; color: #57606a; margin-right: 4px; }
.pager-btn {
  width: 28px; height: 28px;
  border: 1px solid #d0d7de; background: #fff;
  border-radius: 4px; cursor: pointer;
  display: inline-flex; align-items: center; justify-content: center;
  color: #57606a;
}
.pager-btn:hover:not(:disabled) { background: #f6f8fa; }
.pager-btn:disabled { opacity: 0.4; cursor: default; }

.toolbar-right {
  display: inline-flex;
  gap: 6px;
  margin-left: auto;
}
.mini-select {
  padding: 6px 10px;
  font-size: 13px;
  border: 1px solid #d0d7de;
  border-radius: 4px;
  background: #fff;
  color: #1f2328;
}

.owner-block, .changed-block {
  display: flex;
  flex-direction: column;
  margin-left: 16px;
  padding-left: 16px;
  border-left: 1px solid #eaeef2;
  font-size: 12px;
}
.owner-name { font-weight: 600; color: #1f2328; }
.owner-sub { font-size: 11px; }
.muted { color: #8c959f; }

.doc-header {
  display: flex; align-items: center;
  margin-bottom: 20px;
}
.doc-title { display: flex; align-items: center; gap: 8px; font-size: 18px; }
.doc-label { font-weight: 600; }
.num-input {
  padding: 4px 8px;
  border: 1px solid transparent;
  border-radius: 4px;
  font-size: 17px;
  font-weight: 600;
  background: transparent;
  min-width: 240px;
  color: #1f2328;
}
.num-input:hover { border-color: #d0d7de; background: #fff; }
.num-input:focus { border-color: #2c5d9c; background: #fff; outline: none; }
.date-input {
  padding: 4px 8px;
  border: 1px solid transparent;
  border-radius: 4px;
  font-size: 14px;
  background: transparent;
  color: #1f2328;
}
.date-input:hover { border-color: #d0d7de; background: #fff; }
.status-select {
  padding: 4px 8px;
  font-size: 13px;
  border: 1px solid #d0d7de;
  border-radius: 4px;
  background: #fff;
  color: #1f2328;
  margin-left: 12px;
}

.fields {
  display: grid;
  grid-template-columns: 480px 1fr;
  gap: 32px;
}
.field-left, .field-right {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.field {
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: 12px;
  color: #57606a;
}
.field label { font-weight: 500; }
.field input, .field select, .field textarea {
  padding: 6px 10px;
  font-size: 13px;
  border: 1px solid #d0d7de;
  border-radius: 3px;
  background: #fff;
  color: #1f2328;
  font-family: inherit;
}
.field textarea { resize: vertical; }

.combo {
  display: flex;
  align-items: center;
  gap: 6px;
}
.combo select { flex: 1; }
.edit-icon {
  color: #2c5d9c;
  cursor: pointer;
  font-size: 15px;
  padding: 4px;
}
.edit-icon:hover { color: #1e4070; }

.balance-label {
  color: #cf222e;
  font-weight: 600;
}
.balance-value {
  font-size: 14px;
  color: #cf222e;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
}
</style>