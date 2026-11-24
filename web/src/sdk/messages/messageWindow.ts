import type { MessageRange, WindowMessage, WindowPage, WindowPageRequest } from './types'

import { isPrompt } from './range'

export interface MessageWindowOptions<Message extends WindowMessage> {
    load: (request: WindowPageRequest) => Promise<WindowPage<Message>>
    pageSize?: number
    maxSize?: number
}

export interface MessageWindowState<Message extends WindowMessage> {
    messages: readonly Message[]
    hasOlder: boolean
    hasNewer: boolean
    lastPrompt: Message | undefined
    opening: boolean
    loadingOlder: boolean
    loadingNewer: boolean
    requested: MessageRange
    landing: string | undefined
    followingLatest: boolean
    missed: number
    error: unknown
}

export type MessageWindowListener<Message extends WindowMessage> = (
    state: MessageWindowState<Message>,
) => void

export const DEFAULT_PAGE_SIZE = 64
export const DEFAULT_MAX_SIZE = 192

export interface MessageWindow<Message extends WindowMessage> {
    readonly state: MessageWindowState<Message>
    subscribe(listener: MessageWindowListener<Message>): () => void
    open(range: MessageRange): Promise<void>
    refresh(): Promise<void>
    loadOlder(): Promise<void>
    loadNewer(): Promise<void>
    receive(message: Message): void
    update(message: Message): void
    followLatest(): void
    landed(): void
    range(): MessageRange
    contains(messageId: string): boolean
}

function initialState<Message extends WindowMessage>(): MessageWindowState<Message> {
    return {
        messages: [],
        hasOlder: false,
        hasNewer: false,
        lastPrompt: undefined,
        opening: false,
        loadingOlder: false,
        loadingNewer: false,
        requested: {},
        landing: undefined,
        followingLatest: true,
        missed: 0,
        error: undefined,
    }
}

export function createMessageWindow<Message extends WindowMessage>(
    options: MessageWindowOptions<Message>,
): MessageWindow<Message> {
    const pageSize = options.pageSize ?? DEFAULT_PAGE_SIZE
    const maxSize = options.maxSize ?? DEFAULT_MAX_SIZE
    const listeners = new Set<MessageWindowListener<Message>>()
    let state = initialState<Message>()
    let generation = 0

    function set(patch: Partial<MessageWindowState<Message>>) {
        state = { ...state, ...patch }
        for (const listener of listeners) listener(state)
    }

    function newestPrompt(candidate: Message | undefined): Message | undefined {
        const current = state.lastPrompt
        if (!candidate) return current
        if (!current || candidate.createTime >= current.createTime) return candidate
        return current
    }

    async function open(range: MessageRange) {
        const ticket = ++generation
        set({ opening: true, requested: range, error: undefined })
        try {
            const page = await options.load({
                ...range,
                pageSize: range.start ? maxSize : pageSize,
            })
            if (ticket !== generation) return
            set({
                messages: page.messages,
                hasOlder: page.hasOlder,
                hasNewer: page.hasNewer,
                lastPrompt: newestPrompt(page.lastPrompt),
                opening: false,
                loadingOlder: false,
                loadingNewer: false,
                landing: range.start,
                followingLatest: !range.start && !page.hasNewer,
                missed: 0,
            })
        } catch (error) {
            if (ticket !== generation) return
            set({ opening: false, error })
            throw error
        }
    }

    function refresh() {
        const first = state.messages.at(0)
        if (!first) return open(state.requested)
        const last = state.messages.at(-1)
        return reload({ start: first.id, end: state.hasNewer ? last?.id : undefined })
    }

    async function reload(range: MessageRange) {
        const ticket = ++generation
        const page = await options.load({ ...range, pageSize: maxSize })
        if (ticket !== generation) return
        set({
            messages: page.messages,
            hasOlder: page.hasOlder,
            hasNewer: page.hasNewer,
            lastPrompt: newestPrompt(page.lastPrompt),
        })
    }

    async function loadOlder() {
        const first = state.messages.at(0)
        if (!first || !state.hasOlder || state.loadingOlder || state.opening) return
        const ticket = generation
        set({ loadingOlder: true })
        try {
            const page = await options.load({ end: first.id, pageSize: pageSize + 1 })
            if (ticket !== generation) return
            const older = page.messages.filter((message) => !contains(message.id))
            const merged = [...older, ...state.messages]
            const overflow = merged.length > maxSize
            set({
                messages: overflow ? merged.slice(0, maxSize) : merged,
                hasOlder: page.hasOlder,
                hasNewer: state.hasNewer || overflow,
                lastPrompt: newestPrompt(page.lastPrompt),
                loadingOlder: false,
                followingLatest: false,
            })
        } catch (error) {
            if (ticket === generation) set({ loadingOlder: false, error })
            throw error
        }
    }

    async function loadNewer() {
        const last = state.messages.at(-1)
        if (!last || !state.hasNewer || state.loadingNewer || state.opening) return
        const ticket = generation
        set({ loadingNewer: true })
        try {
            const page = await options.load({ start: last.id, pageSize: pageSize + 1 })
            if (ticket !== generation) return
            const newer = page.messages.filter((message) => !contains(message.id))
            const merged = [...state.messages, ...newer]
            const overflow = merged.length > maxSize
            set({
                messages: overflow ? merged.slice(-maxSize) : merged,
                hasOlder: state.hasOlder || overflow,
                hasNewer: page.hasNewer,
                lastPrompt: newestPrompt(page.lastPrompt),
                loadingNewer: false,
                missed: page.hasNewer ? state.missed : 0,
            })
        } catch (error) {
            if (ticket === generation) set({ loadingNewer: false, error })
            throw error
        }
    }

    function receive(message: Message) {
        if (contains(message.id)) return update(message)
        const lastPrompt = isPrompt(message) ? newestPrompt(message) : state.lastPrompt
        if (state.hasNewer || state.opening) {
            set({ lastPrompt, missed: state.missed + 1 })
            return
        }
        set({ messages: [...state.messages, message], lastPrompt })
    }

    function update(message: Message) {
        const lastPrompt = state.lastPrompt?.id === message.id ? message : state.lastPrompt
        if (!contains(message.id)) {
            if (lastPrompt !== state.lastPrompt) set({ lastPrompt })
            return
        }
        set({
            messages: state.messages.map((item) => (item.id === message.id ? message : item)),
            lastPrompt,
        })
    }

    function followLatest() {
        if (state.hasNewer || state.opening) return
        const overflow = state.messages.length > maxSize
        if (state.followingLatest && !overflow) return
        set({
            messages: overflow ? state.messages.slice(-maxSize) : state.messages,
            hasOlder: state.hasOlder || overflow,
            followingLatest: true,
        })
    }

    function landed() {
        if (state.landing) set({ landing: undefined })
    }

    function range(): MessageRange {
        const first = state.messages.at(0)
        const last = state.messages.at(-1)
        if (!first || !last) return state.requested
        if (state.hasNewer) return { start: first.id, end: last.id }
        if (state.followingLatest) return {}
        return { start: first.id }
    }

    function contains(messageId: string) {
        return state.messages.some((message) => message.id === messageId)
    }

    return {
        get state() {
            return state
        },
        subscribe(listener) {
            listeners.add(listener)
            return () => listeners.delete(listener)
        },
        open,
        refresh,
        loadOlder,
        loadNewer,
        receive,
        update,
        followLatest,
        landed,
        range,
        contains,
    }
}
