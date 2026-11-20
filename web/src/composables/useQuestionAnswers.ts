import { computed, ref, toValue, watch } from 'vue'
import type { MaybeRefOrGetter } from 'vue'
import { toast } from '@ui-sdk/vue/toast'

import { useAnswerQuestion } from '@/gen'
import type { Answer, Message } from '@/gen'
import { useComposerSettings } from '@/composables/useComposerSettings'
import type { AgentId } from '@/shared/agents'
import { errorMessage } from '@/shared/errors'

function emptyAnswers(message: Message): Answer[] {
    return (message.body.questions ?? []).map(() => ({ selected: [], custom: '' }))
}

function isAnswered(answer: Answer): boolean {
    return !!answer.selected?.length || !!answer.custom?.trim()
}

/** Answers the user is picking on a question card, and sending them to the agent. */
export function useQuestionAnswers(
    message: MaybeRefOrGetter<Message>,
    agent: MaybeRefOrGetter<AgentId>,
) {
    const settings = useComposerSettings()
    const pending = computed(() => toValue(message).body.status === 'pending')
    const answers = ref<Answer[]>(toValue(message).body.answers ?? emptyAnswers(toValue(message)))

    watch(
        () => toValue(message).body.answers,
        (stored) => {
            if (stored) answers.value = stored
        },
    )

    const complete = computed(() => answers.value.length > 0 && answers.value.every(isAnswered))
    const mutation = useAnswerQuestion({
        mutation: { onError: (error) => toast.error(errorMessage(error)) },
    })

    function submit() {
        if (!pending.value || !complete.value) return
        const current = toValue(message)
        mutation.mutate({
            path: { sessionId: current.sessionId, messageId: current.id },
            body: {
                answers: answers.value.map((answer) => ({
                    selected: answer.selected ?? [],
                    custom: answer.custom?.trim() ?? '',
                })),
                model: settings.modelFor(toValue(agent), current.sessionId),
                mode: settings.modeFor(toValue(agent)),
            },
        })
    }

    return { answers, pending, complete, sending: mutation.isPending, submit }
}
