<script setup lang="ts">
import { computed } from 'vue'

import type { ToolFamily } from '@drift/sdk/toolCalls'
import { firstLine, inputIsPath, shellCommand, splitPath } from '@drift/sdk/toolCalls'

const props = defineProps<{ family: ToolFamily; input: string }>()

const shown = computed(() =>
    firstLine(props.family === 'shell' ? shellCommand(props.input) : props.input),
)
const path = computed(() => (inputIsPath(props.family) ? splitPath(shown.value.line) : null))
</script>

<template>
    <span class="tool-summary">
        <span v-if="path" class="tool-summary-text"
            ><span class="text-muted">{{ path.folder }}</span
            ><span class="text-fg">{{ path.file }}</span></span
        >
        <span v-else class="tool-summary-text" :data-shell="family === 'shell' || undefined">{{
            shown.line
        }}</span>
        <span v-if="shown.more" class="tool-summary-more">+{{ shown.more }}</span>
    </span>
</template>

<style scoped>
.tool-summary {
    display: flex;
    min-width: 0;
    flex: 1;
    align-items: baseline;
    gap: calc(var(--ui-spacing) * 1.5);
    font-family: var(--ui-font-mono);
    font-size: 0.8125rem;
}

.tool-summary-text {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;

    &[data-shell]::before {
        content: '$ ';
        color: var(--ui-muted-foreground);
    }
}

.tool-summary-more {
    flex: none;
    color: var(--ui-muted-foreground);
}
</style>
