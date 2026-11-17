import { onScopeDispose, watch, type Ref } from 'vue'

/** Calls load whenever the sentinel element scrolls into view, for infinite lists. */
export function useInfiniteSentinel(target: Readonly<Ref<HTMLElement | null>>, load: () => void) {
    if (typeof IntersectionObserver === 'undefined') return
    const observer = new IntersectionObserver(
        (entries) => {
            if (entries.some((e) => e.isIntersecting)) load()
        },
        { rootMargin: '120px' },
    )
    watch(
        target,
        (el, prev) => {
            if (prev) observer.unobserve(prev)
            if (el) observer.observe(el)
        },
        { immediate: true },
    )
    onScopeDispose(() => observer.disconnect())
}
