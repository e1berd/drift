import { computed, toValue, watch } from 'vue'

import type { MaybeRefOrGetter } from 'vue'
import type { Message } from '../api'

import { useGetGitSummary } from '../api'

const REFRESH_DELAY_MS = 400

function finishedToolCount(messages: readonly Message[]): number {
    return messages.filter(
        (message) => message.kind === 'tool' && message.body.status !== 'running',
    ).length
}

/** The git summary of a session, refreshed whenever one of the agent's tool calls finishes. */
export function useSessionSummary(
    sessionId: MaybeRefOrGetter<string>,
    messages: MaybeRefOrGetter<readonly Message[]>,
) {
    const path = computed(() => ({ sessionId: toValue(sessionId) }))
    const summary = useGetGitSummary(
        { path },
        { query: { staleTime: 0, refetchOnWindowFocus: 'always' } },
    )
    let timer: ReturnType<typeof setTimeout> | undefined
    watch(
        () => finishedToolCount(toValue(messages)),
        (count, previous) => {
            if (count === previous) return
            clearTimeout(timer)
            timer = setTimeout(() => void summary.refetch(), REFRESH_DELAY_MS)
        },
    )
    return summary
}
