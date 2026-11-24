import { HighlightStyle, syntaxHighlighting } from '@codemirror/language'
import { EditorView } from '@codemirror/view'
import { tags as t } from '@lezer/highlight'

/** Editor chrome driven by the app's CSS tokens, so it follows light and dark themes. */
export const driftTheme = EditorView.theme({
    '&': {
        height: '100%',
        color: 'var(--ui-foreground)',
        backgroundColor: 'transparent',
        fontSize: 'var(--ui-font-size)',
    },
    '.cm-scroller': { fontFamily: 'var(--ui-font-mono)', lineHeight: '1.7' },
    '.cm-content': { caretColor: 'var(--ui-primary)', padding: '10px 0' },
    '.cm-gutters': {
        backgroundColor: 'transparent',
        color: 'var(--ui-muted-foreground)',
        border: 'none',
        paddingLeft: '8px',
    },
    '.cm-activeLine, .cm-activeLineGutter': { backgroundColor: 'var(--ui-muted)' },
    '&.cm-focused': { outline: 'none' },
    '&.cm-focused .cm-selectionBackground, .cm-selectionBackground, ::selection': {
        backgroundColor: 'color-mix(in oklab, var(--ui-primary) 28%, transparent) !important',
    },
    '.cm-foldGutter span': { color: 'var(--ui-muted-foreground)' },
    '.cm-searchMatch': {
        backgroundColor: 'color-mix(in oklab, var(--ui-foreground) 30%, transparent)',
    },
    '.cm-panels': { backgroundColor: 'var(--ui-muted)', color: 'var(--ui-foreground)' },
    '.cm-tooltip': {
        border: '1px solid var(--ui-input)',
        backgroundColor: 'var(--ui-background)',
        borderRadius: '8px',
    },
})

const style = HighlightStyle.define([
    {
        tag: [t.keyword, t.modifier, t.controlKeyword, t.operatorKeyword],
        color: 'light-dark(#8a3ffc, #c4a7ff)',
    },
    { tag: [t.string, t.special(t.string), t.regexp], color: 'light-dark(#1d7a46, #8fe0b0)' },
    { tag: [t.number, t.bool, t.null, t.atom], color: 'light-dark(#b35c00, #f2b86b)' },
    {
        tag: [t.comment, t.lineComment, t.blockComment],
        color: 'var(--ui-muted-foreground)',
        fontStyle: 'italic',
    },
    {
        tag: [t.function(t.variableName), t.function(t.propertyName)],
        color: 'light-dark(#2f5bd3, #9aa5ff)',
    },
    { tag: [t.typeName, t.className, t.namespace], color: 'light-dark(#0d7a75, #6ad1cb)' },
    { tag: [t.propertyName, t.attributeName], color: 'light-dark(#a33a6b, #f29ac3)' },
    { tag: [t.tagName, t.heading], color: 'light-dark(#c5343a, #f7a3a5)', fontWeight: '600' },
    { tag: [t.meta, t.punctuation, t.bracket], color: 'var(--ui-muted-foreground)' },
    { tag: t.invalid, color: 'var(--ui-destructive)' },
])

export const driftHighlight = syntaxHighlighting(style)
