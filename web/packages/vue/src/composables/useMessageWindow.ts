import { computed, inject, onScopeDispose, provide, shallowRef, toValue, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import type { InjectionKey, MaybeRefOrGetter } from 'vue'

import type { Message, WireEvent } from '../api'
import type { MessageRange, MessageWindow, MessageWindowState } from '@drift/sdk/messages'

import { useEventStream } from './useEventStream'
import { listMessages } from '../api'
import { createMessageWindow, rangeFromQuery, sameRange } from '@drift/sdk/messages'
import { reportError } from '@drift/sdk/errors'

const RETAINED_WINDOWS = 6
const windows = new Map<string, MessageWindow<Message>>()
let streaming = false

function loaderFor(sessionId: string) {
    return async (request: { start?: string; end?: string; pageSize: number }) => {
        const { data } = await listMessages({
            path: { sessionId },
            query: { start: request.start, end: request.end, page_size: request.pageSize },
        })
        return data
    }
}

function retain(sessionId: string): MessageWindow<Message> {
    const known = windows.get(sessionId)
    if (known) {
        windows.delete(sessionId)
        windows.set(sessionId, known)
        return known
    }
    const created = createMessageWindow<Message>({ load: loaderFor(sessionId) })
    windows.set(sessionId, created)
    const oldest = windows.keys().next().value
    if (windows.size > RETAINED_WINDOWS && oldest) windows.delete(oldest)
    return created
}

function deliver(event: WireEvent) {
    if (event.type === 'resync') {
        for (const window of windows.values()) window.refresh().catch(reportError)
        return
    }
    if (event.type !== 'message.created' && event.type !== 'message.updated') return
    const message = event.data as Message
    const window = windows.get(message.sessionId)
    if (!window) return
    if (event.type === 'message.created') window.receive(message)
    else window.update(message)
}

function streamIntoWindows() {
    if (streaming) return
    streaming = true
    useEventStream().on(deliver)
}

function routeRange(query: Record<string, unknown>): MessageRange {
    return rangeFromQuery(query)
}

function useRangeInRoute(window: () => MessageWindow<Message>) {
    const route = useRoute()
    const router = useRouter()

    function writeRange(range: MessageRange) {
        if (sameRange(routeRange(route.query), range)) return
        void router.replace({ query: { ...route.query, start: range.start, end: range.end } })
    }

    function followRoute() {
        const current = window()
        const wanted = routeRange(route.query)
        if (sameRange(wanted, current.range()) && current.state.messages.length) return
        if (sameRange(wanted, current.state.requested) && current.state.opening) return
        current.open(wanted).catch((error: unknown) => {
            reportError(error)
            if (wanted.start || wanted.end) writeRange({})
        })
    }

    return { writeRange, followRoute, routeQuery: () => route.query }
}

export function createSessionMessages(sessionId: MaybeRefOrGetter<string>) {
    streamIntoWindows()
    const window = computed(() => retain(toValue(sessionId)))
    const state = shallowRef<MessageWindowState<Message>>(window.value.state)
    const { writeRange, followRoute, routeQuery } = useRangeInRoute(() => window.value)

    let unsubscribe = () => {}
    watch(
        window,
        (current) => {
            unsubscribe()
            state.value = current.state
            unsubscribe = current.subscribe((next) => {
                state.value = next
                if (!next.opening) writeRange(current.range())
            })
            followRoute()
        },
        { immediate: true },
    )
    onScopeDispose(() => unsubscribe())
    watch(() => [routeQuery().start, routeQuery().end], followRoute)

    function reveal(messageId: string) {
        if (window.value.contains(messageId)) return true
        writeRange({ start: messageId })
        return false
    }

    function jumpToLatest() {
        if (!state.value.hasNewer) return false
        writeRange({})
        return true
    }

    return {
        state,
        messages: computed(() => state.value.messages),
        loadOlder: () => window.value.loadOlder().catch(reportError),
        loadNewer: () => window.value.loadNewer().catch(reportError),
        followLatest: () => window.value.followLatest(),
        landed: () => window.value.landed(),
        reveal,
        jumpToLatest,
    }
}

export type SessionMessages = ReturnType<typeof createSessionMessages>

const sessionMessagesKey: InjectionKey<SessionMessages> = Symbol('session-messages')

export function provideSessionMessages(sessionId: MaybeRefOrGetter<string>): SessionMessages {
    const messages = createSessionMessages(sessionId)
    provide(sessionMessagesKey, messages)
    return messages
}

export function useSessionMessages(): SessionMessages {
    const messages = inject(sessionMessagesKey)
    if (!messages) throw new Error('useSessionMessages needs provideSessionMessages in a parent')
    return messages
}
