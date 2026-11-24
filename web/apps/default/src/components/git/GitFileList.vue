<script setup lang="ts">
import { PhArrowCounterClockwise, PhMinus, PhPlus } from '@phosphor-icons/vue'
import { Button } from '@ui-sdk/vue/button'
import { useI18n } from 'vue-i18n'

import type { GitFileChange } from '@drift/vue/api'

import { statusLetter } from '@drift/sdk/git/parseDiff'
import { useSessionGit } from '@drift/vue/composables/useSessionGit'

const props = defineProps<{ title: string; changes: GitFileChange[]; staged: boolean }>()
const emit = defineEmits<{ discard: [change: GitFileChange] }>()
const { t } = useI18n()
const git = useSessionGit()

function fileName(path: string): string {
    return path.split('/').at(-1) ?? path
}

function folder(path: string): string {
    return path.split('/').slice(0, -1).join('/')
}

function isSelected(change: GitFileChange): boolean {
    return git.selected.value?.path === change.path && git.selected.value.staged === props.staged
}

function toggleStage(change: GitFileChange) {
    if (props.staged) git.unstage(change.origPath ? [change.path, change.origPath] : [change.path])
    else git.stage([change.path])
}

function toggleAll() {
    if (props.staged) git.unstage([])
    else git.stage([])
}
</script>

<template>
    <section v-if="changes.length" class="flex flex-col">
        <div class="list-head">
            <b class="font-medium text-fg">{{ title }}</b>
            <span class="tabular-nums">{{ changes.length }}</span>
            <Button
                class="ml-auto"
                :disabled="git.busy.value"
                :aria-label="staged ? t('unstageAll') : t('stageAll')"
                v-tooltip="staged ? t('unstageAll') : t('stageAll')"
                @click="toggleAll"
            >
                <PhMinus v-if="staged" />
                <PhPlus v-else />
            </Button>
        </div>
        <ul class="m-0 list-none p-0">
            <li
                v-for="change in changes"
                :key="change.path"
                class="change-row"
                :data-active="isSelected(change) || undefined"
            >
                <button
                    type="button"
                    class="change-name"
                    v-tooltip="
                        change.origPath ? `${change.origPath} → ${change.path}` : change.path
                    "
                    @click="git.selected.value = { path: change.path, staged }"
                >
                    <span class="truncate text-fg">{{ fileName(change.path) }}</span>
                    <span class="truncate text-muted">{{ folder(change.path) }}</span>
                </button>
                <Button
                    v-if="!staged"
                    class="change-action"
                    :disabled="git.busy.value"
                    :aria-label="t('discard')"
                    v-tooltip="t('discard')"
                    @click="emit('discard', change)"
                >
                    <PhArrowCounterClockwise />
                </Button>
                <Button
                    class="change-action"
                    :disabled="git.busy.value"
                    :aria-label="staged ? t('unstage') : t('stage')"
                    v-tooltip="staged ? t('unstage') : t('stage')"
                    @click="toggleStage(change)"
                >
                    <PhMinus v-if="staged" />
                    <PhPlus v-else />
                </Button>
                <span
                    class="change-letter"
                    :data-status="change.status"
                    v-tooltip="t(`status.${change.status}`)"
                    >{{ statusLetter(change.status) }}</span
                >
            </li>
        </ul>
    </section>
</template>

<style scoped>
.list-head {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 6px 6px 2px 12px;
    color: var(--ui-muted-foreground);
    font-size: 12px;
}

.change-row {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto auto 18px;
    align-items: center;
    gap: 2px;
    min-height: 30px;
    padding-inline: 4px 10px;
    border-radius: var(--ui-radius);
    margin-inline: 4px;

    &:hover,
    &[data-active] {
        background: var(--ui-muted);
    }

    & .change-action {
        opacity: 0;
    }

    &:hover .change-action,
    &:focus-within .change-action,
    &[data-active] .change-action {
        opacity: 1;
    }

    @media (hover: none) {
        & .change-action {
            opacity: 1;
        }
    }
}

.change-name {
    display: flex;
    min-width: 0;
    align-items: baseline;
    gap: 6px;
    padding: 4px 6px;
    border: 0;
    background: none;
    color: inherit;
    font: inherit;
    text-align: start;
    cursor: pointer;

    & > :last-child {
        font-size: 11.5px;
    }
}

.change-letter {
    font-family: var(--ui-font-mono);
    font-size: 12px;
    text-align: center;
    color: var(--ui-muted-foreground);

    &[data-status='added'],
    &[data-status='untracked'] {
        color: light-dark(oklch(0.5 0.14 150), oklch(0.75 0.14 150));
    }

    &[data-status='deleted'],
    &[data-status='conflicted'] {
        color: var(--ui-destructive);
    }
}
</style>

<i18n lang="yaml">
ru:
    stage: Подготовить к коммиту
    unstage: Убрать из подготовленных
    stageAll: Подготовить все
    unstageAll: Убрать все из подготовленных
    discard: Отменить изменения
    status:
        modified: Изменён
        added: Добавлен
        deleted: Удалён
        renamed: Переименован
        copied: Скопирован
        typeChanged: Сменился тип
        conflicted: Конфликт слияния
        untracked: Не отслеживается
en:
    stage: Stage
    unstage: Unstage
    stageAll: Stage all
    unstageAll: Unstage all
    discard: Discard changes
    status:
        modified: Modified
        added: Added
        deleted: Deleted
        renamed: Renamed
        copied: Copied
        typeChanged: Type changed
        conflicted: Merge conflict
        untracked: Untracked
</i18n>
