<script setup lang="ts">
import { computed, defineAsyncComponent, useTemplateRef } from 'vue'
import {
    PhChatCircle,
    PhFile,
    PhGitBranch,
    PhInfo,
    PhTerminal,
    PhTerminalWindow,
} from '@phosphor-icons/vue'
import { Button } from '@ui-sdk/vue/button'
import { Tabs, TabsIndicator, TabsList, TabsPanel, TabsTab } from '@ui-sdk/vue/tabs'
import { toast } from '@ui-sdk/vue/toast'
import { useRoute, useRouter } from 'vue-router'

import RunSpinner from '@/components/RunSpinner.vue'
import TopBar from '@/components/TopBar.vue'
import AgentActivity from '@/components/chat/AgentActivity.vue'
import ChatThread from '@/components/chat/ChatThread.vue'
import PromptComposer from '@/components/composer/PromptComposer.vue'
import SwitchFamilyDialog from '@/components/composer/SwitchFamilyDialog.vue'
import PanePlaceholder from '@/components/PanePlaceholder.vue'
import SessionChatLayout from '@/components/infobar/SessionChatLayout.vue'
import { useInfobar } from '@drift/vue/composables/useInfobar'
import { useComposerSettings } from '@/composables/useComposerSettings'
import { draftKey } from '@drift/vue/composables/useDrafts'
import { useEventStream } from '@drift/vue/composables/useEventStream'
import { useCreateMessage, useStopSession, useUpdateSession } from '@drift/vue/api'
import type { SentPrompt } from '@drift/sdk/attachments'
import { errorMessage } from '@drift/sdk/errors'
import { AGENTS } from '@/shared/agents'
import type { AgentId } from '@/shared/agents'
import { useI18n } from 'vue-i18n'
import { provideSessionMessages } from '@drift/vue/composables/useMessageWindow'
import { useProjects, useSession } from '@drift/vue/queries'

const FilesPane = defineAsyncComponent(() => import('@/components/files/FilesPane.vue'))
const GitPane = defineAsyncComponent(() => import('@/components/git/GitPane.vue'))
const TerminalPane = defineAsyncComponent(() => import('@/components/terminal/TerminalPane.vue'))

const TABS = [
    { id: 'chat', icon: PhChatCircle },
    { id: 'cli', icon: PhTerminal },
    { id: 'files', icon: PhFile },
    { id: 'git', icon: PhGitBranch },
    { id: 'terminal', icon: PhTerminalWindow },
] as const

const { t } = useI18n()
const infobar = useInfobar()
const infobarLabel = computed(() => (infobar.open.value ? t('infobar.hide') : t('infobar.show')))
const route = useRoute()
const router = useRouter()

const sessionId = computed(() => String(route.params.sessionId ?? ''))
const tab = computed({
    get: () =>
        typeof route.params.tab === 'string' && route.params.tab ? route.params.tab : 'chat',
    set: (value: string | null) => {
        if (value)
            void router.replace({
                name: 'session',
                params: { sessionId: sessionId.value, tab: value === 'chat' ? undefined : value },
            })
    },
})

const session = useSession(sessionId)
const { projects } = useProjects()
const project = computed(() =>
    projects.value.find((item) => item.id === session.data.value?.projectId),
)
const sessionMessages = provideSessionMessages(sessionId)
const messages = sessionMessages.messages

const agent = computed(() => session.data.value?.agent ?? 'claude')
const running = computed(() => session.data.value?.status === 'running')
const { activities } = useEventStream()
const activity = computed(() => activities.value.get(sessionId.value))
const waitingForReset = computed(() =>
    messages.value.some(
        (message) => message.kind === 'limit' && message.body.status === 'scheduled',
    ),
)
const placeholder = computed(() => {
    const name = AGENTS[agent.value].name
    if (waitingForReset.value) return t('composer.placeholder.afterReset', { agent: name })
    if (running.value) return t('composer.placeholder.queue', { agent: name })
    return t('composer.placeholder.reply', { agent: name })
})
const runningTool = computed(() => {
    if (sessionMessages.state.value.hasNewer) return undefined
    const last = messages.value.at(-1)
    return last?.kind === 'tool' && last.body.status === 'running' ? last.body.name : undefined
})
const history = computed(() =>
    messages.value.flatMap((message) =>
        message.role === 'user' && message.body.text ? [message.body.text] : [],
    ),
)

const post = useCreateMessage({
    mutation: { onError: (error) => toast.error(errorMessage(error)) },
})

const settings = useComposerSettings()
const switchDialog = useTemplateRef<InstanceType<typeof SwitchFamilyDialog>>('switchDialog')
const update = useUpdateSession({
    mutation: { onError: (error) => toast.error(errorMessage(error)) },
})

async function switchAgent({ agent: next, model }: { agent: AgentId; model: string }) {
    if (!(await switchDialog.value?.ask(next))) return
    settings.setModel(next, model, sessionId.value)
    update.mutate({ path: { sessionId: sessionId.value }, body: { agent: next } })
}

async function send({ text, attachments }: SentPrompt) {
    const attachmentIds = attachments.map((attachment) => attachment.id)
    if (!text && !attachmentIds.length) return
    await post.mutateAsync({
        path: { sessionId: sessionId.value },
        body: {
            text,
            attachmentIds,
            model: settings.modelFor(agent.value, sessionId.value),
            mode: settings.modeFor(agent.value),
        },
    })
}

const stop = useStopSession({
    mutation: { onError: (error) => toast.error(errorMessage(error)) },
})
</script>

<template>
    <Tabs v-model:value="tab" class="flex min-h-0 flex-1 flex-col">
        <TopBar>
            <template #start>
                <template v-if="project">
                    <span class="truncate text-muted">{{ project.name }}</span>
                    <span class="shrink-0 text-muted">/</span>
                </template>
                <span class="truncate font-medium">{{ session.data.value?.title }}</span>
            </template>
            <template #center>
                <TabsList :aria-label="t('tabs.label')" class="session-tabs">
                    <TabsTab
                        v-for="item in TABS"
                        :key="item.id"
                        :value="item.id"
                        v-tooltip="t(`tabs.${item.id}`)"
                    >
                        <component :is="item.icon" />
                        <span class="session-tab-label">{{ t(`tabs.${item.id}`) }}</span>
                    </TabsTab>
                    <TabsIndicator />
                </TabsList>
            </template>
            <template #end>
                <span v-if="running" class="status-tag motion-fade" v-tooltip="AGENTS[agent].name">
                    <RunSpinner /><span class="max-md:sr-only">{{ AGENTS[agent].name }}</span>
                </span>
                <Button
                    v-if="tab === 'chat'"
                    :aria-label="infobarLabel"
                    :aria-pressed="infobar.narrow.value ? undefined : infobar.open.value"
                    :aria-expanded="infobar.narrow.value ? infobar.open.value : undefined"
                    :title="infobarLabel"
                    @click="infobar.toggle()"
                >
                    <PhInfo />
                </Button>
            </template>
        </TopBar>

        <TabsPanel value="chat" class="session-pane">
            <SessionChatLayout :session-id="sessionId" :messages="messages" :running="running">
                <ChatThread :key="sessionId" :agent="agent">
                    <template #after>
                        <AgentActivity
                            v-if="running && session.data.value"
                            :agent="agent"
                            :activity="activity"
                            :running-tool="runningTool"
                            :since="session.data.value.updateTime"
                        />
                    </template>
                </ChatThread>
                <div class="flex-none px-6 pb-4 max-md:px-3 max-md:pb-3">
                    <PromptComposer
                        :draft-key="draftKey(sessionId)"
                        :agent="agent"
                        :session-id="sessionId"
                        :history="history"
                        :placeholder="placeholder"
                        :busy="post.isPending.value || update.isPending.value"
                        :running="running"
                        :stopping="stop.isPending.value"
                        autofocus
                        :submit="send"
                        @stop="stop.mutate({ path: { sessionId } })"
                        @switch-agent="switchAgent"
                    />
                    <SwitchFamilyDialog ref="switchDialog" />
                </div>
            </SessionChatLayout>
        </TabsPanel>
        <TabsPanel value="cli" class="session-pane"
            ><PanePlaceholder :icon="PhTerminal"
        /></TabsPanel>
        <TabsPanel value="files" class="session-pane">
            <FilesPane
                v-if="session.data.value"
                :session-id="sessionId"
                :title="project?.name ?? t('files.noProjectFolder')"
                :hint="project?.path"
            />
        </TabsPanel>
        <TabsPanel value="git" class="session-pane">
            <GitPane v-if="session.data.value" :session-id="sessionId" />
        </TabsPanel>
        <TabsPanel value="terminal" class="session-pane">
            <TerminalPane v-if="session.data.value" :session-id="sessionId" />
        </TabsPanel>
    </Tabs>
</template>

<style scoped>
.session-tabs {
    @media (max-width: 760px) {
        & > [data-slot='tabs-tab'] {
            padding-inline: calc(var(--ui-spacing) * 3);
        }

        & .session-tab-label {
            position: absolute;
            width: 1px;
            height: 1px;
            overflow: hidden;
            clip-path: inset(50%);
            white-space: nowrap;
        }
    }
}

.session-pane {
    display: flex;
    min-height: 0;
    flex: 1;
    flex-direction: column;
    padding: 0;
}
</style>
