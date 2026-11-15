import { ref, type Ref } from 'vue'

export interface ColumnDef {
  key: string
  label: string
  always?: boolean
}

export interface ColumnSettingsApi {
  defs: ColumnDef[]
  visible: Ref<Set<string>>
  isVisible: (key: string) => boolean
  toggle: (key: string) => void
  reset: () => void
}

function loadStorage(pageKey: string, defs: ColumnDef[]): Set<string> {
  try {
    const raw = localStorage.getItem(`rs_cols_${pageKey}`)
    if (!raw) return new Set(defs.map((d) => d.key))
    const arr = JSON.parse(raw) as string[]
    const set = new Set(arr)
    const known = new Set(defs.map((d) => d.key))
    for (const k of Array.from(set)) if (!known.has(k)) set.delete(k)
    for (const d of defs) if (!set.has(d.key) && d.always) set.add(d.key)
    return set
  } catch {
    return new Set(defs.map((d) => d.key))
  }
}

function saveStorage(pageKey: string, visible: Set<string>): void {
  try { localStorage.setItem(`rs_cols_${pageKey}`, JSON.stringify(Array.from(visible))) } catch {}
}

export function useColumnSettings(pageKey: string, defs: ColumnDef[]): ColumnSettingsApi {
  const visible = ref<Set<string>>(loadStorage(pageKey, defs))
  function isVisible(key: string): boolean { return visible.value.has(key) }
  function toggle(key: string): void {
    const def = defs.find((d) => d.key === key)
    if (def?.always) return
    const next = new Set(visible.value)
    if (next.has(key)) next.delete(key); else next.add(key)
    visible.value = next
    saveStorage(pageKey, next)
  }
  function reset(): void {
    const next = new Set(defs.map((d) => d.key))
    visible.value = next
    saveStorage(pageKey, next)
  }
  return { defs, visible, isVisible, toggle, reset }
}