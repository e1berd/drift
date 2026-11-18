import MarkdownIt from 'markdown-it'
import type { RendererRule } from 'markdown-it'

const markdown = new MarkdownIt({ html: false, linkify: true, breaks: true })

const HORIZONTAL_SCROLL_START =
    '<ui-scroll-area class="markdown-scroll"><ui-scroll-area-viewport><ui-scroll-area-content>'
const HORIZONTAL_SCROLL_END =
    '</ui-scroll-area-content></ui-scroll-area-viewport>' +
    '<ui-scroll-area-scrollbar orientation="horizontal"><ui-scroll-area-thumb></ui-scroll-area-thumb></ui-scroll-area-scrollbar>' +
    '</ui-scroll-area>'

function inHorizontalScroll(rule: RendererRule): RendererRule {
    return (...args) => HORIZONTAL_SCROLL_START + rule(...args) + HORIZONTAL_SCROLL_END
}

function renderTokenWith(markup: (html: string) => string): RendererRule {
    return (tokens, index, options, _environment, renderer) =>
        markup(renderer.renderToken(tokens, index, options))
}

const { fence, code_block: codeBlock } = markdown.renderer.rules
if (fence) markdown.renderer.rules.fence = inHorizontalScroll(fence)
if (codeBlock) markdown.renderer.rules.code_block = inHorizontalScroll(codeBlock)
markdown.renderer.rules.table_open = renderTokenWith((html) => HORIZONTAL_SCROLL_START + html)
markdown.renderer.rules.table_close = renderTokenWith((html) => html + HORIZONTAL_SCROLL_END)

markdown.renderer.rules.link_open = (tokens, index, options, _environment, renderer) => {
    tokens[index]?.attrSet('target', '_blank')
    tokens[index]?.attrSet('rel', 'noopener noreferrer')
    return renderer.renderToken(tokens, index, options)
}

export function renderMarkdown(source: string): string {
    return markdown.render(source)
}
