import { reactive } from 'vue'

import type { NoticeOptions, Notifier } from '@drift/vue/notifier'

export type NoticeColor = 'info' | 'success' | 'error' | 'loading'

export interface Notice extends NoticeOptions {
    key: string
    title: string
    color: NoticeColor
}

const DEFAULT_DURATION_MS = 5000
const closeTimers = new Map<string, ReturnType<typeof setTimeout>>()
let sequence = 0

/** Notices on screen, oldest first; SnackbarHost renders them. */
export const notices = reactive<Notice[]>([])

function show(color: NoticeColor, title: string, options: NoticeOptions = {}) {
    const key = options.id ?? `notice-${++sequence}`
    const notice: Notice = { duration: DEFAULT_DURATION_MS, ...options, key, title, color }
    const index = notices.findIndex((existing) => existing.key === key)
    if (index === -1) notices.push(notice)
    else notices.splice(index, 1, notice)
    scheduleClose(notice)
}

function scheduleClose({ key, duration }: Notice) {
    clearTimeout(closeTimers.get(key))
    closeTimers.delete(key)
    if (duration !== undefined && Number.isFinite(duration))
        closeTimers.set(
            key,
            setTimeout(() => closeNotice(key, false), duration),
        )
}

export function closeNotice(key: string, byUser: boolean) {
    const index = notices.findIndex((notice) => notice.key === key)
    if (index === -1) return
    clearTimeout(closeTimers.get(key))
    closeTimers.delete(key)
    const [closed] = notices.splice(index, 1)
    if (byUser) closed?.onUserDismiss?.()
}

export function pressNoticeAction(notice: Notice) {
    notice.action?.onAction()
    if (!notice.action?.keepsNoticeOpen) closeNotice(notice.key, true)
}

export const snackbarNotifier: Notifier = {
    info: (title, options) => show('info', title, options),
    success: (title, options) => show('success', title, options),
    error: (title, options) => show('error', title, options),
    loading: (title, options) => show('loading', title, { duration: Infinity, ...options }),
    describe(id, description) {
        const notice = notices.find((existing) => existing.key === id)
        if (notice) notice.description = description
    },
    dismiss: (id) => closeNotice(id, false),
}
