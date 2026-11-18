import { describe, expect, it } from 'vitest'

import { ancestorsOf, folderOf } from '@/composables/useSessionFiles'

describe('ancestorsOf', () => {
    it('lists parent folders outermost first', () => {
        expect(ancestorsOf('src/app/main.ts')).toEqual(['src', 'src/app'])
    })

    it('is empty for a top-level entry', () => {
        expect(ancestorsOf('README.md')).toEqual([])
    })
})

describe('folderOf', () => {
    it('returns the closest folder', () => {
        expect(folderOf('src/app/main.ts')).toBe('src/app')
    })

    it('returns the root for top-level files and no selection', () => {
        expect(folderOf('README.md')).toBe('')
        expect(folderOf(null)).toBe('')
    })
})
