<script setup lang="ts">
import { computed, nextTick, ref, useTemplateRef, watch } from 'vue'
import { PhFolderSimple, PhPencilSimple, PhTrash } from '@phosphor-icons/vue'
import { Button } from '@ui-sdk/vue/button'
import { Input } from '@ui-sdk/vue/input'
import { Collapsible, CollapsiblePanel, CollapsibleTrigger } from '@ui-sdk/vue/collapsible'
import { toast } from '@ui-sdk/vue/toast'
import { useRoute, useRouter } from 'vue-router'

import type { Project } from '@/gen'

import SessionItem from '@/components/sidebar/SessionItem.vue'
import DraftItem from '@/components/sidebar/DraftItem.vue'
import { draftKey, useDraftIndex } from '@/composables/useDrafts'
import { useOpenProjects } from '@/composables/useOpenProjects'
import { useInfiniteSentinel } from '@/composables/useInfiniteSentinel'
import { useDeleteProject, useUpdateProject } from '@/gen'
import { errorMessage } from '@/shared/errors'
import { useI18n } from 'vue-i18n'
import { useSessions } from '@/shared/queries'

const props = defineProps<{ project: Project; filter: string; isDefault?: boolean }>()
const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const { isOpen, setOpen } = useOpenProjects()
const drafts = useDraftIndex()

const open = computed({
    get: () => isOpen(props.project.id),
    set: (value: boolean) => setOpen(props.project.id, value),
})
const { sessions, isPending, hasNextPage, isFetchingNextPage, fetchNextPage } = useSessions(
    () => props.project.id,
    open,
)

const newKey = computed(() => draftKey(undefined, props.project.id))
const viewingNew = computed(() => {
    if (route.name !== 'new') return false
    const requested = route.query.project
    return typeof requested === 'string' ? requested === props.project.id : !!props.isDefault
})
const showDraft = computed(() => drafts.has(newKey.value) && !viewingNew.value)

const visible = computed(() => {
    if (!props.filter) return sessions.value
    return sessions.value.filter((session) => session.title.toLowerCase().includes(props.filter))
})
const hidden = computed(
    () =>
        !!props.filter &&
        !props.project.name.toLowerCase().includes(props.filter) &&
        !visible.value.length,
)

const mode = ref<'view' | 'rename' | 'confirm'>('view')
const name = ref('')
const renameRow = useTemplateRef<HTMLElement>('renameRow')

const showError = (error: unknown) => toast.error(errorMessage(error))
const rename = useUpdateProject({ mutation: { onError: showError } })
const remove = useDeleteProject({
    mutation: {
        onSuccess: () => {
            if (route.name === 'new' && route.query.project === props.project.id)
                void router.replace('/')
        },
        onError: showError,
    },
})

async function startRename() {
    name.value = props.project.name
    mode.value = 'rename'
    await nextTick()
    renameRow.value?.querySelector('input')?.select()
}

function commitRename() {
    if (mode.value !== 'rename') return
    const value = name.value.trim()
    mode.value = 'view'
    if (value && value !== props.project.name)
        rename.mutate({ path: { projectId: props.project.id }, body: { name: value } })
}

const sentinel = useTemplateRef<HTMLElement>('sentinel')
useInfiniteSentinel(sentinel, () => {
    if (hasNextPage.value && !isFetchingNextPage.value) void fetchNextPage()
})

watch(
    () => props.filter,
    (value) => {
        if (value && !open.value && props.project.name.toLowerCase().includes(value))
            open.value = true
    },
)
</script>

<template>
    <Collapsible v-show="!hidden" v-model:open="open">
        <div v-if="mode === 'confirm'" class="motion-fade flex flex-col gap-2 p-2">
            <span>{{
                t('project.confirmDelete', { name: project.name, count: project.sessionCount })
            }}</span>
            <div class="flex justify-end gap-2">
                <Button @click="mode = 'view'">{{ t('project.cancel') }}</Button>
                <Button
                    data-variant="destructive"
                    :disabled="remove.isPending.value"
                    @click="remove.mutate({ path: { projectId: project.id } })"
                >
                    {{ t('project.delete') }}
                </Button>
            </div>
        </div>

        <div v-else-if="mode === 'rename'" ref="renameRow" class="p-1">
            <Input
                v-model="name"
                maxlength="120"
                :aria-label="t('project.name')"
                @keydown.enter.prevent="commitRename"
                @keydown.esc.prevent="mode = 'view'"
                @blur="commitRename"
            />
        </div>

        <div v-else class="project-row">
            <CollapsibleTrigger class="min-w-0 flex-1">
                <span class="flex min-w-0 flex-1 items-center gap-2">
                    <PhFolderSimple class="shrink-0" />
                    <span class="truncate">{{ project.name }}</span>
                    <span class="ml-auto text-muted tabular-nums">{{ project.sessionCount }}</span>
                </span>
            </CollapsibleTrigger>
            <div class="project-actions">
                <Button
                    :aria-label="t('project.rename')"
                    :title="t('project.rename')"
                    @click="startRename"
                >
                    <PhPencilSimple />
                </Button>
                <Button
                    :aria-label="t('project.delete')"
                    :title="t('project.delete')"
                    @click="mode = 'confirm'"
                >
                    <PhTrash />
                </Button>
            </div>
        </div>

        <CollapsiblePanel>
            <div v-auto-animate class="flex flex-col gap-px py-1 pl-4">
                <DraftItem
                    v-if="showDraft"
                    :key="newKey"
                    :draft-key="newKey"
                    :project-id="project.id"
                />
                <SessionItem v-for="session in visible" :key="session.id" :session="session" />
                <template v-if="isPending">
                    <div v-for="row in 3" :key="`skeleton-${row}`" class="skeleton my-px h-8" />
                </template>
                <p
                    v-else-if="!sessions.length && !showDraft"
                    key="empty"
                    class="px-2 py-1 text-muted"
                >
                    {{ t('nav.noSessions') }}
                </p>
                <div v-if="isFetchingNextPage" key="next" class="skeleton my-px h-8" />
                <Button v-else-if="hasNextPage" key="more" @click="fetchNextPage()">
                    {{ t('nav.showMore') }}
                </Button>
                <div ref="sentinel" key="sentinel" class="h-px" />
            </div>
        </CollapsiblePanel>
    </Collapsible>
</template>

<style scoped>
.project-row {
    display: flex;
    align-items: center;
    gap: 4px;
}

.project-actions {
    display: flex;
    gap: 2px;
    opacity: 0;
    transition: opacity var(--dur-1);
}

.project-row:is(:hover, :focus-within) .project-actions {
    opacity: 1;
}
</style>
