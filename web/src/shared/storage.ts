import { ref, watch } from 'vue'

import type { Ref } from 'vue'

import { reportError } from '@/shared/errors'

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

/** A ref mirrored to localStorage, so the value survives reloads on this device. */
export function storedRef<T>(key: string, fallback: T): Ref<T> {
    const state = ref(readStored(key, fallback)) as Ref<T>
    watch(state, (value) => writeStored(key, value), { deep: true })
    return state
}
