import { inject } from 'vue'

import type { InjectionKey, Plugin } from 'vue'

import { errorMessage } from '@drift/sdk/errors'

export interface NoticeAction {
    label: string
    onAction: () => void
    keepsNoticeOpen?: boolean
}

export interface NoticeOptions {
    id?: string
    description?: string
    duration?: number
    action?: NoticeAction
    onUserDismiss?: () => void
}

/** Short-lived messages to the user; every Drift interface provides one in its own UI kit. */
export interface Notifier {
    info(title: string, options?: NoticeOptions): void
    success(title: string, options?: NoticeOptions): void
    error(title: string, options?: NoticeOptions): void
    loading(title: string, options?: NoticeOptions): void
    describe(id: string, description: string): void
    dismiss(id: string): void
}

const NOTIFIER: InjectionKey<Notifier> = Symbol('drift-notifier')

/** Installs the interface's notifier, so shared composables can show notices with it. */
export function createNotifierPlugin(notifier: Notifier): Plugin {
    return { install: (app) => app.provide(NOTIFIER, notifier) }
}

export function useNotifier(): Notifier {
    const notifier = inject(NOTIFIER)
    if (!notifier)
        throw new Error('useNotifier() needs createNotifierPlugin() installed on the app')
    return notifier
}

/** Shows a failed request or action to the user as an error notice. */
export function useErrorNotice(): (error: unknown) => void {
    const notifier = useNotifier()
    return (error) => notifier.error(errorMessage(error))
}
