<script setup lang="ts">
import { onBeforeUnmount, onMounted, useTemplateRef, watch } from 'vue'
import { basicSetup } from 'codemirror'
import { Compartment, EditorState, Prec } from '@codemirror/state'
import { EditorView, keymap } from '@codemirror/view'
import { LanguageDescription } from '@codemirror/language'
import { languages } from '@codemirror/language-data'

import { driftHighlight, driftTheme } from '@/components/files/codeTheme'
import { reportError } from '@drift/sdk/errors'

const props = defineProps<{ path: string; content: string; readOnly?: boolean }>()
const emit = defineEmits<{ change: [content: string]; save: [] }>()

const host = useTemplateRef<HTMLDivElement>('host')
const language = new Compartment()
let view: EditorView | null = null
let loadToken = 0

function state(doc: string) {
    return EditorState.create({
        doc,
        extensions: [
            basicSetup,
            driftTheme,
            driftHighlight,
            EditorState.readOnly.of(!!props.readOnly),
            EditorView.contentAttributes.of({ 'aria-label': props.path }),
            Prec.high(keymap.of([{ key: 'Mod-s', preventDefault: true, run: requestSave }])),
            EditorView.updateListener.of((update) => {
                if (update.docChanged) emit('change', update.state.doc.toString())
            }),
            language.of([]),
        ],
    })
}

function requestSave() {
    emit('save')
    return true
}

async function applyLanguage(path: string) {
    const token = ++loadToken
    const desc = LanguageDescription.matchFilename(languages, path)
    const support = desc
        ? await desc.load().catch((err: unknown) => (reportError(err), null))
        : null
    if (!view || token !== loadToken) return
    view.dispatch({ effects: language.reconfigure(support ?? []) })
}

onMounted(() => {
    if (!host.value) return
    view = new EditorView({ state: state(props.content), parent: host.value })
    void applyLanguage(props.path)
})

watch(
    () => [props.path, props.content, props.readOnly] as const,
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
</script>

<template>
    <div ref="host" class="min-h-0 flex-1 overflow-hidden" />
</template>
