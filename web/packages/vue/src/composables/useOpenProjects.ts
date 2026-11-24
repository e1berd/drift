import { computed } from 'vue'

import { storedRef } from '../storedRef'

const stored = storedRef<string[]>('drift.open-projects', [])
const open = computed({
    get: () => new Set(stored.value),
    set: (set: Set<string>) => (stored.value = [...set]),
})

/** Which project accordions are expanded on this device. */
export function useOpenProjects() {
    return {
        isOpen: (id: string) => open.value.has(id),
        setOpen(id: string, value: boolean) {
            const next = new Set(open.value)
            if (value) next.add(id)
            else next.delete(id)
            open.value = next
        },
    }
}
