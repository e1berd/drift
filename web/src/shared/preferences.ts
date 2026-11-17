import { createColorScheme, createLocalStorage, createPreference } from '@ui-sdk/core'

/** Light, dark or the system scheme, applied as data-color-scheme on html. */
export const colorScheme = createColorScheme({
    storage: createLocalStorage('drift.color-scheme'),
    global: true,
})

/** Which edge the projects sidebar sits on, applied as data-sidebar-side on html. */
export const sidebarSide = createPreference({
    values: ['right', 'left'],
    defaultValue: 'right',
    attribute: 'data-sidebar-side',
    global: true,
    storage: createLocalStorage('drift.sidebar-side'),
})

/** Interface language, applied as data-locale on html. */
export const locale = createPreference({
    values: ['ru', 'en'],
    defaultValue: 'ru',
    attribute: 'data-locale',
    global: true,
    storage: createLocalStorage('drift.locale'),
})

/** All app-wide preferences, mounted once by the shell. */
export const preferences = [colorScheme, sidebarSide, locale]
