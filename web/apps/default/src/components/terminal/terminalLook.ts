import type { TerminalLook } from '@drift/vue/terminal/useTerminalTheme'

export const terminalLook: TerminalLook = {
    colors: {
        background: 'var(--ui-background)',
        foreground: 'var(--ui-foreground)',
        cursor: 'var(--ui-foreground)',
        cursorAccent: 'var(--ui-background)',
        selectionBackground: 'color-mix(in oklab, var(--ui-primary) 28%, transparent)',
    },
    fontFamily: (host) => getComputedStyle(host).getPropertyValue('--ui-font-mono'),
    schemeAttributes: ['data-color-scheme'],
}
