<script setup lang="ts">
import { computed, defineAsyncComponent } from 'vue'
import {
    PhChatCircle,
    PhFile,
    PhGitBranch,
    PhTerminal,
    PhTerminalWindow,
} from '@phosphor-icons/vue'
import { Tabs, TabsIndicator, TabsList, TabsPanel, TabsTab } from '@ui-sdk/vue/tabs'
import { toast } from '@ui-sdk/vue/toast'
import { useRoute, useRouter } from 'vue-router'

import RunSpinner from '@/components/RunSpinner.vue'
import TopBar from '@/components/TopBar.vue'
import ChatThread from '@/components/chat/ChatThread.vue'
import PromptComposer from '@/components/composer/PromptComposer.vue'
import PanePlaceholder from '@/components/PanePlaceholder.vue'
import { draftKey } from '@/composables/useDrafts'
import { useCreateMessage } from '@/gen'
import { errorMessage } from '@/shared/errors'
import { AGENTS } from '@/shared/agents'
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
const project = computed(() => projects.value.find((p) => p.id === session.data.value?.projectId))
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

function announceUnsupportedAttachments(files: File[]) {
    if (files.length)
        toast.info(t('tabs.soon'), { description: files.map((file) => file.name).join(', ') })
}

function send({ text, files }: { text: string; files: File[] }) {
    announceUnsupportedAttachments(files)
    if (text) post.mutate({ path: { sessionId: sessionId.value }, body: { text } })
}
</script>

<template>
    <Tabs v-model:value="tab" class="flex min-h-0 flex-1 flex-col">
        <TopBar>
            <template #start>
                <span class="shrink-0 whitespace-nowrap text-faint">{{ project?.name }}</span>
                <span class="text-line-strong">/</span>
                <span class="truncate font-medium">{{ session.data.value?.title }}</span>
            </template>
            <template #center>
                <TabsList class="tabs" :aria-label="t('tabs.label')">
                    <TabsIndicator class="tab-ind" />
                    <TabsTab v-for="x in TABS" :key="x.id" :value="x.id" class="tab">
                        <component :is="x.icon" :size="15" />{{ t(`tabs.${x.id}`) }}
                    </TabsTab>
                </TabsList>
            </template>
            <template #end>
                <span v-if="running" class="status motion-fade">
                    <RunSpinner :size="15" />
                    <b class="font-medium text-fg">{{ AGENTS[agent].name }}</b>
                </span>
            </template>
        </TopBar>

        <TabsPanel value="chat" class="pane">
            <ChatThread
                :messages="messages"
                :agent="agent"
                :has-older="!!hasNextPage"
                :loading-older="isFetchingNextPage"
                @load-older="fetchNextPage()"
            />
            <div class="flex-none px-6 pb-[18px] max-md:px-3 max-md:pb-3">
                <PromptComposer
                    :draft-key="draftKey(sessionId)"
                    :agent="agent"
                    :history="history"
                    :placeholder="t('composer.placeholder.reply', { agent: AGENTS[agent].name })"
                    :busy="post.isPending.value"
                    @send="send"
                />
            </div>
        </TabsPanel>
        <TabsPanel value="cli" class="pane"><PanePlaceholder :icon="PhTerminal" /></TabsPanel>
        <TabsPanel value="files" class="pane">
            <FilesPane v-if="project" :project="project" />
        </TabsPanel>
        <TabsPanel value="git" class="pane"><PanePlaceholder :icon="PhGitBranch" /></TabsPanel>
        <TabsPanel value="terminal" class="pane"
            ><PanePlaceholder :icon="PhTerminalWindow"
        /></TabsPanel>
    </Tabs>
</template>

<style scoped>
.tabs {
    position: relative;
    display: flex;
    padding: 3px;
    border: 1px solid var(--ui-border);
    border-radius: 10px;
    background: var(--surface-2);
    overflow-x: auto;
    scrollbar-width: none;
}

.tab-ind {
    position: absolute;
    top: var(--active-tab-top);
    left: var(--active-tab-left);
    width: var(--active-tab-width);
    height: var(--active-tab-height);
    border-radius: 7px;
    background: light-dark(var(--surface), var(--surface-3));
    box-shadow:
        0 1px 2px rgb(0 0 0 / 0.18),
        inset 0 0 0 1px var(--ui-border);
    transition:
        left var(--dur-3) var(--ease-out),
        width var(--dur-3) var(--ease-out);

    &[data-activation-direction='none'] {
        transition: none;
    }
}

.tab {
    position: relative;
    display: inline-flex;
    min-width: 84px;
    height: 30px;
    align-items: center;
    justify-content: center;
    gap: 6px;
    padding: 0 10px;
    border: 0;
    border-radius: 7px;
    background: none;
    color: var(--ui-muted-foreground);
    font-size: 12.5px;
    font-weight: 500;
    cursor: pointer;
    transition: color var(--dur-2);

    &:hover,
    &[data-selected] {
        color: var(--ui-foreground);
    }
}

.pane {
    display: flex;
    min-height: 0;
    flex: 1;
    flex-direction: column;
    animation: drift-fade var(--dur-3) var(--ease-out) both;
}

.status {
    display: inline-flex;
    height: 32px;
    align-items: center;
    gap: 8px;
    padding: 0 10px;
    border: 1px solid var(--ui-border);
    border-radius: 999px;
    background: var(--surface-2);
    font-size: 12.5px;
}

@media (max-width: 760px) {
    .tab {
        min-width: 0;
    }
}
</style>
