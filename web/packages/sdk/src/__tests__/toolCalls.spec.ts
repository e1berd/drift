import { describe, expect, it } from 'vitest'

import { parseOutput, shellCommand, splitPath, toolFamily } from '../toolCalls'

describe('toolCalls', () => {
    it('strips the shell wrapper agents put around commands', () => {
        expect(shellCommand(`/bin/bash -lc 'pnpm test'`)).toBe('pnpm test')
        expect(shellCommand('bash -c "ls -la"')).toBe('ls -la')
        expect(shellCommand('git status')).toBe('git status')
    })

    it('splits a path into folder and file', () => {
        expect(splitPath('web/src/App.vue')).toEqual({ folder: 'web/src/', file: 'App.vue' })
        expect(splitPath('README.md')).toEqual({ folder: '', file: 'README.md' })
    })

    it('reads cat -n output as numbered lines', () => {
        expect(parseOutput(['     1→const a = 1', '     2→'])).toEqual([
            { kind: 'numbered', number: '1', text: 'const a = 1' },
            { kind: 'numbered', number: '2', text: '' },
        ])
    })

    it('colours diff lines only in diff output', () => {
        expect(parseOutput(['-rw-r--r-- 1 file']).at(0)?.kind).toBe('plain')
        expect(parseOutput(['@@ -1 +1 @@', '-old', '+new']).map((line) => line.kind)).toEqual([
            'hunk',
            'del',
            'add',
        ])
    })

    it('groups tools into families', () => {
        expect(toolFamily('Bash')).toBe('shell')
        expect(toolFamily('mcp.custom')).toBe('other')
    })
})
