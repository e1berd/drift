import { computed, ref, toValue } from 'vue'

import type { MaybeRefOrGetter } from 'vue'

import { useAgents } from '../composables/useAgents'
import { useComposerUploads } from '../composables/useComposerUploads'
import { discardSentDraft, useDraft } from '../composables/useDrafts'
import type { AgentId } from '@drift/sdk/agents'
import type { SentPrompt } from '@drift/sdk/attachments'
import { reportError } from '@drift/sdk/errors'

export interface PromptComposerOptions {
    draftKey: MaybeRefOrGetter<string>
    agent: MaybeRefOrGetter<AgentId>
    submit: (prompt: SentPrompt) => Promise<unknown>
    history: MaybeRefOrGetter<string[]>
    busy: MaybeRefOrGetter<boolean>
}

/**
 * The prompt box: the saved draft and its uploads, stepping through earlier prompts, and sending,
 * which gives the files back to the draft when the send fails.
 */
export function usePromptComposer(options: PromptComposerOptions) {
    const agents = useAgents()
    const draft = useDraft(() => toValue(options.draftKey))
    const uploads = useComposerUploads(() => toValue(options.draftKey))
    const historyIndex = ref(-1)
    const sending = ref(false)

    const unavailable = computed(
        () => agents.ready.value && !agents.isConnected(toValue(options.agent)),
    )
    const canSend = computed(
        () =>
            !unavailable.value &&
            !toValue(options.busy) &&
            !sending.value &&
            !uploads.uploading.value &&
            (!!draft.text.value.trim() || uploads.ready.value.length > 0),
    )

    function setText(value: string) {
        draft.set(value)
        draft.external.value++
    }

    function clearSentDraft(key: string, text: string) {
        if (key !== toValue(options.draftKey)) discardSentDraft(key, text)
        else if (draft.text.value.trim() === text) draft.clear()
    }

    async function send() {
        if (!canSend.value) return
        const key = toValue(options.draftKey)
        const prompt = { text: draft.text.value.trim(), attachments: uploads.takeReady() }
        sending.value = true
        try {
            await options.submit(prompt)
            historyIndex.value = -1
            clearSentDraft(key, prompt.text)
        } catch (error) {
            uploads.returnUnsent(key, prompt.attachments)
            reportError(error)
        } finally {
            sending.value = false
        }
    }

    function back() {
        const history = toValue(options.history)
        if (!history.length) return
        const index =
            historyIndex.value === -1 ? history.length - 1 : Math.max(0, historyIndex.value - 1)
        historyIndex.value = index
        setText(history.at(index) ?? '')
    }

    function forward() {
        if (historyIndex.value === -1) return
        const history = toValue(options.history)
        const index = historyIndex.value + 1
        if (index >= history.length) {
            historyIndex.value = -1
            setText('')
            return
        }
        historyIndex.value = index
        setText(history.at(index) ?? '')
    }

    function type(value: string) {
        historyIndex.value = -1
        draft.set(value)
    }

    return {
        draft,
        uploads,
        historyIndex,
        sending,
        unavailable,
        canSend,
        send,
        back,
        forward,
        type,
    }
}
