import { computed, shallowRef, toValue, watch, watchEffect } from 'vue'
import { RestrictionError } from '@uppy/core'
import { toast } from '@ui-sdk/vue/toast'
import { formatFileSize } from '@ui-sdk/vue/file-input'
import { useI18n } from 'vue-i18n'

import type { MaybeRefOrGetter } from 'vue'

import { useGetSettings } from '@/gen'
import type { ReadyAttachment } from '@/shared/attachments'
import { ATTACHMENT_MAX_COUNT } from '@/shared/attachments'
import { reportError } from '@/shared/errors'
import type { DraftUploads, PendingUpload } from '@/shared/uploads/draftUploads'
import { applyMaxUploadBytes, draftUploads } from '@/shared/uploads/draftUploads'
import { isRefusedForGood, uploadFailureStatus } from '@/shared/uploads/failures'
import { forgetPreview } from '@/shared/uploads/previews'

/** Files of a composer draft: added ones upload at once, with progress, pause and resume. */
export function useComposerUploads(draftKey: MaybeRefOrGetter<string>) {
    const { t } = useI18n()
    const settings = useGetSettings({ query: { staleTime: 60_000 } })
    const maxUploadBytes = computed(() => settings.data.value?.maxUploadBytes ?? null)
    watchEffect(() => {
        if (maxUploadBytes.value) applyMaxUploadBytes(maxUploadBytes.value)
    })

    const draft = computed(() => draftUploads(toValue(draftKey)))
    const pending = shallowRef<PendingUpload[]>([])
    const ready = computed(() => draft.value.ready.value)
    const uploading = computed(() => pending.value.length > 0)

    watch(draft, (current, _previous, onCleanup) => onCleanup(followDraft(current)), {
        immediate: true,
    })

    function followDraft({ uploader }: DraftUploads): () => void {
        const sync = () => {
            pending.value = uploader.getFiles()
        }
        const onUploadError = (file: PendingUpload | undefined, error: Error) => {
            announceFailure(file?.name ?? '', error)
            if (file && isRefusedForGood(error)) setTimeout(() => uploader.removeFile(file.id))
        }
        const onRestriction = (file: PendingUpload | undefined, error: Error) =>
            announceRestriction(file, error)
        sync()
        uploader.on('state-update', sync)
        uploader.on('upload-error', onUploadError)
        uploader.on('restriction-failed', onRestriction)
        return () => {
            uploader.off('state-update', sync)
            uploader.off('upload-error', onUploadError)
            uploader.off('restriction-failed', onRestriction)
        }
    }

    function announceFailure(name: string, error: Error) {
        const status = uploadFailureStatus(error)
        if (status === null) reportError(error)
        if (status === 413) toast.error(t('uploads.tooLarge', { name, max: maxSizeLabel() }))
        else if (status === 507) toast.error(t('uploads.noSpace', { name }))
        else toast.error(t('uploads.failed', { name }), { description: error.message })
    }

    function announceRestriction(file: PendingUpload | undefined, error: Error) {
        const name = file?.name ?? ''
        if (file?.size === 0) toast.info(t('uploads.empty', { name }))
        else if (maxUploadBytes.value && (file?.size ?? 0) > maxUploadBytes.value)
            toast.info(t('uploads.tooLarge', { name, max: maxSizeLabel() }))
        else toast.info(error.message)
    }

    function maxSizeLabel(): string {
        return maxUploadBytes.value ? formatFileSize(maxUploadBytes.value) : ''
    }

    function add(files: File[]) {
        const room = Math.max(0, ATTACHMENT_MAX_COUNT - ready.value.length - pending.value.length)
        const skipped = files.slice(room)
        if (skipped.length)
            toast.info(t('uploads.tooMany', { max: ATTACHMENT_MAX_COUNT }), {
                description: skipped.map((file) => file.name).join(', '),
            })
        for (const file of files.slice(0, room)) addOne(file)
    }

    function addOne(file: File) {
        const { uploader } = draft.value
        try {
            const id = uploader.addFile({
                name: file.name,
                type: file.type,
                data: file,
                source: 'Local',
            })
            const added = uploader.getFile(id)
            if (added?.isRestored && !added.progress.uploadComplete)
                uploader.retryUpload(id).catch(reportError)
        } catch (error) {
            if (!(error instanceof RestrictionError)) reportError(error)
        }
    }

    function removePending(id: string) {
        draft.value.uploader.removeFile(id)
        forgetPreview(id)
    }

    function removeReady(id: string) {
        const list = draft.value.ready
        forgetPreview(list.value.find((attachment) => attachment.id === id)?.uploadKey)
        list.value = list.value.filter((attachment) => attachment.id !== id)
    }

    function togglePause(id: string) {
        draft.value.uploader.pauseResume(id)
    }

    function retry(id: string) {
        draft.value.uploader.retryUpload(id).catch(reportError)
    }

    function takeReady(): ReadyAttachment[] {
        const taken = [...ready.value]
        draft.value.ready.value = []
        for (const attachment of taken) forgetPreview(attachment.uploadKey)
        return taken
    }

    function returnUnsent(key: string, attachments: ReadyAttachment[]) {
        const target = draftUploads(key).ready
        const kept = target.value.filter((item) => !attachments.some((sent) => sent.id === item.id))
        target.value = [...attachments, ...kept]
    }

    return {
        pending,
        ready,
        uploading,
        add,
        removePending,
        removeReady,
        togglePause,
        retry,
        takeReady,
        returnUnsent,
    }
}
