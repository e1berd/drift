<script setup lang="ts">
import { ref } from 'vue'
import { mdiFolderOutline, mdiPlus } from '@mdi/js'
import { useI18n } from 'vue-i18n'

import type { Project } from '@drift/vue/api'

import SessionLink from '@/components/nav/SessionLink.vue'
import { useShell } from '@drift/vue/composables/useShell'
import { useSessions } from '@drift/vue/queries'

const props = defineProps<{ project: Project }>()
const { t } = useI18n()
const { closeOnNarrow } = useShell()
const open = ref(false)
const { sessions } = useSessions(() => props.project.id, open)
</script>

<template>
    <v-list-group v-model="open" :value="project.id">
        <template #activator="{ props: activator }">
            <v-list-item
                v-bind="activator"
                :prepend-icon="mdiFolderOutline"
                :title="project.name"
            />
        </template>
        <v-list-item
            :prepend-icon="mdiPlus"
            :title="t('nav.newChatHere')"
            :to="{ name: 'new', query: { project: project.id } }"
            @click="closeOnNarrow"
        />
        <SessionLink v-for="session in sessions" :key="session.id" :session="session" />
    </v-list-group>
</template>
