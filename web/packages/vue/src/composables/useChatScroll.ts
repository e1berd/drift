import { computed, onScopeDispose, ref, watch } from 'vue'

import type { Ref } from 'vue'

import type { Message } from '../api'
import type { SessionMessages } from './useMessageWindow'

import { captureScrollAnchor } from '@drift/sdk/messages'
import { elementScrollController } from '../scrollController'
import type { ScrollController } from '../scrollController'

const BOTTOM_TOLERANCE_PX = 24
const USER_SCROLL_EVENTS = ['wheel', 'touchmove', 'keydown', 'pointerdown'] as const

function isFromAgent(message: Message): boolean {
    return message.role === 'assistant' && message.kind !== 'thinking'
}

function appendedSince(current: readonly Message[], previous: readonly Message[]): Message[] {
    const lastKnown = previous.at(-1)?.id
    if (!lastKnown) return []
    const index = current.findIndex((message) => message.id === lastKnown)
    return index === -1 ? [] : current.slice(index + 1)
}

function sharesMessages(current: readonly Message[], previous: readonly Message[]): boolean {
    const known = new Set(previous.map((message) => message.id))
    return current.some((message) => known.has(message.id))
}

export function flashMessage(element: HTMLElement) {
    element.classList.remove('motion-flash')
    requestAnimationFrame(() => element.classList.add('motion-flash'))
}

export function messageElement(viewport: HTMLElement | null, messageId: string) {
    return viewport?.querySelector<HTMLElement>(`[data-message-id="${CSS.escape(messageId)}"]`)
}

/**
 * Scroll state of the chat log: follows the newest messages, keeps the reading position while
 * the message window grows or drops pages, lands on the window's start, and counts agent
 * messages missed while away from the bottom.
 */
export function useChatScroll(
    viewport: Readonly<Ref<HTMLElement | null>>,
    session: SessionMessages,
    controller: ScrollController = elementScrollController(viewport),
) {
    const atBottom = ref(true)
    const unread = ref(0)
    const olderHeld = ref(false)
    const awayFromEnd = ref(false)

    function releaseOlder() {
        olderHeld.value = false
    }

    function distanceFromBottom(element: HTMLElement): number {
        return element.scrollHeight - element.scrollTop - element.clientHeight
    }

    function measure() {
        const element = viewport.value
        if (!element) return
        atBottom.value = distanceFromBottom(element) <= BOTTOM_TOLERANCE_PX
        if (!atBottom.value || session.state.value.hasNewer) return
        unread.value = 0
        awayFromEnd.value = false
        session.followLatest()
    }

    const resizes = typeof ResizeObserver === 'undefined' ? null : new ResizeObserver(measure)
    watch(
        viewport,
        (element, previous) => {
            previous?.removeEventListener('scroll', measure)
            for (const type of USER_SCROLL_EVENTS) previous?.removeEventListener(type, releaseOlder)
            resizes?.disconnect()
            if (!element) return
            element.addEventListener('scroll', measure, { passive: true })
            for (const type of USER_SCROLL_EVENTS)
                element.addEventListener(type, releaseOlder, { passive: true })
            resizes?.observe(element)
            if (element.firstElementChild) resizes?.observe(element.firstElementChild)
            measure()
        },
        { immediate: true },
    )
    onScopeDispose(() => {
        viewport.value?.removeEventListener('scroll', measure)
        for (const type of USER_SCROLL_EVENTS)
            viewport.value?.removeEventListener(type, releaseOlder)
        resizes?.disconnect()
    })

    watch(
        () => session.state.value.hasNewer,
        (hasNewer) => {
            if (hasNewer) awayFromEnd.value = true
        },
        { immediate: true },
    )

    let restoreAnchor: (() => void) | null = null
    let followed = true
    let hadNewer = session.state.value.hasNewer

    watch(
        session.messages,
        () => {
            const element = viewport.value
            if (!element) return
            followed =
                !hadNewer && (controller.isSticking() || controller.isAtEnd() || atBottom.value)
            restoreAnchor = followed ? null : captureScrollAnchor(element)
        },
        { flush: 'pre' },
    )

    function land(messageId: string) {
        const target = messageElement(viewport.value ?? null, messageId)
        session.landed()
        if (!target) return
        olderHeld.value = true
        awayFromEnd.value = true
        target.scrollIntoView({ block: 'start', behavior: 'instant' })
        flashMessage(target)
    }

    watch(
        session.messages,
        (current, previous) => {
            const wasLive = !hadNewer
            hadNewer = session.state.value.hasNewer
            const element = viewport.value
            if (!element) return
            const { landing, hasNewer } = session.state.value
            if (landing) return land(landing)
            if (!hasNewer && previous.length && !sharesMessages(current, previous)) {
                unread.value = 0
                return controller.scrollToEnd({ behavior: 'instant' })
            }
            const appended = appendedSince(current, previous)
            if (wasLive && (followed || appended.some((message) => message.role === 'user')))
                return controller.scrollToEnd({ behavior: 'instant' })
            restoreAnchor?.()
            unread.value += appended.filter(isFromAgent).length
        },
        { flush: 'post' },
    )

    function scrollToEnd() {
        unread.value = 0
        if (session.jumpToLatest()) return
        controller.scrollToEnd()
    }

    return {
        atBottom: computed(() => atBottom.value),
        unread: computed(() => unread.value + session.state.value.missed),
        scrollToEnd,
        olderHeld: computed(() => olderHeld.value),
        stickToEnd: computed(() => !session.state.value.hasNewer && !awayFromEnd.value),
    }
}
