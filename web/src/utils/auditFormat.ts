import type { AuditLog } from '../api/audit'
import type { User } from '../api/users'

export function formatUserShort(u: User | null | undefined): string {
  if (!u) return 'Система'
  const ln = (u.last_name ?? '').trim()
  const fn = (u.first_name ?? '').trim()
  const mn = (u.middle_name ?? '').trim()
  if (ln || fn) {
    const a = fn ? fn.charAt(0) + '.' : ''
    const b = mn ? ' ' + mn.charAt(0) + '.' : ''
    return ((ln || '') + ' ' + a + b).trim()
  }
  return u.full_name || u.email || 'Сотрудник'
}

export function formatDateTimeRu(iso: string): string {
  const d = new Date(iso)
  const p = (n: number) => String(n).padStart(2, '0')
  const yy = String(d.getFullYear()).slice(2)
  return p(d.getDate()) + '.' + p(d.getMonth() + 1) + '.' + yy + ' в ' + p(d.getHours()) + ':' + p(d.getMinutes())
}

interface Body {
  name?: string
  number?: string
  incoming_number?: string
  email?: string
  full_name?: string
  [k: string]: unknown
}

function parseJson(s?: string): Body | null {
  if (!s) return null
  try { return JSON.parse(s) as Body } catch { return null }
}

export function formatEventRu(l: AuditLog): string {
  const p = (l.path || '').toLowerCase()
  const method = (l.method || '').toUpperCase()
  const b = parseJson(l.request_body)
  const num = (b?.number || b?.incoming_number || '') as string
  const nm = (b?.name || '') as string

  if (p.includes('/internal-orders')) {
    if (p.endsWith('/post')) return 'провёл(а) внутренний заказ №' + num
    if (p.endsWith('/cancel')) return 'отменил(а) внутренний заказ №' + num
    if (method === 'POST' && /\/internal-orders\/?$/.test(p)) return 'создал(а) внутренний заказ №' + num
    if (method === 'PUT' || method === 'PATCH') return 'изменил(а) внутренний заказ №' + num
  }
  if (p.includes('/documents')) {
    if (p.endsWith('/post')) return 'провёл(а) документ №' + num
    if (p.endsWith('/cancel')) return 'отменил(а) документ №' + num
    if (method === 'POST' && /\/documents\/?$/.test(p)) return 'создал(а) документ №' + num
  }
  if (p.includes('/orders')) {
    if (p.endsWith('/confirm')) return 'подтвердил(а) заказ №' + num
    if (p.endsWith('/ship')) return 'отгрузил(а) заказ №' + num
    if (p.endsWith('/cancel')) return 'отменил(а) заказ №' + num
    if (method === 'POST' && /\/orders\/?$/.test(p)) return 'создал(а) заказ №' + num
  }
  if (p.includes('/products')) {
    if (method === 'POST') return 'создал(а) товар «' + nm + '»'
    if (method === 'PUT' || method === 'PATCH') return 'изменил(а) товар «' + nm + '»'
    if (method === 'DELETE') return 'удалил(а) товар «' + nm + '»'
  }
  if (p.includes('/users')) {
    const u = (b?.full_name || b?.email || '') as string
    if (method === 'POST') return 'создал(а) сотрудника ' + u
    if (method === 'PUT' || method === 'PATCH') return 'изменил(а) сотрудника ' + u
  }
  return method + ' ' + (l.path || '')
}
