import { reportError } from './errors'

const COPIED_MS = 1500

/** Handles a click inside rendered markdown: a code block's copy button copies its code. */
export function copyCodeOnClick(event: MouseEvent): void {
    const button = (event.target as Element | null)?.closest<HTMLElement>('[data-copy]')
    const code = button?.closest('.code-block')?.querySelector('pre')?.textContent
    if (!button || code === undefined || code === null) return
    navigator.clipboard
        .writeText(code)
        .then(() => {
            button.dataset.copied = ''
            setTimeout(() => delete button.dataset.copied, COPIED_MS)
        })
        .catch(reportError)
}
