import { describe, expect, it } from 'vitest'

import { renderMarkdown } from '@/shared/markdown'

describe('renderMarkdown', () => {
    it('renders emphasis, lists and code', () => {
        const html = renderMarkdown('**1. Bold** text\n\n- item\n\n```go\nfmt.Println()\n```')

        expect(html).toContain('<strong>1. Bold</strong>')
        expect(html).toContain('<li>item</li>')
        expect(html).toContain('<pre><code class="language-go">fmt.Println()\n</code></pre>')
    })

    it('wraps code blocks and tables in a horizontal scroll area', () => {
        const html = renderMarkdown('```\nwide\n```\n\n| a |\n| - |\n| b |')

        expect(html.match(/<ui-scroll-area class="markdown-scroll">/g)).toHaveLength(2)
        expect(html).toContain('<ui-scroll-area-content><pre>')
        expect(html).toContain('<ui-scroll-area-content><table>')
        expect(html).toContain('</table>\n</ui-scroll-area-content>')
    })

    it('escapes raw html', () => {
        expect(renderMarkdown('<img src=x onerror=alert(1)>')).not.toContain('<img')
    })

    it('drops javascript links', () => {
        expect(renderMarkdown('[x](javascript:alert(1))')).not.toContain('href')
    })

    it('opens links in a new tab', () => {
        expect(renderMarkdown('https://example.com')).toContain(
            '<a href="https://example.com" target="_blank" rel="noopener noreferrer">',
        )
    })

    it('keeps single line breaks', () => {
        expect(renderMarkdown('one\ntwo')).toContain('one<br>\ntwo')
    })
})
