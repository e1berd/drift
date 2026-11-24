<script setup lang="ts">
import { mdiFileOutline } from '@mdi/js'

import type { useComposerUploads } from '@drift/vue/composables/useComposerUploads'

defineProps<{ uploads: ReturnType<typeof useComposerUploads> }>()
</script>

<template>
    <div
        v-if="uploads.pending.value.length || uploads.ready.value.length"
        class="d-flex flex-wrap ga-2 px-3 pt-3"
    >
        <v-chip
            v-for="file in uploads.pending.value"
            :key="file.id"
            size="small"
            closable
            @click:close="uploads.removePending(file.id)"
        >
            <template #prepend>
                <v-progress-circular
                    :model-value="file.progress.percentage ?? 0"
                    size="16"
                    width="2"
                    class="me-2"
                />
            </template>
            {{ file.name }}
        </v-chip>
        <v-chip
            v-for="attachment in uploads.ready.value"
            :key="attachment.id"
            size="small"
            :prepend-icon="mdiFileOutline"
            closable
            @click:close="uploads.removeReady(attachment.id)"
            >{{ attachment.name }}</v-chip
        >
    </div>
</template>
