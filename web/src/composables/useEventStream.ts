import { ref } from 'vue'

import type { InfiniteData, QueryClient } from '@tanstack/vue-query'
import type {
    Draft,
    ListMessagesResponse,
    ListProjectsResponse,
    ListSessionsResponse,
    Message,
    Project,
    Session,
    SessionActivity,
    WireEvent,
} from '@/gen'

import {
    getSessionQueryKey,
    listMessagesInfiniteQueryKey,
    listProjectsInfiniteQueryKey,
    listSessionsInfiniteQueryKey,
    listStandaloneSessionsInfiniteQueryKey,
    listSubscriptionsQueryKey,
    searchSessionsInfiniteQueryKey,
} from '@/gen'
import { applyDraftEvent, loadDrafts } from '@/composables/useDrafts'
import { deviceId } from '@/shared/device'
import { API_BASE } from '@/shared/http/client'
import { reportError } from '@/shared/errors'

type Listener = (event: WireEvent) => void
type Pages<T> = InfiniteData<T, string>

const connected = ref(true)
const freshSessions = ref(new Set<string>())
const activities = ref(new Map<string, SessionActivity>())
const listeners = new Set<Listener>()
const REOPEN_DELAY_MS = 3000

let source: EventSource | null = null
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

function onMessageCreated(qc: QueryClient, message: Message) {
    qc.setQueryData<Pages<ListMessagesResponse>>(
        listMessagesInfiniteQueryKey({ path: { sessionId: message.sessionId } }),
        (data) =>
            mapPages(
                data,
                (page) => page.messages,
                (page, messages) => ({ ...page, messages }),
                (items, index) =>
                    index === 0 && !items.some((item) => item.id === message.id)
                        ? [message, ...items]
                        : items,
            ),
    )
}

function onMessageUpdated(qc: QueryClient, message: Message) {
    qc.setQueryData<Pages<ListMessagesResponse>>(
        listMessagesInfiniteQueryKey({ path: { sessionId: message.sessionId } }),
        (data) =>
            mapPages(
                data,
                (page) => page.messages,
                (page, messages) => ({ ...page, messages }),
                (items) => items.map((item) => (item.id === message.id ? message : item)),
            ),
    )
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
            onAgentMessage(event.data as Message)
            return onMessageCreated(qc, event.data as Message)
        case 'message.updated':
            return onMessageUpdated(qc, event.data as Message)
        case 'draft.updated':
        case 'draft.deleted':
            return applyDraftEvent(event.type, event.data as Partial<Draft>, event.origin)
        case 'subscription.created':
        case 'subscription.updated':
        case 'subscription.deleted':
            return void qc.invalidateQueries({ queryKey: listSubscriptionsQueryKey() })
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

function reopenClosedStream(qc: QueryClient, closed: EventSource) {
    if (closed.readyState !== EventSource.CLOSED || source !== closed) return
    source = null
    setTimeout(() => startEventStream(qc), REOPEN_DELAY_MS)
}

/** Opens the single per-tab SSE connection and folds every event into the query cache. */
export function startEventStream(qc: QueryClient) {
    if (source || typeof EventSource === 'undefined') return
    const stream = new EventSource(eventStreamUrl())
    source = stream
    stream.addEventListener('open', () => (connected.value = true))
    stream.addEventListener('error', () => {
        connected.value = false
        reopenClosedStream(qc, stream)
    })
    stream.addEventListener('message', (message: MessageEvent<string>) => {
        if (message.lastEventId) lastEventId = message.lastEventId
        const event = parseEvent(message)
        if (!event) return
        apply(qc, event)
        for (const listener of listeners) listener(event)
    })
}

/** Connection state, sessions created elsewhere since this tab opened, and live agent activity per session. */
export function useEventStream() {
    function on(listener: Listener) {
        listeners.add(listener)
        return () => listeners.delete(listener)
    }
    return { connected, freshSessions, activities, on }
}
