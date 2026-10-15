// Умный поиск номенклатуры: нормализация + скелет + токены + fuzzy + ранжирование.
// Без внешних зависимостей, работает в браузере и в Node (для тестов).

const HOMOGLYPHS: Record<string, string> = {
  'а': 'a', 'е': 'e', 'о': 'o', 'с': 'c', 'р': 'p', 'х': 'x',
  'у': 'y', 'к': 'k', 'м': 'm', 'т': 't', 'н': 'h', 'в': 'b',
  'и': 'u', 'і': 'i', 'ё': 'e', 'й': 'i',
}
const MULT_SIGNS = '×✕✖*xхXХ'
const SEPARATORS_RE = /[\s\-_./\\|,'"`´’]+/g

export function normalize(input: string | null | undefined): string {
  if (!input) return ''
  let out = ''
  for (const ch of String(input).toLowerCase()) {
    if (HOMOGLYPHS[ch]) out += HOMOGLYPHS[ch]
    else if (MULT_SIGNS.includes(ch)) out += 'x'
    else out += ch
  }
  out = out.replace(SEPARATORS_RE, ' ')
  out = out.replace(/\s+/g, ' ').trim()
  return out
}

export function skeleton(input: string | null | undefined): string {
  const n = normalize(input)
  const alnum = n.replace(/[^a-z0-9]/g, '')
  // 'x' между цифрами — знак умножения, а не буква: "1x2x0.75" => "12075"
  return alnum.replace(/(?<=\d)x|x(?=\d)/g, '')
}

export function tokenize(input: string | null | undefined): string[] {
  const n = normalize(input)
  return n ? n.split(' ').filter(Boolean) : []
}

export function levenshtein(a: string, b: string, max = Infinity): number {
  if (a === b) return 0
  const al = a.length, bl = b.length
  if (al === 0) return bl
  if (bl === 0) return al
  if (Math.abs(al - bl) > max) return max + 1
  let prev = new Array<number>(bl + 1)
  let curr = new Array<number>(bl + 1)
  for (let j = 0; j <= bl; j++) prev[j] = j
  for (let i = 1; i <= al; i++) {
    curr[0] = i
    let rowMin = i
    for (let j = 1; j <= bl; j++) {
      const cost = a[i - 1] === b[j - 1] ? 0 : 1
      curr[j] = Math.min(prev[j] + 1, curr[j - 1] + 1, prev[j - 1] + cost)
      if (curr[j] < rowMin) rowMin = curr[j]
    }
    if (rowMin > max) return max + 1
    const t = prev; prev = curr; curr = t
  }
  return prev[bl]
}

export type FieldName = 'name' | 'sku' | 'barcode' | 'external_id'
export interface Hit { field: FieldName; start: number; end: number }
export interface Scored<T> { item: T; score: number; hits: Hit[] }

export interface SmartSearchOptions<T> {
  /** Порядок: [name, sku, barcode, external_id] — как в Product. */
  getFields: (item: T) => Array<string | null | undefined>
  fieldOrder?: FieldName[]
  limit?: number
  extraScore?: (item: T) => number
  minChars?: number
}

const W = {
  exact: 1000,
  startsWithCode: 850,
  includesCode: 750,
  startsWithName: 650,
  includesName: 550,
  skeletonStart: 400,
  skeletonInclude: 300,
  allTokens: 200,
  fuzzy: 100,
}

function exactRange(original: string, queryLower: string): { start: number; end: number } | null {
  if (!original || !queryLower) return null
  const idx = original.toLowerCase().indexOf(queryLower)
  if (idx < 0) return null
  return { start: idx, end: idx + queryLower.length }
}

export function smartSearch<T>(items: T[], query: string, opts: SmartSearchOptions<T>): Scored<T>[] {
  const q = (query ?? '').trim()
  if (q.length < (opts.minChars ?? 1)) return []

  const qNorm = normalize(q)
  const qSkel = skeleton(q)
  const qTokens = tokenize(q)
  const qLower = q.toLowerCase()
  const limit = opts.limit ?? 20
  const extra = opts.extraScore

  const results: Scored<T>[] = []

  for (const item of items) {
    const f = opts.getFields(item)
    const name = f[0] ?? ''
    const sku = f[1] ?? ''
    const barcode = f[2] ?? ''
    const ext = f[3] ?? ''

    const nName = normalize(name)
    const nSku = normalize(sku)
    const nBar = normalize(barcode)
    const nExt = normalize(ext)

    let best = 0
    const hits: Hit[] = []

    // 1) точное
    if (name && nName === qNorm) { best = W.exact; const r = exactRange(name, qLower); if (r) hits.push({ field: 'name', ...r }) }
    if (sku && nSku === qNorm) { best = W.exact; const r = exactRange(sku, qLower); if (r) hits.push({ field: 'sku', ...r }) }
    if (barcode && nBar === qNorm) { best = W.exact; const r = exactRange(barcode, qLower); if (r) hits.push({ field: 'barcode', ...r }) }
    if (ext && nExt === qNorm) { best = W.exact }

    // 2) коды: startsWith / includes
    const codes: Array<[FieldName, string, string]> = [
      ['sku', sku, nSku], ['barcode', barcode, nBar], ['external_id', ext, nExt],
    ]
    for (const [fname, raw, norm] of codes) {
      if (!raw || !norm) continue
      if (norm.startsWith(qNorm)) {
        if (W.startsWithCode > best) best = W.startsWithCode
        const r = exactRange(raw, qLower); if (r) hits.push({ field: fname, ...r })
      } else if (norm.includes(qNorm)) {
        if (W.includesCode > best) best = W.includesCode
      }
    }

    // 3) имя
    if (nName) {
      if (nName.startsWith(qNorm)) {
        if (W.startsWithName > best) best = W.startsWithName
        const r = exactRange(name, qLower); if (r) hits.push({ field: 'name', ...r })
      } else if (nName.includes(qNorm)) {
        if (W.includesName > best) best = W.includesName
        const r = exactRange(name, qLower); if (r) hits.push({ field: 'name', ...r })
      }
    }

    // 4) скелет
    if (best < W.includesName && qSkel.length >= 2) {
      const skels = [skeleton(name), skeleton(sku), skeleton(barcode), skeleton(ext)]
      for (const sk of skels) {
        if (!sk) continue
        if (sk.startsWith(qSkel)) { if (W.skeletonStart > best) best = W.skeletonStart; break }
        if (sk.includes(qSkel)) { if (W.skeletonInclude > best) best = W.skeletonInclude; break }
      }
    }

    // 5) AND по токенам
    if (best < W.skeletonInclude && qTokens.length > 1) {
      const hay = nName + ' ' + nSku + ' ' + nBar + ' ' + nExt
      if (qTokens.every((t) => hay.includes(t))) {
        if (W.allTokens > best) best = W.allTokens
      }
    }

    // 6) fuzzy
    if (best < W.allTokens && qNorm.length >= 5) {
      const maxDist = qNorm.length >= 8 ? 2 : 1
      for (const cand of [nName, nSku, nBar, nExt]) {
        if (!cand) continue
        const win = qNorm.length
        if (cand.length < win - maxDist) continue
        if (levenshtein(qNorm, cand.slice(0, win), maxDist) <= maxDist) {
          if (W.fuzzy > best) best = W.fuzzy
          break
        }
      }
    }

    if (best <= 0) continue
    let score = best + (extra ? extra(item) : 0)
    // бонус за короткое имя (тай-брейк)
    score -= name.length * 0.001
    results.push({ item, score, hits })
  }

  results.sort((a, b) => b.score - a.score)
  return results.slice(0, limit)
}

export function smartScoreOne<T>(query: string, item: T, opts: SmartSearchOptions<T>): Scored<T> | null {
  const r = smartSearch([item], query, { ...opts, limit: 1 })
  return r[0] ?? null
}