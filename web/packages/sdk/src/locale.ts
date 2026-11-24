import { readCookie, writeCookie } from './cookies'

export const locales = ['ru', 'en'] as const
export type Locale = (typeof locales)[number]
export const defaultLocale: Locale = 'ru'

const LOCALE_COOKIE = 'lang'

export function isLocale(value: unknown): value is Locale {
    return locales.includes(value as Locale)
}

/** The saved language, else the first browser language we support, else the default. */
export function detectLocale(): Locale {
    const saved = readCookie(LOCALE_COOKIE)
    if (isLocale(saved)) return saved
    for (const tag of navigator.languages ?? []) {
        const base = tag.toLowerCase().split('-').at(0)
        if (isLocale(base)) return base
    }
    return defaultLocale
}

/** Saves the language for the server and every Drift interface. */
export function saveLocale(locale: Locale): void {
    writeCookie(LOCALE_COOKIE, locale)
}
