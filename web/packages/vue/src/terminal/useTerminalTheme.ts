import { onScopeDispose, watch } from 'vue'

import type { ITheme, Terminal } from '@xterm/xterm'

import { useMediaQuery } from '../composables/useMediaQuery'

export interface TerminalLook {
    colors: ITheme
    fontFamily: (host: HTMLElement) => string
    schemeAttributes: string[]
}

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

function themeFor(host: HTMLElement, colors: ITheme): ITheme {
    return Object.fromEntries(
        Object.entries(colors).map(([name, value]) => [name, resolveColor(host, String(value))]),
    )
}

/**
 * xterm draws on a canvas and cannot read CSS variables, so the interface's color tokens are
 * resolved to concrete colors and re-resolved whenever the light/dark scheme changes.
 */
export function useTerminalTheme(
    terminal: Terminal,
    host: () => HTMLElement | null,
    look: TerminalLook,
) {
    function apply() {
        const element = host()
        if (!element) return
        terminal.options.theme = themeFor(element, look.colors)
        terminal.options.fontFamily = look.fontFamily(element)
    }

    const prefersDark = useMediaQuery('(prefers-color-scheme: dark)')
    watch(prefersDark, apply)
    const schemeObserver = new MutationObserver(apply)
    schemeObserver.observe(document.documentElement, { attributeFilter: look.schemeAttributes })
    onScopeDispose(() => schemeObserver.disconnect())

    return { apply }
}
