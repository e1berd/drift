import { ref, readonly } from 'vue'
import { toast } from '@ui-sdk/vue/toast'
import { useI18n } from 'vue-i18n'

import { uploadAttachment } from '@/gen'
import { partitionAttachable } from '@/shared/attachments'
import { errorMessage, reportError } from '@/shared/errors'

/** Uploads the images of a message before it is sent and reports files that cannot go along. */
export function useAttachmentUpload() {
    const { t } = useI18n()
    const uploading = ref(false)

    function announceRejected(files: File[]) {
        if (files.length)
            toast.info(t('composer.imagesOnly'), {
                description: files.map((file) => file.name).join(', '),
            })
    }

    /** Resolves to the attachment ids in file order, or null when an upload failed. */
    async function upload(files: File[]): Promise<string[] | null> {
        const { accepted, rejected } = partitionAttachable(files)
        announceRejected(rejected)
        if (!accepted.length) return []
        uploading.value = true
        try {
            const uploaded = await Promise.all(
                accepted.map((file) =>
                    uploadAttachment({ body: file, query: { name: file.name } }),
                ),
            )
            return uploaded.map(({ data }) => data.id)
        } catch (error) {
            reportError(error)
            toast.error(errorMessage(error))
            return null
        } finally {
            uploading.value = false
        }
    }

    return { uploading: readonly(uploading), upload, announceRejected }
}
