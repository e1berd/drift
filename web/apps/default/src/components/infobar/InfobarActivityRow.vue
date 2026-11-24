<script setup lang="ts">
import { computed } from 'vue'
import {
    PhArrowSquareOut,
    PhFilePlus,
    PhGitBranch,
    PhTrash,
    PhTreeStructure,
} from '@phosphor-icons/vue'
import { useI18n } from 'vue-i18n'

import { useActivityActions } from '@drift/vue/composables/useActivityActions'
import { relativeTime } from '@drift/sdk/time'
import type { ActivityEntry } from '@drift/vue/api'

const props = defineProps<{ entry: ActivityEntry; active: boolean }>()
const { t, locale } = useI18n()
const actions = useActivityActions()

const ICONS = { branch: PhGitBranch, worktree: PhTreeStructure, file: PhFilePlus } as const

const title = computed(() =>
    props.entry.kind === 'file'
        ? (props.entry.name.split('/').at(-1) ?? props.entry.name)
        : props.entry.name,
)
const removable = computed(() => !props.entry.removed && props.entry.kind !== 'branch')

function branchMeta(entry: ActivityEntry): string[] {
    const parts = [t('commits', { count: entry.commits ?? 0 })]
    if (entry.current) parts.unshift(t('current'))
    parts.push(entry.upstream ? t('unpushed', { count: entry.unpushed ?? 0 }) : t('local'))
    return parts
}

function presentMeta(entry: ActivityEntry): string[] {
    if (entry.kind === 'branch') return branchMeta(entry)
    if (entry.kind === 'worktree')
        return [
            entry.branch ?? t('detached'),
            entry.changes ? t('changed', { count: entry.changes }) : t('clean'),
        ]
    const folder = entry.name.split('/').slice(0, -1).join('/')
    return [folder || t('sessionFolder')]
}

const meta = computed(() => {
    const entry = props.entry
    if (!entry.removed) return presentMeta(entry).join(' · ')
    if (entry.removedBy !== 'user') return t('removed')
    const when = entry.removeTime ? relativeTime(entry.removeTime, locale.value) : ''
    return [t('removedByYou'), when].filter(Boolean).join(' · ')
})
</script>

<template>
    <li
        class="activity-row"
        :data-link-target="entry.key"
        :data-link-source="entry.messageId"
        :data-active="active || undefined"
        :data-removed="entry.removed || undefined"
    >
        <component :is="ICONS[entry.kind]" class="activity-icon" aria-hidden="true" />
        <button
            type="button"
            class="activity-main"
            :title="entry.name"
            @click="actions.reveal(entry.messageId)"
        >
            <span class="activity-name truncate">{{ title }}</span>
            <span class="truncate text-muted">{{ meta }}</span>
        </button>
        <span class="activity-actions">
            <button
                v-if="actions.canOpen(entry)"
                type="button"
                class="row-action"
                :aria-label="t('open', { name: entry.name })"
                @click="actions.open(entry)"
            >
                <PhArrowSquareOut />
            </button>
            <button
                v-if="removable"
                type="button"
                class="row-action"
                :aria-label="t('remove', { name: entry.name })"
                @click="actions.remove(entry)"
            >
                <PhTrash />
            </button>
        </span>
    </li>
</template>

<style scoped>
.activity-row {
    display: grid;
    grid-template-columns: 16px minmax(0, 1fr) auto;
    align-items: start;
    gap: calc(var(--ui-spacing) * 2);
    padding: calc(var(--ui-spacing) * 1.5) calc(var(--ui-spacing) * 2);
    border: 1px solid transparent;
    border-radius: var(--ui-radius);
    transition:
        background-color var(--dur-2) var(--ease-out),
        border-color var(--dur-2) var(--ease-out);

    &:hover,
    &[data-active] {
        border-color: color-mix(in oklab, var(--ui-primary) 45%, transparent);
        background: color-mix(in oklab, var(--ui-primary) 10%, transparent);
    }

    &[data-removed] {
        color: var(--ui-muted-foreground);

        & .activity-name {
            text-decoration: line-through;
        }
    }

    @media (pointer: coarse) {
        min-height: 44px;
    }
}

.activity-icon {
    margin-top: 2px;
    color: var(--ui-muted-foreground);
}

.activity-main {
    display: grid;
    min-width: 0;
    gap: 2px;
    padding: 0;
    border: 0;
    background: none;
    color: inherit;
    font: inherit;
    text-align: start;
    cursor: pointer;

    &:focus-visible {
        border-radius: 2px;
        outline: 2px solid var(--ui-ring);
        outline-offset: 2px;
    }
}

.activity-name {
    font-weight: 500;
}

.activity-actions {
    display: flex;
    gap: 2px;
    opacity: 0;
    transition: opacity var(--dur-2) var(--ease-out);

    .activity-row:hover &,
    .activity-row:focus-within & {
        opacity: 1;
    }

    @media (pointer: coarse) {
        opacity: 1;

        & .row-action {
            width: 32px;
            height: 32px;
        }
    }
}
</style>

<i18n lang="yaml">
ru:
    current: текущая
    commits: 'коммитов: {count}'
    unpushed: 'не запушено: {count}'
    local: не запушена
    changed: 'изменений: {count}'
    clean: без изменений
    detached: detached HEAD
    sessionFolder: папка сессии
    removed: удалено
    removedByYou: удалено вами
    open: Открыть {name}
    remove: Удалить {name}
en:
    current: current
    commits: '{count} commits'
    unpushed: '{count} unpushed'
    local: not pushed
    changed: '{count} changes'
    clean: clean
    detached: detached HEAD
    sessionFolder: session folder
    removed: removed
    removedByYou: removed by you
    open: Open {name}
    remove: Remove {name}
</i18n>
