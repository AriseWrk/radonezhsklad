import { describe, it, expect } from 'vitest'
import { normalize, skeleton, tokenize, levenshtein, smartSearch } from './smartSearch'

interface P { id: string; name: string; sku?: string; barcode?: string; external_id?: string }

function mk(partial: Partial<P> & { name: string }): P {
  return { id: partial.name, ...partial }
}

const opts = {
  getFields: (p: P) => [p.name, p.sku ?? null, p.barcode ?? null, p.external_id ?? null],
  limit: 20,
}

describe('normalize', () => {
  it('lower + гомоглифы + разделители', () => {
    expect(normalize('  С2000-МК  ')).toBe('c2000 mk')
    expect(normalize('Батарея_12.7')).toBe(normalize('батарея 12 7'))
    expect(normalize('foo|bar\\baz')).toBe('foo bar baz')
  })
  it('знаки умножения все → x', () => {
    expect(normalize('1×2×0.75')).toBe('1x2x0 75')
    expect(normalize('1х2х0.75')).toBe('1x2x0 75')
    expect(normalize('1*2*0.75')).toBe('1x2x0 75')
  })
})

describe('skeleton', () => {
  it('1x2x0.75 == 1×2×0.75 == 1х2х0.75 == 1 2 0 75', () => {
    const a = skeleton('1x2x0.75')
    expect(skeleton('1×2×0.75')).toBe(a)
    expect(skeleton('1х2х0.75')).toBe(a)
    expect(skeleton('1 2 0 75')).toBe(a)
    expect(a).toBe('12075')
  })
  it('буква x между буквами — не удаляется', () => {
    expect(skeleton('Sony XZ')).toBe('sonyxz')
  })
})

describe('tokenize', () => {
  it('с2000-м / с2000 м дают одинаковые токены', () => {
    expect(tokenize('С2000-М')).toEqual(['c2000', 'm'])
    expect(tokenize('С2000 М')).toEqual(['c2000', 'm'])
  })
})

describe('levenshtein', () => {
  it('простые случаи', () => {
    expect(levenshtein('kitten', 'sitting')).toBe(3)
    expect(levenshtein('abc', 'abc')).toBe(0)
    expect(levenshtein('abc', 'abcd')).toBe(1)
  })
})

describe('smartSearch — кейсы ТЗ', () => {
  const items: P[] = [
    mk({ name: 'Кабель 1×2×0.75', sku: 'KAB-1X2X075' }),
    mk({ name: 'С2000-М', sku: 'C2000M' }),
    mk({ name: 'С2000-МК', sku: 'C2000-MK' }),
    mk({ name: 'Батарея 12-7', sku: 'BAT-12-7' }),
    mk({ name: 'Sony XZ-100', sku: 'SONY-XZ100' }),
  ]

  function top(q: string): string | undefined {
    return smartSearch(items, q, opts)[0]?.item.name
  }

  it('кабель: 1x2x0.75 / 1×2×0.75 / 1х2х0.75 / 1 2 0 75', () => {
    for (const q of ['1x2x0.75', '1×2×0.75', '1х2х0.75', '1 2 0 75', 'кабель 1х2х0.75']) {
      expect(top(q)).toBe('Кабель 1×2×0.75')
    }
  })

  it('с2000м / с2000-м / с2000 м / C2000-MK / С2000 МК', () => {
    expect(top('с2000м')).toBe('С2000-М')
    expect(top('с2000-м')).toBe('С2000-М')
    expect(top('с2000 м')).toBe('С2000-М')
    expect(top('C2000-MK')).toBe('С2000-МК')
    expect(top('С2000 МК')).toBe('С2000-МК')
  })

  it('батарея 12-7 / батарея 12 7 / батарея 127', () => {
    expect(top('батарея 12-7')).toBe('Батарея 12-7')
    expect(top('батарея 12 7')).toBe('Батарея 12-7')
    expect(top('батарея 127')).toBe('Батарея 12-7')
  })

  it('опечатка ловится fuzzy', () => {
    expect(top('баттарея 12-7')).toBe('Батарея 12-7')
  })

  it('SKU-совпадение выше name-совпадения', () => {
    const arr = [mk({ name: 'кабель обычный', sku: 'ZZZ' }), mk({ name: 'что-то', sku: 'KAB-1X2X075' })]
    const r = smartSearch(arr, 'kab-1x2x075', { ...opts, getFields: (p: P) => [p.name, p.sku ?? null, null, null] })
    expect(r[0].item.name).toBe('что-то')
  })
})
