import { afterEach, describe, expect, it, vi } from 'vitest'

import { closeNotice, notices, pressNoticeAction, snackbarNotifier } from '../snackbarNotifier'

describe('snackbarNotifier', () => {
    afterEach(() => {
        notices.splice(0)
        vi.useRealTimers()
    })

    it('replaces a notice with the same id and closes it after its duration', () => {
        vi.useFakeTimers()
        snackbarNotifier.loading('Lost', { id: 'link', duration: 1000 })
        snackbarNotifier.success('Back', { id: 'link', duration: 1000 })
        expect(notices.map((notice) => [notice.key, notice.title, notice.color])).toEqual([
            ['link', 'Back', 'success'],
        ])
        vi.advanceTimersByTime(1000)
        expect(notices).toHaveLength(0)
    })

    it('keeps a loading notice until dismissed and updates its description', () => {
        snackbarNotifier.loading('Lost', { id: 'link' })
        snackbarNotifier.describe('link', 'No network')
        expect(notices.at(0)?.description).toBe('No network')
        snackbarNotifier.dismiss('link')
        expect(notices).toHaveLength(0)
    })

    it('stays open after an action that asks to, and reports only user dismissals', () => {
        const onAction = vi.fn()
        const onUserDismiss = vi.fn()
        snackbarNotifier.loading('Lost', {
            id: 'link',
            action: { label: 'Retry', onAction, keepsNoticeOpen: true },
            onUserDismiss,
        })
        const notice = notices.at(0)!
        pressNoticeAction(notice)
        expect(onAction).toHaveBeenCalledOnce()
        expect(notices).toHaveLength(1)
        snackbarNotifier.dismiss('link')
        expect(onUserDismiss).not.toHaveBeenCalled()
        snackbarNotifier.loading('Lost', { id: 'link', onUserDismiss })
        closeNotice('link', true)
        expect(onUserDismiss).toHaveBeenCalledOnce()
    })
})
