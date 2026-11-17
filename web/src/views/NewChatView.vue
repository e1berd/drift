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

import AgentLogo from '@/components/AgentLogo.vue'
import BrandMark from '@/components/BrandMark.vue'
import TopBar from '@/components/TopBar.vue'
import PromptComposer from '@/components/composer/PromptComposer.vue'
import { draftKey } from '@/composables/useDrafts'
import { useAgents } from '@/composables/useAgents'
import { useCreateSession } from '@/gen'
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

const agent = ref<AgentId>(agents.connected.value.at(0) ?? 'claude')
const projectId = computed({
    get: () =>
        typeof route.query.project === 'string'
            ? route.query.project
            : (projects.value.at(0)?.id ?? ''),
    set: (id: string | null) => {
        if (id) void router.replace({ name: 'new', query: { project: id } })
    },
})
const projectItems = computed(() =>
    projects.value.map((project) => ({ value: project.id, label: project.name })),
)
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

const create = useCreateSession({
    mutation: {
        onSuccess: ({ session }) =>
            router.push({ name: 'session', params: { sessionId: session.id } }),
        onError: (error) => toast.error(errorMessage(error)),
    },
})

function announceUnsupportedAttachments(files: File[]) {
    if (files.length)
        toast.info(t('tabs.soon'), { description: files.map((file) => file.name).join(', ') })
}

function send({ text, files }: { text: string; files: File[] }) {
    announceUnsupportedAttachments(files)
    if (text)
        create.mutate({ path: { projectId: projectId.value }, body: { agent: agent.value, text } })
}
</script>

<template>
    <TopBar>
        <template #start><span class="truncate font-medium">{{ t('nav.newChat') }}</span></template>
    </TopBar>
    <section
        class="motion-fade flex min-h-0 flex-1 items-center justify-center overflow-auto p-6 max-md:p-3"
    >
        <div class="flex w-full max-w-[720px] flex-col items-center gap-5">
            <BrandMark :size="32" />
            <h1 class="m-0 text-2xl font-semibold">{{ t('empty.title') }}</h1>
            <p class="m-0 text-center text-muted">{{ t('empty.lead') }}</p>

            <SegmentedControl v-model:value="agentValue" :aria-label="t('composer.model')">
                <SegmentedControlIndicator />
                <SegmentedControlItem
                    v-for="item in agents.all"
                    :key="item.id"
                    :value="item.id"
                    :disabled="!agents.isConnected(item.id)"
                    :title="agents.isConnected(item.id) ? undefined : t('empty.notConnected')"
                >
                    <AgentLogo :agent="item.id" decorative />{{ item.name }}
                </SegmentedControlItem>
            </SegmentedControl>

            <p v-if="!projects.length" class="m-0 text-muted">{{ t('empty.noProject') }}</p>
            <PromptComposer
                v-else
                :draft-key="draftKey(undefined, projectId)"
                :agent="agent"
                :placeholder="t('composer.placeholder.new', { agent: AGENTS[agent].name })"
                :busy="create.isPending.value"
                @send="send"
            >
                <template #start>
                    <Select v-model:value="projectId" :items="projectItems">
                        <SelectTrigger :aria-label="t('composer.project')"
                            ><SelectValue
                        /></SelectTrigger>
                        <SelectContent side="top" align="start">
                            <SelectItem
                                v-for="project in projects"
                                :key="project.id"
                                :value="project.id"
                            >
                                <SelectItemText>{{ project.name }}</SelectItemText>
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
    </section>
</template>
