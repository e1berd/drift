import type { GitChangeStatusKey } from '../api'

export type DiffLineKind = 'add' | 'del' | 'hunk' | 'meta' | 'context'

export interface DiffLine {
    kind: DiffLineKind
    text: string
    oldLine?: number
    newLine?: number
}

const HUNK = /^@@ -(?<old>\d+)(?:,\d+)? \+(?<new>\d+)(?:,\d+)? @@/

/** Lines of a unified diff with their kind and old/new line numbers; the file header becomes `meta`. */
export function parseDiff(diff: string): DiffLine[] {
    const lines: DiffLine[] = []
    let oldLine = 0
    let newLine = 0
    let inHunk = false
    for (const text of diff.split('\n')) {
        const hunk = HUNK.exec(text)
        if (hunk) {
            oldLine = Number(hunk.groups?.old)
            newLine = Number(hunk.groups?.new)
            inHunk = true
            lines.push({ kind: 'hunk', text })
        } else if (!inHunk || text.startsWith('diff --git ')) {
            inHunk = false
            if (text) lines.push({ kind: 'meta', text })
        } else if (text.startsWith('+')) {
            lines.push({ kind: 'add', text, newLine: newLine++ })
        } else if (text.startsWith('-')) {
            lines.push({ kind: 'del', text, oldLine: oldLine++ })
        } else if (text.startsWith(' ')) {
            lines.push({ kind: 'context', text, oldLine: oldLine++, newLine: newLine++ })
        } else if (text.startsWith('\\')) {
            lines.push({ kind: 'meta', text })
        }
    }
    return lines
}

const LETTERS: Record<GitChangeStatusKey, string> = {
    modified: 'M',
    added: 'A',
    deleted: 'D',
    renamed: 'R',
    copied: 'C',
    typeChanged: 'T',
    conflicted: 'U',
    untracked: '?',
}

/** One-letter badge as `git status --short` prints it. */
export function statusLetter(status: GitChangeStatusKey): string {
    return LETTERS[status]
}
