import type { ReadyAttachment } from '@/shared/attachments'
import type { PendingUpload } from '@/shared/uploads/draftUploads'

/** One attachment in the composer, still uploading or ready; the key survives the switch. */
export type ComposerTile =
    | { key: string; ready: ReadyAttachment; pending?: undefined }
    | { key: string; pending: PendingUpload; ready?: undefined }
