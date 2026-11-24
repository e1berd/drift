import MarkdownIt from 'markdown-it'
import type { RendererRule } from 'markdown-it'

const markdown = new MarkdownIt({ html: false, linkify: true, breaks: true })
const { escapeHtml } = markdown.utils

const COPY_ICON =
    '<svg viewBox="0 0 256 256" width="14" height="14" fill="currentColor" aria-hidden="true"><path d="M216 32H88a8 8 0 0 0-8 8v40H40a8 8 0 0 0-8 8v128a8 8 0 0 0 8 8h128a8 8 0 0 0 8-8v-40h40a8 8 0 0 0 8-8V40a8 8 0 0 0-8-8Zm-56 176H48V96h112Zm48-48h-32V88a8 8 0 0 0-8-8H96V48h112Z"/></svg>'

/** Words the rendered markdown shows in the reader's language. */
export interface MarkdownLabels {
    copy: string
}

function codeLanguage(info: string): string {
    return info.trim().split(/\s+/).at(0) ?? ''
}

function inCodeBlock(rule: RendererRule): RendererRule {
    return (tokens, index, options, environment, renderer) => {
        const language = codeLanguage(tokens[index]?.info ?? '')
        const copy = escapeHtml((environment?.labels as MarkdownLabels | undefined)?.copy ?? '')
        return (
            '<div class="code-block"><div class="code-block-head">' +
            `<span class="code-block-language">${escapeHtml(language)}</span>` +
            `<button type="button" class="code-block-copy" data-copy aria-label="${copy}" title="${copy}">${COPY_ICON}</button>` +
            `</div><div class="code-block-scroll">${rule(tokens, index, options, environment, renderer)}</div></div>`
        )
    }
}

function renderTokenWith(markup: (html: string) => string): RendererRule {
    return (tokens, index, options, _environment, renderer) =>
        markup(renderer.renderToken(tokens, index, options))
}

const { fence, code_block: codeBlock } = markdown.renderer.rules
if (fence) markdown.renderer.rules.fence = inCodeBlock(fence)
if (codeBlock) markdown.renderer.rules.code_block = inCodeBlock(codeBlock)
markdown.renderer.rules.table_open = renderTokenWith(
    (html) => '<div class="markdown-scroll">' + html,
)
markdown.renderer.rules.table_close = renderTokenWith((html) => html + '</div>')

markdown.renderer.rules.link_open = (tokens, index, options, _environment, renderer) => {
    tokens[index]?.attrSet('target', '_blank')
    tokens[index]?.attrSet('rel', 'noopener noreferrer')
    return renderer.renderToken(tokens, index, options)
}

export function renderMarkdown(source: string, labels: MarkdownLabels): string {
    return markdown.render(source, { labels })
}
