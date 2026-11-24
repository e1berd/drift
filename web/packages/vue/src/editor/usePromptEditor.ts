import { onBeforeUnmount, toValue, watch } from 'vue'
import { Extension } from '@tiptap/core'
import { Placeholder } from '@tiptap/extensions'
import StarterKit from '@tiptap/starter-kit'
import { useEditor } from '@tiptap/vue-3'

import type { JSONContent } from '@tiptap/core'
import type { Component, MaybeRefOrGetter } from 'vue'

import { PastedText } from './pastedText'

export interface PromptEditorOptions {
    text: MaybeRefOrGetter<string>
    revision: MaybeRefOrGetter<number>
    placeholder: MaybeRefOrGetter<string>
    label: string
    disabled: MaybeRefOrGetter<boolean | undefined>
    pastedTextView: Component
    onChange: (text: string) => void
    onSubmit: () => void
    onHistoryBack: () => void
    onEscape: () => void
    onFiles: (files: File[]) => void
}

const PLAIN_TEXT_ONLY = {
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
} as const

function toDoc(text: string): JSONContent {
    return {
        type: 'doc',
        content: text.split('\n').map((line) => ({
            type: 'paragraph',
            content: line ? [{ type: 'text', text: line }] : [],
        })),
    }
}

function promptKeys(options: PromptEditorOptions) {
    return Extension.create({
        name: 'driftKeys',
        addKeyboardShortcuts() {
            return {
                Enter: () => {
                    options.onSubmit()
                    return true
                },
                'Mod-Enter': () => {
                    options.onSubmit()
                    return true
                },
                ArrowUp: ({ editor }) => {
                    if (!editor.isEmpty) return false
                    options.onHistoryBack()
                    return true
                },
                Escape: () => {
                    options.onEscape()
                    return false
                },
            }
        },
    })
}

/**
 * The Tiptap editor behind the prompt box: plain paragraphs, large pastes folded into chips,
 * Enter to send and files taken from paste and drop. The interface renders `editor` itself.
 */
export function usePromptEditor(options: PromptEditorOptions) {
    function takeFiles(list: FileList | null | undefined): boolean {
        const files = [...(list ?? [])]
        if (!files.length) return false
        options.onFiles(files)
        return true
    }

    const editor = useEditor({
        content: toDoc(toValue(options.text)),
        editable: !toValue(options.disabled),
        extensions: [
            StarterKit.configure(PLAIN_TEXT_ONLY),
            Placeholder.configure({ placeholder: () => toValue(options.placeholder) }),
            PastedText.configure({ nodeView: options.pastedTextView }),
            promptKeys(options),
        ],
        editorProps: {
            handlePaste: (_view, event) => takeFiles(event.clipboardData?.files),
            handleDrop: (_view, event) => takeFiles(event.dataTransfer?.files),
            attributes: {
                class: 'prompt',
                'aria-label': options.label,
                'aria-multiline': 'true',
                role: 'textbox',
            },
        },
        onUpdate: ({ editor: current }) =>
            options.onChange(current.getText({ blockSeparator: '\n' })),
    })

    watch(
        () => toValue(options.revision),
        () => {
            const current = editor.value
            const text = toValue(options.text)
            if (!current || current.getText({ blockSeparator: '\n' }) === text) return
            current.commands.setContent(toDoc(text), { emitUpdate: false })
            current.commands.focus('end')
        },
    )

    watch(
        () => toValue(options.placeholder),
        () => editor.value?.view.dispatch(editor.value.state.tr),
    )

    watch(
        () => toValue(options.disabled),
        (disabled) => editor.value?.setEditable(!disabled, false),
    )

    onBeforeUnmount(() => editor.value?.destroy())

    return { editor, focus: () => editor.value?.commands.focus('end') }
}
