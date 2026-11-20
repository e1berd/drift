<script setup lang="ts">
import { computed } from 'vue'
import { PhSidebarSimple } from '@phosphor-icons/vue'
import { Button } from '@ui-sdk/vue/button'
import { usePreference } from '@ui-sdk/vue/preferences'
import { useI18n } from 'vue-i18n'

import { useShell } from '@/composables/useShell'
import { sidebarSide } from '@/shared/preferences'

const { t } = useI18n()
const { sidebarOpen, toggleSidebar } = useShell()
const side = usePreference(sidebarSide)
const left = computed(() => side.value.value === 'left')
const label = computed(() => (sidebarOpen.value ? t('nav.toggle.hide') : t('nav.toggle.show')))
</script>

<template>
    <header class="topbar">
        <div class="flex min-w-0 items-center gap-2">
            <Button
                v-if="left"
                :aria-label="label"
                :aria-expanded="sidebarOpen"
                v-tooltip="label"
                @click="toggleSidebar"
            >
                <PhSidebarSimple />
            </Button>
            <div class="crumb"><slot name="start" /></div>
        </div>
        <div class="flex justify-center"><slot name="center" /></div>
        <div class="flex min-w-0 items-center justify-end gap-2">
            <slot name="end" />
            <Button
                v-if="!left"
                :aria-label="label"
                :aria-expanded="sidebarOpen"
                v-tooltip="label"
                @click="toggleSidebar"
            >
                <PhSidebarSimple class="-scale-x-100" />
            </Button>
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

    @media (max-width: 1024px) {
        grid-template-columns: auto minmax(0, 1fr) auto;
    }
}

.crumb {
    display: flex;
    min-width: 0;
    overflow: hidden;
    align-items: center;
    gap: 8px;
    padding-inline-start: 6px;

    @media (max-width: 1024px) {
        display: none;
    }
}
</style>
