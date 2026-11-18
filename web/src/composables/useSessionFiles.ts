import { computed, inject, provide, reactive, ref, toValue, watch } from 'vue'

import type { InjectionKey, MaybeRefOrGetter } from 'vue'
import type { TreeEntry } from '@/gen'

import { listSessionTree, useListSessionTree } from '@/gen'

/** Ancestor folders of a folder-relative path, outermost first: `a/b/c.ts` → `a`, `a/b`. */
export function ancestorsOf(entryPath: string): string[] {
    const parts = entryPath.split('/').slice(0, -1)
    return parts.map((_, index) => parts.slice(0, index + 1).join('/'))
}

/** Folder that holds a folder-relative path, `''` for the root. */
export function folderOf(entryPath: string | null): string {
    return entryPath ? (ancestorsOf(entryPath).at(-1) ?? '') : ''
}

function createSessionFiles(sessionId: MaybeRefOrGetter<string>) {
    const children = reactive(new Map<string, TreeEntry[]>())
    const expanded = ref<string[]>([])
    const selected = ref<string | null>(null)
    const drafts = reactive(new Map<string, string>())

    const root = useListSessionTree({
        path: computed(() => ({ sessionId: toValue(sessionId) })),
        query: { path: '.' },
    })

    watch(
        () => toValue(sessionId),
        () => {
            children.clear()
            drafts.clear()
            expanded.value = []
            selected.value = null
        },
    )

    async function loadChildren(folder: string, options: { signal?: AbortSignal } = {}) {
        const { data } = await listSessionTree({
            path: { sessionId: toValue(sessionId) },
            query: { path: folder },
            signal: options.signal,
        })
        children.set(folder, data.entries)
    }

    async function reveal(entry: TreeEntry) {
        const folders = ancestorsOf(entry.path)
        await Promise.all([root.refetch(), ...folders.map((folder) => loadChildren(folder))])
        expanded.value = [...new Set([...expanded.value, ...folders])]
        if (!entry.dir) selected.value = entry.path
    }

    return { root, children, expanded, selected, drafts, loadChildren, reveal }
}

export type SessionFiles = ReturnType<typeof createSessionFiles>

const KEY: InjectionKey<SessionFiles> = Symbol('drift.sessionFiles')

/** Tree, selection and unsaved edits of the folder a session's agent works in; created once by FilesPane. */
export function provideSessionFiles(sessionId: MaybeRefOrGetter<string>): SessionFiles {
    const files = createSessionFiles(sessionId)
    provide(KEY, files)
    return files
}

/** Files state for components inside FilesPane. */
export function useSessionFiles(): SessionFiles {
    const files = inject(KEY)
    if (!files) throw new Error('useSessionFiles() called outside FilesPane')
    return files
}
