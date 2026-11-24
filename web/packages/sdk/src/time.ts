const MINUTE = 60
const HOUR = 60 * MINUTE
const DAY = 24 * HOUR
const WEEK = 7 * DAY

const units: [Intl.NumberFormatOptions['unit'], number][] = [
    ['week', WEEK],
    ['day', DAY],
    ['hour', HOUR],
    ['minute', MINUTE],
]

const cache = new Map<string, Intl.NumberFormat | Intl.RelativeTimeFormat | Intl.DateTimeFormat>()

function cached<T extends Intl.NumberFormat | Intl.RelativeTimeFormat | Intl.DateTimeFormat>(
    key: string,
    make: () => T,
): T {
    let f = cache.get(key) as T | undefined
    if (!f) {
        f = make()
        cache.set(key, f)
    }
    return f
}

/** Compact age for list rows: "сейчас", "5 мин", "3 ч", "вчера", "2 нед", then a date. */
export function relativeTime(iso: string, lang: string, now = Date.now()): string {
    const seconds = Math.max(0, Math.round((now - Date.parse(iso)) / 1000))
    if (seconds < MINUTE) {
        return cached(
            `rel:${lang}`,
            () => new Intl.RelativeTimeFormat(lang, { numeric: 'auto' }),
        ).format(0, 'second')
    }
    if (seconds >= DAY && seconds < 2 * DAY) {
        return cached(
            `rel:${lang}`,
            () => new Intl.RelativeTimeFormat(lang, { numeric: 'auto' }),
        ).format(-1, 'day')
    }
    if (seconds >= 5 * WEEK) {
        return cached(
            `date:${lang}`,
            () => new Intl.DateTimeFormat(lang, { day: 'numeric', month: 'short' }),
        ).format(Date.parse(iso))
    }
    for (const [unit, size] of units) {
        if (seconds >= size) {
            return cached(
                `num:${lang}:${unit}`,
                () => new Intl.NumberFormat(lang, { style: 'unit', unit, unitDisplay: 'short' }),
            ).format(Math.floor(seconds / size))
        }
    }
    return ''
}
