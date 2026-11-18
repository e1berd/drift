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

    failPermanently() {
        this.readyState = FakeEventSource.CLOSED
        this.dispatchEvent(new Event('error'))
    }

    close() {
        this.readyState = FakeEventSource.CLOSED
    }
}

async function loadEventStream() {
    vi.resetModules()
    return import('@/composables/useEventStream')
}

describe('startEventStream', () => {
    beforeEach(() => {
        vi.useFakeTimers()
        FakeEventSource.instances = []
        vi.stubGlobal('EventSource', FakeEventSource)
    })

    afterEach(() => {
        vi.useRealTimers()
        vi.unstubAllGlobals()
    })

    it('reopens a stream the browser gave up on and resumes after the last event', async () => {
        const { startEventStream, useEventStream } = await loadEventStream()
        const { connected } = useEventStream()

        startEventStream(new QueryClient())
        const first = FakeEventSource.instances.at(0)!
        first.open()
        first.receive('42', { type: 'noop' })
        first.failPermanently()

        expect(connected.value).toBe(false)
        expect(FakeEventSource.instances).toHaveLength(1)

        vi.advanceTimersByTime(3000)

        const second = FakeEventSource.instances.at(1)!
        expect(new URL(second.url).searchParams.get('last_event_id')).toBe('42')
        second.open()
        expect(connected.value).toBe(true)
    })

    it('leaves reconnecting to the browser while the stream is not closed', async () => {
        const { startEventStream } = await loadEventStream()

        startEventStream(new QueryClient())
        FakeEventSource.instances.at(0)!.dispatchEvent(new Event('error'))
        vi.advanceTimersByTime(3000)

        expect(FakeEventSource.instances).toHaveLength(1)
    })
})
