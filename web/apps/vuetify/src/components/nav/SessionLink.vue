<script setup lang="ts">
import type { Session } from '@drift/vue/api'

import AgentAvatar from '@/components/AgentAvatar.vue'
import { useShell } from '@drift/vue/composables/useShell'

defineProps<{ session: Session }>()
const { closeOnNarrow } = useShell()
</script>

<template>
    <v-list-item
        :title="session.title"
        :to="{ name: 'session', params: { sessionId: session.id } }"
        @click="closeOnNarrow"
    >
        <template #prepend>
            <AgentAvatar :agent="session.agent" :size="20" class="me-3" />
        </template>
        <template v-if="session.status === 'running'" #append>
            <v-progress-circular indeterminate size="14" width="2" color="primary" />
        </template>
    </v-list-item>
</template>
