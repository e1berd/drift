<script setup lang="ts">
import { computed, useTemplateRef, watch } from 'vue'
import {
    ScrollArea,
    ScrollAreaContent,
    ScrollAreaScrollbar,
    ScrollAreaThumb,
    ScrollAreaViewport,
} from '@ui-sdk/vue/scroll-area'
import { useI18n } from 'vue-i18n'

import type { AgentId } from '@/shared/agents'

import RunSpinner from '@/components/RunSpinner.vue'
import JumpToLatest from '@/components/chat/JumpToLatest.vue'
import LastPromptBar from '@/components/chat/LastPromptBar.vue'
import MessageItem from '@/components/chat/MessageItem.vue'
import { useChatScroll } from '@/composables/useChatScroll'
import { useInfiniteSentinel } from '@/composables/useInfiniteSentinel'
import { useSessionMessages } from '@/composables/useMessageWindow'

defineProps<{ agent: AgentId }>()
const { t } = useI18n()
const session = useSessionMessages()
const state = session.state

const top = useTemplateRef<HTMLElement>('top')
const bottom = useTemplateRef<HTMLElement>('bottom')
const viewport = computed(
    () => top.value?.closest<HTMLElement>('[data-slot="scroll-area-viewport"]') ?? null,
)
const { controller, atBottom, unread, scrollToEnd, olderHeld, stickToEnd } = useChatScroll(
    viewport,
    session,
)

function loadOlderUnlessHeld() {
    if (!olderHeld.value) void session.loadOlder()
}

useInfiniteSentinel(top, loadOlderUnlessHeld)
useInfiniteSentinel(bottom, () => void session.loadNewer())
watch(olderHeld, (held) => {
    if (!held && top.value && viewport.value && isNearTop(top.value, viewport.value))
        loadOlderUnlessHeld()
})

function isNearTop(sentinel: HTMLElement, scroller: HTMLElement) {
    return sentinel.getBoundingClientRect().bottom >= scroller.getBoundingClientRect().top - 120
}
</script>

<template>
    <ScrollArea
        :stick-to-end="stickToEnd"
        :controller="controller"
        class="scroll-pane chat-log min-h-0 flex-1"
    >
        <ScrollAreaViewport role="log" :aria-label="t('tabs.chat')">
            <ScrollAreaContent class="px-4 py-3">
                <div class="mx-auto flex max-w-[780px] flex-col py-6">
                    <div ref="top" class="h-px" />
                    <div v-if="state.hasOlder" class="older" aria-live="polite">
                        <span v-if="state.loadingOlder" class="status-tag">
                            <RunSpinner :size="13" fast />{{ t('chat.olderLoading') }}
                        </span>
                    </div>
                    <div class="messages">
                        <MessageItem
                            v-for="message in state.messages"
                            :key="message.id"
                            :data-message-id="message.id"
                            :data-kind="message.role === 'user' ? 'user' : message.kind"
                            :message="message"
                            :agent="agent"
                        />
                        <slot v-if="!state.hasNewer" name="after" />
                    </div>
                    <div v-if="state.hasNewer" class="older" aria-live="polite">
                        <span v-if="state.loadingNewer" class="status-tag">
                            <RunSpinner :size="13" fast />{{ t('chat.newerLoading') }}
                        </span>
                    </div>
                    <div ref="bottom" class="h-px" />
                </div>
            </ScrollAreaContent>
        </ScrollAreaViewport>
        <ScrollAreaScrollbar orientation="vertical"><ScrollAreaThumb /></ScrollAreaScrollbar>
        <LastPromptBar :viewport="viewport" />
        <JumpToLatest :visible="!atBottom || state.hasNewer" :unread="unread" @jump="scrollToEnd" />
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
