<script setup lang="ts">
import { mdiCog, mdiPlus, mdiWaves } from '@mdi/js'
import { useI18n } from 'vue-i18n'

import ProjectGroup from '@/components/nav/ProjectGroup.vue'
import SessionLink from '@/components/nav/SessionLink.vue'
import { useShell } from '@drift/vue/composables/useShell'
import { useProjects, useStandaloneSessions } from '@drift/vue/queries'

const { t } = useI18n()
const { closeOnNarrow } = useShell()
const { projects } = useProjects()
const { sessions: chats } = useStandaloneSessions()
</script>

<template>
    <div class="d-flex flex-column fill-height">
        <v-list-item :prepend-icon="mdiWaves" title="Drift" class="py-3" />
        <div class="px-3 pb-2">
            <v-btn
                block
                variant="tonal"
                color="primary"
                :prepend-icon="mdiPlus"
                to="/"
                @click="closeOnNarrow"
                >{{ t('nav.newChat') }}</v-btn
            >
        </div>
        <v-list density="compact" nav class="flex-grow-1 overflow-y-auto">
            <v-list-subheader>{{ t('nav.projects') }}</v-list-subheader>
            <ProjectGroup v-for="project in projects" :key="project.id" :project="project" />
            <v-list-item v-if="!projects.length" :subtitle="t('nav.noProjects')" disabled />
            <v-list-subheader>{{ t('nav.chats') }}</v-list-subheader>
            <SessionLink v-for="chat in chats" :key="chat.id" :session="chat" />
            <v-list-item v-if="!chats.length" :subtitle="t('nav.noChats')" disabled />
        </v-list>
        <v-divider />
        <v-list density="compact" nav>
            <v-list-item
                :prepend-icon="mdiCog"
                :title="t('nav.settings')"
                to="/settings"
                @click="closeOnNarrow"
            />
        </v-list>
    </div>
</template>
