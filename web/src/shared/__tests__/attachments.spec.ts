import { describe, expect, it } from 'vitest'

import {
    ATTACHMENT_MAX_BYTES,
    ATTACHMENT_MAX_COUNT,
    attachmentURL,
    isAttachable,
    partitionAttachable,
} from '../attachments'

const image = (name: string, size = 10, type = 'image/png') =>
    new File([new Uint8Array(size)], name, { type })

describe('attachments', () => {
    it('accepts only supported images within the size limit', () => {
        expect(isAttachable(image('a.png'))).toBe(true)
        expect(isAttachable(image('a.webp', 10, 'image/webp'))).toBe(true)
        expect(isAttachable(image('a.svg', 10, 'image/svg+xml'))).toBe(false)
        expect(isAttachable(image('a.txt', 10, 'text/plain'))).toBe(false)
        expect(isAttachable(image('big.png', ATTACHMENT_MAX_BYTES + 1))).toBe(false)
        expect(isAttachable(image('empty.png', 0))).toBe(false)
    })

    it('rejects files past the per-message limit', () => {
        const files = Array.from({ length: ATTACHMENT_MAX_COUNT + 2 }, (_, i) => image(`${i}.png`))
        const { accepted, rejected } = partitionAttachable([
            image('doc.pdf', 10, 'application/pdf'),
            ...files,
        ])
        expect(accepted).toHaveLength(ATTACHMENT_MAX_COUNT)
        expect(rejected.map((file) => file.name)).toEqual(['doc.pdf', '10.png', '11.png'])
    })

    it('builds the content URL', () => {
        expect(attachmentURL('a b')).toBe('/api/attachments/a%20b/content')
    })
})
