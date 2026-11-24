<script setup lang="ts">
import { computed, useTemplateRef } from 'vue'
import { mdiArrowDown } from '@mdi/js'
import { useI18n } from 'vue-i18n'

import type { AgentId } from '@drift/sdk/agents'

import MessageCard from '@/components/chat/MessageCard.vue'
import { useChatThread } from '@drift/vue/chat/useChatThread'

defineProps<{ agent: AgentId; running: boolean }>()
const { t } = useI18n()
const scroller = useTemplateRef<HTMLElement>('scroller')
const top = useTemplateRef<HTMLElement>('top')
const bottom = useTemplateRef<HTMLElement>('bottom')
const { session, state, atBottom, unread, scrollToEnd } = useChatThread(scroller, top, bottom)
const messages = session.messages
const showJump = computed(() => !atBottom.value || state.value.hasNewer)
</script>

<template>
    <div class="thread">
        <div ref="scroller" class="thread-scroll">
            <div class="thread-column">
                <div ref="top" />
                <v-progress-linear v-if="state.loadingOlder" indeterminate />
                <MessageCard
                    v-for="message in messages"
                    :key="message.id"
                    :message="message"
                    :agent="agent"
                    :data-message-id="message.id"
                />
                <v-progress-linear v-if="state.loadingNewer" indeterminate />
                <div v-if="running" class="d-flex align-center ga-2 text-medium-emphasis">
                    <v-progress-circular indeterminate size="16" width="2" />
                    {{ t('chat.working') }}
                </div>
                <div ref="bottom" />
            </div>
        </div>
        <v-btn
            v-if="showJump"
            class="thread-jump"
            variant="elevated"
            rounded="pill"
            :prepend-icon="mdiArrowDown"
            @click="scrollToEnd"
            >{{ unread ? t('chat.unread', { count: unread }) : t('chat.latest') }}</v-btn
        >
    </div>
</template>

<style scoped>
.thread {
    position: relative;
    flex: 1;
    min-block-size: 0;
}

.thread-scroll {
    block-size: 100%;
    overflow-y: auto;
}

.thread-column {
    display: flex;
    flex-direction: column;
    gap: 16px;
    max-inline-size: 820px;
    margin-inline: auto;
    padding: 16px;
}

.thread-jump {
    position: absolute;
    inset-block-end: 12px;
    inset-inline-start: 50%;
    translate: -50% 0;
}
</style>
