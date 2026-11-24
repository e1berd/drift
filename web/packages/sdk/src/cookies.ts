const YEAR_SECONDS = 31_536_000

/** A cookie value, or null when it is not set. */
export function readCookie(name: string): string | null {
    const prefix = `${encodeURIComponent(name)}=`
    const entry = document.cookie.split('; ').find((part) => part.startsWith(prefix))
    return entry === undefined ? null : decodeURIComponent(entry.slice(prefix.length))
}

/** Saves a site-wide cookie for a year, readable by the server that renders the auth pages. */
export function writeCookie(name: string, value: string): void {
    const parts = [
        `${encodeURIComponent(name)}=${encodeURIComponent(value)}`,
        'path=/',
        `max-age=${YEAR_SECONDS}`,
        'samesite=lax',
    ]
    if (location.protocol === 'https:') parts.push('secure')
    document.cookie = parts.join('; ')
}
