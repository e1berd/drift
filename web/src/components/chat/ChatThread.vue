<script setup lang="ts">
import { useTemplateRef } from 'vue'
import { PhArrowDown } from '@phosphor-icons/vue'
import {
    ScrollArea,
    ScrollAreaContent,
    ScrollAreaScrollToEnd,
    ScrollAreaScrollbar,
    ScrollAreaThumb,
    ScrollAreaViewport,
} from '@ui-sdk/vue/scroll-area'

import MessageItem from '@/components/chat/MessageItem.vue'
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
</script>

<template>
    <ScrollArea stick-to-end class="relative min-h-0 flex-1">
        <ScrollAreaViewport role="log" :aria-label="t('tabs.chat')" class="h-full w-full">
            <ScrollAreaContent>
                <div
                    class="mx-auto flex max-w-[780px] flex-col gap-3.5 px-6 pt-7 pb-7 max-md:px-3.5"
                >
                    <div ref="top" class="h-px" />
                    <p v-if="loadingOlder" class="self-center text-[12px] text-faint">
                        {{ t('chat.olderLoading') }}
                    </p>
                    <div v-auto-animate class="flex flex-col gap-3.5">
                        <MessageItem
                            v-for="m in messages"
                            :key="m.id"
                            :message="m"
                            :agent="agent"
                        />
                    </div>
                    <slot name="after" />
                </div>
            </ScrollAreaContent>
        </ScrollAreaViewport>
        <ScrollAreaScrollbar orientation="vertical" class="chat-scrollbar">
            <ScrollAreaThumb class="chat-thumb" />
        </ScrollAreaScrollbar>
        <ScrollAreaScrollToEnd class="chat-to-end" :aria-label="t('chat.jumpToEnd')">
            <PhArrowDown :size="16" />
        </ScrollAreaScrollToEnd>
    </ScrollArea>
</template>

<style>
.chat-scrollbar {
    display: flex;
    flex-direction: column;
    width: 8px;
    padding: 2px;
    opacity: 0;
    transition: opacity var(--dur-2);

    &[data-visible],
    &:hover {
        opacity: 1;
    }
}

.chat-thumb {
    border-radius: 4px;
    background: var(--border-strong);
}

.chat-to-end {
    position: absolute;
    bottom: 14px;
    left: 50%;
    display: grid;
    place-items: center;
    width: 34px;
    height: 34px;
    border: 1px solid var(--border-strong);
    border-radius: 50%;
    background: var(--surface);
    box-shadow: var(--shadow);
    color: var(--ui-foreground);
    opacity: 0;
    pointer-events: none;
    translate: -50% 8px;
    transition:
        opacity var(--dur-2) var(--ease-out),
        translate var(--dur-3) var(--ease-out);

    &[data-visible] {
        opacity: 1;
        pointer-events: auto;
        translate: -50% 0;
    }
}
</style>
