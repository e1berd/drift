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
    <div v-if="message.role === 'user'" class="flex flex-col items-end gap-1.5">
        <div v-if="message.body.attachments?.length" class="flex flex-wrap justify-end gap-2">
            <span
                v-for="a in message.body.attachments"
                :key="a.name"
                class="flex h-[52px] max-w-[230px] items-center gap-2.5 rounded-[9px] border border-line-strong bg-surface px-3"
            >
                <PhFile :size="18" class="text-accent" />
                <span class="truncate text-[12.5px] font-medium">{{ a.name }}</span>
            </span>
        </div>
        <div
            v-if="message.body.text"
            class="max-w-[82%] rounded-[14px_14px_4px_14px] border border-line bg-surface-2 px-3.5 py-2.5 leading-normal whitespace-pre-wrap"
        >
            {{ message.body.text }}
        </div>
    </div>

    <ThinkingBlock v-else-if="message.kind === 'thinking'" :body="message.body" />
    <ToolCall v-else-if="message.kind === 'tool'" :body="message.body" />

    <div
        v-else-if="message.kind === 'note'"
        class="flex items-center gap-2 self-center text-[12px] text-faint before:h-px before:w-10 before:bg-line-strong after:h-px after:w-10 after:bg-line-strong"
    >
        {{ message.body.text }}
    </div>

    <div v-else class="grid grid-cols-[22px_minmax(0,1fr)] gap-3.5 leading-relaxed">
        <span class="mt-0.5 grid h-[22px] w-[22px] place-items-center"
            ><AgentLogo :agent="agent" decorative
        /></span>
        <p class="m-0 whitespace-pre-wrap">{{ message.body.text }}</p>
    </div>
</template>
