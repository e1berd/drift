<script setup lang="ts">
import { computed } from 'vue'
import { NodeViewWrapper, nodeViewProps } from '@tiptap/vue-3'
import { mdiTextBoxOutline } from '@mdi/js'
import { useI18n } from 'vue-i18n'

import { countLines } from '@drift/vue/editor/pastedText'

const props = defineProps(nodeViewProps)
const { t } = useI18n()
const lines = computed(() => countLines(String(props.node.attrs.text)))
</script>

<template>
    <NodeViewWrapper as="span">
        <v-chip
            size="small"
            :prepend-icon="mdiTextBoxOutline"
            :closable="editor.isEditable"
            :variant="selected ? 'flat' : 'tonal'"
            @click:close="deleteNode()"
            >{{ t('composer.pasted', { lines }) }}</v-chip
        >
    </NodeViewWrapper>
</template>
