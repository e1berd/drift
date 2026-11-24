import { readCookie, writeCookie } from './cookies'

const FRONTEND_COOKIE = 'ui'
const NAME_PREFIX = 'frontends/'

export const DEFAULT_FRONTEND = `${NAME_PREFIX}default`

export function frontendName(id: string): string {
    return NAME_PREFIX + id
}

export function frontendId(name: string): string {
    return name.slice(NAME_PREFIX.length)
}

/** The interface this browser picked, as a resource name, or null to follow the instance default. */
export function pickedFrontend(): string | null {
    const id = readCookie(FRONTEND_COOKIE)
    return id ? frontendName(id) : null
}

/** Saves the pick in the cookie the server reads to choose which interface a page loads. */
export function pickFrontend(name: string): void {
    writeCookie(FRONTEND_COOKIE, frontendId(name))
}
