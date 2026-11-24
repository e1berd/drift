import { QueryClient } from '@tanstack/vue-query'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

class FakeEventSource extends EventTarget {
    static readonly CONNECTING = 0
    static readonly OPEN = 1
    static readonly CLOSED = 2
    static instances: FakeEventSource[] = []

    readyState = FakeEventSource.CONNECTING
    readonly url: string

    constructor(url: URL | string) {
        super()
        this.url = String(url)
        FakeEventSource.instances.push(this)
    }

    open() {
        this.readyState = FakeEventSource.OPEN
        this.dispatchEvent(new Event('open'))
    }

    receive(id: string, data: unknown) {
        this.dispatchEvent(
            new MessageEvent('message', { data: JSON.stringify(data), lastEventId: id }),
        )
    }

    signal(type: 'ready' | 'ping', id = '') {
        this.dispatchEvent(new MessageEvent(type, { data: '{}', lastEventId: id }))
    }

    fail() {
        this.dispatchEvent(new Event('error'))
    }

    close() {
        this.readyState = FakeEventSource.CLOSED
    }
}

let disconnect: (() => void) | undefined

async function loadEventStream() {
    vi.resetModules()
    const connection = await import('../realtimeConnection')
    disconnect = connection.disconnectRealtime
    const stream = await import('../useEventStream')
    return { ...stream, ...connection }
}

function latestSource() {
    return FakeEventSource.instances.at(-1)!
}

describe('startEventStream', () => {
    beforeEach(() => {
        vi.useFakeTimers()
        vi.spyOn(Math, 'random').mockReturnValue(1)
        FakeEventSource.instances = []
        vi.stubGlobal('EventSource', FakeEventSource)
    })

    afterEach(() => {
        disconnect?.()
        vi.restoreAllMocks()
        vi.useRealTimers()
        vi.unstubAllGlobals()
    })

    it('reconnects after an error and resumes after the last event', async () => {
        const { startEventStream, useEventStream } = await loadEventStream()
        const { connected } = useEventStream()

        startEventStream(new QueryClient())
        const first = latestSource()
        first.open()
        first.receive('42', { type: 'noop' })
        first.fail()

        expect(connected.value).toBe(false)
        expect(first.readyState).toBe(FakeEventSource.CLOSED)

        vi.advanceTimersByTime(1000)

        const second = latestSource()
        expect(second).not.toBe(first)
        expect(new URL(second.url).searchParams.get('last_event_id')).toBe('42')
        second.open()
        expect(connected.value).toBe(true)
    })

    it('resumes from the ready cursor when no change event arrived yet', async () => {
        const { startEventStream } = await loadEventStream()

        startEventStream(new QueryClient())
        const first = latestSource()
        first.open()
        first.signal('ready', 'boot-7')
        first.fail()
        vi.advanceTimersByTime(1000)

        expect(new URL(latestSource().url).searchParams.get('last_event_id')).toBe('boot-7')
    })

    it('backs off exponentially while the server stays unreachable', async () => {
        const { startEventStream } = await loadEventStream()

        startEventStream(new QueryClient())
        latestSource().fail()
        vi.advanceTimersByTime(1000)
        expect(FakeEventSource.instances).toHaveLength(2)

        latestSource().fail()
        vi.advanceTimersByTime(1999)
        expect(FakeEventSource.instances).toHaveLength(2)
        vi.advanceTimersByTime(1)
        expect(FakeEventSource.instances).toHaveLength(3)
    })

    it('drops a connection that went silent and opens a new one', async () => {
        const { startEventStream, useRealtimeConnection } = await loadEventStream()
        const { status } = useRealtimeConnection()

        startEventStream(new QueryClient())
        const first = latestSource()
        first.open()
        vi.advanceTimersByTime(30_000)
        first.signal('ping')
        vi.advanceTimersByTime(30_000)
        expect(status.value).toBe('open')

        vi.advanceTimersByTime(15_000)
        expect(status.value).toBe('lost')
        expect(first.readyState).toBe(FakeEventSource.CLOSED)

        vi.advanceTimersByTime(1000)
        expect(latestSource()).not.toBe(first)
    })

    it('reconnects immediately when the browser comes back online', async () => {
        const { startEventStream } = await loadEventStream()

        startEventStream(new QueryClient())
        latestSource().open()
        window.dispatchEvent(new Event('offline'))
        expect(FakeEventSource.instances).toHaveLength(1)

        window.dispatchEvent(new Event('online'))
        expect(FakeEventSource.instances).toHaveLength(2)
    })
})
