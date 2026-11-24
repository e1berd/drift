<script setup lang="ts">
import { computed, useTemplateRef } from 'vue'
import { PhChatsCircle, PhFolderSimple } from '@phosphor-icons/vue'
import { Button } from '@ui-sdk/vue/button'
import { useI18n } from 'vue-i18n'

import type { Session } from '@/gen'

import SessionItem from '@/components/sidebar/SessionItem.vue'
import { useInfiniteSentinel } from '@/composables/useInfiniteSentinel'
import { useSidebarTab } from '@/composables/useSidebarTab'
import { errorMessage } from '@/shared/errors'
import { useProjects, useSessionSearch } from '@/shared/queries'

interface ResultGroup {
    key: string
    projectId?: string
    sessions: Session[]
}

const props = defineProps<{ text: string }>()
const { t } = useI18n()
const { projects } = useProjects()
const {
    sessions,
    isPending,
    isPlaceholderData,
    error,
    hasNextPage,
    isFetchingNextPage,
    fetchNextPage,
} = useSessionSearch(() => props.text)
const { revealSessionPlace } = useSidebarTab()

const projectNames = computed(
    () => new Map(projects.value.map((project) => [project.id, project.name])),
)

const groups = computed(() => {
    const byProject = new Map<string, ResultGroup>()
    for (const session of sessions.value) {
        const key = session.projectId ?? 'chats'
        const group = byProject.get(key) ?? { key, projectId: session.projectId, sessions: [] }
        group.sessions.push(session)
        byProject.set(key, group)
    }
    return [...byProject.values()]
})

function groupTitle(group: ResultGroup) {
    if (!group.projectId) return t('chats')
    return projectNames.value.get(group.projectId) ?? t('project')
}

const sentinel = useTemplateRef<HTMLElement>('sentinel')
useInfiniteSentinel(sentinel, () => {
    if (hasNextPage.value && !isFetchingNextPage.value) void fetchNextPage()
})
</script>

<template>
    <section
        class="px-2 pt-2 pb-2"
        :aria-label="t('label')"
        :aria-busy="isPending || isPlaceholderData"
    >
        <div v-auto-animate class="flex flex-col">
            <template v-if="isPending">
                <div v-for="row in 4" :key="`skeleton-${row}`" class="skeleton my-px h-8" />
            </template>
            <p v-else-if="error" key="error" class="m-0 px-2 py-1 text-destructive">
                {{ errorMessage(error) }}
            </p>
            <p v-else-if="!groups.length" key="empty" class="m-0 px-2 py-1 text-muted">
                {{ t('empty', { text }) }}
            </p>
            <div v-for="group in groups" :key="group.key" class="pb-1">
                <div class="flex items-center gap-2 px-2 pt-2 pb-1 text-muted">
                    <PhFolderSimple v-if="group.projectId" class="shrink-0" />
                    <PhChatsCircle v-else class="shrink-0" />
                    <span class="truncate">{{ groupTitle(group) }}</span>
                </div>
                <div
                    v-auto-animate
                    class="flex flex-col gap-px"
                    @click="revealSessionPlace(group.projectId)"
                >
                    <SessionItem
                        v-for="session in group.sessions"
                        :key="session.id"
                        :session="session"
                    />
                </div>
            </div>
            <div v-if="isFetchingNextPage" key="next" class="skeleton my-px h-8" />
            <Button v-else-if="hasNextPage" key="more" @click="fetchNextPage()">
                {{ t('nav.showMore') }}
            </Button>
            <div ref="sentinel" key="sentinel" class="h-px" />
        </div>
    </section>
</template>

<i18n lang="yaml">
ru:
    label: 'Результаты поиска'
    chats: 'Чаты'
    project: 'Проект'
    empty: 'Нет сессий с «{text}» в названии'
en:
    label: 'Search results'
    chats: 'Chats'
    project: 'Project'
    empty: 'No sessions with “{text}” in the title'
</i18n>
