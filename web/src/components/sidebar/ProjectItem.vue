<script setup lang="ts">
import { computed, nextTick, ref, useTemplateRef, watch } from 'vue'
import { PhCaretRight, PhFolderSimple, PhPencilSimple, PhTrash } from '@phosphor-icons/vue'
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
import { useI18n } from '@/shared/i18n'
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
const renameInput = useTemplateRef<HTMLInputElement>('renameInput')

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
    renameInput.value?.select()
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
    (f) => {
        if (f && !open.value && props.project.name.toLowerCase().includes(f)) open.value = true
    },
)
</script>

<template>
    <Collapsible v-show="!hidden" v-model:open="open" class="project">
        <div v-if="mode === 'confirm'" class="confirm motion-fade">
            <span class="min-w-0 flex-1">{{
                t('project.confirmDelete', { name: project.name, count: project.sessionCount })
            }}</span>
            <button class="btn h-[26px] border-0 bg-transparent px-2" @click="mode = 'view'">
                {{ t('project.cancel') }}
            </button>
            <button
                class="btn danger h-[26px] px-2"
                :disabled="remove.isPending.value"
                @click="remove.mutate({ path: { projectId: project.id } })"
            >
                {{ t('project.delete') }}
            </button>
        </div>

        <div v-else-if="mode === 'rename'" class="flex h-[34px] items-center pr-1.5 pl-[30px]">
            <label class="sr-only" :for="`rename-${project.id}`">{{ t('project.name') }}</label>
            <input
                :id="`rename-${project.id}`"
                ref="renameInput"
                v-model="name"
                class="field h-7"
                maxlength="120"
                @keydown.enter.prevent="commitRename"
                @keydown.esc.prevent="mode = 'view'"
                @blur="commitRename"
            />
        </div>

        <div v-else class="row">
            <CollapsibleTrigger class="row-btn">
                <PhCaretRight :size="14" class="chev" />
                <PhFolderSimple :size="15" class="text-muted" />
                <span class="min-w-0 flex-1 truncate">{{ project.name }}</span>
                <span class="count">{{ project.sessionCount }}</span>
            </CollapsibleTrigger>
            <div class="actions">
                <button
                    class="icon-btn small"
                    :aria-label="t('project.rename')"
                    :title="t('project.rename')"
                    @click="startRename"
                >
                    <PhPencilSimple :size="14" />
                </button>
                <button
                    class="icon-btn small"
                    :aria-label="t('project.delete')"
                    :title="t('project.delete')"
                    @click="mode = 'confirm'"
                >
                    <PhTrash :size="14" />
                </button>
            </div>
        </div>

        <CollapsiblePanel class="panel">
            <div v-auto-animate class="sessions">
                <DraftItem
                    v-if="showDraft"
                    :key="newKey"
                    :draft-key="newKey"
                    :project-id="project.id"
                />
                <SessionItem v-for="s in visible" :key="s.id" :session="s" />
                <template v-if="isPending">
                    <div
                        v-for="i in 3"
                        :key="`sk${i}`"
                        class="skeleton my-px h-[30px]"
                        :style="{ opacity: 1 - i * 0.25 }"
                    />
                </template>
                <p
                    v-else-if="!sessions.length && !showDraft"
                    key="empty"
                    class="px-2.5 pt-1.5 pb-2 text-[12.5px] text-faint"
                >
                    {{ t('nav.noSessions') }}
                </p>
                <template v-if="isFetchingNextPage">
                    <div v-for="i in 2" :key="`nx${i}`" class="skeleton my-px h-[30px]" />
                </template>
                <button v-else-if="hasNextPage" key="more" class="more" @click="fetchNextPage()">
                    {{ t('nav.showMore') }}
                </button>
                <div ref="sentinel" key="sentinel" class="h-px" />
            </div>
        </CollapsiblePanel>
    </Collapsible>
</template>

<style scoped>
.row {
    position: relative;
    display: flex;
    align-items: center;
    border-radius: 8px;
    transition: background var(--dur-1);

    &:hover {
        background: var(--surface-2);
    }
}

.row-btn {
    display: flex;
    min-width: 0;
    flex: 1;
    align-items: center;
    gap: 8px;
    height: 34px;
    padding: 0 8px;
    border: 0;
    border-radius: 8px;
    background: none;
    color: inherit;
    font-size: 13px;
    font-weight: 500;
    text-align: left;
    cursor: pointer;
}

.chev {
    color: var(--faint);
    transition: transform var(--dur-3) var(--ease-out);
}

.row-btn[data-open] .chev {
    transform: rotate(90deg);
}

.count {
    color: var(--faint);
    font-size: 11.5px;
    font-variant-numeric: tabular-nums;
    transition: opacity var(--dur-1);
}

.actions {
    position: absolute;
    top: 3px;
    right: 4px;
    display: flex;
    gap: 1px;
    opacity: 0;
    transition: opacity var(--dur-1);
}

.row:is(:hover, :focus-within) .actions {
    opacity: 1;
}

.row:is(:hover, :focus-within) .count {
    opacity: 0;
}

.confirm {
    display: flex;
    align-items: center;
    gap: 6px;
    min-height: 34px;
    padding: 4px 4px 4px 10px;
    border-radius: 8px;
    background: var(--del-bg);
    font-size: 12.5px;
}

.panel {
    height: var(--collapsible-panel-height);
    overflow: hidden;
    transition: height var(--dur-3) var(--ease-out);

    &:is([data-starting-style], [data-ending-style]) {
        height: 0;
    }

    &[data-ending-style] {
        transition-timing-function: var(--ease-in-out);
    }
}

.sessions {
    position: relative;
    display: flex;
    flex-direction: column;
    gap: 1px;
    padding: 2px 0 6px 20px;

    &::before {
        content: '';
        position: absolute;
        top: 2px;
        bottom: 8px;
        left: 15px;
        width: 1px;
        background: var(--ui-border);
    }
}

.more {
    height: 28px;
    padding: 0 10px;
    border: 0;
    border-radius: 7px;
    background: none;
    color: var(--faint);
    font-size: 12.5px;
    text-align: left;
    cursor: pointer;

    &:hover {
        background: var(--surface-2);
        color: var(--ui-foreground);
    }
}
</style>
