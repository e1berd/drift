import { readCookie, writeCookie } from './cookies'

export const colorSchemes = ['system', 'light', 'dark'] as const
export type ColorScheme = (typeof colorSchemes)[number]

const COLOR_SCHEME_COOKIE = 'color-scheme'

function isColorScheme(value: unknown): value is ColorScheme {
    return colorSchemes.includes(value as ColorScheme)
}

/** The saved scheme, shared by every interface and the server-rendered auth pages. */
export function readColorScheme(): ColorScheme {
    const saved = readCookie(COLOR_SCHEME_COOKIE)
    return isColorScheme(saved) ? saved : 'system'
}

export function saveColorScheme(scheme: ColorScheme): void {
    writeCookie(COLOR_SCHEME_COOKIE, scheme)
}
