import type { ReadyAttachment } from '@drift/sdk/attachments'
import type { PendingUpload } from './draftUploads'

/** One attachment in the composer, still uploading or ready; the key survives the switch. */
export type ComposerTile =
    | { key: string; ready: ReadyAttachment; pending?: undefined }
    | { key: string; pending: PendingUpload; ready?: undefined }
