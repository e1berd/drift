<script setup lang="ts">
import { computed } from 'vue'

import ComposerAttachmentTile from '@/components/composer/ComposerAttachmentTile.vue'
import type { ComposerTile } from '@drift/vue/uploads/composerTile'
import type { useComposerUploads } from '@drift/vue/composables/useComposerUploads'

const props = defineProps<{ uploads: ReturnType<typeof useComposerUploads> }>()
const emit = defineEmits<{ open: [url: string, name: string] }>()

const tiles = computed<ComposerTile[]>(() => {
    const ready = props.uploads.ready.value
    const finished = new Set(ready.map((attachment) => attachment.uploadKey))
    return [
        ...ready.map((attachment) => ({
            key: attachment.uploadKey ?? attachment.id,
            ready: attachment,
        })),
        ...props.uploads.pending.value
            .filter((upload) => !finished.has(upload.id))
            .map((upload) => ({ key: upload.id, pending: upload })),
    ]
})
</script>

<template>
    <div v-if="tiles.length" class="composer-attachments">
        <ComposerAttachmentTile
            v-for="tile in tiles"
            :key="tile.key"
            :tile="tile"
            @remove="
                tile.ready
                    ? uploads.removeReady(tile.ready.id)
                    : uploads.removePending(tile.pending.id)
            "
            @toggle-pause="tile.pending && uploads.togglePause(tile.pending.id)"
            @retry="tile.pending && uploads.retry(tile.pending.id)"
            @reselect="(file) => uploads.add([file])"
            @open="(url, name) => emit('open', url, name)"
        />
    </div>
</template>

<style scoped>
.composer-attachments {
    display: flex;
    flex-wrap: wrap;
    align-items: flex-start;
    gap: calc(var(--ui-spacing) * 2);
    padding: calc(var(--ui-spacing) * 3) calc(var(--ui-spacing) * 3) 0;
}
</style>
