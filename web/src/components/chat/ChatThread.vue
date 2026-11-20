<script setup lang="ts">
import { computed, toRef, useTemplateRef } from 'vue'
import {
    ScrollArea,
    ScrollAreaContent,
    ScrollAreaScrollbar,
    ScrollAreaThumb,
    ScrollAreaViewport,
} from '@ui-sdk/vue/scroll-area'

import RunSpinner from '@/components/RunSpinner.vue'
import JumpToLatest from '@/components/chat/JumpToLatest.vue'
import LastPromptBar from '@/components/chat/LastPromptBar.vue'
import MessageItem from '@/components/chat/MessageItem.vue'
import { useChatScroll } from '@/composables/useChatScroll'
import { useInfiniteSentinel } from '@/composables/useInfiniteSentinel'
import type { Message } from '@/gen'
import type { AgentId } from '@/shared/agents'
import { useI18n } from 'vue-i18n'

const props = defineProps<{
    messages: Message[]
    agent: AgentId
    hasOlder: boolean
    loadingOlder: boolean
}>()
const emit = defineEmits<{ loadOlder: [] }>()
const { t } = useI18n()

const top = useTemplateRef<HTMLElement>('top')
useInfiniteSentinel(top, () => {
    if (props.hasOlder && !props.loadingOlder) emit('loadOlder')
})

const viewport = computed(
    () => top.value?.closest<HTMLElement>('[data-slot="scroll-area-viewport"]') ?? null,
)
const { controller, atBottom, unread, scrollToEnd } = useChatScroll(
    viewport,
    toRef(props, 'messages'),
)
const lastPrompt = computed(() =>
    props.messages.findLast(
        (message) =>
            message.role === 'user' &&
            (!!message.body.text?.trim() || !!message.body.attachments?.length),
    ),
)
</script>

<template>
    <ScrollArea stick-to-end :controller="controller" class="scroll-pane chat-log min-h-0 flex-1">
        <ScrollAreaViewport role="log" :aria-label="t('tabs.chat')">
            <ScrollAreaContent class="px-4 py-3">
                <div class="mx-auto flex max-w-[780px] flex-col py-6">
                    <div ref="top" class="h-px" />
                    <div v-if="hasOlder" class="older" aria-live="polite">
                        <span v-if="loadingOlder" class="status-tag">
                            <RunSpinner :size="13" fast />{{ t('chat.olderLoading') }}
                        </span>
                    </div>
                    <div class="messages">
                        <MessageItem
                            v-for="message in messages"
                            :key="message.id"
                            :data-message-id="message.id"
                            :data-kind="message.role === 'user' ? 'user' : message.kind"
                            :message="message"
                            :agent="agent"
                        />
                    </div>
                    <slot name="after" />
                </div>
            </ScrollAreaContent>
        </ScrollAreaViewport>
        <ScrollAreaScrollbar orientation="vertical"><ScrollAreaThumb /></ScrollAreaScrollbar>
        <LastPromptBar :message="lastPrompt" :viewport="viewport" />
        <JumpToLatest :visible="!atBottom" :unread="unread" @jump="scrollToEnd" />
    </ScrollArea>
</template>

<style scoped>
.older {
    display: grid;
    height: 32px;
    place-items: center;
}

.messages {
    display: flex;
    flex-direction: column;
    padding-bottom: calc(var(--ui-spacing) * 4);

    & > * + * {
        margin-top: calc(var(--ui-spacing) * 4);
    }

    & > [data-kind='tool'] + [data-kind='tool'],
    & > [data-kind='thinking'] + [data-kind='tool'],
    & > [data-kind='tool'] + [data-kind='thinking'] {
        margin-top: calc(var(--ui-spacing) * 1);
    }

    & > [data-kind='user'] {
        scroll-margin-top: calc(var(--ui-spacing) * 20);
    }
}

.chat-log :deep([data-slot='scroll-area-viewport']) {
    overflow-anchor: none;
}
</style>
