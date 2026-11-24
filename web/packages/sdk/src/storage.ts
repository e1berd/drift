import { reportError } from './errors'

/** Reads a JSON value saved in localStorage, falling back when it is missing or unreadable. */
export function readStored<T>(key: string, fallback: T): T {
    try {
        const raw = localStorage.getItem(key)
        return raw === null ? fallback : (JSON.parse(raw) as T)
    } catch (error) {
        reportError(error)
        return fallback
    }
}

/** Saves a JSON value to localStorage. */
export function writeStored(key: string, value: unknown): void {
    try {
        localStorage.setItem(key, JSON.stringify(value))
    } catch (error) {
        reportError(error)
    }
}
