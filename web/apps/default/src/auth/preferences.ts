// Language and theme switchers of the server-rendered auth pages (#language-switch and
// #theme-switch in app/internal/auth/templates/layout.html).
import { applyColorScheme, colorScheme } from '@/shared/preferences'
import { locales, saveLocale } from '@drift/sdk/locale'

import type { Locale } from '@drift/sdk/locale'

type ValueChange = CustomEvent<{ value: string | null }>

const themes = ['system', 'light', 'dark'] as const

export function bindPreferenceSwitches(root: ParentNode, reload = () => window.location.reload()) {
    applyColorScheme()

    root.querySelector('#theme-switch')?.addEventListener('valuechange', (event) => {
        const value = (event as ValueChange).detail.value
        if (themes.includes(value as (typeof themes)[number])) {
            colorScheme.set(value as (typeof themes)[number], { event })
        }
    })

    // The pages are rendered on the server, so a new language needs a fresh render.
    root.querySelector('#language-switch')?.addEventListener('valuechange', (event) => {
        const value = (event as ValueChange).detail.value
        if (locales.includes(value as Locale) && value !== document.documentElement.lang) {
            saveLocale(value as Locale)
            reload()
        }
    })
}
