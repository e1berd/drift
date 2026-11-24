<script setup lang="ts">
import { PhArrowRight } from '@phosphor-icons/vue'
import { useI18n } from 'vue-i18n'

defineProps<{ current: string; next: string }>()
const { t } = useI18n()
</script>

<template>
    <div class="version-change">
        <span class="flex flex-col">
            <span class="text-muted">{{ t('current') }}</span>
            <code>{{ current }}</code>
        </span>
        <PhArrowRight class="text-muted" />
        <span class="flex flex-col">
            <span class="text-muted">{{ t('next') }}</span>
            <code class="font-semibold">{{ next }}</code>
        </span>
    </div>
</template>

<style scoped>
.version-change {
    display: flex;
    align-items: center;
    gap: 1rem;
    padding: 0.75rem 1rem;
    border: 1px solid var(--ui-border);
    border-radius: var(--ui-radius);
}
</style>

<i18n lang="yaml">
ru:
    current: Текущая версия
    next: Новая версия
en:
    current: Current version
    next: New version
</i18n>
