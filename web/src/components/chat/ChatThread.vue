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
    <ScrollArea stick-to-end class="scroll-pane min-h-0 flex-1">
        <ScrollAreaViewport role="log" :aria-label="t('tabs.chat')">
            <ScrollAreaContent class="px-4 py-3">
                <div class="mx-auto flex max-w-[780px] flex-col gap-4 py-6">
                    <div ref="top" class="h-px" />
                    <p v-if="loadingOlder" class="self-center text-muted">
                        {{ t('chat.olderLoading') }}
                    </p>
                    <div v-auto-animate class="flex flex-col gap-4">
                        <MessageItem
                            v-for="message in messages"
                            :key="message.id"
                            :message="message"
                            :agent="agent"
                        />
                    </div>
                    <slot name="after" />
                </div>
            </ScrollAreaContent>
        </ScrollAreaViewport>
        <ScrollAreaScrollbar orientation="vertical"><ScrollAreaThumb /></ScrollAreaScrollbar>
        <ScrollAreaScrollToEnd :aria-label="t('chat.jumpToEnd')" class="jump-to-end">
            <PhArrowDown />
        </ScrollAreaScrollToEnd>
    </ScrollArea>
</template>

<style scoped>
.jump-to-end {
    position: absolute;
    bottom: 12px;
    left: 50%;
    translate: -50% 0;
}
</style>
