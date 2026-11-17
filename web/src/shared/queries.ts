import { computed, toValue } from 'vue'

import type { MaybeRefOrGetter } from 'vue'

import {
    useGetSession,
    useListMessagesInfinite,
    useListProjectsInfinite,
    useListSessionsInfinite,
} from '@/gen'

/** All projects across loaded pages, oldest first. */
export function useProjects() {
    const query = useListProjectsInfinite()
    const projects = computed(() => query.data.value?.pages.flatMap((page) => page.projects) ?? [])
    return { ...query, projects }
}

/** Sessions of one project, latest activity first, fetched only while the project is expanded. */
export function useSessions(
    projectId: MaybeRefOrGetter<string>,
    enabled: MaybeRefOrGetter<boolean>,
) {
    const query = useListSessionsInfinite(
        { path: computed(() => ({ projectId: toValue(projectId) })) },
        { query: { enabled: () => toValue(enabled) } },
    )
    const sessions = computed(() => query.data.value?.pages.flatMap((page) => page.sessions) ?? [])
    return { ...query, sessions }
}

/** One session by id. */
export function useSession(sessionId: MaybeRefOrGetter<string>) {
    return useGetSession(
        { path: computed(() => ({ sessionId: toValue(sessionId) })) },
        { query: { enabled: () => !!toValue(sessionId) } },
    )
}

/** Messages of a session; pages go back in time, the list is returned oldest first. */
export function useMessages(sessionId: MaybeRefOrGetter<string>) {
    const query = useListMessagesInfinite(
        { path: computed(() => ({ sessionId: toValue(sessionId) })) },
        { query: { enabled: () => !!toValue(sessionId) } },
    )
    const messages = computed(() =>
        (query.data.value?.pages.flatMap((page) => page.messages) ?? []).toReversed(),
    )
    return { ...query, messages }
}
