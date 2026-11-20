export type ToolFamily =
    | 'shell'
    | 'edit'
    | 'write'
    | 'read'
    | 'search'
    | 'web'
    | 'todo'
    | 'agent'
    | 'other'

const FAMILIES: Record<string, ToolFamily> = {
    Bash: 'shell',
    BashOutput: 'shell',
    KillShell: 'shell',
    Edit: 'edit',
    MultiEdit: 'edit',
    NotebookEdit: 'edit',
    Write: 'write',
    Read: 'read',
    Grep: 'search',
    Glob: 'search',
    LS: 'search',
    WebFetch: 'web',
    WebSearch: 'web',
    TodoWrite: 'todo',
    Task: 'agent',
    Agent: 'agent',
}

const PATH_FAMILIES: ReadonlySet<ToolFamily> = new Set(['edit', 'write', 'read'])
const SHELL_WRAPPER = /^(?:\/(?:usr\/)?bin\/)?(?:ba|z)?sh\s+-l?c\s+(['"])([\s\S]*)\1$/
const NUMBERED_LINE = /^\s*(\d+)(?:→|\t)(.*)$/

export function toolFamily(name: string): ToolFamily {
    return FAMILIES[name] ?? 'other'
}

/** The command a shell tool ran, without the `bash -lc '…'` wrapper some agents add. */
export function shellCommand(input: string): string {
    const wrapped = SHELL_WRAPPER.exec(input.trim())
    return wrapped?.at(2) ?? input.trim()
}

/** A file path split for display: the folder dims, the file name stands out. */
export function splitPath(path: string): { folder: string; file: string } {
    const slash = path.lastIndexOf('/')
    return slash === -1
        ? { folder: '', file: path }
        : { folder: path.slice(0, slash + 1), file: path.slice(slash + 1) }
}

/** Whether the input of the tool is a file path, possibly one per line with a change kind. */
export function inputIsPath(family: ToolFamily): boolean {
    return PATH_FAMILIES.has(family)
}

/** The first line of a multi-line input and how many more lines follow it. */
export function firstLine(text: string): { line: string; more: number } {
    const lines = text.split('\n')
    return { line: lines.at(0) ?? '', more: lines.length - 1 }
}

export type OutputLine =
    | { kind: 'numbered'; number: string; text: string }
    | { kind: 'add' | 'del' | 'hunk' | 'plain'; text: string }

function diffKind(line: string): 'add' | 'del' | 'hunk' | 'plain' {
    if (line.startsWith('@@')) return 'hunk'
    if (line.startsWith('+') && !line.startsWith('+++')) return 'add'
    if (line.startsWith('-') && !line.startsWith('---')) return 'del'
    return 'plain'
}

/** Output lines typed for display: `cat -n` style numbered lines, diff lines, or plain text. */
export function parseOutput(lines: string[]): OutputLine[] {
    const numbered = lines.filter((line) => NUMBERED_LINE.test(line)).length
    const looksNumbered = numbered > 0 && numbered >= lines.length / 2
    const looksDiff = lines.some((line) => line.startsWith('@@') || line.startsWith('diff --git'))
    return lines.map((line) => {
        const match = looksNumbered ? NUMBERED_LINE.exec(line) : null
        if (match) return { kind: 'numbered', number: match.at(1) ?? '', text: match.at(2) ?? '' }
        return { kind: looksDiff ? diffKind(line) : 'plain', text: line }
    })
}
