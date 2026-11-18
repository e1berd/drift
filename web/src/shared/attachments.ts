import { API_BASE } from '@/shared/http/client'

/** Image types the server accepts as attachments (it checks the bytes, not the name). */
export const ATTACHMENT_TYPES = ['image/png', 'image/jpeg', 'image/gif', 'image/webp'] as const

/** Largest image the server accepts. */
export const ATTACHMENT_MAX_BYTES = 5 * 1024 * 1024

/** Largest number of images one message can carry. */
export const ATTACHMENT_MAX_COUNT = 10

/** Whether a picked, pasted or dropped file can be attached to a message. */
export function isAttachable(file: File): boolean {
    return (
        (ATTACHMENT_TYPES as readonly string[]).includes(file.type) &&
        file.size > 0 &&
        file.size <= ATTACHMENT_MAX_BYTES
    )
}

/** Splits files into those that can be attached (up to the per-message limit) and the rest. */
export function partitionAttachable(files: File[]): { accepted: File[]; rejected: File[] } {
    const accepted: File[] = []
    const rejected: File[] = []
    for (const file of files) {
        if (isAttachable(file) && accepted.length < ATTACHMENT_MAX_COUNT) accepted.push(file)
        else rejected.push(file)
    }
    return { accepted, rejected }
}

/** URL of an attachment's image bytes. */
export function attachmentURL(id: string): string {
    const url = new URL(
        `${API_BASE}/attachments/${encodeURIComponent(id)}/content`,
        globalThis.location?.origin ?? 'http://localhost',
    )
    return url.pathname
}
