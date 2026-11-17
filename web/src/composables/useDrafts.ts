import { computed, reactive, ref, toValue, watch } from 'vue'

import type { MaybeRefOrGetter } from 'vue'
import type { Draft } from '@/gen'

import { deleteDraft, listDrafts, updateDraft } from '@/gen'
import { deviceId } from '@/shared/device'
import { reportError } from '@/shared/errors'

const SAVE_DELAY = 400
const TYPING_WINDOW = 2500

const drafts = reactive(new Map<string, Draft>())
const remoteTyping = reactive(new Set<string>())
const typingTimers = new Map<string, ReturnType<typeof setTimeout>>()
const pending = new Map<string, { text: string; timer: ReturnType<typeof setTimeout> }>()
const loaded = ref(false)

function saveDraft(key: string, text: string, keepalive = false) {
    const request = text
        ? updateDraft({ path: { key }, body: { text }, keepalive })
        : deleteDraft({ path: { key }, keepalive })
    return request.catch(reportError)
}

function flushPendingDrafts(keepalive: boolean) {
    for (const [key, entry] of pending) {
        clearTimeout(entry.timer)
        void saveDraft(key, entry.text, keepalive)
    }
    pending.clear()
}

function flushWhenPageHides() {
    window.addEventListener('pagehide', () => flushPendingDrafts(true))
    document.addEventListener('visibilitychange', () => {
        if (document.visibilityState === 'hidden') flushPendingDrafts(true)
    })
}

if (typeof window !== 'undefined') flushWhenPageHides()

/** Loads every draft of the user once; later changes arrive as realtime events. */
export async function loadDrafts() {
    const { data } = await listDrafts()
    drafts.clear()
    for (const draft of data.drafts) drafts.set(draft.key, draft)
    loaded.value = true
}

function markRemoteTyping(key: string) {
    remoteTyping.add(key)
    clearTimeout(typingTimers.get(key))
    typingTimers.set(
        key,
        setTimeout(() => remoteTyping.delete(key), TYPING_WINDOW),
    )
}

/** Applies a realtime draft event; edits from other devices light up the typing indicator. */
export function applyDraftEvent(
    type: 'draft.updated' | 'draft.deleted',
    data: Partial<Draft>,
    origin?: string,
) {
    const key = data.key
    if (!key || origin === deviceId()) return
    if (type === 'draft.deleted') {
        drafts.delete(key)
        remoteTyping.delete(key)
        return
    }
    drafts.set(key, data as Draft)
    markRemoteTyping(key)
}

/** Draft key of a session, or of a new chat inside a project. */
export function draftKey(sessionId?: string, projectId?: string): string {
    if (sessionId) return `session:${sessionId}`
    return projectId ? `new:${projectId}` : ''
}

function isFromOtherDevice(draft: Draft | undefined): boolean {
    return !!draft && draft.deviceId !== deviceId()
}

/** Read-only view of all drafts for the sidebar. */
export function useDraftIndex() {
    return {
        has: (key: string) => !!drafts.get(key)?.text,
        get: (key: string) => drafts.get(key),
        typing: (key: string) => remoteTyping.has(key),
        fromOther: (key: string) => isFromOtherDevice(drafts.get(key)),
        all: computed(() => [...drafts.values()]),
    }
}

function storeLocalDraft(key: string, text: string) {
    if (text)
        drafts.set(key, { key, text, deviceId: deviceId(), updateTime: new Date().toISOString() })
    else drafts.delete(key)
}

function scheduleSave(key: string, text: string) {
    const previous = pending.get(key)
    if (previous) clearTimeout(previous.timer)
    pending.set(key, {
        text,
        timer: setTimeout(() => {
            pending.delete(key)
            void saveDraft(key, text)
        }, SAVE_DELAY),
    })
}

/** The editable draft for one key, saved with a short debounce and flushed when the page hides. */
export function useDraft(key: MaybeRefOrGetter<string>) {
    const text = ref('')
    const external = ref(0)

    function replaceText(value: string) {
        text.value = value
        external.value++
    }

    watch(
        () => [toValue(key), loaded.value] as const,
        ([currentKey]) => replaceText(currentKey ? (drafts.get(currentKey)?.text ?? '') : ''),
        { immediate: true },
    )

    watch(
        () => {
            const currentKey = toValue(key)
            return currentKey ? drafts.get(currentKey) : undefined
        },
        (draft) => {
            const currentKey = toValue(key)
            if (!draft && text.value && currentKey && !pending.has(currentKey)) replaceText('')
            else if (draft && isFromOtherDevice(draft) && draft.text !== text.value)
                replaceText(draft.text)
        },
    )

    function set(value: string) {
        const currentKey = toValue(key)
        text.value = value
        if (!currentKey) return
        storeLocalDraft(currentKey, value)
        scheduleSave(currentKey, value)
    }

    function clear() {
        set('')
        external.value++
    }

    const remote = computed(() => {
        const currentKey = toValue(key)
        return {
            typing: !!currentKey && remoteTyping.has(currentKey),
            fromOther: !!currentKey && isFromOtherDevice(drafts.get(currentKey)),
        }
    })

    return { text, external, remote, set, clear }
}
