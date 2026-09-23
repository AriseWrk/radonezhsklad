<template>
  <div class="card-page">
    <div class="ms-doc-head">
      <div class="ms-title">
        <button class="ms-help" title="Справка"><MsIcon name="help" :size="14" /></button>
        <span>Договор</span>
        <button class="ms-refresh" @click="load" title="Обновить"><MsIcon name="refresh" :size="14" /></button>
      </div>
      <div class="ms-status">
        <span class="st" :class="form.archived ? 'st-cancelled' : 'st-posted'">
          {{ form.archived ? 'Архивный' : 'Активный' }}
        </span>
      </div>
    </div>

    <div class="ms-toolbar">
      <MsButton variant="primary" icon="save" @click="save" :disabled="saving">
        {{ saving ? 'Сохранение...' : 'Сохранить' }}
      </MsButton>
      <MsButton icon="trash" @click="onArchive">
        {{ form.archived ? 'Вернуть' : 'В архив' }}
      </MsButton>
      <MsButton icon="print">Печать</MsButton>
      <MsButton icon="send">Отправить</MsButton>
      <div class="pager-nav" v-if="neighbors">
        <button class="pg" :disabled="!neighbors.prev_id" @click="goTo(neighbors.prev_id)">‹</button>
        <span class="pg-info">{{ neighbors.index }} из {{ neighbors.total }}</span>
        <button class="pg" :disabled="!neighbors.next_id" @click="goTo(neighbors.next_id)">›</button>
      </div>
      <div class="toolbar-spacer"></div>
      <MsButton variant="icon" icon="close" @click="close" title="Закрыть" />
    </div>

    <div v-if="error" class="error-box">{{ error }}</div>

    <div class="ms-doc-fields">
      <div class="doc-left">
        <div class="row">
          <label>Номер</label>
          <input v-model="form.number" />
        </div>
        <div class="row">
          <label>Дата</label>
          <input v-model="form.doc_date" type="datetime-local" />
        </div>
        <div class="row">
          <label>Организация</label>
          <select v-model="form.organization_id">
            <option value="">—</option>
            <option v-for="o in organizations" :key="o.id" :value="o.id">{{ o.name }}</option>
          </select>
        </div>
        <div class="row">
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
      </div>
      <div class="doc-right">
        <div class="row">
          <label>Контрагент</label>
          <select v-model="form.customer_id">
            <option value="">—</option>
            <option v-for="c in customers" :key="c.id" :value="c.id">{{ c.name }}</option>
          </select>
        </div>
        <div class="row">
          <label>Код</label>
          <input v-model="form.code" />
        </div>
        <div class="row">
          <label>Сумма договора</label>
          <input v-model.number="form.amount" type="number" step="0.01" min="0" />
        </div>
        <div class="row">
          <label>Баланс</label>
          <span class="ro">{{ formatMoney(customerBalance) }} руб</span>
        </div>
      </div>
    </div>

    <div class="bottom-row">
      <div class="comment-box">
        <label class="lbl">Комментарий</label>
        <textarea v-model="form.comment" rows="5"></textarea>
      </div>
      <div class="totals-box">
        <div class="total-row big">
          <span>Сумма:</span>
          <span class="total-num">{{ formatMoney(form.amount) }}</span>
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
import MsButton from '../components/MsButton.vue'
import MsIcon from '../components/MsIcon.vue'

const route = useRoute()
const router = useRouter()

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

const customerBalance = computed(() => form.amount - form.paid)

function formatMoney(n: number) {
  return n.toLocaleString('ru-RU', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
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
.card-page { font-size: 13px; }
.ms-doc-head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 10px; }
.ms-title { display: flex; align-items: center; gap: 8px; font-size: 20px; font-weight: 600; color: #1f2328; }
.ms-help { width: 20px; height: 20px; border-radius: 50%; border: 1px solid #b8c0c8; background: transparent; color: #57606a; cursor: pointer; display: inline-flex; align-items: center; justify-content: center; padding: 0; }
.ms-help:hover { background: #f0f2f5; }
.ms-refresh { width: 24px; height: 24px; padding: 0; display: inline-flex; align-items: center; justify-content: center; border: none; background: transparent; color: #57606a; cursor: pointer; }
.ms-refresh:hover { color: #2c5d9c; }
.ms-status { display: flex; align-items: center; gap: 12px; }
.st { display: inline-block; padding: 3px 12px; border-radius: 3px; font-size: 12px; font-weight: 500; }
.st-posted    { background: #d8e8c8; color: #2d6a1e; }
.st-cancelled { background: #f5d5d5; color: #a01c1c; }

.ms-toolbar { display: flex; align-items: center; gap: 6px; margin-bottom: 12px; flex-wrap: wrap; padding-bottom: 10px; border-bottom: 1px solid #eaeef2; }
.toolbar-spacer { flex: 1; }

.pager-nav { display: inline-flex; align-items: center; gap: 4px; margin-left: 8px; }
.pg { width: 24px; height: 24px; border: 1px solid #d0d7de; background: #fff; border-radius: 3px; cursor: pointer; font-size: 12px; color: #1f2328; }
.pg:disabled { opacity: 0.4; cursor: default; }
.pg-info { font-size: 12px; color: #57606a; margin: 0 6px; }

.ms-doc-fields { display: grid; grid-template-columns: 1fr 1fr; gap: 6px 40px; margin-bottom: 14px; max-width: 1100px; }
.doc-left, .doc-right { display: flex; flex-direction: column; gap: 6px; }
.row { display: grid; grid-template-columns: 160px 1fr; align-items: center; gap: 10px; }
.row > label { font-size: 13px; color: #57606a; }
.row > input, .row > select, .row > .ro {
  height: 28px; padding: 0 8px; font-size: 13px; line-height: 28px;
  border: 1px solid #d0d7de; border-radius: 3px;
  background: #fff; color: #1f2328; width: 100%;
}
.row > .ro { border-color: transparent; background: transparent; }

.bottom-row { display: grid; grid-template-columns: 1fr 380px; gap: 24px; align-items: start; margin-top: 12px; }
.comment-box { display: flex; flex-direction: column; gap: 4px; }
.comment-box .lbl { font-size: 12px; color: #57606a; }
.comment-box textarea { font-family: inherit; font-size: 13px; padding: 8px 10px; border: 1px solid #d0d7de; border-radius: 3px; background: #fff; color: #1f2328; resize: vertical; }
.totals-box { background: #f6f8fa; border: 1px solid #eaeef2; border-radius: 4px; padding: 12px 16px; }
.total-row { display: flex; justify-content: space-between; align-items: center; font-size: 15px; font-weight: 600; color: #1f2328; }
.total-num { font-variant-numeric: tabular-nums; }
</style>