<script setup lang="ts">
import { PhCopy } from '@phosphor-icons/vue'
import { Button } from '@ui-sdk/vue/button'
import { toast } from '@ui-sdk/vue/toast'
import { useI18n } from 'vue-i18n'

import type { Message } from '@drift/vue/api'
import type { AgentId } from '@/shared/agents'

import AgentLogo from '@/components/AgentLogo.vue'
import MarkdownText from '@/components/chat/MarkdownText.vue'
import MessageAttachments from '@/components/chat/MessageAttachments.vue'
import ThinkingBlock from '@/components/chat/ThinkingBlock.vue'
import LimitCard from '@/components/chat/LimitCard.vue'
import QuestionCard from '@/components/chat/question/QuestionCard.vue'
import SystemNote from '@/components/chat/SystemNote.vue'
import UsageWarning from '@/components/chat/UsageWarning.vue'
import ToolCall from '@/components/chat/ToolCall.vue'
import { reportError } from '@drift/sdk/errors'

const props = defineProps<{ message: Message; agent: AgentId }>()
const { t } = useI18n()

function copyMessageText() {
    if (!props.message.body.text) return
    navigator.clipboard
        .writeText(props.message.body.text)
        .then(() => toast.success(t('copied')))
        .catch(reportError)
}
</script>

<template>
    <div v-if="message.role === 'user'" class="flex flex-col items-end gap-2">
        <MessageAttachments
            v-if="message.body.attachments?.length"
            :attachments="message.body.attachments"
        />
        <div
            v-if="message.body.text"
            class="max-w-[82%] rounded-(--ui-radius) bg-surface px-3 py-2 whitespace-pre-wrap"
        >
            {{ message.body.text }}
        </div>
        <Button
            v-if="message.body.text"
            :aria-label="t('copy')"
            :title="t('copy')"
            @click="copyMessageText"
        >
            <PhCopy />
        </Button>
    </div>

    <ThinkingBlock v-else-if="message.kind === 'thinking'" :body="message.body" />
    <ToolCall
        v-else-if="message.kind === 'tool'"
        :body="message.body"
        :data-message-id="message.id"
    />

    <QuestionCard v-else-if="message.kind === 'question'" :message="message" :agent="agent" />
    <LimitCard v-else-if="message.kind === 'limit'" :message="message" />
    <UsageWarning
        v-else-if="message.kind === 'note' && message.body.code === 'usageWarning'"
        :body="message.body"
    />
    <SystemNote v-else-if="message.kind === 'note'" :body="message.body" />

    <div v-else class="grid grid-cols-[20px_minmax(0,1fr)] gap-4">
        <AgentLogo :agent="agent" :size="20" decorative />
        <MarkdownText :source="message.body.text ?? ''" />
    </div>
</template>

<i18n lang="yaml">
ru:
    copy: Скопировать сообщение
    copied: Сообщение скопировано
en:
    copy: Copy message
    copied: Message copied
</i18n>
