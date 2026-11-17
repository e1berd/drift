import { inject, provide, type InjectionKey, type Ref } from 'vue'

/** Sidebar state the shell shares with top bars of every view. */
export interface ShellContext {
    sidebarOpen: Readonly<Ref<boolean>>
    toggleSidebar: () => void
    closeOnNarrow: () => void
}

const KEY: InjectionKey<ShellContext> = Symbol('drift.shell')

/** Called once by AppShell. */
export function provideShell(ctx: ShellContext) {
    provide(KEY, ctx)
}

/** Sidebar controls for components inside the shell. */
export function useShell(): ShellContext {
    const ctx = inject(KEY)
    if (!ctx) throw new Error('useShell() called outside AppShell')
    return ctx
}
