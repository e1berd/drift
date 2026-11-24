import { computed, inject, provide, ref, toValue, watch } from 'vue'
import { useQueryClient } from '@tanstack/vue-query'
import { toast } from '@ui-sdk/vue/toast'

import type { InjectionKey, MaybeRefOrGetter } from 'vue'
import type { GitFileChange, GitStatus } from '@drift/vue/api'

import {
    getGitStatusQueryKey,
    useCheckoutGitBranch,
    useCommitGitChanges,
    useDiscardGitChanges,
    useGetGitStatus,
    useInitGitRepository,
    usePullGitBranch,
    usePushGitBranch,
    useStageGitPaths,
    useUnstageGitPaths,
} from '@drift/vue/api'
import { errorMessage } from '@drift/sdk/errors'

const STATUS_URL = '/sessions/:sessionId/git'

export interface GitSelection {
    path: string
    staged: boolean
}

/** Whether a query key belongs to one of the git endpoints of a session (status, diff, branches, commits). */
export function isSessionGitQuery(
    queryKey: readonly unknown[],
    sessionId: string,
    { withStatus = true } = {},
): boolean {
    const head = queryKey.at(0)
    if (typeof head !== 'object' || head === null) return false
    const { url, params } = head as { url?: unknown; params?: { sessionId?: unknown } }
    if (typeof url !== 'string' || params?.sessionId !== sessionId) return false
    if (url === STATUS_URL) return withStatus
    return url.startsWith(`${STATUS_URL}/`)
}

/** Selection after a status change: the same path on the same side, else on the other side, else nothing. */
export function followSelection(
    selection: GitSelection | null,
    status: GitStatus,
): GitSelection | null {
    if (!selection) return null
    const has = (list: GitFileChange[]) => list.some((change) => change.path === selection.path)
    const [same, other] = selection.staged
        ? [status.staged, status.unstaged]
        : [status.unstaged, status.staged]
    if (has(same)) return selection
    if (has(other)) return { path: selection.path, staged: !selection.staged }
    return null
}

function createSessionGit(sessionId: MaybeRefOrGetter<string>) {
    const qc = useQueryClient()
    const path = computed(() => ({ sessionId: toValue(sessionId) }))
    const status = useGetGitStatus(
        { path },
        { query: { staleTime: 0, refetchOnWindowFocus: 'always' } },
    )
    const selected = ref<GitSelection | null>(null)
    const repository = computed(() => !!status.data.value?.repository)

    watch(
        () => status.data.value,
        (next) => {
            if (next) selected.value = followSelection(selected.value, next)
        },
    )

    function applyStatus(next: GitStatus) {
        qc.setQueryData(getGitStatusQueryKey({ path: path.value }), next)
        void qc.invalidateQueries({
            predicate: (query) =>
                isSessionGitQuery(query.queryKey, path.value.sessionId, { withStatus: false }),
        })
    }

    function refresh() {
        return qc.invalidateQueries({
            predicate: (query) => isSessionGitQuery(query.queryKey, path.value.sessionId),
        })
    }

    const mutation = {
        mutation: {
            onSuccess: applyStatus,
            onError: (error: unknown) => toast.error(errorMessage(error)),
        },
    }
    const actions = {
        init: useInitGitRepository(mutation),
        stage: useStageGitPaths(mutation),
        unstage: useUnstageGitPaths(mutation),
        discard: useDiscardGitChanges(mutation),
        commit: useCommitGitChanges(mutation),
        checkout: useCheckoutGitBranch(mutation),
        pull: usePullGitBranch(mutation),
        push: usePushGitBranch(mutation),
    }
    const busy = computed(() => Object.values(actions).some((action) => action.isPending.value))

    return {
        sessionId: computed(() => path.value.sessionId),
        status,
        repository,
        selected,
        busy,
        refresh,
        init: () => actions.init.mutate({ path: path.value }),
        stage: (paths: string[]) => actions.stage.mutate({ path: path.value, body: { paths } }),
        unstage: (paths: string[]) => actions.unstage.mutate({ path: path.value, body: { paths } }),
        discard: (paths: string[]) => actions.discard.mutate({ path: path.value, body: { paths } }),
        commit: (message: string) =>
            actions.commit.mutateAsync({ path: path.value, body: { message } }),
        checkout: (branch: string, create = false) =>
            actions.checkout.mutateAsync({ path: path.value, body: { branch, create } }),
        pull: () => actions.pull.mutate({ path: path.value }),
        push: () => actions.push.mutate({ path: path.value }),
        pending: {
            commit: actions.commit.isPending,
            pull: actions.pull.isPending,
            push: actions.push.isPending,
            init: actions.init.isPending,
        },
    }
}

export type SessionGit = ReturnType<typeof createSessionGit>

const KEY: InjectionKey<SessionGit> = Symbol('drift.sessionGit')

/** Git status, selection and actions of the folder a session's agent works in; created once by GitPane. */
export function provideSessionGit(sessionId: MaybeRefOrGetter<string>): SessionGit {
    const git = createSessionGit(sessionId)
    provide(KEY, git)
    return git
}

/** Git state for components inside GitPane. */
export function useSessionGit(): SessionGit {
    const git = inject(KEY)
    if (!git) throw new Error('useSessionGit() called outside GitPane')
    return git
}
