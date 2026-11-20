import { onScopeDispose, ref, watch, type CSSProperties, type Ref } from 'vue'

const MARKER_SELECTOR = '[data-project-marker]'
const VIEWPORT_SELECTOR = '[data-slot="scroll-area-viewport"]'

function headOf(marker: Element | undefined) {
    return marker?.nextElementSibling as HTMLElement | null | undefined
}

export function useStickyProjects(list: Readonly<Ref<HTMLElement | null>>) {
    const currentIndex = ref(-1)
    const previousHeadHeight = ref(0)
    let viewport: HTMLElement | null = null
    let frame = 0

    function findCurrentIndex(markers: HTMLElement[], viewportTop: number) {
        let found = -1
        markers.forEach((marker, index) => {
            const stackAbove = index > 0 ? (headOf(markers[index - 1])?.offsetHeight ?? 0) : 0
            if (marker.getBoundingClientRect().top <= viewportTop + stackAbove + 0.5) found = index
        })
        return found
    }

    function measure() {
        frame = 0
        if (!list.value || !viewport) return
        const markers = [...list.value.querySelectorAll<HTMLElement>(MARKER_SELECTOR)]
        currentIndex.value = findCurrentIndex(markers, viewport.getBoundingClientRect().top)
        previousHeadHeight.value = headOf(markers[currentIndex.value - 1])?.offsetHeight ?? 0
    }

    function scheduleMeasure() {
        if (!frame) frame = requestAnimationFrame(measure)
    }

    const resizeObserver =
        typeof ResizeObserver === 'undefined' ? null : new ResizeObserver(scheduleMeasure)

    function detach() {
        viewport?.removeEventListener('scroll', scheduleMeasure)
        resizeObserver?.disconnect()
        viewport = null
    }

    watch(
        list,
        (element) => {
            detach()
            viewport = element?.closest<HTMLElement>(VIEWPORT_SELECTOR) ?? null
            if (!element || !viewport) return
            viewport.addEventListener('scroll', scheduleMeasure, { passive: true })
            resizeObserver?.observe(element)
            resizeObserver?.observe(viewport)
            scheduleMeasure()
        },
        { immediate: true, flush: 'post' },
    )

    onScopeDispose(() => {
        detach()
        cancelAnimationFrame(frame)
    })

    function pinStyle(index: number, count: number): CSSProperties {
        const current = currentIndex.value
        if (index > current) return { bottom: 0, zIndex: count - index }
        const belowPrevious = index === current && current > 0
        return {
            top: belowPrevious ? `${previousHeadHeight.value}px` : 0,
            zIndex: count + 1 + index,
        }
    }

    return { pinStyle }
}
