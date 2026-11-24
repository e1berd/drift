// The color scheme lives in a cookie so the Go server renders /login and /setup in the chosen
// scheme (see app/internal/auth/templates.go); the language is in @drift/sdk/locale.
import {
    createColorScheme,
    createCookieStorage,
    createLocalStorage,
    createPreference,
    type ColorSchemeValue,
} from '@ui-sdk/core'

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
