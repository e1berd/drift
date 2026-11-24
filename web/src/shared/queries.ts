import { computed, toValue } from 'vue'
import { keepPreviousData } from '@tanstack/vue-query'

import type { MaybeRefOrGetter } from 'vue'

import {
    useGetSession,
    useListProjectsInfinite,
    useListSessionsInfinite,
    useListStandaloneSessionsInfinite,
    useSearchSessionsInfinite,
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

/** Sessions without a project, latest activity first. */
export function useStandaloneSessions() {
    const query = useListStandaloneSessionsInfinite()
    const sessions = computed(() => query.data.value?.pages.flatMap((page) => page.sessions) ?? [])
    return { ...query, sessions }
}

/** Sessions of all projects and chats whose title contains `text`; idle while `text` is empty. */
export function useSessionSearch(text: MaybeRefOrGetter<string>) {
    const query = useSearchSessionsInfinite(
        { query: () => ({ query: toValue(text) }) },
        { query: { enabled: () => !!toValue(text), placeholderData: keepPreviousData } },
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
