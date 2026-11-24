import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { effectScope } from 'vue'

const api = vi.hoisted(() => ({
    listDrafts: vi.fn(),
    updateDraft: vi.fn(),
    deleteDraft: vi.fn(),
}))

vi.mock('../../api', () => api)
vi.mock('@drift/sdk/errors', () => ({ reportError: vi.fn() }))

async function freshDrafts() {
    vi.resetModules()
    return import('../useDrafts')
}

describe('useDrafts offline safety', () => {
    beforeEach(() => {
        vi.useFakeTimers()
        localStorage.clear()
        api.listDrafts.mockResolvedValue({ data: { drafts: [] } })
        api.updateDraft.mockRejectedValue(new TypeError('Failed to fetch'))
        api.deleteDraft.mockResolvedValue({})
    })
    afterEach(() => {
        vi.useRealTimers()
        vi.restoreAllMocks()
    })

    it('keeps a draft the server never received across a reload', async () => {
        const first = await freshDrafts()
        const scope = effectScope()
        const draft = scope.run(() => first.useDraft('session:1'))!
        draft.set('a very long prompt')
        await vi.runAllTimersAsync()
        scope.stop()

        const reloaded = await freshDrafts()
        await reloaded.loadDrafts()
        const restored = effectScope().run(() => reloaded.useDraft('session:1'))!
        expect(restored.text.value).toBe('a very long prompt')
    })

    it('forgets the backup once the server confirms the draft', async () => {
        api.updateDraft.mockResolvedValue({})
        const drafts = await freshDrafts()
        const draft = effectScope().run(() => drafts.useDraft('session:1'))!
        draft.set('saved text')
        await vi.runAllTimersAsync()

        expect(localStorage.getItem('drift.unsyncedDrafts')).toBe('{}')
    })

    it('discards a sent draft only when it was not edited after sending', async () => {
        const drafts = await freshDrafts()
        const draft = effectScope().run(() => drafts.useDraft('new:chat'))!
        draft.set('first prompt and more')
        drafts.discardSentDraft('new:chat', 'first prompt')
        expect(drafts.useDraftIndex().has('new:chat')).toBe(true)

        drafts.discardSentDraft('new:chat', 'first prompt and more')
        expect(drafts.useDraftIndex().has('new:chat')).toBe(false)
    })
})
