<script setup lang="ts">
import { ref } from 'vue'
import { PhDownloadSimple } from '@phosphor-icons/vue'
import { Button } from '@ui-sdk/vue/button'
import { formatFileSize } from '@ui-sdk/vue/file-input'

import AttachmentTile from '@/components/attachments/AttachmentTile.vue'
import ImagePreview from '@/components/composer/ImagePreview.vue'
import type { Attachment } from '@/gen'
import { attachmentURL, isInlineImage } from '@/shared/attachments'

defineProps<{ attachments: Attachment[] }>()

const preview = ref<{ url: string; name: string } | null>(null)
</script>

<template>
    <div class="message-attachments">
        <template v-for="attachment in attachments" :key="attachment.id">
            <AttachmentTile
                v-if="isInlineImage(attachment.type)"
                v-tooltip="attachment.name"
                :name="attachment.name"
                :image-url="attachmentURL(attachment.id)"
            >
                <template #lead>
                    <Button
                        :aria-label="attachment.name"
                        @click="
                            preview = { url: attachmentURL(attachment.id), name: attachment.name }
                        "
                    />
                </template>
            </AttachmentTile>
            <a
                v-else
                v-tooltip="attachment.name"
                :href="attachmentURL(attachment.id)"
                :download="attachment.name"
                class="message-file"
            >
                <AttachmentTile :name="attachment.name" :detail="formatFileSize(attachment.size)">
                    <template #corner>
                        <span class="message-file-download"><PhDownloadSimple /></span>
                    </template>
                </AttachmentTile>
            </a>
        </template>
    </div>
    <ImagePreview v-if="preview" :url="preview.url" :name="preview.name" @close="preview = null" />
</template>

<style scoped>
.message-attachments {
    display: flex;
    max-width: 82%;
    flex-wrap: wrap;
    justify-content: flex-end;
    gap: calc(var(--ui-spacing) * 2);

    & :deep(.tile) {
        --tile-size: 72px;
    }

    & :deep(.tile-lead-image > [data-slot='button']) {
        cursor: zoom-in;
    }
}

.message-file {
    display: contents;
    color: var(--ui-foreground);
    text-decoration: none;

    & > :deep(.tile) {
        transition: background var(--dur-1);
    }

    &:hover > :deep(.tile) {
        background: var(--ui-muted);
    }

    & :deep(.tile-corner) {
        top: 50%;
        right: calc(var(--ui-spacing) * 3);
        translate: 0 -50%;
    }
}

.message-file-download {
    display: grid;
    color: var(--ui-muted-foreground);
}
</style>
