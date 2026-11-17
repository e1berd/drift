// Language and color scheme preferences, shared by the SPA and the server-rendered auth pages.
// Both live in cookies so the Go server renders /login and /setup in the chosen language and
// scheme (see app/internal/i18n and app/internal/auth/templates.go).
import {
    createColorScheme,
    createCookieStorage,
    createLocalStorage,
    createPreference,
    type ColorSchemeValue,
} from '@ui-sdk/core'

export const locales = ['ru', 'en'] as const
export type Locale = (typeof locales)[number]
export const defaultLocale: Locale = 'ru'

const localeCookie = createCookieStorage('lang')

function isLocale(value: unknown): value is Locale {
    return locales.includes(value as Locale)
}

/** The saved language, else the first browser language we support, else the default. */
export function detectLocale(): Locale {
    const saved = localeCookie.get()
    if (isLocale(saved)) return saved
    for (const tag of navigator.languages ?? []) {
        const base = tag.toLowerCase().split('-').at(0)
        if (isLocale(base)) return base
    }
    return defaultLocale
}

/** Saves the language for the server and the SPA. */
export function saveLocale(locale: Locale): void {
    localeCookie.set(locale)
}

export const colorScheme = createColorScheme({
    storage: createCookieStorage('color-scheme'),
})

/** Keeps `data-color-scheme` on <html> in step with the saved choice; base.css maps it to `color-scheme`. */
export function applyColorScheme(root: HTMLElement = document.documentElement): () => void {
    const apply = ({ value }: { value: ColorSchemeValue }) => {
        root.dataset.colorScheme = value
    }
    apply(colorScheme.get())
    return colorScheme.subscribe(apply)
}

/** Which edge the projects sidebar sits on, applied as data-sidebar-side on <html>. */
export const sidebarSide = createPreference({
    values: ['right', 'left'],
    defaultValue: 'right',
    attribute: 'data-sidebar-side',
    global: true,
    storage: createLocalStorage('drift.sidebar-side'),
})

/** Preferences mounted once by the app shell. */
export const shellPreferences = [sidebarSide]
