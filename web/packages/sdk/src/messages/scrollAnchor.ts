export interface ScrollAnchorOptions {
    itemSelector?: string
    idOf?: (element: HTMLElement) => string | undefined
    candidates?: number
}

interface AnchorPoint {
    id: string
    offset: number
}

const DEFAULT_ITEM_SELECTOR = '[data-message-id]'

function defaultIdOf(element: HTMLElement) {
    return element.dataset.messageId
}

function visibleAnchors(
    viewport: HTMLElement,
    items: HTMLElement[],
    idOf: (element: HTMLElement) => string | undefined,
    limit: number,
): AnchorPoint[] {
    const top = viewport.getBoundingClientRect().top
    const anchors: AnchorPoint[] = []
    for (const item of items) {
        const rect = item.getBoundingClientRect()
        const id = idOf(item)
        if (!id || rect.bottom <= top) continue
        anchors.push({ id, offset: rect.top - top })
        if (anchors.length >= limit) break
    }
    return anchors
}

/**
 * Remembers where the first visible items sit in the viewport; the returned function scrolls so
 * the first item that still exists is back at the same spot after the list above it changed.
 */
export function captureScrollAnchor(
    viewport: HTMLElement,
    options: ScrollAnchorOptions = {},
): () => void {
    const selector = options.itemSelector ?? DEFAULT_ITEM_SELECTOR
    const idOf = options.idOf ?? defaultIdOf
    const items = [...viewport.querySelectorAll<HTMLElement>(selector)]
    const anchors = visibleAnchors(viewport, items, idOf, options.candidates ?? 4)

    return function restore() {
        const top = viewport.getBoundingClientRect().top
        const byId = new Map(
            [...viewport.querySelectorAll<HTMLElement>(selector)].map((item) => [idOf(item), item]),
        )
        for (const anchor of anchors) {
            const element = byId.get(anchor.id)
            if (!element) continue
            const drift = element.getBoundingClientRect().top - top - anchor.offset
            if (drift) viewport.scrollTop += drift
            return
        }
    }
}
