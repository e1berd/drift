<script setup lang="ts">
import { AGENTS } from '@drift/sdk/agents'
import type { AgentId } from '@drift/sdk/agents'

defineProps<{ agent: AgentId; size?: number }>()
</script>

<template>
    <v-avatar :size="size ?? 24" rounded="0">
        <v-img :src="AGENTS[agent].logo" :alt="AGENTS[agent].name" />
    </v-avatar>
</template>
