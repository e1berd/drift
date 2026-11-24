<script setup lang="ts">
import { useTemplateRef } from 'vue'

import { driftHighlight, driftTheme } from '@/components/files/codeTheme'
import { useCodeEditor } from '@drift/vue/code/useCodeEditor'

const props = defineProps<{ path: string; content: string; readOnly?: boolean }>()
const emit = defineEmits<{ change: [content: string]; save: [] }>()

const host = useTemplateRef<HTMLDivElement>('host')
useCodeEditor(host, {
    path: () => props.path,
    content: () => props.content,
    readOnly: () => props.readOnly,
    look: [driftTheme, driftHighlight],
    onChange: (content) => emit('change', content),
    onSave: () => emit('save'),
})
</script>

<template>
    <div ref="host" class="min-h-0 flex-1 overflow-hidden" />
</template>
