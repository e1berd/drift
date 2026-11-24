import { describe, expect, it } from 'vitest'

import type { GitStatus } from '@drift/vue/api'

import { followSelection, isSessionGitQuery } from '@/composables/useSessionGit'
import {
    getGitDiffQueryKey,
    getGitStatusQueryKey,
    getSessionQueryKey,
    listGitCommitsInfiniteQueryKey,
} from '@drift/vue/api'

const SESSION = '0b6f6a52-6f3c-4a8e-9d2c-3e0f5a1b2c3d'

function status(staged: string[], unstaged: string[]): GitStatus {
    return {
        repository: true,
        detached: false,
        ahead: 0,
        behind: 0,
        staged: staged.map((path) => ({ path, status: 'modified' })),
        unstaged: unstaged.map((path) => ({ path, status: 'modified' })),
    }
}

describe('isSessionGitQuery', () => {
    const path = { sessionId: SESSION }

    it('matches every git endpoint of the session', () => {
        expect(isSessionGitQuery(getGitStatusQueryKey({ path }), SESSION)).toBe(true)
        expect(
            isSessionGitQuery(getGitDiffQueryKey({ path, query: { path: 'a.txt' } }), SESSION),
        ).toBe(true)
        expect(isSessionGitQuery(listGitCommitsInfiniteQueryKey({ path }), SESSION)).toBe(true)
    })

    it('can leave the status out', () => {
        const options = { withStatus: false }
        expect(isSessionGitQuery(getGitStatusQueryKey({ path }), SESSION, options)).toBe(false)
        expect(
            isSessionGitQuery(
                getGitDiffQueryKey({ path, query: { path: 'a.txt' } }),
                SESSION,
                options,
            ),
        ).toBe(true)
    })

    it('ignores other sessions and other endpoints', () => {
        expect(isSessionGitQuery(getGitStatusQueryKey({ path }), 'other')).toBe(false)
        expect(isSessionGitQuery(getSessionQueryKey({ path }), SESSION)).toBe(false)
        expect(isSessionGitQuery(['plain'], SESSION)).toBe(false)
    })
})

describe('followSelection', () => {
    it('keeps a selection that is still there', () => {
        const selection = { path: 'a.txt', staged: false }
        expect(followSelection(selection, status([], ['a.txt']))).toBe(selection)
    })

    it('follows a file to the other side after staging', () => {
        expect(followSelection({ path: 'a.txt', staged: false }, status(['a.txt'], []))).toEqual({
            path: 'a.txt',
            staged: true,
        })
    })

    it('drops a file that has no changes left', () => {
        expect(followSelection({ path: 'a.txt', staged: true }, status([], ['b.txt']))).toBeNull()
        expect(followSelection(null, status([], ['a.txt']))).toBeNull()
    })
})
