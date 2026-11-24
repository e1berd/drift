import { onBeforeUnmount, onMounted, toValue, watch } from 'vue'
import { basicSetup } from 'codemirror'
import { Compartment, EditorState, Prec } from '@codemirror/state'
import { EditorView, keymap } from '@codemirror/view'
import { LanguageDescription } from '@codemirror/language'
import { languages } from '@codemirror/language-data'

import type { Extension } from '@codemirror/state'
import type { MaybeRefOrGetter, Ref } from 'vue'

import { reportError } from '@drift/sdk/errors'

export interface CodeEditorOptions {
    path: MaybeRefOrGetter<string>
    content: MaybeRefOrGetter<string>
    readOnly: MaybeRefOrGetter<boolean | undefined>
    look: Extension[]
    onChange: (content: string) => void
    onSave: () => void
}

/**
 * A CodeMirror view mounted into `host`: language picked from the file name, Mod-S to save,
 * and a fresh state whenever another file or mode comes in. `look` holds the interface's theme.
 */
export function useCodeEditor(host: Readonly<Ref<HTMLElement | null>>, options: CodeEditorOptions) {
    const language = new Compartment()
    let view: EditorView | null = null
    let loadToken = 0

    function requestSave() {
        options.onSave()
        return true
    }

    function state(doc: string) {
        return EditorState.create({
            doc,
            extensions: [
                basicSetup,
                ...options.look,
                EditorState.readOnly.of(!!toValue(options.readOnly)),
                EditorView.contentAttributes.of({ 'aria-label': toValue(options.path) }),
                Prec.high(keymap.of([{ key: 'Mod-s', preventDefault: true, run: requestSave }])),
                EditorView.updateListener.of((update) => {
                    if (update.docChanged) options.onChange(update.state.doc.toString())
                }),
                language.of([]),
            ],
        })
    }

    async function applyLanguage(path: string) {
        const token = ++loadToken
        const description = LanguageDescription.matchFilename(languages, path)
        const support = description
            ? await description.load().catch((error: unknown) => (reportError(error), null))
            : null
        if (!view || token !== loadToken) return
        view.dispatch({ effects: language.reconfigure(support ?? []) })
    }

    onMounted(() => {
        if (!host.value) return
        view = new EditorView({ state: state(toValue(options.content)), parent: host.value })
        void applyLanguage(toValue(options.path))
    })

    watch(
        () => [toValue(options.path), toValue(options.content), toValue(options.readOnly)] as const,
        ([path, content, readOnly], [previousPath, , previousReadOnly]) => {
            if (!view) return
            const unchanged = path === previousPath && readOnly === previousReadOnly
            if (unchanged && view.state.doc.toString() === content) return
            view.setState(state(content))
            void applyLanguage(path)
        },
    )

    onBeforeUnmount(() => {
        view?.destroy()
        view = null
    })
}
