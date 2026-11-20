import { inject, provide } from 'vue'

import type { InjectionKey } from 'vue'
import type { ActivityEntry } from '@/gen'

/** What a row of the activity log can do; provided by the chat layout that owns the session. */
export interface ActivityActions {
    reveal: (messageId: string) => void
    open: (entry: ActivityEntry) => void
    remove: (entry: ActivityEntry) => void
    canOpen: (entry: ActivityEntry) => boolean
}

const KEY: InjectionKey<ActivityActions> = Symbol('drift.activity-actions')

export function provideActivityActions(actions: ActivityActions) {
    provide(KEY, actions)
}

export function useActivityActions(): ActivityActions {
    const actions = inject(KEY)
    if (!actions) throw new Error('useActivityActions() called outside SessionChatLayout')
    return actions
}
