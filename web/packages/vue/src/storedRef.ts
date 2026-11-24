import { ref, watch } from 'vue'

import type { Ref } from 'vue'

import { readStored, writeStored } from '@drift/sdk/storage'

/** A ref mirrored to localStorage, so the value survives reloads on this device. */
export function storedRef<T>(key: string, fallback: T): Ref<T> {
    const state = ref(readStored(key, fallback)) as Ref<T>
    watch(state, (value) => writeStored(key, value), { deep: true })
    return state
}
