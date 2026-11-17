import { onScopeDispose, ref } from 'vue'

/** Reactive matchMedia result. */
export function useMediaQuery(query: string) {
    const mql = typeof window !== 'undefined' && window.matchMedia ? window.matchMedia(query) : null
    const matches = ref(mql?.matches ?? false)
    const update = (e: MediaQueryListEvent) => {
        matches.value = e.matches
    }
    mql?.addEventListener('change', update)
    onScopeDispose(() => mql?.removeEventListener('change', update))
    return matches
}
