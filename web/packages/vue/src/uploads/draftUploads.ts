import Uppy from '@uppy/core'
import GoldenRetriever from '@uppy/golden-retriever'
import Tus from '@uppy/tus'

import type { Body, Meta, UppyFile } from '@uppy/core'
import type { Ref } from 'vue'

import type { ReadyAttachment } from '@drift/sdk/attachments'
import { attachmentIdFromUploadURL, uploadEndpoint } from '@drift/sdk/attachments'
import { reportError } from '@drift/sdk/errors'
import { storedRef } from '../storedRef'

export type Uploader = Uppy<Meta, Body>
export type PendingUpload = UppyFile<Meta, Body>

/** Uploads of one composer draft: files still travelling and attachments ready to send. */
export interface DraftUploads {
    uploader: Uploader
    ready: Ref<ReadyAttachment[]>
}

const CHUNK_BYTES = 64 * 1024 * 1024
const PARALLEL_UPLOADS = 3
const RETRY_DELAYS_MS = [0, 1_000, 3_000, 5_000, 10_000, 20_000, 30_000, 60_000]
const RECOVERY_MS = 7 * 24 * 60 * 60 * 1000

const drafts = new Map<string, DraftUploads>()

/**
 * The uploads of a draft live as long as the page: switching chats keeps them running, and
 * Golden Retriever restores unfinished ones after a reload.
 */
export function draftUploads(draftKey: string): DraftUploads {
    let entry = drafts.get(draftKey)
    if (!entry) {
        entry = createDraftUploads(draftKey)
        drafts.set(draftKey, entry)
    }
    return entry
}

/** Applies the instance upload limit to every draft. */
export function applyMaxUploadBytes(maxFileSize: number): void {
    currentMaxFileSize = maxFileSize
    for (const { uploader } of drafts.values()) setRestrictions(uploader)
}

let currentMaxFileSize: number | null = null

function setRestrictions(uploader: Uploader) {
    uploader.setOptions({ restrictions: { minFileSize: 1, maxFileSize: currentMaxFileSize } })
}

function createDraftUploads(draftKey: string): DraftUploads {
    const ready = storedRef<ReadyAttachment[]>(`drift.attachments.${draftKey}`, [])
    const uploader: Uploader = new Uppy<Meta, Body>({
        id: `drift-uploads-${draftKey}`,
        autoProceed: true,
        allowMultipleUploadBatches: true,
    })
        .use(Tus, {
            endpoint: uploadEndpoint(),
            chunkSize: CHUNK_BYTES,
            limit: PARALLEL_UPLOADS,
            retryDelays: RETRY_DELAYS_MS,
            removeFingerprintOnSuccess: true,
            allowedMetaFields: ['name', 'type'],
        })
        .use(GoldenRetriever, { expires: RECOVERY_MS })
    setRestrictions(uploader)
    uploader.on('upload-success', (file, response) => {
        if (file) moveToReady(uploader, ready, file, response.uploadURL)
    })
    uploader.on('restored', () => resumeRestored(uploader, ready))
    globalThis.addEventListener?.('online', () => retryFailed(uploader))
    return { uploader, ready }
}

function moveToReady(
    uploader: Uploader,
    ready: Ref<ReadyAttachment[]>,
    file: PendingUpload,
    uploadURL: string | undefined,
) {
    const id = attachmentIdFromUploadURL(uploadURL ?? file.tus?.uploadUrl ?? undefined)
    if (id && !ready.value.some((attachment) => attachment.id === id))
        ready.value.push({
            id,
            name: file.name ?? id,
            size: file.size ?? 0,
            type: file.type || 'application/octet-stream',
            uploadKey: file.id,
        })
    setTimeout(() => {
        if (uploader.getFile(file.id)) uploader.removeFile(file.id)
    })
}

function resumeRestored(uploader: Uploader, ready: Ref<ReadyAttachment[]>) {
    uploader.setState({ currentUploads: {}, recoveredState: null })
    for (const file of uploader.getFiles()) {
        if (file.progress.uploadComplete)
            moveToReady(uploader, ready, file, file.response?.uploadURL)
        else if (!file.isGhost) uploader.retryUpload(file.id).catch(reportError)
    }
}

function retryFailed(uploader: Uploader) {
    for (const file of uploader.getFiles()) {
        if (file.error && !file.isGhost) uploader.retryUpload(file.id).catch(reportError)
    }
}
