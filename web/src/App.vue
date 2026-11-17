<script setup lang="ts">
import { onMounted } from 'vue'
import { RouterView } from 'vue-router'

import AppShell from '@/components/AppShell.vue'
import { useSession } from '@/composables/useSession'

const { session, refresh } = useSession()

onMounted(refresh)
</script>

<template>
    <AppShell v-if="session">
        <RouterView v-slot="{ Component, route }">
            <component
                :is="Component"
                :key="
                    route.name === 'session' ? String(route.params.sessionId) : String(route.name)
                "
            />
        </RouterView>
    </AppShell>
</template>
