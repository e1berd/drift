<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useQueryClient } from '@tanstack/vue-query'
import { Drawer, DrawerBackdrop, DrawerContent, DrawerTitle } from '@ui-sdk/vue/drawer'
import { Preferences, usePreference } from '@ui-sdk/vue/preferences'
import {
    Resizable,
    ResizableHandle,
    ResizablePanel,
    createResizableController,
} from '@ui-sdk/vue/resizable'

import AppSidebar from '@/components/sidebar/AppSidebar.vue'
import AppToasts from '@/components/AppToasts.vue'
import DeliveryPrompt from '@/components/delivery/DeliveryPrompt.vue'
import { useMediaQuery } from '@drift/vue/composables/useMediaQuery'
import { provideShell } from '@drift/vue/composables/useShell'
import { startEventStream } from '@drift/vue/composables/useEventStream'
import { loadDrafts } from '@drift/vue/composables/useDrafts'
import { useRealtimeNotice } from '@drift/vue/composables/useRealtimeNotice'
import { reportError } from '@drift/sdk/errors'
import { useI18n } from 'vue-i18n'
import { shellPreferences, sidebarSide } from '@/shared/preferences'

const { t } = useI18n()
const qc = useQueryClient()
const side = usePreference(sidebarSide)
const narrow = useMediaQuery('(max-width: 760px)')
useRealtimeNotice()

const shell = createResizableController()
const collapsed = ref(false)
const drawerOpen = ref(false)

const order = computed(() =>
    side.value.value === 'left' ? ['sidebar', 'main'] : ['main', 'sidebar'],
)
const sidebarOpen = computed(() => (narrow.value ? drawerOpen.value : !collapsed.value))

function syncCollapsed() {
    collapsed.value = shell.isCollapsed('sidebar')
}

function toggleSidebar() {
    if (narrow.value) {
        drawerOpen.value = !drawerOpen.value
        return
    }
    if (shell.isCollapsed('sidebar')) shell.expand('sidebar')
    else shell.collapse('sidebar')
    syncCollapsed()
}

provideShell({
    sidebarOpen,
    toggleSidebar,
    closeOnNarrow: () => {
        if (narrow.value) drawerOpen.value = false
    },
})

onMounted(() => {
    syncCollapsed()
    startEventStream(qc)
    loadDrafts().catch(reportError)
})
</script>

<template>
    <Preferences :preferences="shellPreferences">
        <div class="shell h-dvh">
            <template v-if="narrow">
                <main class="flex h-full min-w-0 flex-col"><slot /></main>
                <Drawer v-model:open="drawerOpen" :side="side.value === 'left' ? 'left' : 'right'">
                    <DrawerBackdrop />
                    <DrawerContent class="sidebar-drawer">
                        <DrawerTitle class="sr-only">{{ t('nav.projects') }}</DrawerTitle>
                        <AppSidebar />
                    </DrawerContent>
                </Drawer>
            </template>

            <Resizable
                v-else
                :key="side.value"
                auto-save-id="drift-shell"
                :controller="shell"
                class="shell-panels"
                @layout-change="syncCollapsed"
            >
                <template v-for="(id, index) in order" :key="id">
                    <ResizableHandle v-if="index === 1" :aria-label="t('nav.resize')" />
                    <ResizablePanel
                        v-if="id === 'sidebar'"
                        id="sidebar"
                        class="sidebar-panel"
                        default-size="300px"
                        min-size="240px"
                        max-size="520px"
                        collapsible
                    >
                        <div class="sidebar-inner"><AppSidebar /></div>
                    </ResizablePanel>
                    <ResizablePanel v-else id="main" :min-size="40" class="min-w-0">
                        <main class="flex h-full min-w-0 flex-col"><slot /></main>
                    </ResizablePanel>
                </template>
            </Resizable>
        </div>
        <AppToasts />
        <DeliveryPrompt />
    </Preferences>
</template>

<style scoped>
.shell-panels {
    height: 100%;
    border: 0;
    border-radius: 0;
}

.shell :deep([data-slot='resizable-panel']) {
    padding: 0;
    transition: flex-basis var(--dur-4) var(--ease-out);
}

.shell :deep([data-slot='resizable'][data-dragging] [data-slot='resizable-panel']) {
    transition: none;
}

.sidebar-panel {
    overflow: hidden;
}

.sidebar-inner {
    height: 100%;
    min-width: 240px;
    transition: opacity var(--dur-3) var(--ease-out);
}

.sidebar-panel[data-collapsed] .sidebar-inner {
    opacity: 0;
}

.sidebar-drawer {
    width: min(88vw, 340px);
    padding: 0;
}
</style>
