import { describe, expect, it } from 'vitest'

import {
    formatResetClock,
    formatResetDate,
    headlineWindows,
    secondsUntil,
    usageLevel,
    usageWindowKind,
} from '../usage'

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

describe('formatResetClock', () => {
    it('shows only the time for a reset later today', () => {
        const now = new Date(2026, 10, 18, 10, 0).getTime()
        expect(formatResetClock(new Date(2026, 10, 18, 15, 0).toISOString(), 'en-GB', now)).toBe(
            '15:00',
        )
    })

    it('adds the weekday for a reset within the week', () => {
        const now = new Date(2026, 10, 18, 10, 0).getTime()
        expect(formatResetClock(new Date(2026, 10, 20, 9, 30).toISOString(), 'en-GB', now)).toBe(
            'Fri 09:30',
        )
    })
})

describe('formatResetDate', () => {
    it('shows the reset date and time in the local timezone', () => {
        const iso = new Date(2026, 10, 19, 15, 0).toISOString()
        expect(formatResetDate(iso, 'ru')).toBe('19 нояб., 15:00')
        expect(formatResetDate(iso, 'en')).toBe('Nov 19, 15:00')
    })
})

describe('headlineWindows', () => {
    const windows = [
        { id: 'five_hour', usedPercent: 40 },
        { id: 'seven_day', usedPercent: 20 },
        { id: 'seven_day_opus', usedPercent: 70 },
        { id: 'seven_day_sonnet', usedPercent: 5 },
    ]

    it('takes the tighter of the shared and the model week', () => {
        expect(headlineWindows(windows, 'opus')).toEqual({
            session: windows.at(0),
            week: windows.at(2),
        })
        expect(headlineWindows(windows, 'sonnet').week).toBe(windows.at(1))
        expect(headlineWindows(windows, 'haiku').week).toBe(windows.at(1))
    })

    it('finds Codex windows by length', () => {
        const codex = [
            { id: 'primary', windowMinutes: 300, usedPercent: 12 },
            { id: 'secondary', windowMinutes: 10080, usedPercent: 3 },
        ]
        expect(headlineWindows(codex, 'default')).toEqual({
            session: codex.at(0),
            week: codex.at(1),
        })
    })

    it('leaves a window out when the agent did not report it', () => {
        expect(headlineWindows([], 'default')).toEqual({ session: undefined, week: undefined })
    })
})

describe('usageLevel', () => {
    it('raises the level at 75 and 90 percent', () => {
        expect(usageLevel(74)).toBe('normal')
        expect(usageLevel(75)).toBe('high')
        expect(usageLevel(90)).toBe('critical')
    })
})
