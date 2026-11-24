<script setup lang="ts">
import { mdiDockRight } from '@mdi/js'
import { useI18n } from 'vue-i18n'

import { useShell } from '@drift/vue/composables/useShell'

defineProps<{ title: string; subtitle?: string }>()
const { t } = useI18n()
const { sidebarOpen, toggleSidebar } = useShell()
</script>

<template>
    <v-toolbar density="compact" color="surface" border="b">
        <v-toolbar-title>
            {{ title }}
            <span v-if="subtitle" class="text-medium-emphasis text-body-2 ms-2">{{
                subtitle
            }}</span>
        </v-toolbar-title>
        <slot />
        <v-btn
            :icon="mdiDockRight"
            :aria-label="sidebarOpen ? t('nav.hide') : t('nav.show')"
            @click="toggleSidebar"
        />
    </v-toolbar>
</template>
