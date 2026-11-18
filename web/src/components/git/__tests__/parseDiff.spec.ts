import { describe, expect, it } from 'vitest'

import { parseDiff, statusLetter } from '@/components/git/parseDiff'

const DIFF = [
    'diff --git a/a.txt b/a.txt',
    'index 5626abf..f719efd 100644',
    '--- a/a.txt',
    '+++ b/a.txt',
    '@@ -1,3 +1,3 @@ heading',
    ' one',
    '-two',
    '+zwei',
    '+--- not a header',
    ' three',
    '\\ No newline at end of file',
    '',
].join('\n')

describe('parseDiff', () => {
    it('keeps the header as meta and numbers both sides', () => {
        const lines = parseDiff(DIFF)
        expect(lines.slice(0, 4).map((line) => line.kind)).toEqual(['meta', 'meta', 'meta', 'meta'])
        expect(
            lines.slice(4).map(({ kind, oldLine, newLine }) => [kind, oldLine, newLine]),
        ).toEqual([
            ['hunk', undefined, undefined],
            ['context', 1, 1],
            ['del', 2, undefined],
            ['add', undefined, 2],
            ['add', undefined, 3],
            ['context', 3, 4],
            ['meta', undefined, undefined],
        ])
    })

    it('starts counting at the hunk position and handles one-line hunks', () => {
        const lines = parseDiff('@@ -10 +12,2 @@\n-a\n+b\n+c\n')
        expect(lines.map((line) => line.oldLine ?? line.newLine)).toEqual([undefined, 10, 12, 13])
    })

    it('is empty for an empty diff', () => {
        expect(parseDiff('')).toEqual([])
    })
})

describe('statusLetter', () => {
    it('prints letters like git status --short', () => {
        expect(statusLetter('modified')).toBe('M')
        expect(statusLetter('untracked')).toBe('?')
        expect(statusLetter('conflicted')).toBe('U')
    })
})
