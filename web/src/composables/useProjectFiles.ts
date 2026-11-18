import { computed, inject, provide, reactive, ref, toValue, watch } from 'vue'

import type { InjectionKey, MaybeRefOrGetter } from 'vue'
import type { TreeEntry } from '@/gen'

import { listTree, useListTree } from '@/gen'

/** Ancestor folders of a project-relative path, outermost first: `a/b/c.ts` → `a`, `a/b`. */
export function ancestorsOf(entryPath: string): string[] {
    const parts = entryPath.split('/').slice(0, -1)
    return parts.map((_, index) => parts.slice(0, index + 1).join('/'))
}

/** Folder that holds a project-relative path, `''` for the project root. */
export function folderOf(entryPath: string | null): string {
    return entryPath ? (ancestorsOf(entryPath).at(-1) ?? '') : ''
}

function createProjectFiles(projectId: MaybeRefOrGetter<string>) {
    const children = reactive(new Map<string, TreeEntry[]>())
    const expanded = ref<string[]>([])
    const selected = ref<string | null>(null)
    const drafts = reactive(new Map<string, string>())

    const root = useListTree({
        path: computed(() => ({ projectId: toValue(projectId) })),
        query: { path: '.' },
    })

    watch(
        () => toValue(projectId),
        () => {
            children.clear()
            drafts.clear()
            expanded.value = []
            selected.value = null
        },
    )

    async function loadChildren(folder: string, options: { signal?: AbortSignal } = {}) {
        const { data } = await listTree({
            path: { projectId: toValue(projectId) },
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

export type ProjectFiles = ReturnType<typeof createProjectFiles>

const KEY: InjectionKey<ProjectFiles> = Symbol('drift.projectFiles')

/** Tree, selection and unsaved edits of one project's files; created once by FilesPane. */
export function provideProjectFiles(projectId: MaybeRefOrGetter<string>): ProjectFiles {
    const files = createProjectFiles(projectId)
    provide(KEY, files)
    return files
}

/** Files state for components inside FilesPane. */
export function useProjectFiles(): ProjectFiles {
    const files = inject(KEY)
    if (!files) throw new Error('useProjectFiles() called outside FilesPane')
    return files
}
