<script setup lang="ts">
import { PhFile } from '@phosphor-icons/vue'

import AgentLogo from '@/components/AgentLogo.vue'
import ThinkingBlock from '@/components/chat/ThinkingBlock.vue'
import ToolCall from '@/components/chat/ToolCall.vue'
import type { Message } from '@/gen'
import type { AgentId } from '@/shared/agents'

defineProps<{ message: Message; agent: AgentId }>()
</script>

<template>
    <div v-if="message.role === 'user'" class="flex flex-col items-end gap-2">
        <div v-if="message.body.attachments?.length" class="flex flex-wrap justify-end gap-2">
            <span
                v-for="attachment in message.body.attachments"
                :key="attachment.name"
                class="status-tag"
            >
                <PhFile />{{ attachment.name }}
            </span>
        </div>
        <div
            v-if="message.body.text"
            class="max-w-[82%] rounded-(--ui-radius) bg-surface px-3 py-2 whitespace-pre-wrap"
        >
            {{ message.body.text }}
        </div>
    </div>

    <ThinkingBlock v-else-if="message.kind === 'thinking'" :body="message.body" />
    <ToolCall v-else-if="message.kind === 'tool'" :body="message.body" />

    <p v-else-if="message.kind === 'note'" class="m-0 self-center text-muted">
        {{ message.body.text }}
    </p>

    <div v-else class="grid grid-cols-[20px_minmax(0,1fr)] gap-4">
        <AgentLogo :agent="agent" :size="20" decorative />
        <p class="m-0 whitespace-pre-wrap">{{ message.body.text }}</p>
    </div>
</template>
