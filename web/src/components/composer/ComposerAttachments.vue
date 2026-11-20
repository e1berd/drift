<script setup lang="ts">
import PendingUploadChip from '@/components/composer/PendingUploadChip.vue'
import ReadyAttachmentChip from '@/components/composer/ReadyAttachmentChip.vue'
import type { useComposerUploads } from '@/composables/useComposerUploads'

defineProps<{ uploads: ReturnType<typeof useComposerUploads> }>()
const emit = defineEmits<{ open: [url: string, name: string] }>()
</script>

<template>
    <div
        v-if="uploads.ready.value.length || uploads.pending.value.length"
        v-auto-animate
        class="flex flex-wrap items-start gap-2 px-3 pt-3"
    >
        <ReadyAttachmentChip
            v-for="attachment in uploads.ready.value"
            :key="attachment.id"
            :attachment="attachment"
            @remove="uploads.removeReady(attachment.id)"
            @open="(url) => emit('open', url, attachment.name)"
        />
        <PendingUploadChip
            v-for="upload in uploads.pending.value"
            :key="upload.id"
            :upload="upload"
            @remove="uploads.removePending(upload.id)"
            @toggle-pause="uploads.togglePause(upload.id)"
            @retry="uploads.retry(upload.id)"
            @reselect="(file) => uploads.add([file])"
        />
    </div>
</template>
