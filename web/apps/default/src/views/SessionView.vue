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
import { draftKey } from '@drift/vue/composables/useDrafts'
import { AGENTS } from '@drift/sdk/agents'
import type { AgentChoice } from '@drift/vue/chat/useNewChat'
import { useI18n } from 'vue-i18n'
import { useSessionChat } from '@drift/vue/chat/useSessionChat'

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

const {
    session,
    project,
    messages,
    agent,
    running,
    activity,
    replyState,
    runningTool,
    history,
    busy,
    stopping,
    switchAgent: applyAgentSwitch,
    send,
    stop,
} = useSessionChat(sessionId)
const placeholder = computed(() =>
    t(`composer.placeholder.${replyState.value}`, { agent: AGENTS[agent.value].name }),
)

const switchDialog = useTemplateRef<InstanceType<typeof SwitchFamilyDialog>>('switchDialog')

async function switchAgent(choice: AgentChoice) {
    if (await switchDialog.value?.ask(choice.agent)) applyAgentSwitch(choice)
}
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
                        :busy="busy"
                        :running="running"
                        :stopping="stopping"
                        autofocus
                        :submit="send"
                        @stop="stop"
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
