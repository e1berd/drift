import { mergeAttributes, Node } from '@tiptap/core'
import { Plugin, PluginKey } from '@tiptap/pm/state'
import { VueNodeViewRenderer } from '@tiptap/vue-3'

import type { Component } from 'vue'

export interface PastedTextOptions {
    minChars: number
    minLines: number
    nodeView: Component | null
}

export function countLines(text: string): number {
    return text.split('\n').length
}

export function isLargePaste(
    text: string,
    { minChars, minLines }: Pick<PastedTextOptions, 'minChars' | 'minLines'>,
): boolean {
    return text.length >= minChars || countLines(text) >= minLines
}

export function normalizePaste(text: string): string {
    return text.replaceAll(/\r\n?/g, '\n').replaceAll(/^\n+|\s+$/g, '')
}

export const PastedText = Node.create<PastedTextOptions>({
    name: 'pastedText',
    group: 'inline',
    inline: true,
    atom: true,
    selectable: true,
    draggable: false,

    addOptions() {
        return { minChars: 1500, minLines: 20, nodeView: null }
    },

    addAttributes() {
        return {
            text: {
                default: '',
                parseHTML: (el) => el.getAttribute('data-pasted-text') ?? '',
                renderHTML: (attrs) => ({ 'data-pasted-text': attrs.text }),
            },
        }
    },

    parseHTML() {
        return [{ tag: 'span[data-pasted-text]' }]
    },

    renderHTML({ HTMLAttributes }) {
        return ['span', mergeAttributes(HTMLAttributes)]
    },

    renderText({ node }) {
        return node.attrs.text
    },

    addNodeView() {
        return this.options.nodeView ? VueNodeViewRenderer(this.options.nodeView) : null
    },

    addProseMirrorPlugins() {
        const { type, options } = this
        return [
            new Plugin({
                key: new PluginKey('pastedText'),
                props: {
                    handlePaste(view, event) {
                        if (event.clipboardData?.files.length) return false
                        const text = normalizePaste(
                            event.clipboardData?.getData('text/plain') ?? '',
                        )
                        if (!text || !isLargePaste(text, options)) return false
                        view.dispatch(
                            view.state.tr
                                .replaceSelectionWith(type.create({ text }), false)
                                .scrollIntoView(),
                        )
                        return true
                    },
                },
            }),
        ]
    },
})
