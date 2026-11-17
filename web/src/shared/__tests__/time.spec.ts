import { describe, expect, it } from 'vitest'

import { relativeTime } from '@/shared/time'

const now = Date.parse('2026-11-17T12:00:00Z')
const ago = (seconds: number) => new Date(now - seconds * 1000).toISOString()

describe('relativeTime', () => {
    it('formats recent times without a minus sign', () => {
        expect(relativeTime(ago(4 * 60), 'ru', now)).toBe('4 мин')
        expect(relativeTime(ago(3 * 3600), 'en', now)).toBe('3 hr')
    })

    it('says yesterday for one day ago', () => {
        expect(relativeTime(ago(26 * 3600), 'ru', now)).toBe('вчера')
    })

    it('treats future and very recent times as now', () => {
        expect(relativeTime(ago(-30), 'ru', now)).toBe('сейчас')
    })
})
