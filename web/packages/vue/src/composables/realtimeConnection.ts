import { computed, readonly, ref } from 'vue'

export type RealtimeStatus = 'connecting' | 'open' | 'lost'

export interface RealtimeHandlers {
    url(): URL
    onMessage(message: MessageEvent<string>): void
    onCursor(id: string): void
}

const RETRY_BASE_MS = 1000
const RETRY_MAX_MS = 15_000
const STALE_AFTER_MS = 40_000
const WATCHDOG_INTERVAL_MS = 5000

const status = ref<RealtimeStatus>('connecting')
const browserOnline = ref(typeof navigator === 'undefined' ? true : navigator.onLine)
const nextAttemptAt = ref<number | null>(null)

let handlers: RealtimeHandlers | null = null
let source: EventSource | null = null
let failedAttempts = 0
let lastSignalAt = 0
let retryTimer: ReturnType<typeof setTimeout> | undefined
let watchdogTimer: ReturnType<typeof setInterval> | undefined

function markAlive() {
    lastSignalAt = Date.now()
}

function isStale() {
    return Date.now() - lastSignalAt > STALE_AFTER_MS
}

function retryDelay(attempt: number) {
    const ceiling = Math.min(RETRY_MAX_MS, RETRY_BASE_MS * 2 ** attempt)
    return Math.round(ceiling / 2 + Math.random() * (ceiling / 2))
}

function cancelScheduledRetry() {
    clearTimeout(retryTimer)
    retryTimer = undefined
    nextAttemptAt.value = null
}

function closeSource() {
    source?.close()
    source = null
}

function scheduleRetry() {
    cancelScheduledRetry()
    const delay = retryDelay(failedAttempts)
    failedAttempts++
    nextAttemptAt.value = Date.now() + delay
    retryTimer = setTimeout(openSource, delay)
}

function dropConnection() {
    closeSource()
    status.value = 'lost'
    scheduleRetry()
}

function onOpen() {
    failedAttempts = 0
    status.value = 'open'
    markAlive()
}

function openSource() {
    if (!handlers) return
    cancelScheduledRetry()
    closeSource()
    const current = handlers
    const stream = new EventSource(current.url())
    source = stream
    if (status.value !== 'lost') status.value = 'connecting'
    stream.addEventListener('open', onOpen)
    stream.addEventListener('error', () => {
        if (source === stream) dropConnection()
    })
    stream.addEventListener('ping', markAlive)
    stream.addEventListener('ready', (event: MessageEvent<string>) => {
        markAlive()
        if (event.lastEventId) current.onCursor(event.lastEventId)
    })
    stream.addEventListener('message', (message: MessageEvent<string>) => {
        markAlive()
        current.onMessage(message)
    })
}

/** Reconnects right away, skipping any pending backoff. */
export function reconnectNow() {
    failedAttempts = 0
    openSource()
}

function reconnectIfUnhealthy() {
    if (status.value !== 'open' || isStale()) reconnectNow()
}

function checkHeartbeat() {
    if (status.value === 'open' && isStale()) dropConnection()
}

function onBrowserOnline() {
    browserOnline.value = true
    reconnectNow()
}

function onBrowserOffline() {
    browserOnline.value = false
    if (source) dropConnection()
}

function onVisibilityChange() {
    if (document.visibilityState === 'visible') reconnectIfUnhealthy()
}

function onPageShow(event: PageTransitionEvent) {
    if (event.persisted) reconnectIfUnhealthy()
}

function watchBrowserSignals() {
    window.addEventListener('online', onBrowserOnline)
    window.addEventListener('offline', onBrowserOffline)
    window.addEventListener('pageshow', onPageShow)
    document.addEventListener('visibilitychange', onVisibilityChange)
    watchdogTimer = setInterval(checkHeartbeat, WATCHDOG_INTERVAL_MS)
}

/** Opens the single per-tab realtime connection and keeps it alive for the life of the tab. */
export function connectRealtime(next: RealtimeHandlers) {
    if (handlers || typeof EventSource === 'undefined') return
    handlers = next
    watchBrowserSignals()
    openSource()
}

/** Status of the realtime connection, whether the browser reports network access, and when the next retry fires. */
export function useRealtimeConnection() {
    return {
        status: readonly(status),
        connected: computed(() => status.value !== 'lost'),
        browserOnline: readonly(browserOnline),
        nextAttemptAt: readonly(nextAttemptAt),
        reconnectNow,
    }
}

export function disconnectRealtime() {
    window.removeEventListener('online', onBrowserOnline)
    window.removeEventListener('offline', onBrowserOffline)
    window.removeEventListener('pageshow', onPageShow)
    document.removeEventListener('visibilitychange', onVisibilityChange)
    clearInterval(watchdogTimer)
    cancelScheduledRetry()
    closeSource()
    handlers = null
}
