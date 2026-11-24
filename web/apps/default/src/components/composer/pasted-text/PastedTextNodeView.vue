<script setup lang="ts">
import { NodeViewWrapper, nodeViewProps } from '@tiptap/vue-3'

import PastedTextChip from './PastedTextChip.vue'

defineProps(nodeViewProps)
</script>

<template>
    <NodeViewWrapper as="span" :data-selected="selected || undefined">
        <PastedTextChip
            :text="node.attrs.text"
            :removable="editor.isEditable"
            @remove="deleteNode()"
        />
    </NodeViewWrapper>
</template>
