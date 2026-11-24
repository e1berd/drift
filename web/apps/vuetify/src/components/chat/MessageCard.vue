<script setup lang="ts">
import { useI18n } from 'vue-i18n'

import type { Message } from '@drift/vue/api'
import type { AgentId } from '@drift/sdk/agents'

import AgentAvatar from '@/components/AgentAvatar.vue'
import LimitCard from '@/components/chat/LimitCard.vue'
import MarkdownText from '@/components/chat/MarkdownText.vue'
import QuestionCard from '@/components/chat/QuestionCard.vue'
import ToolLine from '@/components/chat/ToolLine.vue'

defineProps<{ message: Message; agent: AgentId }>()
const { t } = useI18n()
</script>

<template>
    <div v-if="message.role === 'user'" class="d-flex flex-column align-end ga-1">
        <v-chip v-for="attachment in message.body.attachments" :key="attachment.id" size="small">{{
            attachment.name
        }}</v-chip>
        <v-sheet
            v-if="message.body.text"
            color="surface-light"
            rounded="lg"
            class="user-text px-4 py-2"
            >{{ message.body.text }}</v-sheet
        >
    </div>
    <v-expansion-panels v-else-if="message.kind === 'thinking'" variant="accordion" flat>
        <v-expansion-panel :title="t('chat.thinking')" class="bg-transparent">
            <v-expansion-panel-text class="text-medium-emphasis">
                <MarkdownText :source="message.body.text ?? ''" />
            </v-expansion-panel-text>
        </v-expansion-panel>
    </v-expansion-panels>
    <ToolLine v-else-if="message.kind === 'tool'" :body="message.body" />
    <QuestionCard v-else-if="message.kind === 'question'" :message="message" :agent="agent" />
    <LimitCard v-else-if="message.kind === 'limit'" :message="message" />
    <div v-else-if="message.kind === 'note'" class="text-caption text-medium-emphasis text-center">
        {{ message.body.code ? t(`notes.${message.body.code}`) : message.body.text }}
    </div>
    <div v-else class="d-flex ga-3">
        <AgentAvatar :agent="agent" :size="24" class="mt-1" />
        <MarkdownText :source="message.body.text ?? ''" class="flex-grow-1" />
    </div>
</template>

<style scoped>
.user-text {
    max-inline-size: 82%;
    white-space: pre-wrap;
}
</style>
