import { toValue } from 'vue'
import { useQueryClient } from '@tanstack/vue-query'
import { useErrorNotice } from '../notifier'

import type { MaybeRefOrGetter } from 'vue'
import type { ActivityEntry, GitSummary } from '../api'

import { getGitSummaryQueryKey, useRemoveActivityFile, useRemoveActivityWorktree } from '../api'
import { reportError } from '@drift/sdk/errors'

export interface RemoveChoice {
    deleteBranch: boolean
}

type AskRemoval = (entry: ActivityEntry) => Promise<RemoveChoice | null>

/** Removes files and worktrees from the activity log after the user confirms in a dialog. */
export function useActivityRemoval(
    sessionId: MaybeRefOrGetter<string>,
    ask: () => AskRemoval | undefined,
) {
    const qc = useQueryClient()
    const showError = useErrorNotice()
    const mutation = {
        mutation: {
            onSuccess: (summary: GitSummary) =>
                qc.setQueryData(
                    getGitSummaryQueryKey({ path: { sessionId: toValue(sessionId) } }),
                    summary,
                ),
            onError: showError,
        },
    }
    const removeFile = useRemoveActivityFile(mutation)
    const removeWorktree = useRemoveActivityWorktree(mutation)

    function removeConfirmed(entry: ActivityEntry, choice: RemoveChoice) {
        const path = { sessionId: toValue(sessionId) }
        if (entry.kind === 'file') return removeFile.mutate({ path, body: { name: entry.name } })
        removeWorktree.mutate({
            path,
            body: { name: entry.name, force: !!entry.changes, deleteBranch: choice.deleteBranch },
        })
    }

    function remove(entry: ActivityEntry) {
        const askRemoval = ask()
        if (!askRemoval) return
        askRemoval(entry)
            .then((choice) => {
                if (choice) removeConfirmed(entry, choice)
            })
            .catch(reportError)
    }

    return { remove }
}
