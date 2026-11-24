import type { Ref } from 'vue'

/** What chat scrolling needs from a scroll container, whichever UI kit draws it. */
export interface ScrollController {
    isAtEnd(): boolean
    isSticking(): boolean
    scrollToEnd(options?: { behavior?: ScrollBehavior }): void
}

const END_TOLERANCE_PX = 1

/** A controller over a plain overflowing element, for interfaces without their own scroll area. */
export function elementScrollController(
    viewport: Readonly<Ref<HTMLElement | null>>,
): ScrollController {
    function isAtEnd(): boolean {
        const element = viewport.value
        if (!element) return true
        return element.scrollHeight - element.scrollTop - element.clientHeight <= END_TOLERANCE_PX
    }

    return {
        isAtEnd,
        isSticking: isAtEnd,
        scrollToEnd(options) {
            const element = viewport.value
            element?.scrollTo({ top: element.scrollHeight, behavior: options?.behavior })
        },
    }
}
