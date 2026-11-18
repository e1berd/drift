<script setup lang="ts">
import { PhArrowSquareOut } from '@phosphor-icons/vue'
import { useI18n } from 'vue-i18n'

import type { Release } from '@/gen'

import MarkdownText from '@/components/chat/MarkdownText.vue'

defineProps<{ release: Release }>()
const { t } = useI18n()
</script>

<template>
    <section class="flex min-h-0 flex-col gap-2">
        <div class="flex items-center justify-between gap-2">
            <span class="font-medium">{{ t('title', { version: release.version }) }}</span>
            <a :href="release.url" target="_blank" rel="noopener noreferrer" class="release-link">
                {{ t('open') }}<PhArrowSquareOut />
            </a>
        </div>
        <div class="release-notes">
            <MarkdownText v-if="release.notes.trim()" :source="release.notes" />
            <p v-else class="m-0 text-muted">{{ t('empty') }}</p>
        </div>
    </section>
</template>

<style scoped>
.release-link {
    display: inline-flex;
    align-items: center;
    gap: 0.25rem;
}

.release-notes {
    max-height: min(40dvh, 22rem);
    overflow: auto;
    overscroll-behavior: contain;
    padding: 0.5rem 0.75rem;
    border: 1px solid var(--ui-border);
    border-radius: var(--ui-radius);
}
</style>

<i18n lang="yaml">
ru:
    title: 'Что нового в {version}'
    open: Release notes на GitHub
    empty: Описание релиза пустое — подробности на GitHub.
en:
    title: "What's new in {version}"
    open: Release notes on GitHub
    empty: The release has no notes here; see GitHub for details.
</i18n>
