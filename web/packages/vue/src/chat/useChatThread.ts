import { watch } from 'vue'

import type { Ref } from 'vue'

import { useChatScroll } from '../composables/useChatScroll'
import { useInfiniteSentinel } from '../composables/useInfiniteSentinel'
import { useSessionMessages } from '../composables/useMessageWindow'
import type { ScrollController } from '../scrollController'

const NEAR_TOP_PX = 120

function isNearTop(sentinel: HTMLElement, scroller: HTMLElement) {
    return (
        sentinel.getBoundingClientRect().bottom >=
        scroller.getBoundingClientRect().top - NEAR_TOP_PX
    )
}

/**
 * The message log of the provided session: scrolling (see useChatScroll) plus loading older and
 * newer pages when the `top` and `bottom` sentinels come into view.
 */
export function useChatThread(
    viewport: Readonly<Ref<HTMLElement | null>>,
    top: Readonly<Ref<HTMLElement | null>>,
    bottom: Readonly<Ref<HTMLElement | null>>,
    controller?: ScrollController,
) {
    const session = useSessionMessages()
    const scroll = useChatScroll(viewport, session, controller)

    function loadOlderUnlessHeld() {
        if (!scroll.olderHeld.value) void session.loadOlder()
    }

    useInfiniteSentinel(top, loadOlderUnlessHeld)
    useInfiniteSentinel(bottom, () => void session.loadNewer())
    watch(scroll.olderHeld, (held) => {
        if (!held && top.value && viewport.value && isNearTop(top.value, viewport.value))
            loadOlderUnlessHeld()
    })

    return { session, state: session.state, ...scroll }
}
