import { describe, expect, it } from 'vitest'

import {
    attachmentIdFromUploadURL,
    attachmentURL,
    isInlineImage,
    uploadEndpoint,
} from '../attachments'

describe('attachments', () => {
    it('shows only the common raster images inline', () => {
        expect(isInlineImage('image/png')).toBe(true)
        expect(isInlineImage('image/webp')).toBe(true)
        expect(isInlineImage('image/svg+xml')).toBe(false)
        expect(isInlineImage('application/zip')).toBe(false)
    })

    it('builds the content URL and the upload endpoint', () => {
        expect(attachmentURL('a b')).toBe('/api/attachments/a%20b/content')
        expect(uploadEndpoint()).toBe('/api/uploads')
    })

    it('takes the attachment id from the tus upload URL', () => {
        const id = '6f1c1d1e-8a5b-4c1b-9d2e-0f1a2b3c4d5e'
        expect(attachmentIdFromUploadURL(`https://drift.example/api/uploads/${id}`)).toBe(id)
        expect(attachmentIdFromUploadURL(`/api/uploads/${id}`)).toBe(id)
        expect(attachmentIdFromUploadURL(undefined)).toBeNull()
    })
})
