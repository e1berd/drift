<script setup lang="ts">
import { PhChatCircleDots } from '@phosphor-icons/vue'
import { Button } from '@ui-sdk/vue/button'
import { useI18n } from 'vue-i18n'

import QuestionField from '@/components/chat/question/QuestionField.vue'
import { useQuestionAnswers } from '@/composables/useQuestionAnswers'
import type { Message } from '@drift/vue/api'
import type { AgentId } from '@/shared/agents'
import { AGENTS } from '@/shared/agents'

const props = defineProps<{ message: Message; agent: AgentId }>()
const { t } = useI18n()

const { answers, pending, complete, sending, submit } = useQuestionAnswers(
    () => props.message,
    () => props.agent,
)
</script>

<template>
    <form class="chat-card" :data-closed="!pending || undefined" @submit.prevent="submit">
        <div class="chat-card-head">
            <PhChatCircleDots aria-hidden="true" />
            {{ t('title', { agent: AGENTS[agent].name }) }}
        </div>
        <QuestionField
            v-for="(question, index) in message.body.questions ?? []"
            :key="index"
            v-model="answers[index]!"
            :question="question"
            :disabled="!pending || sending"
        />
        <div class="chat-card-actions">
            <Button
                v-if="pending"
                type="submit"
                data-variant="solid"
                :disabled="!complete || sending"
            >
                {{ t('send') }}
            </Button>
            <span v-else-if="message.body.status === 'answered'" class="text-muted">{{
                t('answered')
            }}</span>
            <span v-else class="text-muted">{{ t('dismissed') }}</span>
        </div>
    </form>
</template>

<i18n lang="yaml">
ru:
    title: '{agent} спрашивает'
    send: Ответить
    answered: Ответ отправлен агенту.
    dismissed: Вопрос закрыт без ответа.
en:
    title: '{agent} asks'
    send: Answer
    answered: The answer went to the agent.
    dismissed: Closed without an answer.
</i18n>
