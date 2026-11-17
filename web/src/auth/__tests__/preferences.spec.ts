import { afterEach, describe, expect, it, vi } from 'vitest'

import { bindPreferenceSwitches } from '../preferences'

function page(lang = 'ru') {
  document.documentElement.lang = lang
  document.body.innerHTML = `
    <ui-select id="language-switch"></ui-select>
    <ui-select id="theme-switch"></ui-select>`
  return document.body
}

function change(id: string, value: string) {
  document.getElementById(id)!.dispatchEvent(new CustomEvent('valuechange', { detail: { value } }))
}

afterEach(() => {
  document.cookie = 'lang=; max-age=0; path=/'
  document.cookie = 'color-scheme=; max-age=0; path=/'
})

describe('auth preference switches', () => {
  it('saves the language and reloads the server-rendered page', () => {
    const reload = vi.fn()
    bindPreferenceSwitches(page('ru'), reload)

    change('language-switch', 'en')

    expect(document.cookie).toContain('lang=en')
    expect(reload).toHaveBeenCalledOnce()
  })

  it('ignores unknown languages and the current one', () => {
    const reload = vi.fn()
    bindPreferenceSwitches(page('ru'), reload)

    change('language-switch', 'ru')
    change('language-switch', 'de')

    expect(reload).not.toHaveBeenCalled()
  })

  it('applies and saves the color scheme without a reload', () => {
    const reload = vi.fn()
    bindPreferenceSwitches(page(), reload)

    change('theme-switch', 'dark')

    expect(document.documentElement.dataset.colorScheme).toBe('dark')
    expect(document.cookie).toContain('color-scheme=dark')
    expect(reload).not.toHaveBeenCalled()
  })
})
