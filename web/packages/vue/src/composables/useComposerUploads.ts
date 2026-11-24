import { computed, shallowRef, toValue, watch, watchEffect } from 'vue'
import { RestrictionError } from '@uppy/core'
import { useI18n } from 'vue-i18n'

import type { MaybeRefOrGetter } from 'vue'

import { useGetSettings } from '../api'
import type { ReadyAttachment } from '@drift/sdk/attachments'
import { ATTACHMENT_MAX_COUNT } from '@drift/sdk/attachments'
import { reportError } from '@drift/sdk/errors'
import type { DraftUploads, PendingUpload } from '../uploads/draftUploads'
import { applyMaxUploadBytes, draftUploads } from '../uploads/draftUploads'
import { isRefusedForGood, uploadFailureStatus } from '@drift/sdk/uploads/failures'
import { forgetPreview } from '@drift/sdk/uploads/previews'
import { formatFileSize } from '@drift/sdk/fileSize'
import { useNotifier } from '../notifier'

/** Files of a composer draft: added ones upload at once, with progress, pause and resume. */
export function useComposerUploads(draftKey: MaybeRefOrGetter<string>) {
    const { t } = useI18n()
    const notifier = useNotifier()
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
        if (status === 413)
            notifier.error(t('drift.uploads.tooLarge', { name, max: maxSizeLabel() }))
        else if (status === 507) notifier.error(t('drift.uploads.noSpace', { name }))
        else notifier.error(t('drift.uploads.failed', { name }), { description: error.message })
    }

    function announceRestriction(file: PendingUpload | undefined, error: Error) {
        const name = file?.name ?? ''
        if (file?.size === 0) notifier.info(t('drift.uploads.empty', { name }))
        else if (maxUploadBytes.value && (file?.size ?? 0) > maxUploadBytes.value)
            notifier.info(t('drift.uploads.tooLarge', { name, max: maxSizeLabel() }))
        else notifier.info(error.message)
    }

    function maxSizeLabel(): string {
        return maxUploadBytes.value ? formatFileSize(maxUploadBytes.value) : ''
    }

    function add(files: File[]) {
        const room = Math.max(0, ATTACHMENT_MAX_COUNT - ready.value.length - pending.value.length)
        const skipped = files.slice(room)
        if (skipped.length)
            notifier.info(t('drift.uploads.tooMany', { max: ATTACHMENT_MAX_COUNT }), {
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
