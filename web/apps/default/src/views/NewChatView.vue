<script setup lang="ts">
import { computed, ref } from 'vue'
import {
    Select,
    SelectContent,
    SelectItem,
    SelectItemIndicator,
    SelectItemText,
    SelectTrigger,
    SelectValue,
} from '@ui-sdk/vue/select'
import { RouterLink, useRoute, useRouter } from 'vue-router'

import ScrollPane from '@/components/ScrollPane.vue'
import AgentPicker from '@/components/AgentPicker.vue'
import BrandMark from '@/components/BrandMark.vue'
import TopBar from '@/components/TopBar.vue'
import PromptComposer from '@/components/composer/PromptComposer.vue'
import { useMediaQuery } from '@drift/vue/composables/useMediaQuery'
import { draftKey } from '@drift/vue/composables/useDrafts'
import { AGENTS } from '@drift/sdk/agents'
import type { AgentId } from '@drift/sdk/agents'
import { useI18n } from 'vue-i18n'
import { useProjects } from '@drift/vue/queries'
import { useNewChat } from '@drift/vue/chat/useNewChat'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const { projects } = useProjects()
const narrow = useMediaQuery('(max-width: 760px)')
const composerFocused = ref(false)
const zen = computed(() => narrow.value && composerFocused.value)

function onComposerFocusOut(event: FocusEvent) {
    const wrapper = event.currentTarget as HTMLElement
    if (!wrapper.contains(event.relatedTarget as Node | null)) composerFocused.value = false
}

const NO_PROJECT = 'none'

const projectId = computed(() =>
    typeof route.query.project === 'string' ? route.query.project : '',
)
const projectChoice = computed({
    get: () => projectId.value || NO_PROJECT,
    set: (id: string | null) => {
        if (!id) return
        void router.replace(
            id === NO_PROJECT ? { name: 'new' } : { name: 'new', query: { project: id } },
        )
    },
})
const projectItems = computed(() => [
    { value: NO_PROJECT, label: t('composer.noProject') },
    ...projects.value.map((project) => ({ value: project.id, label: project.name })),
])
const { agent, agents, creating, selectAgent, switchAgent, send } = useNewChat(
    projectId,
    (session) => router.push({ name: 'session', params: { sessionId: session.id } }),
)
const agentValue = computed({
    get: () => agent.value,
    set: (value: string | null) => value && selectAgent(value as AgentId),
})
</script>

<template>
    <TopBar>
        <template #start><span class="truncate font-medium">{{ t('nav.newChat') }}</span></template>
    </TopBar>
    <ScrollPane
        class="motion-fade min-h-0 flex-1"
        :content-class="
            zen
                ? 'flex min-h-full items-start justify-center p-3'
                : 'flex min-h-full items-center justify-center p-6 max-md:p-3'
        "
    >
        <div class="flex w-full max-w-[720px] flex-col items-center gap-5">
            <div v-show="!zen" class="contents">
                <BrandMark :size="32" />
                <h1 class="m-0 text-2xl font-semibold">
                    {{ t('empty.title') }}
                </h1>
                <p class="m-0 text-center text-muted">{{ t('empty.lead') }}</p>

                <AgentPicker
                    v-if="agents.available.value.length"
                    v-model="agentValue"
                    :agents="agents.available.value"
                    :label="t('composer.model')"
                />
            </div>

            <div class="contents" @focusin="composerFocused = true" @focusout="onComposerFocusOut">
                <PromptComposer
                    :draft-key="draftKey(undefined, projectId)"
                    :agent="agent"
                    :placeholder="t('composer.placeholder.new', { agent: AGENTS[agent].name })"
                    :busy="creating"
                    :autofocus="!narrow"
                    :submit="send"
                    @switch-agent="switchAgent"
                >
                    <template #start>
                        <Select v-model:value="projectChoice" :items="projectItems">
                            <SelectTrigger
                                class="composer-select"
                                :aria-label="t('composer.project')"
                                ><SelectValue
                            /></SelectTrigger>
                            <SelectContent side="top" align="start">
                                <SelectItem
                                    v-for="item in projectItems"
                                    :key="item.value"
                                    :value="item.value"
                                >
                                    <SelectItemText>{{ item.label }}</SelectItemText>
                                    <SelectItemIndicator />
                                </SelectItem>
                            </SelectContent>
                        </Select>
                    </template>
                </PromptComposer>
            </div>

            <p v-show="!zen" class="m-0 text-muted">
                {{ t('empty.connectHint') }}
                <RouterLink to="/settings" class="underline underline-offset-4">{{
                    t('nav.settings').toLowerCase()
                }}</RouterLink>
            </p>
        </div>
    </ScrollPane>
</template>
