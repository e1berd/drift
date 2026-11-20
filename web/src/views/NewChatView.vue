<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import {
    SegmentedControl,
    SegmentedControlIndicator,
    SegmentedControlItem,
} from '@ui-sdk/vue/segmented-control'
import {
    Select,
    SelectContent,
    SelectItem,
    SelectItemIndicator,
    SelectItemText,
    SelectTrigger,
    SelectValue,
} from '@ui-sdk/vue/select'
import { toast } from '@ui-sdk/vue/toast'
import { RouterLink, useRoute, useRouter } from 'vue-router'

import ScrollPane from '@/components/ScrollPane.vue'
import AgentLogo from '@/components/AgentLogo.vue'
import BrandMark from '@/components/BrandMark.vue'
import TopBar from '@/components/TopBar.vue'
import PromptComposer from '@/components/composer/PromptComposer.vue'
import { useComposerSettings } from '@/composables/useComposerSettings'
import { draftKey } from '@/composables/useDrafts'
import { useAgents } from '@/composables/useAgents'
import { useCreateSession, useCreateStandaloneSession } from '@/gen'
import type { CreateSessionResponse } from '@/gen'
import type { ReadyAttachment } from '@/shared/attachments'
import { errorMessage } from '@/shared/errors'
import { AGENTS } from '@/shared/agents'
import type { AgentId } from '@/shared/agents'
import { useI18n } from 'vue-i18n'
import { useProjects } from '@/shared/queries'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const { projects } = useProjects()
const agents = useAgents()
const settings = useComposerSettings()

const agent = ref<AgentId>(agents.connected.value.at(0) ?? 'claude')
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
const agentValue = computed({
    get: () => agent.value,
    set: (value: string | null) => {
        if (value && agents.isConnected(value as AgentId)) agent.value = value as AgentId
    },
})

watch(agents.connected, (list) => {
    const first = list.at(0)
    if (!list.includes(agent.value) && first) agent.value = first
})

const created = {
    mutation: {
        onSuccess: ({ session }: CreateSessionResponse) =>
            router.push({ name: 'session', params: { sessionId: session.id } }),
        onError: (error: unknown) => toast.error(errorMessage(error)),
    },
}
const createInProject = useCreateSession(created)
const createStandalone = useCreateStandaloneSession(created)
const creating = computed(() => createInProject.isPending.value || createStandalone.isPending.value)

function switchAgent(choice: { agent: AgentId; model: string }) {
    settings.setModel(choice.agent, choice.model)
    agent.value = choice.agent
}

function send({ text, attachments }: { text: string; attachments: ReadyAttachment[] }) {
    if (!agents.isConnected(agent.value)) return
    const attachmentIds = attachments.map((attachment) => attachment.id)
    if (!text && !attachmentIds.length) return
    const body = {
        agent: agent.value,
        text,
        title: text ? undefined : attachments.at(0)?.name.slice(0, 200),
        attachmentIds,
        model: settings.modelFor(agent.value),
        mode: settings.mode(),
    }
    if (projectId.value) createInProject.mutate({ path: { projectId: projectId.value }, body })
    else createStandalone.mutate({ body })
}
</script>

<template>
    <TopBar>
        <template #start><span class="truncate font-medium">{{ t('nav.newChat') }}</span></template>
    </TopBar>
    <ScrollPane
        class="motion-fade min-h-0 flex-1"
        content-class="flex min-h-full items-center justify-center p-6 max-md:p-3"
    >
        <div class="flex w-full max-w-[720px] flex-col items-center gap-5">
            <BrandMark :size="32" />
            <h1 class="m-0 text-2xl font-semibold">{{ t('empty.title') }}</h1>
            <p class="m-0 text-center text-muted">{{ t('empty.lead') }}</p>

            <SegmentedControl
                v-if="agents.available.value.length"
                v-model:value="agentValue"
                :aria-label="t('composer.model')"
            >
                <SegmentedControlIndicator />
                <SegmentedControlItem
                    v-for="item in agents.available.value"
                    :key="item.id"
                    :value="item.id"
                >
                    <AgentLogo :agent="item.id" decorative />{{ item.name }}
                </SegmentedControlItem>
            </SegmentedControl>

            <PromptComposer
                :draft-key="draftKey(undefined, projectId)"
                :agent="agent"
                :placeholder="t('composer.placeholder.new', { agent: AGENTS[agent].name })"
                :busy="creating"
                autofocus
                @send="send"
                @switch-agent="switchAgent"
            >
                <template #start>
                    <Select v-model:value="projectChoice" :items="projectItems">
                        <SelectTrigger class="composer-select" :aria-label="t('composer.project')"
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

            <p class="m-0 text-muted">
                {{ t('empty.connectHint') }}
                <RouterLink to="/settings" class="underline underline-offset-4">{{
                    t('nav.settings').toLowerCase()
                }}</RouterLink>
            </p>
        </div>
    </ScrollPane>
</template>
