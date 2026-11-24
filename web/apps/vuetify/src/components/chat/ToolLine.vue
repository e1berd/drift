<script setup lang="ts">
import { computed } from 'vue'
import {
    mdiConsole,
    mdiFileDocumentOutline,
    mdiFilePlusOutline,
    mdiFormatListChecks,
    mdiMagnify,
    mdiPencilOutline,
    mdiRobotOutline,
    mdiWeb,
    mdiWrenchOutline,
} from '@mdi/js'

import type { MessageBody } from '@drift/vue/api'
import { firstLine, shellCommand, toolFamily } from '@drift/sdk/toolCalls'
import type { ToolFamily } from '@drift/sdk/toolCalls'

const ICONS: Record<ToolFamily, string> = {
    shell: mdiConsole,
    edit: mdiPencilOutline,
    write: mdiFilePlusOutline,
    read: mdiFileDocumentOutline,
    search: mdiMagnify,
    web: mdiWeb,
    todo: mdiFormatListChecks,
    agent: mdiRobotOutline,
    other: mdiWrenchOutline,
}

const props = defineProps<{ body: MessageBody }>()
const family = computed(() => toolFamily(props.body.name ?? ''))
const summary = computed(() => {
    const input = props.body.input ?? ''
    return firstLine(family.value === 'shell' ? shellCommand(input) : input).line
})
</script>

<template>
    <v-chip
        size="small"
        variant="tonal"
        :prepend-icon="ICONS[family]"
        :color="body.status === 'error' ? 'error' : undefined"
        class="tool-line"
    >
        <strong class="me-2">{{ body.name }}</strong>
        <span class="text-truncate">{{ summary }}</span>
        <template v-if="body.status === 'running'" #append>
            <v-progress-circular indeterminate size="12" width="2" class="ms-2" />
        </template>
    </v-chip>
</template>

<style scoped>
.tool-line {
    max-inline-size: 100%;
}
</style>
