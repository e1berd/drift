<script setup lang="ts">
import { PhChatsCircle, PhFolderSimple } from '@phosphor-icons/vue'
import { Tabs, TabsIndicator, TabsList, TabsPanel, TabsTab } from '@ui-sdk/vue/tabs'
import { useI18n } from 'vue-i18n'

import ScrollPane from '@/components/ScrollPane.vue'
import ProjectsNav from '@/components/sidebar/ProjectsNav.vue'
import StandaloneChats from '@/components/sidebar/StandaloneChats.vue'
import { useSidebarTab } from '@/composables/useSidebarTab'

const { t } = useI18n()
const { tab } = useSidebarTab()
</script>

<template>
    <Tabs v-model:value="tab" class="flex min-h-0 flex-1 flex-col">
        <TabsList :aria-label="t('label')" class="sidebar-tabs">
            <TabsTab value="projects"><PhFolderSimple />{{ t('nav.projects') }}</TabsTab>
            <TabsTab value="chats"><PhChatsCircle />{{ t('nav.chats') }}</TabsTab>
            <TabsIndicator />
        </TabsList>
        <TabsPanel value="projects" class="sidebar-pane">
            <ScrollPane class="min-h-0 flex-1"><ProjectsNav /></ScrollPane>
        </TabsPanel>
        <TabsPanel value="chats" class="sidebar-pane">
            <ScrollPane class="min-h-0 flex-1"><StandaloneChats /></ScrollPane>
        </TabsPanel>
    </Tabs>
</template>

<style scoped>
.sidebar-tabs {
    flex: none;
    margin-top: calc(var(--ui-spacing) * 2);
    padding-inline: calc(var(--ui-spacing) * 2);

    & > :deep([data-slot='tabs-tab']) {
        flex: 1;
        justify-content: center;
    }
}

.sidebar-pane {
    display: flex;
    min-height: 0;
    flex: 1;
    flex-direction: column;
    padding: 0;
}
</style>

<i18n lang="yaml">
ru:
    label: 'Разделы боковой панели'
en:
    label: 'Sidebar sections'
</i18n>
