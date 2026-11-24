<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'

import ViewBar from '@/components/ViewBar.vue'
import ChatThread from '@/components/chat/ChatThread.vue'
import PromptBox from '@/components/composer/PromptBox.vue'
import { useSessionChat } from '@drift/vue/chat/useSessionChat'
import { draftKey } from '@drift/vue/composables/useDrafts'
import { AGENTS } from '@drift/sdk/agents'

const { t } = useI18n()
const route = useRoute()
const sessionId = computed(() => String(route.params.sessionId ?? ''))
const { session, project, agent, running, replyState, history, busy, stopping, send, stop } =
    useSessionChat(sessionId)

const placeholder = computed(() =>
    t(`composer.placeholder.${replyState.value}`, { agent: AGENTS[agent.value].name }),
)
</script>

<template>
    <ViewBar :title="session.data.value?.title ?? ''" :subtitle="project?.name" />
    <ChatThread :agent="agent" :running="running" />
    <div class="session-composer">
        <PromptBox
            :draft-key="draftKey(sessionId)"
            :agent="agent"
            :session-id="sessionId"
            :placeholder="placeholder"
            :history="history"
            :busy="busy"
            :running="running"
            :stopping="stopping"
            :submit="send"
            autofocus
            @stop="stop"
        />
    </div>
</template>

<style scoped>
.session-composer {
    inline-size: min(820px, 100%);
    margin-inline: auto;
    padding: 0 16px 16px;
}
</style>
