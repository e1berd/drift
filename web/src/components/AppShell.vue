<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useMediaQuery } from '@/composables/useMediaQuery'
import { useQueryClient } from '@tanstack/vue-query'
import { Preferences, usePreference } from '@ui-sdk/vue/preferences'
import {
    Resizable,
    ResizableHandle,
    ResizablePanel,
    createResizableController,
} from '@ui-sdk/vue/resizable'

import AppSidebar from '@/components/sidebar/AppSidebar.vue'
import AppToasts from '@/components/AppToasts.vue'
import { provideShell } from '@/composables/useShell'
import { startEventStream } from '@/composables/useEventStream'
import { loadDrafts } from '@/composables/useDrafts'
import { reportError } from '@/shared/errors'
import { useI18n } from 'vue-i18n'
import { shellPreferences, sidebarSide } from '@/shared/preferences'

const { t } = useI18n()
const qc = useQueryClient()
const side = usePreference(sidebarSide)
const narrow = useMediaQuery('(max-width: 760px)')

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
        <div class="shell h-dvh" :data-side="side.value">
            <div v-if="narrow" class="relative h-full">
                <main class="flex h-full min-w-0 flex-col"><slot /></main>
                <button
                    v-if="drawerOpen"
                    class="drawer-scrim"
                    :aria-label="t('nav.toggle.hide')"
                    tabindex="-1"
                    @click="drawerOpen = false"
                />
                <aside
                    class="drawer"
                    :data-open="drawerOpen || undefined"
                    :inert="!drawerOpen || undefined"
                >
                    <AppSidebar />
                </aside>
            </div>

            <Resizable
                v-else
                :key="side.value"
                auto-save-id="drift-shell"
                :controller="shell"
                class="h-full"
                @layout-change="syncCollapsed"
            >
                <template v-for="(id, i) in order" :key="id">
                    <ResizableHandle v-if="i === 1" class="handle" :aria-label="t('nav.resize')" />
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
    </Preferences>
</template>

<style scoped>
.shell :deep([data-slot='resizable-panel']) {
    transition: flex-basis var(--dur-4) var(--ease-out);
}

.shell :deep([data-slot='resizable'][data-dragging] [data-slot='resizable-panel']) {
    transition: none;
}

.sidebar-panel {
    overflow: hidden;
    background: var(--surface);
}

.sidebar-inner {
    height: 100%;
    min-width: 240px;
    transition:
        opacity var(--dur-3) var(--ease-out),
        transform var(--dur-4) var(--ease-out);
}

.sidebar-panel[data-collapsed] .sidebar-inner {
    opacity: 0;
    transform: translateX(var(--sidebar-shift));
    transition:
        opacity var(--dur-2) var(--ease-in),
        transform var(--dur-3) var(--ease-in);
}

.shell[data-side='right'] {
    --sidebar-shift: 20px;
    --sidebar-edge: inset 1px 0 0 var(--ui-border);
}

.shell[data-side='left'] {
    --sidebar-shift: -20px;
    --sidebar-edge: inset -1px 0 0 var(--ui-border);
}

.sidebar-panel:not([data-collapsed]) {
    box-shadow: var(--sidebar-edge);
}

.handle {
    position: relative;
    z-index: 3;
    width: 0;
    cursor: col-resize;
    outline: none;

    &::before {
        content: '';
        position: absolute;
        inset-block: 0;
        left: -4px;
        width: 8px;
    }

    &::after {
        content: '';
        position: absolute;
        inset-block: 0;
        left: -1px;
        width: 2px;
        background: var(--accent);
        opacity: 0;
        transition: opacity var(--dur-2);
    }

    &[data-hovering]::after,
    &[data-dragging]::after,
    &:focus-visible::after {
        opacity: 1;
    }
}

.drawer-scrim {
    position: absolute;
    inset: 0;
    z-index: 60;
    border: 0;
    background: rgb(0 0 0 / 0.4);
    animation: drift-fade var(--dur-3) var(--ease-out) both;
}

.drawer {
    position: absolute;
    inset-block: 0;
    z-index: 70;
    width: min(88vw, 340px);
    background: var(--surface);
    box-shadow: var(--shadow);
    visibility: hidden;
    transition:
        transform var(--dur-3) var(--ease-in),
        visibility 0s linear var(--dur-3);
}

.shell[data-side='right'] .drawer {
    right: 0;
    transform: translateX(100%);
}

.shell[data-side='left'] .drawer {
    left: 0;
    transform: translateX(-100%);
}

.drawer[data-open] {
    transform: none;
    visibility: visible;
    transition:
        transform var(--dur-4) var(--ease-out),
        visibility 0s;
}
</style>
