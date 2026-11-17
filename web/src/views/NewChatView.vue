<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { PhCaretDown, PhCheck, PhFolderSimple } from '@phosphor-icons/vue'
import {
    Select,
    SelectContent,
    SelectItem,
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
import { useI18n } from '@/shared/i18n'
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
const projectItems = computed(() => projects.value.map((p) => ({ value: p.id, label: p.name })))

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
        <div class="flex w-full max-w-[720px] flex-col gap-[18px]">
            <div class="flex flex-col items-center gap-2.5 text-center">
                <span class="grid h-11 w-11 place-items-center rounded-xl bg-accent-soft"
                    ><BrandMark :size="24"
                /></span>
                <h1 class="m-0 text-[28px] leading-tight font-semibold tracking-tight">
                    {{ t('empty.title') }}
                </h1>
                <p class="m-0 max-w-[460px] leading-normal text-muted">{{ t('empty.lead') }}</p>
            </div>

            <div
                class="flex flex-wrap justify-center gap-2"
                role="group"
                :aria-label="t('composer.model')"
            >
                <button
                    v-for="a in agents.all"
                    :key="a.id"
                    type="button"
                    class="agent"
                    :data-on="agent === a.id || undefined"
                    :disabled="!agents.isConnected(a.id)"
                    :aria-pressed="agent === a.id"
                    @click="agent = a.id"
                >
                    <AgentLogo :agent="a.id" decorative />
                    {{ a.name }}
                    <small
                        v-if="!agents.isConnected(a.id)"
                        class="text-[11.5px] font-normal text-faint"
                        >{{ t('empty.notConnected') }}</small
                    >
                </button>
            </div>

            <p v-if="!projects.length" class="m-0 text-center text-[13px] text-faint">
                {{ t('empty.noProject') }}
            </p>
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
                        <SelectTrigger class="chip" :aria-label="t('composer.project')">
                            <PhFolderSimple :size="15" />
                            <SelectValue />
                            <PhCaretDown :size="13" />
                        </SelectTrigger>
                        <SelectContent class="popup w-[240px]" side="top" align="start">
                            <div class="popup-label">{{ t('composer.project') }}</div>
                            <SelectItem
                                v-for="p in projects"
                                :key="p.id"
                                :value="p.id"
                                class="option"
                            >
                                <PhCheck :size="15" class="check" />
                                <b><SelectItemText>{{ p.name }}</SelectItemText></b>
                                <span />
                            </SelectItem>
                        </SelectContent>
                    </Select>
                </template>
            </PromptComposer>

            <p class="m-0 text-center text-[12.5px] text-faint">
                {{ t('empty.connectHint') }}
                <RouterLink to="/settings" class="text-accent underline-offset-4 hover:underline">{{
                    t('nav.settings').toLowerCase()
                }}</RouterLink>
            </p>
        </div>
    </section>
</template>

<style scoped>
.agent {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    height: 38px;
    padding: 0 14px 0 10px;
    border: 1px solid var(--border-strong);
    border-radius: 999px;
    background: var(--surface);
    font-size: 13px;
    font-weight: 500;
    cursor: pointer;
    transition:
        border-color var(--dur-2),
        background var(--dur-2),
        transform var(--dur-1);

    &:hover {
        border-color: var(--faint);
    }

    &:active {
        transform: scale(0.97);
    }

    &[data-on] {
        border-color: var(--accent);
        background: var(--accent-soft);
    }

    &:disabled {
        border-style: dashed;
        opacity: 0.5;
        cursor: not-allowed;
    }
}
</style>
