import { describe, expect, it } from 'vitest'

import { renderMarkdown } from '../markdown'

const render = (source: string) => renderMarkdown(source, { copy: 'Copy' })

describe('renderMarkdown', () => {
    it('renders emphasis, lists and code', () => {
        const html = render('**1. Bold** text\n\n- item\n\n```go\nfmt.Println()\n```')

        expect(html).toContain('<strong>1. Bold</strong>')
        expect(html).toContain('<li>item</li>')
        expect(html).toContain('<pre><code class="language-go">fmt.Println()\n</code></pre>')
    })

    it('wraps code blocks in a card with the language and a copy button', () => {
        const html = renderMarkdown('```bash\nls -la\n```', { copy: 'Copy "code"' })

        expect(html).toContain('<span class="code-block-language">bash</span>')
        expect(html).toContain('aria-label="Copy &quot;code&quot;"')
        expect(html).toContain('<div class="code-block-scroll"><pre>')
    })

    it('wraps tables in a horizontal scroll container', () => {
        const html = render('| a |\n| - |\n| b |')

        expect(html).toContain('<div class="markdown-scroll"><table>')
        expect(html).toContain('</table>\n</div>')
    })

    it('escapes raw html', () => {
        expect(render('<img src=x onerror=alert(1)>')).not.toContain('<img')
    })

    it('drops javascript links', () => {
        expect(render('[x](javascript:alert(1))')).not.toContain('href')
    })

    it('opens links in a new tab', () => {
        expect(render('https://example.com')).toContain(
            '<a href="https://example.com" target="_blank" rel="noopener noreferrer">',
        )
    })

    it('keeps single line breaks', () => {
        expect(render('one\ntwo')).toContain('one<br>\ntwo')
    })
})
