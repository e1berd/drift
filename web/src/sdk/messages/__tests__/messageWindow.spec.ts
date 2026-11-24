import { describe, expect, it } from 'vitest'

import type { WindowMessage, WindowPage, WindowPageRequest } from '../types'

import { createMessageWindow } from '../messageWindow'

function transcript(count: number): WindowMessage[] {
    return Array.from({ length: count }, (_, index) => ({
        id: `m${index}`,
        role: index % 10 === 0 ? 'user' : 'assistant',
        createTime: new Date(Date.UTC(2026, 0, 1, 0, 0, 0, index)).toISOString(),
        body: { text: `text ${index}` },
    }))
}

function serve(all: WindowMessage[]) {
    const requests: WindowPageRequest[] = []
    async function load(request: WindowPageRequest): Promise<WindowPage<WindowMessage>> {
        requests.push(request)
        const indexOf = (id: string | undefined) => all.findIndex((message) => message.id === id)
        const lastPrompt = all.findLast((message) => message.role === 'user')
        if (request.start) {
            const from = indexOf(request.start)
            const until = request.end ? indexOf(request.end) + 1 : all.length
            const to = Math.min(from + request.pageSize, until)
            return {
                messages: all.slice(from, to),
                hasOlder: from > 0,
                hasNewer: to < all.length,
                lastPrompt,
            }
        }
        const to = request.end ? indexOf(request.end) + 1 : all.length
        const from = Math.max(0, to - request.pageSize)
        return {
            messages: all.slice(from, to),
            hasOlder: from > 0,
            hasNewer: to < all.length,
            lastPrompt,
        }
    }
    return { load, requests }
}

function ids(window: { state: { messages: readonly WindowMessage[] } }) {
    const { messages } = window.state
    return [messages.at(0)?.id, messages.at(-1)?.id, messages.length]
}

describe('createMessageWindow', () => {
    it('opens on the newest page and keeps the URL clean', async () => {
        const window = createMessageWindow({ load: serve(transcript(300)).load })
        await window.open({})
        expect(ids(window)).toEqual(['m236', 'm299', 64])
        expect(window.state.hasOlder).toBe(true)
        expect(window.state.hasNewer).toBe(false)
        expect(window.state.lastPrompt?.id).toBe('m290')
        expect(window.range()).toEqual({})
    })

    it('slides up page by page and drops the bottom past the limit', async () => {
        const window = createMessageWindow({ load: serve(transcript(300)).load })
        await window.open({})
        await window.loadOlder()
        expect(ids(window)).toEqual(['m172', 'm299', 128])
        expect(window.range()).toEqual({ start: 'm172' })
        await window.loadOlder()
        await window.loadOlder()
        expect(ids(window)).toEqual(['m44', 'm235', 192])
        expect(window.state.hasNewer).toBe(true)
        expect(window.range()).toEqual({ start: 'm44', end: 'm235' })
    })

    it('slides back down and returns to the newest messages', async () => {
        const window = createMessageWindow({ load: serve(transcript(300)).load })
        await window.open({ start: 'm10', end: 'm73' })
        expect(ids(window)).toEqual(['m10', 'm73', 64])
        expect(window.state.landing).toBe('m10')
        await window.loadNewer()
        await window.loadNewer()
        await window.loadNewer()
        expect(ids(window)).toEqual(['m74', 'm265', 192])
        expect(window.state.hasOlder).toBe(true)
        await window.loadNewer()
        expect(ids(window)).toEqual(['m108', 'm299', 192])
        expect(window.state.hasNewer).toBe(false)
        expect(window.range()).toEqual({ start: 'm108' })
        window.followLatest()
        expect(window.range()).toEqual({})
    })

    it('lands on a prompt that is outside the window', async () => {
        const window = createMessageWindow({ load: serve(transcript(300)).load, pageSize: 5 })
        await window.open({})
        expect(window.contains('m290')).toBe(false)
        await window.open({ start: 'm290' })
        expect(ids(window)).toEqual(['m290', 'm299', 10])
        expect(window.state.landing).toBe('m290')
        window.landed()
        expect(window.state.landing).toBeUndefined()
    })

    it('appends live messages only while the newest end is loaded', async () => {
        const all = transcript(300)
        const window = createMessageWindow({ load: serve(all).load })
        await window.open({ start: 'm0' })
        const prompt = {
            id: 'm300',
            role: 'user',
            createTime: '2027-01-01T00:00:00Z',
            body: { text: 'hi' },
        }
        window.receive(prompt)
        expect(window.contains('m300')).toBe(false)
        expect(window.state.missed).toBe(1)
        expect(window.state.lastPrompt?.id).toBe('m300')

        await window.open({})
        window.receive(prompt)
        expect(window.state.messages.at(-1)?.id).toBe('m300')
        window.update({ ...prompt, body: { text: 'edited' } })
        expect(window.state.messages.at(-1)?.body.text).toBe('edited')
        expect(window.state.lastPrompt?.body.text).toBe('edited')
    })

    it('ignores a response that a newer open superseded', async () => {
        const { load } = serve(transcript(300))
        const window = createMessageWindow({ load })
        const first = window.open({ start: 'm0' })
        const second = window.open({})
        await Promise.all([first, second])
        expect(ids(window)).toEqual(['m236', 'm299', 64])
    })
})
