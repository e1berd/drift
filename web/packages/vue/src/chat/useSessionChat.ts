import { computed, toValue } from 'vue'

import type { MaybeRefOrGetter } from 'vue'

import { useCreateMessage, useStopSession, useUpdateSession } from '../api'
import { useComposerSettings } from '../composables/useComposerSettings'
import { useEventStream } from '../composables/useEventStream'
import { provideSessionMessages } from '../composables/useMessageWindow'
import { useErrorNotice } from '../notifier'
import { useProjects, useSession } from '../queries'
import type { AgentChoice } from './useNewChat'
import type { SentPrompt } from '@drift/sdk/attachments'

export type ReplyState = 'reply' | 'queue' | 'afterReset'

/**
 * One open session: its messages (provided to the thread below), agent and run state, the
 * prompts sent so far, and the actions on it. Ask the user before calling `switchAgent`.
 */
export function useSessionChat(sessionId: MaybeRefOrGetter<string>) {
    const showError = useErrorNotice()
    const settings = useComposerSettings()
    const session = useSession(sessionId)
    const { projects } = useProjects()
    const sessionMessages = provideSessionMessages(sessionId)
    const messages = sessionMessages.messages
    const { activities } = useEventStream()

    const project = computed(() =>
        projects.value.find((item) => item.id === session.data.value?.projectId),
    )
    const agent = computed(() => session.data.value?.agent ?? 'claude')
    const running = computed(() => session.data.value?.status === 'running')
    const activity = computed(() => activities.value.get(toValue(sessionId)))
    const waitingForReset = computed(() =>
        messages.value.some(
            (message) => message.kind === 'limit' && message.body.status === 'scheduled',
        ),
    )
    const replyState = computed<ReplyState>(() => {
        if (waitingForReset.value) return 'afterReset'
        return running.value ? 'queue' : 'reply'
    })
    const runningTool = computed(() => {
        if (sessionMessages.state.value.hasNewer) return undefined
        const last = messages.value.at(-1)
        return last?.kind === 'tool' && last.body.status === 'running' ? last.body.name : undefined
    })
    const history = computed(() =>
        messages.value.flatMap((message) =>
            message.role === 'user' && message.body.text ? [message.body.text] : [],
        ),
    )

    const post = useCreateMessage({ mutation: { onError: showError } })
    const update = useUpdateSession({ mutation: { onError: showError } })
    const stopMutation = useStopSession({ mutation: { onError: showError } })

    function switchAgent({ agent: next, model }: AgentChoice) {
        settings.setModel(next, model, toValue(sessionId))
        update.mutate({ path: { sessionId: toValue(sessionId) }, body: { agent: next } })
    }

    async function send({ text, attachments }: SentPrompt) {
        const attachmentIds = attachments.map((attachment) => attachment.id)
        if (!text && !attachmentIds.length) return
        await post.mutateAsync({
            path: { sessionId: toValue(sessionId) },
            body: {
                text,
                attachmentIds,
                model: settings.modelFor(agent.value, toValue(sessionId)),
                mode: settings.modeFor(agent.value),
            },
        })
    }

    function stop() {
        stopMutation.mutate({ path: { sessionId: toValue(sessionId) } })
    }

    return {
        session,
        project,
        messages,
        sessionMessages,
        agent,
        running,
        activity,
        replyState,
        runningTool,
        history,
        busy: computed(() => post.isPending.value || update.isPending.value),
        stopping: stopMutation.isPending,
        switchAgent,
        send,
        stop,
    }
}
