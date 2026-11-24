import { ref } from 'vue'

import type { InfiniteData, QueryClient } from '@tanstack/vue-query'
import type {
    Delivery,
    Draft,
    ListProjectsResponse,
    ListSessionsResponse,
    Message,
    Project,
    Session,
    SessionActivity,
    WireEvent,
} from '../api'

import {
    getDeliveryQueryKey,
    getSessionQueryKey,
    listProjectsInfiniteQueryKey,
    listSessionsInfiniteQueryKey,
    listStandaloneSessionsInfiniteQueryKey,
    listSubscriptionsQueryKey,
    searchSessionsInfiniteQueryKey,
} from '../api'
import { applyDraftEvent, loadDrafts } from './useDrafts'
import { connectRealtime, useRealtimeConnection } from './realtimeConnection'
import { deviceId } from '@drift/sdk/device'
import { API_BASE } from '@drift/sdk/http/client'
import { reportError } from '@drift/sdk/errors'

type Listener = (event: WireEvent) => void
type Pages<T> = InfiniteData<T, string>

const freshSessions = ref(new Set<string>())
const activities = ref(new Map<string, SessionActivity>())
const listeners = new Set<Listener>()
let lastEventId = ''

function mapPages<TPage, TItem>(
    data: Pages<TPage> | undefined,
    read: (page: TPage) => TItem[],
    write: (page: TPage, items: TItem[]) => TPage,
    change: (items: TItem[], index: number) => TItem[],
): Pages<TPage> | undefined {
    if (!data) return data
    return {
        ...data,
        pages: data.pages.map((page, index) => write(page, change(read(page), index))),
    }
}

function updateProjects(
    qc: QueryClient,
    change: (items: Project[], index: number, total: number) => Project[],
) {
    qc.setQueryData<Pages<ListProjectsResponse>>(listProjectsInfiniteQueryKey(), (data) =>
        mapPages(
            data,
            (page) => page.projects,
            (page, projects) => ({ ...page, projects }),
            (items, index) => change(items, index, data?.pages.length ?? 0),
        ),
    )
}

function upsertProject(qc: QueryClient, project: Project) {
    const key = listProjectsInfiniteQueryKey()
    const known = qc
        .getQueryData<Pages<ListProjectsResponse>>(key)
        ?.pages.some((page) => page.projects.some((item) => item.id === project.id))
    updateProjects(qc, (items, index, total) => {
        if (known) return items.map((item) => (item.id === project.id ? project : item))
        return index === total - 1 ? [...items, project] : items
    })
}

function bumpSessionCount(qc: QueryClient, projectId: string | undefined, delta: number) {
    if (!projectId) return
    updateProjects(qc, (items) =>
        items.map((item) =>
            item.id === projectId
                ? { ...item, sessionCount: Math.max(0, item.sessionCount + delta) }
                : item,
        ),
    )
}

function sessionsKey(projectId: string | undefined) {
    return projectId
        ? listSessionsInfiniteQueryKey({ path: { projectId } })
        : listStandaloneSessionsInfiniteQueryKey()
}

function updateSessions(
    qc: QueryClient,
    projectId: string | undefined,
    change: (items: Session[], index: number) => Session[],
) {
    qc.setQueryData<Pages<ListSessionsResponse>>(sessionsKey(projectId), (data) =>
        mapPages(
            data,
            (page) => page.sessions,
            (page, sessions) => ({ ...page, sessions }),
            change,
        ),
    )
}

const searchResultsKey = searchSessionsInfiniteQueryKey({ query: { query: '' } }).slice(0, 1)

function updateSearchResults(qc: QueryClient, change: (items: Session[]) => Session[]) {
    qc.setQueriesData<Pages<ListSessionsResponse>>({ queryKey: searchResultsKey }, (data) =>
        mapPages(
            data,
            (page) => page.sessions,
            (page, sessions) => ({ ...page, sessions }),
            change,
        ),
    )
}

function refreshSearchResults(qc: QueryClient) {
    void qc.invalidateQueries({ queryKey: searchResultsKey })
}

function putSessionFirst(qc: QueryClient, session: Session) {
    updateSessions(qc, session.projectId, (items, index) => {
        const rest = items.filter((item) => item.id !== session.id)
        return index === 0 ? [session, ...rest] : rest
    })
    qc.setQueryData(getSessionQueryKey({ path: { sessionId: session.id } }), session)
}

function isKnownSession(qc: QueryClient, session: Session): boolean {
    return !!qc
        .getQueryData<Pages<ListSessionsResponse>>(sessionsKey(session.projectId))
        ?.pages.some((page) => page.sessions.some((item) => item.id === session.id))
}

function onSessionCreated(qc: QueryClient, session: Session, mine: boolean) {
    const known = isKnownSession(qc, session)
    putSessionFirst(qc, session)
    if (!known) bumpSessionCount(qc, session.projectId, 1)
    refreshSearchResults(qc)
    if (!mine) freshSessions.value = new Set(freshSessions.value).add(session.id)
}

function onSessionDeleted(qc: QueryClient, { id, projectId }: { id: string; projectId?: string }) {
    updateSessions(qc, projectId, (items) => items.filter((item) => item.id !== id))
    updateSearchResults(qc, (items) => items.filter((item) => item.id !== id))
    qc.removeQueries({ queryKey: getSessionQueryKey({ path: { sessionId: id } }) })
    bumpSessionCount(qc, projectId, -1)
}

function onProjectDeleted(qc: QueryClient, { id }: { id: string }) {
    updateProjects(qc, (items) => items.filter((item) => item.id !== id))
    qc.removeQueries({ queryKey: sessionsKey(id) })
    updateSearchResults(qc, (items) => items.filter((item) => item.projectId !== id))
}

function setActivity(sessionId: string, activity: SessionActivity | null) {
    const next = new Map(activities.value)
    if (activity) next.set(sessionId, activity)
    else next.delete(sessionId)
    activities.value = next
}

function onSessionUpdated(qc: QueryClient, session: Session) {
    if (session.status !== 'running' && activities.value.has(session.id))
        setActivity(session.id, null)
    putSessionFirst(qc, session)
    updateSearchResults(qc, (items) =>
        items.map((item) => (item.id === session.id ? session : item)),
    )
}

function onAgentMessage(message: Message) {
    if (message.role === 'assistant' && message.kind === 'text')
        setActivity(message.sessionId, { sessionId: message.sessionId, phase: 'thinking' })
}

function resync(qc: QueryClient) {
    void qc.invalidateQueries()
    loadDrafts().catch(reportError)
}

function apply(qc: QueryClient, event: WireEvent) {
    const mine = event.origin === deviceId()
    switch (event.type) {
        case 'resync':
            return resync(qc)
        case 'project.created':
        case 'project.updated':
            return upsertProject(qc, event.data as Project)
        case 'project.deleted':
            return onProjectDeleted(qc, event.data as { id: string })
        case 'session.created':
            return onSessionCreated(qc, event.data as Session, mine)
        case 'session.updated':
            return onSessionUpdated(qc, event.data as Session)
        case 'session.activity': {
            const activity = event.data as SessionActivity
            return setActivity(activity.sessionId, activity)
        }
        case 'session.deleted':
            return onSessionDeleted(qc, event.data as { id: string; projectId?: string })
        case 'message.created':
            return onAgentMessage(event.data as Message)
        case 'draft.updated':
        case 'draft.deleted':
            return applyDraftEvent(event.type, event.data as Partial<Draft>, event.origin)
        case 'subscription.created':
        case 'subscription.updated':
        case 'subscription.deleted':
            return void qc.invalidateQueries({ queryKey: listSubscriptionsQueryKey() })
        case 'delivery.updated':
            return void qc.setQueryData(getDeliveryQueryKey(), event.data as Delivery)
    }
}

function parseEvent(message: MessageEvent<string>): WireEvent | null {
    try {
        return JSON.parse(message.data) as WireEvent
    } catch (error) {
        reportError(error)
        return null
    }
}

function eventStreamUrl() {
    const url = new URL(`${API_BASE}/events`, location.origin)
    if (lastEventId) url.searchParams.set('last_event_id', lastEventId)
    return url
}

function receive(qc: QueryClient, message: MessageEvent<string>) {
    if (message.lastEventId) lastEventId = message.lastEventId
    const event = parseEvent(message)
    if (!event) return
    apply(qc, event)
    for (const listener of listeners) listener(event)
}

/** Opens the single per-tab SSE connection and folds every event into the query cache. */
export function startEventStream(qc: QueryClient) {
    connectRealtime({
        url: eventStreamUrl,
        onMessage: (message) => receive(qc, message),
        onCursor: (id) => (lastEventId = id),
    })
}

/** Connection state, sessions created elsewhere since this tab opened, and live agent activity per session. */
export function useEventStream() {
    function on(listener: Listener) {
        listeners.add(listener)
        return () => listeners.delete(listener)
    }
    const { connected } = useRealtimeConnection()
    return { connected, freshSessions, activities, on }
}
