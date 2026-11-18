import { describe, expect, it } from 'vitest'

import { secondsUntil, usageWindowKind } from '@/shared/usage'

describe('usageWindowKind', () => {
    it('names Claude windows by id', () => {
        expect(usageWindowKind({ id: 'five_hour' })).toBe('session')
        expect(usageWindowKind({ id: 'seven_day_opus', windowMinutes: 10080 })).toBe('weekOpus')
    })

    it('names Codex windows by length', () => {
        expect(usageWindowKind({ id: 'primary', windowMinutes: 300 })).toBe('session')
        expect(usageWindowKind({ id: 'secondary', windowMinutes: 10080 })).toBe('week')
        expect(usageWindowKind({ id: 'secondary', windowMinutes: 60 })).toBe('other')
    })
})

describe('secondsUntil', () => {
    it('counts down and stops at zero', () => {
        const now = Date.parse('2026-11-18T10:00:00Z')
        expect(secondsUntil('2026-11-18T12:00:00Z', now)).toBe(7200)
        expect(secondsUntil('2026-11-18T09:00:00Z', now)).toBe(0)
        expect(secondsUntil(undefined, now)).toBeNull()
    })
})
