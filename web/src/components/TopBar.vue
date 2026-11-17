<script setup lang="ts">
import { computed } from 'vue'
import { PhSidebarSimple } from '@phosphor-icons/vue'
import { usePreference } from '@ui-sdk/vue/preferences'

import { useShell } from '@/composables/useShell'
import { useI18n } from 'vue-i18n'
import { sidebarSide } from '@/shared/preferences'

const { t } = useI18n()
const { sidebarOpen, toggleSidebar } = useShell()
const side = usePreference(sidebarSide)
const left = computed(() => side.value.value === 'left')
const label = computed(() => (sidebarOpen.value ? t('nav.toggle.hide') : t('nav.toggle.show')))
</script>

<template>
    <header class="topbar">
        <div class="flex min-w-0 items-center gap-2 text-[13px]">
            <button
                v-if="left"
                class="icon-btn"
                :data-pressed="sidebarOpen || undefined"
                :aria-label="label"
                :title="label"
                :aria-expanded="sidebarOpen"
                @click="toggleSidebar"
            >
                <PhSidebarSimple :size="18" />
            </button>
            <div class="crumb"><slot name="start" /></div>
        </div>
        <div class="flex justify-center"><slot name="center" /></div>
        <div class="flex min-w-0 items-center justify-end gap-1.5">
            <slot name="end" />
            <button
                v-if="!left"
                class="icon-btn"
                :data-pressed="sidebarOpen || undefined"
                :aria-label="label"
                :title="label"
                :aria-expanded="sidebarOpen"
                @click="toggleSidebar"
            >
                <PhSidebarSimple :size="18" class="-scale-x-100" />
            </button>
        </div>
    </header>
</template>

<style scoped>
.topbar {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr);
    align-items: center;
    gap: 12px;
    height: var(--topbar-h);
    flex: none;
    padding: 0 10px;
    border-bottom: 1px solid var(--ui-border);
}

.crumb {
    display: flex;
    min-width: 0;
    align-items: center;
    gap: 8px;
    padding-inline-start: 10px;
}

@media (max-width: 760px) {
    .topbar {
        grid-template-columns: auto minmax(0, 1fr) auto;
    }

    .crumb {
        display: none;
    }
}
</style>
