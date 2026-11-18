import type { UsageWindow } from '@/gen'

export type UsageWindowKind = 'session' | 'week' | 'weekOpus' | 'weekSonnet' | 'other'

const NAMED: Record<string, UsageWindowKind> = {
    five_hour: 'session',
    seven_day: 'week',
    seven_day_opus: 'weekOpus',
    seven_day_sonnet: 'weekSonnet',
}

/** Which label a limit window gets: Claude names its windows, Codex only gives their length. */
export function usageWindowKind(
    window: Pick<UsageWindow, 'id' | 'windowMinutes'>,
): UsageWindowKind {
    const named = NAMED[window.id]
    if (named) return named
    if (window.windowMinutes === 300) return 'session'
    if (window.windowMinutes === 7 * 24 * 60) return 'week'
    return 'other'
}

/** Seconds until a window resets, never negative; null when the agent gave no time. */
export function secondsUntil(iso: string | undefined, now = Date.now()): number | null {
    if (!iso) return null
    return Math.max(0, Math.round((Date.parse(iso) - now) / 1000))
}

/** "через 3 часа" / "in 2 days": the reset moment in the largest whole unit. */
export function formatReset(seconds: number, lang: string): string {
    const format = new Intl.RelativeTimeFormat(lang, { numeric: 'auto' })
    if (seconds < 3600) return format.format(Math.max(1, Math.ceil(seconds / 60)), 'minute')
    if (seconds < 48 * 3600) return format.format(Math.round(seconds / 3600), 'hour')
    return format.format(Math.round(seconds / 86_400), 'day')
}

/** "15:00" today, "пт, 15:00" later in the week, "21 нояб., 15:00" beyond it. */
export function formatResetClock(iso: string, lang: string, now = Date.now()): string {
    const reset = new Date(iso)
    const days = (reset.getTime() - now) / 86_400_000
    const sameDay = reset.toDateString() === new Date(now).toDateString()
    const options: Intl.DateTimeFormatOptions = sameDay
        ? { hour: '2-digit', minute: '2-digit' }
        : days < 6
          ? { weekday: 'short', hour: '2-digit', minute: '2-digit' }
          : { day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' }
    return new Intl.DateTimeFormat(lang, options).format(reset)
}

export type UsageLevel = 'normal' | 'high' | 'critical'

export function usageLevel(usedPercent: number): UsageLevel {
    if (usedPercent >= 90) return 'critical'
    if (usedPercent >= 75) return 'high'
    return 'normal'
}

const MODEL_WEEKS: Partial<Record<string, UsageWindowKind>> = {
    opus: 'weekOpus',
    sonnet: 'weekSonnet',
}

export function headlineWindows<
    Window extends Pick<UsageWindow, 'id' | 'windowMinutes' | 'usedPercent'>,
>(windows: Window[], model: string): { session?: Window; week?: Window } {
    const weekKinds: UsageWindowKind[] = ['week']
    const modelWeek = MODEL_WEEKS[model]
    if (modelWeek) weekKinds.push(modelWeek)
    const session = windows.find((window) => usageWindowKind(window) === 'session')
    const week = windows
        .filter((window) => weekKinds.includes(usageWindowKind(window)))
        .reduce<Window | undefined>(
            (tightest, window) =>
                !tightest || window.usedPercent > tightest.usedPercent ? window : tightest,
            undefined,
        )
    return { session, week }
}
