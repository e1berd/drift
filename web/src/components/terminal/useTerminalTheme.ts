import { onScopeDispose, watch } from 'vue'

import type { ITheme, Terminal } from '@xterm/xterm'

import { useMediaQuery } from '@/composables/useMediaQuery'

const THEME_TOKENS = {
    background: 'var(--ui-background)',
    foreground: 'var(--ui-foreground)',
    cursor: 'var(--ui-foreground)',
    cursorAccent: 'var(--ui-background)',
    selectionBackground: 'color-mix(in oklab, var(--ui-primary) 28%, transparent)',
} satisfies ITheme

function toRgba(cssColor: string): string {
    const context = document.createElement('canvas').getContext('2d', { willReadFrequently: true })
    if (!context) return cssColor
    context.fillStyle = cssColor
    context.fillRect(0, 0, 1, 1)
    const [red, green, blue, alpha] = context.getImageData(0, 0, 1, 1).data
    return `rgba(${red}, ${green}, ${blue}, ${(alpha ?? 255) / 255})`
}

function resolveColor(host: HTMLElement, value: string): string {
    const probe = document.createElement('span')
    probe.style.color = value
    host.append(probe)
    const computed = getComputedStyle(probe).color
    probe.remove()
    return toRgba(computed)
}

function themeFor(host: HTMLElement): ITheme {
    return Object.fromEntries(
        Object.entries(THEME_TOKENS).map(([name, value]) => [name, resolveColor(host, value)]),
    )
}

/**
 * xterm draws on a canvas and cannot read CSS variables, so the ui-sdk tokens are resolved to
 * concrete colors and re-resolved whenever the light/dark scheme changes.
 */
export function useTerminalTheme(terminal: Terminal, host: () => HTMLElement | null) {
    function apply() {
        const element = host()
        if (!element) return
        terminal.options.theme = themeFor(element)
        terminal.options.fontFamily = getComputedStyle(element).getPropertyValue('--ui-font-mono')
    }

    const prefersDark = useMediaQuery('(prefers-color-scheme: dark)')
    watch(prefersDark, apply)
    const schemeObserver = new MutationObserver(apply)
    schemeObserver.observe(document.documentElement, { attributeFilter: ['data-color-scheme'] })
    onScopeDispose(() => schemeObserver.disconnect())

    return { apply }
}
