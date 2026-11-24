<script setup lang="ts">
import { EditorContent } from '@tiptap/vue-3'

import ScrollPane from '@/components/ScrollPane.vue'
import PastedTextNodeView from '@/components/composer/pasted-text/PastedTextNodeView.vue'
import { usePromptEditor } from '@drift/vue/editor/usePromptEditor'

const props = defineProps<{
    modelValue: string
    placeholder: string
    revision: number
    label: string
    disabled?: boolean
}>()

const emit = defineEmits<{
    'update:modelValue': [value: string]
    submit: []
    historyBack: []
    escape: []
    files: [files: File[]]
}>()

const { editor, focus } = usePromptEditor({
    text: () => props.modelValue,
    revision: () => props.revision,
    placeholder: () => props.placeholder,
    label: props.label,
    disabled: () => props.disabled,
    pastedTextView: PastedTextNodeView,
    onChange: (text) => emit('update:modelValue', text),
    onSubmit: () => emit('submit'),
    onHistoryBack: () => emit('historyBack'),
    onEscape: () => emit('escape'),
    onFiles: (files) => emit('files', files),
})

defineExpose({ focus })
</script>

<template>
    <ScrollPane class="editor-scroll">
        <EditorContent :editor="editor" class="editor" />
    </ScrollPane>
</template>

<style scoped>
.editor-scroll {
    max-height: 240px;
}

.editor :deep(.prompt) {
    min-height: 56px;
    padding: 12px 12px 4px;
    outline: none;
    white-space: pre-wrap;
    word-break: break-word;
}

.editor :deep(.prompt p) {
    margin: 0;
}

.editor :deep(.prompt p.is-editor-empty:first-child::before) {
    content: attr(data-placeholder);
    float: left;
    height: 0;
    color: var(--ui-muted-foreground);
    pointer-events: none;
}
</style>
