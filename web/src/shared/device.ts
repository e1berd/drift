import { reportError } from '@/shared/errors'

const KEY = 'drift.device'

let cached: string | null = null

/** Stable id of this browser, sent with every write so its own realtime echoes can be skipped. */
export function deviceId(): string {
    if (cached) return cached
    try {
        cached = localStorage.getItem(KEY)
        if (!cached) {
            cached = crypto.randomUUID()
            localStorage.setItem(KEY, cached)
        }
    } catch (error) {
        reportError(error)
        cached ??= crypto.randomUUID()
    }
    return cached
}

/** Rough device class from the user agent, shown next to drafts written elsewhere. */
export function deviceKind(): 'phone' | 'desktop' {
    return /Android|iPhone|iPad|Mobile/i.test(navigator.userAgent) ? 'phone' : 'desktop'
}
