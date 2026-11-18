<script setup lang="ts">
import { computed, defineAsyncComponent, useTemplateRef } from 'vue'
import {
    PhChatCircle,
    PhStop,
    PhFile,
    PhGitBranch,
    PhTerminal,
    PhTerminalWindow,
} from '@phosphor-icons/vue'
import { Button } from '@ui-sdk/vue/button'
import { Tabs, TabsIndicator, TabsList, TabsPanel, TabsTab } from '@ui-sdk/vue/tabs'
import { toast } from '@ui-sdk/vue/toast'
import { useRoute, useRouter } from 'vue-router'

import RunSpinner from '@/components/RunSpinner.vue'
import TopBar from '@/components/TopBar.vue'
import ChatThread from '@/components/chat/ChatThread.vue'
import PromptComposer from '@/components/composer/PromptComposer.vue'
import SwitchFamilyDialog from '@/components/composer/SwitchFamilyDialog.vue'
import PanePlaceholder from '@/components/PanePlaceholder.vue'
import { useComposerSettings } from '@/composables/useComposerSettings'
import { draftKey } from '@/composables/useDrafts'
import { useCreateMessage, useStopSession, useUpdateSession } from '@/gen'
import { errorMessage } from '@/shared/errors'
import { AGENTS } from '@/shared/agents'
import type { AgentId } from '@/shared/agents'
import { useI18n } from 'vue-i18n'
import { useMessages, useProjects, useSession } from '@/shared/queries'

const FilesPane = defineAsyncComponent(() => import('@/components/files/FilesPane.vue'))

const TABS = [
    { id: 'chat', icon: PhChatCircle },
    { id: 'cli', icon: PhTerminal },
    { id: 'files', icon: PhFile },
    { id: 'git', icon: PhGitBranch },
    { id: 'terminal', icon: PhTerminalWindow },
] as const

const { t } = useI18n()
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
const { messages, hasNextPage, isFetchingNextPage, fetchNextPage } = useMessages(sessionId)

const agent = computed(() => session.data.value?.agent ?? 'claude')
const running = computed(() => session.data.value?.status === 'running')
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
    settings.setModel(next, model)
    update.mutate({ path: { sessionId: sessionId.value }, body: { agent: next } })
}

function announceUnsupportedAttachments(files: File[]) {
    if (files.length)
        toast.info(t('tabs.soon'), { description: files.map((file) => file.name).join(', ') })
}

function send({ text, files }: { text: string; files: File[] }) {
    announceUnsupportedAttachments(files)
    if (!text) return
    post.mutate({
        path: { sessionId: sessionId.value },
        body: { text, model: settings.modelFor(agent.value), mode: settings.mode() },
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
                    <span class="shrink-0 whitespace-nowrap text-muted">{{ project.name }}</span>
                    <span class="text-muted">/</span>
                </template>
                <span class="truncate font-medium">{{ session.data.value?.title }}</span>
            </template>
            <template #center>
                <TabsList :aria-label="t('tabs.label')">
                    <TabsTab v-for="item in TABS" :key="item.id" :value="item.id">
                        <component :is="item.icon" />{{ t(`tabs.${item.id}`) }}
                    </TabsTab>
                    <TabsIndicator />
                </TabsList>
            </template>
            <template #end>
                <span v-if="running" class="status-tag motion-fade">
                    <RunSpinner />{{ AGENTS[agent].name }}
                </span>
            </template>
        </TopBar>

        <TabsPanel value="chat" class="session-pane">
            <ChatThread
                :messages="messages"
                :agent="agent"
                :has-older="!!hasNextPage"
                :loading-older="isFetchingNextPage"
                @load-older="fetchNextPage()"
            />
            <div class="flex-none px-6 pb-4 max-md:px-3 max-md:pb-3">
                <PromptComposer
                    :draft-key="draftKey(sessionId)"
                    :agent="agent"
                    :history="history"
                    :placeholder="t('composer.placeholder.reply', { agent: AGENTS[agent].name })"
                    :busy="post.isPending.value || update.isPending.value"
                    autofocus
                    @send="send"
                    @switch-agent="switchAgent"
                >
                    <template v-if="running" #action>
                        <Button
                            data-variant="solid"
                            :disabled="stop.isPending.value"
                            :aria-label="t('composer.stop')"
                            :title="t('composer.stop')"
                            @click="stop.mutate({ path: { sessionId } })"
                        >
                            <PhStop weight="fill" />
                        </Button>
                    </template>
                </PromptComposer>
                <SwitchFamilyDialog ref="switchDialog" />
            </div>
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
        <TabsPanel value="git" class="session-pane"
            ><PanePlaceholder :icon="PhGitBranch"
        /></TabsPanel>
        <TabsPanel value="terminal" class="session-pane"
            ><PanePlaceholder :icon="PhTerminalWindow"
        /></TabsPanel>
    </Tabs>
</template>

<style scoped>
.session-pane {
    display: flex;
    min-height: 0;
    flex: 1;
    flex-direction: column;
    padding: 0;
}
</style>
