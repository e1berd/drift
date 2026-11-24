<script setup lang="ts">
import { useI18n } from 'vue-i18n'

import RunSpinner from '@/components/RunSpinner.vue'
import InfobarActivity from '@/components/infobar/InfobarActivity.vue'
import InfobarChanges from '@/components/infobar/InfobarChanges.vue'
import type { GitSummary } from '@drift/vue/api'

const { placement = 'docked' } = defineProps<{
    placement?: 'docked' | 'drawer'
    summary?: GitSummary
    loading: boolean
    failed: boolean
    running: boolean
    activeMessageId: string | null
}>()
const { t } = useI18n()
</script>

<template>
    <aside class="infobar" data-infobar :data-placement="placement" :aria-label="t('label')">
        <div class="infobar-inner">
            <header class="infobar-header">
                <span class="font-medium">{{ t('title') }}</span>
                <span v-if="running" class="status-tag"><RunSpinner />{{ t('running') }}</span>
                <span v-else-if="summary?.base" class="status-tag" :title="t('baseHint')">
                    {{ t('base', { commit: summary.base.slice(0, 7) }) }}
                </span>
            </header>
            <div v-if="loading" class="infobar-top flex flex-col gap-2">
                <div v-for="i in 4" :key="i" class="skeleton h-3.5" />
            </div>
            <p v-else-if="failed" class="infobar-top m-0 text-muted">{{ t('failed') }}</p>
            <template v-else-if="summary">
                <div class="infobar-top">
                    <InfobarChanges v-if="summary.repository" :changes="summary.changes" />
                    <p v-else class="m-0 text-muted">{{ t('notRepository') }}</p>
                </div>
                <InfobarActivity
                    :activity="summary.activity"
                    :active-message-id="activeMessageId"
                />
            </template>
        </div>
    </aside>
</template>

<i18n lang="yaml">
ru:
    label: Статус сессии
    title: Статус сессии
    running: работает
    base: от {commit}
    baseHint: Коммит, с которого началась сессия; изменения считаются от него
    failed: Не удалось получить состояние сессии.
    notRepository: Папка сессии не git-репозиторий, поэтому изменения кода не считаются.
en:
    label: Session status
    title: Session status
    running: running
    base: from {commit}
    baseHint: The commit the session started at; changes are counted from it
    failed: Could not read the session state.
    notRepository: The session folder is not a git repository, so code changes are not counted.
</i18n>

<style scoped>
.infobar {
    min-width: 0;
    overflow: hidden;
    border-inline-end: 1px solid var(--ui-border);
    transition: border-color var(--dur-4) var(--ease-out);

    &[inert] {
        border-color: transparent;
    }

    &[data-placement='drawer'] {
        height: 100%;
        border: 0;
    }
}

.infobar-inner {
    display: flex;
    width: var(--infobar-width);
    height: 100%;
    flex-direction: column;
    animation: infobar-in var(--dur-3) var(--ease-out) both;
    transition:
        opacity var(--dur-3) var(--ease-out),
        translate var(--dur-3) var(--ease-out);

    .infobar[data-placement='drawer'] & {
        width: 100%;
        animation: none;
    }

    .infobar[inert] & {
        opacity: 0;
        translate: -8px 0;
    }

    @media (prefers-reduced-motion: reduce) {
        animation-name: infobar-fade;
        translate: none;
    }
}

.infobar-header {
    display: flex;
    min-height: 44px;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    padding-inline: calc(var(--ui-spacing) * 4);
    border-bottom: 1px solid var(--ui-border);
}

.infobar-top {
    flex: none;
    padding: calc(var(--ui-spacing) * 4);
}

@keyframes infobar-in {
    from {
        opacity: 0;
        translate: -8px 0;
    }
}

@keyframes infobar-fade {
    from {
        opacity: 0;
    }
}
</style>
