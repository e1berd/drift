import { computed, onScopeDispose, ref, watch } from 'vue'
import { createScrollAreaController } from '@ui-sdk/vue/scroll-area'

import type { Ref } from 'vue'

import type { Message } from '@/gen'

const BOTTOM_TOLERANCE_PX = 24

function isFromAgent(message: Message): boolean {
    return message.role === 'assistant' && message.kind !== 'thinking'
}

/**
 * Scroll state of the chat log: keeps the bottom in view while new messages arrive, keeps the
 * reading position when older pages are prepended, and counts agent messages missed while away.
 */
export function useChatScroll(
    viewport: Readonly<Ref<HTMLElement | null>>,
    messages: Readonly<Ref<Message[]>>,
) {
    const controller = createScrollAreaController()
    const atBottom = ref(true)
    const unread = ref(0)

    function distanceFromBottom(element: HTMLElement): number {
        return element.scrollHeight - element.scrollTop - element.clientHeight
    }

    function measure() {
        const element = viewport.value
        if (!element) return
        atBottom.value = distanceFromBottom(element) <= BOTTOM_TOLERANCE_PX
        if (atBottom.value) unread.value = 0
    }

    const resizes = typeof ResizeObserver === 'undefined' ? null : new ResizeObserver(measure)
    watch(
        viewport,
        (element, previous) => {
            previous?.removeEventListener('scroll', measure)
            resizes?.disconnect()
            if (!element) return
            element.addEventListener('scroll', measure, { passive: true })
            resizes?.observe(element)
            if (element.firstElementChild) resizes?.observe(element.firstElementChild)
            measure()
        },
        { immediate: true },
    )
    onScopeDispose(() => {
        viewport.value?.removeEventListener('scroll', measure)
        resizes?.disconnect()
    })

    let offsetFromBottom: number | null = null
    let followed = true

    watch(
        messages,
        (current, previous) => {
            const element = viewport.value
            if (!element) return
            followed = controller.isSticking() || controller.isAtEnd() || atBottom.value
            offsetFromBottom =
                previous.length && current.at(0)?.id !== previous.at(0)?.id
                    ? element.scrollHeight - element.scrollTop
                    : null
        },
        { flush: 'pre' },
    )

    watch(
        messages,
        (current, previous) => {
            const element = viewport.value
            if (!element) return
            if (offsetFromBottom !== null && !followed) {
                element.scrollTop = element.scrollHeight - offsetFromBottom
                return
            }
            const appended = appendedSince(current, previous)
            if (followed || appended.some((message) => message.role === 'user')) {
                controller.scrollToEnd({ behavior: 'instant' })
                return
            }
            unread.value += appended.filter(isFromAgent).length
        },
        { flush: 'post' },
    )

    function appendedSince(current: Message[], previous: Message[]): Message[] {
        const lastKnown = previous.at(-1)?.id
        if (!lastKnown) return []
        const index = current.findIndex((message) => message.id === lastKnown)
        return index === -1 ? [] : current.slice(index + 1)
    }

    function scrollToEnd() {
        unread.value = 0
        controller.scrollToEnd()
    }

    return {
        controller,
        atBottom: computed(() => atBottom.value),
        unread: computed(() => unread.value),
        scrollToEnd,
    }
}
