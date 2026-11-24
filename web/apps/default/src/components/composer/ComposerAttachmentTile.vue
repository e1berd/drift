<script setup lang="ts">
import { computed, useTemplateRef } from 'vue'
import {
    PhArrowClockwise,
    PhArrowsOut,
    PhFolderOpen,
    PhPause,
    PhPlay,
    PhX,
} from '@phosphor-icons/vue'
import { Button } from '@ui-sdk/vue/button'
import { formatFileSize } from '@ui-sdk/vue/file-input'
import { useI18n } from 'vue-i18n'

import AttachmentTile from '@/components/attachments/AttachmentTile.vue'
import type { AttachmentState } from '@drift/sdk/attachmentState'
import type { ComposerTile } from '@drift/vue/uploads/composerTile'
import { attachmentURL, isInlineImage } from '@drift/sdk/attachments'
import { knownPreview, localPreview } from '@drift/sdk/uploads/previews'

const props = defineProps<{ tile: ComposerTile }>()
const emit = defineEmits<{
    remove: []
    togglePause: []
    retry: []
    reselect: [file: File]
    open: [url: string, name: string]
}>()
const { t } = useI18n()
const picker = useTemplateRef<HTMLInputElement>('picker')

const name = computed(() => props.tile.ready?.name ?? props.tile.pending?.name ?? '')
const type = computed(() => props.tile.ready?.type ?? props.tile.pending?.type ?? '')
const total = computed(() => props.tile.ready?.size ?? props.tile.pending?.size ?? 0)
const sent = computed(() => props.tile.pending?.progress.bytesUploaded || 0)
const percent = computed(() => {
    if (props.tile.ready) return 100
    return total.value ? Math.min(100, (sent.value / total.value) * 100) : 0
})

const state = computed<AttachmentState>(() => {
    const upload = props.tile.pending
    if (!upload) return 'ready'
    if (upload.isGhost) return 'ghost'
    if (upload.error) return 'failed'
    if (upload.isPaused) return 'paused'
    return 'uploading'
})

const imageUrl = computed(() => {
    if (!isInlineImage(type.value)) return undefined
    const { ready, pending } = props.tile
    if (pending) return localPreview(pending.id, pending.data)
    return knownPreview(ready?.uploadKey) ?? (ready ? attachmentURL(ready.id) : undefined)
})

const detail = computed(() => {
    switch (state.value) {
        case 'ghost':
            return t('ghostShort')
        case 'failed':
            return t('failedShort')
        case 'paused':
            return `${t('uploads.paused')} · ${Math.floor(percent.value)}%`
        case 'uploading':
            return `${formatFileSize(sent.value)} / ${formatFileSize(total.value)}`
        default:
            return formatFileSize(total.value)
    }
})

const leadAction = computed(() => {
    switch (state.value) {
        case 'ghost':
            return { label: t('uploads.reselect', { name: name.value }), icon: PhFolderOpen }
        case 'failed':
            return { label: t('uploads.retry'), icon: PhArrowClockwise }
        case 'paused':
            return { label: t('uploads.resume'), icon: PhPlay }
        case 'uploading':
            return { label: t('uploads.pause'), icon: PhPause }
        default:
            return imageUrl.value ? { label: name.value, icon: PhArrowsOut } : null
    }
})

function runLeadAction() {
    switch (state.value) {
        case 'ghost':
            return picker.value?.click()
        case 'failed':
            return emit('retry')
        case 'paused':
        case 'uploading':
            return emit('togglePause')
        default:
            if (props.tile.ready) emit('open', attachmentURL(props.tile.ready.id), name.value)
    }
}

function onPicked(event: Event) {
    const input = event.target as HTMLInputElement
    const file = input.files?.item(0)
    if (file) emit('reselect', file)
    input.value = ''
}
</script>

<template>
    <AttachmentTile
        v-tooltip="state === 'ghost' ? t('uploads.ghost') : name"
        :name="name"
        :state="state"
        :image-url="imageUrl"
        :percent="percent"
        :detail="detail"
        :data-action="leadAction ? (state === 'ready' ? 'open' : 'control') : undefined"
    >
        <template #lead>
            <Button v-if="leadAction" :aria-label="leadAction.label" @click="runLeadAction">
                <component :is="leadAction.icon" />
            </Button>
            <input ref="picker" type="file" hidden @change="onPicked" />
        </template>
        <template #corner>
            <Button :aria-label="t('composer.removeAttachment', { name })" @click="emit('remove')">
                <PhX />
            </Button>
        </template>
    </AttachmentTile>
</template>

<style scoped>
:is([data-action='open'], [data-state='uploading'])[data-kind='image']
    :deep(.tile-lead-image > [data-slot='button']) {
    opacity: 0;
    transition: opacity var(--dur-1);

    &:is(:hover, :focus-visible) {
        opacity: 1;
    }
}
</style>

<i18n lang="yaml">
ru:
    ghostShort: Выберите файл ещё раз
    failedShort: Не загрузился
en:
    ghostShort: Pick the file again
    failedShort: Upload failed
</i18n>
