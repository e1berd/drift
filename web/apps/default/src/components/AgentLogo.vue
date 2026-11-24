<script setup lang="ts">
import { computed } from 'vue'

import type { AgentId } from '@drift/sdk/agents'
import { AGENTS } from '@drift/sdk/agents'

const props = withDefaults(defineProps<{ agent: AgentId; size?: number; decorative?: boolean }>(), {
    size: 16,
    decorative: false,
})

const meta = computed(() => AGENTS[props.agent])
</script>

<template>
    <img
        :src="meta.logo"
        :width="size"
        :height="size"
        :alt="decorative ? '' : meta.name"
        v-tooltip="decorative ? undefined : meta.name"
        class="shrink-0 rounded-[4px] select-none"
        draggable="false"
    />
</template>
