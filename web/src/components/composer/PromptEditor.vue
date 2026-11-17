<script setup lang="ts">
import { onBeforeUnmount, watch } from 'vue'
import { Extension, type JSONContent } from '@tiptap/core'
import { Placeholder } from '@tiptap/extensions'
import StarterKit from '@tiptap/starter-kit'
import { EditorContent, useEditor } from '@tiptap/vue-3'

const props = defineProps<{
    modelValue: string
    placeholder: string
    revision: number
    label: string
}>()

const emit = defineEmits<{
    'update:modelValue': [value: string]
    submit: []
    historyBack: []
    escape: []
    files: [files: File[]]
}>()

function toDoc(text: string): JSONContent {
    return {
        type: 'doc',
        content: text.split('\n').map((line) => ({
            type: 'paragraph',
            content: line ? [{ type: 'text', text: line }] : [],
        })),
    }
}

function emitFiles(list: FileList | null | undefined): boolean {
    const files = [...(list ?? [])]
    if (!files.length) return false
    emit('files', files)
    return true
}

const Keys = Extension.create({
    name: 'driftKeys',
    addKeyboardShortcuts() {
        return {
            Enter: () => {
                emit('submit')
                return true
            },
            'Mod-Enter': () => {
                emit('submit')
                return true
            },
            ArrowUp: ({ editor }) => {
                if (!editor.isEmpty) return false
                emit('historyBack')
                return true
            },
            Escape: () => {
                emit('escape')
                return false
            },
        }
    },
})

const editor = useEditor({
    content: toDoc(props.modelValue),
    extensions: [
        StarterKit.configure({
            heading: false,
            blockquote: false,
            bulletList: false,
            orderedList: false,
            listItem: false,
            listKeymap: false,
            horizontalRule: false,
            codeBlock: false,
            link: false,
            underline: false,
        }),
        Placeholder.configure({ placeholder: () => props.placeholder }),
        Keys,
    ],
    editorProps: {
        handlePaste: (_view, event) => emitFiles(event.clipboardData?.files),
        handleDrop: (_view, event) => emitFiles(event.dataTransfer?.files),
        attributes: {
            class: 'prompt',
            'aria-label': props.label,
            'aria-multiline': 'true',
            role: 'textbox',
        },
    },
    onUpdate: ({ editor: e }) => emit('update:modelValue', e.getText({ blockSeparator: '\n' })),
})

watch(
    () => props.revision,
    () => {
        const e = editor.value
        if (!e || e.getText({ blockSeparator: '\n' }) === props.modelValue) return
        e.commands.setContent(toDoc(props.modelValue), { emitUpdate: false })
        e.commands.focus('end')
    },
)

watch(
    () => props.placeholder,
    () => editor.value?.view.dispatch(editor.value.state.tr),
)

defineExpose({ focus: () => editor.value?.commands.focus('end') })

onBeforeUnmount(() => editor.value?.destroy())
</script>

<template>
    <EditorContent :editor="editor" class="editor" />
</template>

<style scoped>
.editor :deep(.prompt) {
    min-height: 56px;
    max-height: 240px;
    overflow-y: auto;
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
