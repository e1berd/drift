import { computed, reactive, ref, toValue, watch } from 'vue'

import type { MaybeRefOrGetter } from 'vue'
import type { Draft } from '@/gen'

import { deleteDraft, listDrafts, updateDraft } from '@/gen'
import { deviceId } from '@/shared/device'
import { forgetSyncedDraft, readUnsyncedDrafts, rememberUnsyncedDraft } from '@/shared/draftBackup'
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
    return request.then(() => forgetSyncedDraft(key, text)).catch(reportError)
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

function restoreUnsyncedDrafts() {
    for (const [key, { text, updateTime }] of Object.entries(readUnsyncedDrafts())) {
        if (text) drafts.set(key, { key, text, deviceId: deviceId(), updateTime })
        else drafts.delete(key)
    }
}

function resendUnsyncedDrafts() {
    for (const [key, { text }] of Object.entries(readUnsyncedDrafts())) {
        if (!pending.has(key)) void saveDraft(key, text)
    }
}

if (typeof window !== 'undefined') {
    flushWhenPageHides()
    restoreUnsyncedDrafts()
    window.addEventListener('online', resendUnsyncedDrafts)
}

/** Loads every draft of the user; text this device has not synced yet wins over the server copy. */
export async function loadDrafts() {
    const { data } = await listDrafts()
    drafts.clear()
    for (const draft of data.drafts) drafts.set(draft.key, draft)
    restoreUnsyncedDrafts()
    loaded.value = true
    resendUnsyncedDrafts()
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

/** Draft of a new chat that belongs to no project. */
export const STANDALONE_DRAFT_KEY = 'new:chat'

/** Draft key of a session, of a new chat inside a project, or of a new standalone chat. */
export function draftKey(sessionId?: string, projectId?: string): string {
    if (sessionId) return `session:${sessionId}`
    return projectId ? `new:${projectId}` : STANDALONE_DRAFT_KEY
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
    rememberUnsyncedDraft(key, text)
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

/** Drops a sent draft that is no longer open, unless it was edited after sending. */
export function discardSentDraft(key: string, sentText: string) {
    if ((drafts.get(key)?.text ?? '').trim() !== sentText) return
    storeLocalDraft(key, '')
    scheduleSave(key, '')
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
