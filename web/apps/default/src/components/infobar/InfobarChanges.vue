<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

import type { GitChangeSummary } from '@drift/vue/api'

const props = defineProps<{ changes: GitChangeSummary }>()
const { t } = useI18n()

const total = computed(() => props.changes.additions + props.changes.deletions)
const addedShare = computed(() => (total.value ? props.changes.additions / total.value : 0))
const hiddenFiles = computed(() => props.changes.fileCount - props.changes.files.length)
</script>

<template>
    <section class="infobar-section">
        <h3 class="infobar-heading">
            <span>{{ t('title') }}</span>
            <span class="tabular-nums">{{ t('files', { count: changes.fileCount }) }}</span>
        </h3>
        <p v-if="!changes.fileCount" class="m-0 text-muted">{{ t('none') }}</p>
        <template v-else>
            <div class="flex gap-3 text-base font-medium tabular-nums">
                <span class="text-(--add-fg)">+{{ changes.additions }}</span>
                <span class="text-(--del-fg)">−{{ changes.deletions }}</span>
            </div>
            <div class="change-bar" aria-hidden="true">
                <span class="change-bar-fill" :style="{ scale: `${addedShare} 1` }" />
            </div>
            <ul class="m-0 flex list-none flex-col gap-1 p-0">
                <li v-for="file in changes.files" :key="file.path" class="file-row">
                    <span class="truncate" :title="file.path">{{ file.path }}</span>
                    <span class="text-muted tabular-nums"
                        >+{{ file.additions
                        }}<template v-if="file.deletions"> −{{ file.deletions }}</template></span
                    >
                </li>
            </ul>
            <p v-if="hiddenFiles > 0" class="m-0 text-muted">
                {{ t('more', { count: hiddenFiles }) }}
            </p>
        </template>
    </section>
</template>

<style scoped>
.change-bar {
    height: 4px;
    overflow: hidden;
    border-radius: 2px;
    background: var(--del-fg);
}

.change-bar-fill {
    display: block;
    height: 100%;
    background: var(--add-fg);
    transform-origin: left;
    transition: scale var(--dur-3) var(--ease-out);
}

.file-row {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    gap: 8px;
}
</style>

<i18n lang="yaml">
ru:
    title: Изменения
    files: 'файлов: {count}'
    none: С начала сессии код не менялся.
    more: 'и ещё файлов: {count}'
en:
    title: Changes
    files: '{count} files'
    none: No code changed since the session started.
    more: '{count} more files'
</i18n>
