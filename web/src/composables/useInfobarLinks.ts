import { onBeforeUnmount, onMounted, shallowRef, toValue, watch } from 'vue'

import type { MaybeRefOrGetter, Ref } from 'vue'

import { routeLinks } from '@/shared/linkRoutes'
import type { LinkEnds, LinkRoute } from '@/shared/linkRoutes'

export interface InfobarLink {
    key: string
    messageId: string
}

export interface DrawnLink extends LinkRoute {
    messageId: string
}

const CARD_HEAD_OFFSET = 18
const EDGE_PADDING = 2
const FALLBACK_GUTTER = 40

interface Frame {
    top: number
    bottom: number
}

function relative(rect: DOMRect, origin: DOMRect) {
    return {
        top: rect.top - origin.top,
        bottom: rect.bottom - origin.top,
        left: rect.left - origin.left,
        right: rect.right - origin.left,
    }
}

function clampToFrame(y: number, frame: Frame): { y: number; hidden: boolean } {
    const low = frame.top + EDGE_PADDING
    const high = frame.bottom - EDGE_PADDING
    if (y < low) return { y: low, hidden: true }
    if (y > high) return { y: high, hidden: true }
    return { y, hidden: false }
}

function measure(container: HTMLElement, links: InfobarLink[]): DrawnLink[] {
    const infobar = container.querySelector<HTMLElement>('[data-infobar-scroll]')
    const chat = container.querySelector<HTMLElement>('[role="log"]')
    if (!infobar || !chat || !links.length) return []
    const origin = container.getBoundingClientRect()
    const infobarBox = relative(infobar.getBoundingClientRect(), origin)
    const chatBox = relative(chat.getBoundingClientRect(), origin)
    const ends: LinkEnds[] = []
    const cardLefts: number[] = []
    const shown: InfobarLink[] = []
    for (const link of links) {
        const row = container.querySelector<HTMLElement>(
            `[data-link-target="${CSS.escape(link.key)}"]`,
        )
        if (!row) continue
        const rowBox = relative(row.getBoundingClientRect(), origin)
        const target = clampToFrame((rowBox.top + rowBox.bottom) / 2, infobarBox)
        const card = chat.querySelector<HTMLElement>(
            `[data-message-id="${CSS.escape(link.messageId)}"]`,
        )
        const cardBox = card ? relative(card.getBoundingClientRect(), origin) : null
        const source = clampToFrame(
            cardBox
                ? cardBox.top + Math.min(CARD_HEAD_OFFSET, (cardBox.bottom - cardBox.top) / 2)
                : chatBox.top,
            chatBox,
        )
        if (cardBox && !source.hidden) cardLefts.push(cardBox.left)
        shown.push(link)
        ends.push({
            id: link.key,
            sourceY: source.y,
            targetY: target.y,
            sourceHidden: source.hidden,
            targetHidden: target.hidden,
        })
    }
    const left = infobarBox.right
    const right = cardLefts.length ? Math.min(...cardLefts) : left + FALLBACK_GUTTER
    return routeLinks(ends, { left, right: Math.max(right, left + FALLBACK_GUTTER / 2) }).map(
        (route, index) => ({ ...route, messageId: shown[index]!.messageId }),
    )
}

function messageIdUnder(target: EventTarget | null): string | null {
    if (!(target instanceof Element)) return null
    const element = target.closest<HTMLElement>('[data-message-id],[data-link-source]')
    return element?.dataset.messageId ?? element?.dataset.linkSource ?? null
}

/** Live routes between info bar rows and the chat cards that created them, redrawn on every frame of scroll or resize. */
export function useInfobarLinks(
    container: Readonly<Ref<HTMLElement | null>>,
    links: MaybeRefOrGetter<InfobarLink[]>,
) {
    const routes = shallowRef<DrawnLink[]>([])
    const activeMessageId = shallowRef<string | null>(null)
    let frame = 0
    const resize = new ResizeObserver(schedule)

    function redraw() {
        frame = 0
        const element = container.value
        routes.value = element ? measure(element, toValue(links)) : []
    }

    function schedule() {
        if (!frame) frame = requestAnimationFrame(redraw)
    }

    function observe(element: HTMLElement) {
        resize.disconnect()
        resize.observe(element)
        element
            .querySelectorAll<HTMLElement>(
                '[data-infobar-scroll], [role="log"], [data-slot="scroll-area-content"]',
            )
            .forEach((child) => resize.observe(child))
    }

    function onPointerOver(event: PointerEvent) {
        activeMessageId.value = messageIdUnder(event.target)
    }

    function onPointerLeave() {
        activeMessageId.value = null
    }

    function attach(element: HTMLElement) {
        element.addEventListener('scroll', schedule, { capture: true, passive: true })
        element.addEventListener('pointerover', onPointerOver)
        element.addEventListener('pointerleave', onPointerLeave)
        observe(element)
        schedule()
    }

    function detach(element: HTMLElement) {
        element.removeEventListener('scroll', schedule, { capture: true })
        element.removeEventListener('pointerover', onPointerOver)
        element.removeEventListener('pointerleave', onPointerLeave)
    }

    onMounted(() => {
        watch(
            container,
            (element, previous) => {
                if (previous) detach(previous)
                if (element) attach(element)
            },
            { immediate: true },
        )
        watch(() => toValue(links), schedule, { deep: true, flush: 'post' })
    })

    onBeforeUnmount(() => {
        if (container.value) detach(container.value)
        resize.disconnect()
        cancelAnimationFrame(frame)
    })

    return { routes, activeMessageId, refresh: schedule }
}
