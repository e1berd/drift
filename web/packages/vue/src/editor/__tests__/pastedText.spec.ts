import { afterEach, describe, expect, it } from 'vitest'
import { Editor } from '@tiptap/core'
import StarterKit from '@tiptap/starter-kit'

import { isLargePaste, normalizePaste, PastedText } from '../pastedText'

const options = { minChars: 50, minLines: 5 }

function paste(editor: Editor, text: string) {
    const clipboardData = {
        files: [] as unknown as FileList,
        getData: (type: string) => (type === 'text/plain' ? text : ''),
    }
    const event = Object.assign(new Event('paste', { cancelable: true }), { clipboardData })
    return editor.view.someProp('handlePaste', (f) =>
        f(editor.view, event as ClipboardEvent, editor.view.state.selection.content()),
    )
}

describe('isLargePaste', () => {
    it('triggers on chars or lines', () => {
        expect(isLargePaste('short', options)).toBe(false)
        expect(isLargePaste('x'.repeat(50), options)).toBe(true)
        expect(isLargePaste('a\nb\nc\nd\ne', options)).toBe(true)
    })
})

describe('normalizePaste', () => {
    it('unifies line endings and trims edge blank lines', () => {
        expect(normalizePaste('\n\n  a\r\nb\rc  \n\n')).toBe('  a\nb\nc')
    })
})

describe('PastedText', () => {
    let editor: Editor
    afterEach(() => editor?.destroy())

    function make() {
        editor = new Editor({
            extensions: [StarterKit, PastedText.configure(options)],
            content: {
                type: 'doc',
                content: [{ type: 'paragraph', content: [{ type: 'text', text: 'see ' }] }],
            },
        })
        editor.commands.focus('end')
        return editor
    }

    it('turns a large paste into a node and keeps its text in getText', () => {
        make()
        const big = Array.from({ length: 10 }, (_, i) => `line ${i}`).join('\n')
        expect(paste(editor, big)).toBe(true)
        expect(editor.getJSON().content?.at(0)?.content?.at(1)).toMatchObject({
            type: 'pastedText',
            attrs: { text: big },
        })
        expect(editor.getText({ blockSeparator: '\n' })).toBe(`see ${big}`)
    })

    it('leaves a small paste to the default handler', () => {
        make()
        expect(paste(editor, 'tiny')).toBeFalsy()
    })
})
