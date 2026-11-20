import { API_BASE } from '@/shared/http/client'

/** Largest number of files one message can carry. */
export const ATTACHMENT_MAX_COUNT = 10

const INLINE_IMAGE_TYPES: readonly string[] = ['image/png', 'image/jpeg', 'image/gif', 'image/webp']

/** A finished upload waiting in the composer for its message. */
export interface ReadyAttachment {
    id: string
    name: string
    size: number
    type: string
}

/** Whether the server shows the attachment inline as an image rather than offering a download. */
export function isInlineImage(type: string): boolean {
    return INLINE_IMAGE_TYPES.includes(type)
}

function apiPath(path: string): string {
    return new URL(`${API_BASE}${path}`, globalThis.location?.origin ?? 'http://localhost').pathname
}

/** URL of an attachment's bytes. */
export function attachmentURL(id: string): string {
    return apiPath(`/attachments/${encodeURIComponent(id)}/content`)
}

/** The tus endpoint new uploads are created at. */
export function uploadEndpoint(): string {
    return apiPath('/uploads')
}

/** The attachment id the server gives a finished upload: the last segment of its tus URL. */
export function attachmentIdFromUploadURL(uploadURL: string | undefined): string | null {
    if (!uploadURL) return null
    const id = new URL(uploadURL, 'http://localhost').pathname.split('/').at(-1)
    return id || null
}
