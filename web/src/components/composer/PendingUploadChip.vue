<script setup lang="ts">
import { computed, useTemplateRef } from 'vue'
import {
    PhArrowClockwise,
    PhFolderOpen,
    PhPause,
    PhPlay,
    PhWarningCircle,
    PhX,
} from '@phosphor-icons/vue'
import { Button } from '@ui-sdk/vue/button'
import { formatFileSize } from '@ui-sdk/vue/file-input'
import { Progress, ProgressIndicator, ProgressTrack } from '@ui-sdk/vue/progress'
import { useI18n } from 'vue-i18n'

import type { PendingUpload } from '@/shared/uploads/draftUploads'

const props = defineProps<{ upload: PendingUpload }>()
const emit = defineEmits<{
    remove: []
    togglePause: []
    retry: []
    reselect: [file: File]
}>()
const { t } = useI18n()
const picker = useTemplateRef<HTMLInputElement>('picker')

const name = computed(() => props.upload.name ?? '')
const total = computed(() => props.upload.size ?? 0)
const sent = computed(() => props.upload.progress.bytesUploaded || 0)
const percent = computed(() => (total.value ? Math.min(100, (sent.value / total.value) * 100) : 0))
const state = computed(() => {
    if (props.upload.isGhost) return 'ghost'
    if (props.upload.error) return 'failed'
    if (props.upload.isPaused) return 'paused'
    return 'uploading'
})

function onPicked(event: Event) {
    const input = event.target as HTMLInputElement
    const file = input.files?.item(0)
    if (file) emit('reselect', file)
    input.value = ''
}
</script>

<template>
    <div class="pending" :data-state="state" v-tooltip="name">
        <div class="flex min-w-0 items-center gap-1">
            <PhWarningCircle v-if="state === 'failed'" class="shrink-0 text-destructive" />
            <span class="min-w-0 flex-1 truncate">{{ name }}</span>
            <Button
                v-if="state === 'ghost'"
                :aria-label="t('uploads.reselect', { name })"
                v-tooltip="t('uploads.reselect', { name })"
                @click="picker?.click()"
            >
                <PhFolderOpen />
            </Button>
            <Button
                v-else-if="state === 'failed'"
                :aria-label="t('uploads.retry')"
                v-tooltip="t('uploads.retry')"
                @click="emit('retry')"
            >
                <PhArrowClockwise />
            </Button>
            <Button
                v-else
                :aria-label="state === 'paused' ? t('uploads.resume') : t('uploads.pause')"
                v-tooltip="state === 'paused' ? t('uploads.resume') : t('uploads.pause')"
                @click="emit('togglePause')"
            >
                <PhPlay v-if="state === 'paused'" />
                <PhPause v-else />
            </Button>
            <Button
                :aria-label="t('composer.removeAttachment', { name })"
                v-tooltip="t('composer.removeAttachment', { name })"
                @click="emit('remove')"
            >
                <PhX />
            </Button>
        </div>
        <Progress :value="percent" :max="100" :aria-label="name">
            <ProgressTrack><ProgressIndicator /></ProgressTrack>
        </Progress>
        <span class="status-tag">
            {{
                state === 'ghost'
                    ? t('uploads.ghost')
                    : t('uploads.progress', {
                          sent: formatFileSize(sent),
                          total: formatFileSize(total),
                          percent: Math.floor(percent),
                      })
            }}
            <template v-if="state === 'paused'"> · {{ t('uploads.paused') }}</template>
        </span>
        <input ref="picker" type="file" hidden @change="onPicked" />
    </div>
</template>

<style scoped>
.pending {
    display: flex;
    flex-direction: column;
    gap: 4px;
    width: min(260px, 100%);
    padding: calc(var(--ui-spacing) * 1.5);
    border: 1px solid var(--ui-border);
    border-radius: var(--ui-radius);

    &[data-state='failed'] {
        border-color: var(--ui-destructive);
    }

    &[data-state='ghost'] {
        border-style: dashed;
    }
}
</style>
