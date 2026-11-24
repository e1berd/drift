import { ApiError } from './http/ApiError'

function isExpected(error: unknown): boolean {
    if (error instanceof DOMException && error.name === 'AbortError') return true
    return error instanceof ApiError
}

/** Single place every caught error goes to; expected noise (aborts, HTTP errors) is dropped here. */
export function reportError(error: unknown): void {
    if (isExpected(error)) return
    if (typeof globalThis.reportError === 'function') globalThis.reportError(error)
    else queueMicrotask(() => Promise.reject(error))
}

/** Human-readable text of an error for toasts and inline messages. */
export function errorMessage(error: unknown): string {
    return error instanceof Error ? error.message : String(error)
}
