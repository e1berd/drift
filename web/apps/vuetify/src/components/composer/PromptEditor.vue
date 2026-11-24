<script setup lang="ts">
import { EditorContent } from '@tiptap/vue-3'

import PastedTextChip from '@/components/composer/PastedTextChip.vue'
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
    pastedTextView: PastedTextChip,
    onChange: (text) => emit('update:modelValue', text),
    onSubmit: () => emit('submit'),
    onHistoryBack: () => emit('historyBack'),
    onEscape: () => emit('escape'),
    onFiles: (files) => emit('files', files),
})

defineExpose({ focus })
</script>

<template>
    <EditorContent :editor="editor" class="prompt-editor" />
</template>

<style scoped>
.prompt-editor {
    max-block-size: 240px;
    overflow-y: auto;

    & :deep(.prompt) {
        min-block-size: 56px;
        padding: 12px 16px 4px;
        outline: none;
        white-space: pre-wrap;
        word-break: break-word;

        & p {
            margin: 0;
        }

        & p.is-editor-empty:first-child::before {
            content: attr(data-placeholder);
            float: inline-start;
            block-size: 0;
            pointer-events: none;
            color: rgb(var(--v-theme-on-surface) / var(--v-medium-emphasis-opacity));
        }
    }
}
</style>
