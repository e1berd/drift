<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { useDisplay } from 'vuetify'

import AgentAvatar from '@/components/AgentAvatar.vue'
import ViewBar from '@/components/ViewBar.vue'
import PromptBox from '@/components/composer/PromptBox.vue'
import { useNewChat } from '@drift/vue/chat/useNewChat'
import { draftKey } from '@drift/vue/composables/useDrafts'
import { useProjects } from '@drift/vue/queries'
import { AGENTS } from '@drift/sdk/agents'
import type { AgentId } from '@drift/sdk/agents'

const NO_PROJECT = ''

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const { mobile } = useDisplay()
const { projects } = useProjects()

const projectId = computed({
    get: () => (typeof route.query.project === 'string' ? route.query.project : NO_PROJECT),
    set: (id: string) =>
        void router.replace(id ? { name: 'new', query: { project: id } } : { name: 'new' }),
})
const projectItems = computed(() => [
    { value: NO_PROJECT, title: t('composer.noProject') },
    ...projects.value.map((project) => ({ value: project.id, title: project.name })),
])

const { agent, agents, creating, selectAgent, send } = useNewChat(projectId, (session) =>
    router.push({ name: 'session', params: { sessionId: session.id } }),
)
const agentChoice = computed({
    get: () => agent.value,
    set: (next: AgentId) => selectAgent(next),
})
</script>

<template>
    <ViewBar :title="t('nav.newChat')" />
    <div class="new-chat">
        <h1 class="text-h5 text-center">{{ t('newChat.title') }}</h1>
        <p class="text-medium-emphasis text-center">{{ t('newChat.lead') }}</p>
        <v-btn-toggle
            v-if="agents.available.value.length"
            v-model="agentChoice"
            mandatory
            divided
            variant="outlined"
            density="comfortable"
        >
            <v-btn v-for="item in agents.available.value" :key="item.id" :value="item.id">
                <AgentAvatar :agent="item.id" :size="18" class="me-2" />
                {{ item.name }}
            </v-btn>
        </v-btn-toggle>
        <PromptBox
            :draft-key="draftKey(undefined, projectId)"
            :agent="agent"
            :placeholder="t('composer.placeholder.new', { agent: AGENTS[agent].name })"
            :busy="creating"
            :autofocus="!mobile"
            :submit="send"
        >
            <template #start>
                <v-select
                    v-model="projectId"
                    :items="projectItems"
                    :aria-label="t('composer.project')"
                    variant="plain"
                    class="project-select"
                />
            </template>
        </PromptBox>
    </div>
</template>

<style scoped>
.new-chat {
    display: flex;
    flex: 1;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 16px;
    inline-size: min(760px, 100%);
    margin-inline: auto;
    padding: clamp(12px, 3vw, 32px);
    overflow-y: auto;
}

.project-select {
    flex: 0 1 auto;
    min-inline-size: 140px;
}
</style>
