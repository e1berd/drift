<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

import { parseDiff } from '@/components/git/parseDiff'
import { useSessionGit } from '@/composables/useSessionGit'
import { useGetGitDiff } from '@/gen'
import { errorMessage } from '@/shared/errors'

const { t } = useI18n()
const git = useSessionGit()
const selected = computed(() => git.selected.value)

const diff = useGetGitDiff(
    {
        path: () => ({ sessionId: git.sessionId.value }),
        query: () => ({ path: selected.value?.path ?? '', staged: selected.value?.staged }),
    },
    {
        query: {
            enabled: () => !!selected.value,
            staleTime: 0,
            refetchOnWindowFocus: 'always',
        },
    },
)
const lines = computed(() => parseDiff(diff.data.value?.diff ?? ''))
const binary = computed(() => lines.value.some((line) => line.text.startsWith('Binary files ')))
</script>

<template>
    <section class="flex min-h-0 min-w-0 flex-col">
        <div class="diff-head">
            <b class="truncate font-mono font-medium text-fg">{{ selected?.path ?? t('pick') }}</b>
            <span v-if="selected" class="shrink-0">
                {{ selected.staged ? t('staged') : t('worktree') }}
            </span>
        </div>
        <div v-if="!selected" class="grid flex-1 place-items-center p-6 text-center text-muted">
            {{ t('pick') }}
        </div>
        <div v-else-if="diff.isPending.value" class="flex flex-col gap-1.5 p-4">
            <div
                v-for="i in 8"
                :key="i"
                class="skeleton h-3.5"
                :style="{ width: `${40 + ((i * 37) % 50)}%` }"
            />
        </div>
        <div v-else-if="diff.isError.value" class="p-4 text-danger">
            {{ errorMessage(diff.error.value) }}
        </div>
        <div v-else-if="!lines.length" class="grid flex-1 place-items-center text-muted">
            {{ t('empty') }}
        </div>
        <div v-else class="diff-body">
            <p v-if="binary" class="m-0 px-4 py-2 text-muted">{{ t('binary') }}</p>
            <div class="diff-lines">
                <div
                    v-for="(line, index) in lines"
                    :key="index"
                    class="diff-line"
                    :data-line="line.kind"
                >
                    <span class="diff-number">{{ line.oldLine ?? '' }}</span>
                    <span class="diff-number">{{ line.newLine ?? '' }}</span>
                    <span class="diff-text">{{ line.text || ' ' }}</span>
                </div>
            </div>
            <p v-if="diff.data.value?.truncated" class="m-0 px-4 py-2 text-muted">
                {{ t('truncated') }}
            </p>
        </div>
    </section>
</template>

<style scoped>
.diff-head {
    display: flex;
    height: 40px;
    flex: none;
    align-items: center;
    gap: 8px;
    padding: 0 14px;
    border-bottom: 1px solid var(--ui-border);
    color: var(--ui-muted-foreground);
    font-size: 12.5px;
}

.diff-body {
    min-height: 0;
    flex: 1;
    overflow: auto;
}

.diff-lines {
    display: grid;
    min-width: max-content;
    padding-block: 4px;
    font-family: var(--ui-font-mono);
    font-size: 12.5px;
    line-height: 1.55;
}

.diff-line {
    display: grid;
    grid-template-columns: 4.5ch 4.5ch auto;

    &[data-line='add'] {
        background: var(--add-bg);
    }

    &[data-line='del'] {
        background: var(--del-bg);
    }

    &[data-line='hunk'],
    &[data-line='meta'] {
        color: var(--ui-muted-foreground);
    }

    &[data-line='hunk'] {
        background: var(--ui-muted);
    }
}

.diff-number {
    padding-inline-end: 1ch;
    color: var(--ui-muted-foreground);
    text-align: end;
    user-select: none;
}

.diff-text {
    padding-inline: 1ch 2ch;
    white-space: pre;
    tab-size: 4;
}
</style>

<i18n lang="yaml">
ru:
    pick: Выберите файл, чтобы увидеть изменения
    staged: подготовлено
    worktree: рабочая копия
    empty: Изменений нет
    binary: Двоичный файл — построчное сравнение недоступно.
    truncated: Изменения обрезаны на 512 КБ.
en:
    pick: Pick a file to see its changes
    staged: staged
    worktree: working tree
    empty: No changes
    binary: Binary file; no line-by-line diff.
    truncated: The diff is cut at 512 KB.
</i18n>
