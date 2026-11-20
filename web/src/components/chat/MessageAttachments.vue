<script setup lang="ts">
import { computed, ref } from 'vue'
import { PhDownloadSimple } from '@phosphor-icons/vue'
import { Button } from '@ui-sdk/vue/button'
import { formatFileSize } from '@ui-sdk/vue/file-input'

import ImagePreview from '@/components/composer/ImagePreview.vue'
import type { Attachment } from '@/gen'
import { attachmentURL, isInlineImage } from '@/shared/attachments'

const props = defineProps<{ attachments: Attachment[] }>()

const preview = ref<{ url: string; name: string } | null>(null)
const images = computed(() => props.attachments.filter((item) => isInlineImage(item.type)))
const files = computed(() => props.attachments.filter((item) => !isInlineImage(item.type)))
</script>

<template>
    <div class="flex flex-col items-end gap-2">
        <div v-if="images.length" class="flex flex-wrap justify-end gap-2">
            <Button
                v-for="attachment in images"
                :key="attachment.id"
                :aria-label="attachment.name"
                :title="attachment.name"
                class="thumb"
                @click="preview = { url: attachmentURL(attachment.id), name: attachment.name }"
            >
                <img :src="attachmentURL(attachment.id)" alt="" loading="lazy" decoding="async" />
            </Button>
        </div>
        <a
            v-for="attachment in files"
            :key="attachment.id"
            :href="attachmentURL(attachment.id)"
            :download="attachment.name"
            v-tooltip="attachment.name"
            class="file"
        >
            <PhDownloadSimple class="shrink-0" />
            <span class="min-w-0 truncate">{{ attachment.name }}</span>
            <span class="shrink-0 text-muted">{{ formatFileSize(attachment.size) }}</span>
        </a>
    </div>
    <ImagePreview v-if="preview" :url="preview.url" :name="preview.name" @close="preview = null" />
</template>

<style scoped>
.thumb {
    width: auto;
    height: 96px;
    padding: 0;
    overflow: hidden;

    & img {
        height: 100%;
        max-width: 240px;
        object-fit: cover;
    }
}

.file {
    display: inline-flex;
    align-items: center;
    gap: calc(var(--ui-spacing) * 2);
    max-width: min(360px, 100%);
    padding: calc(var(--ui-spacing) * 1.5) calc(var(--ui-spacing) * 2.5);
    border: 1px solid var(--ui-border);
    border-radius: var(--ui-radius);
    color: var(--ui-foreground);
    text-decoration: none;

    &:hover {
        background: var(--ui-muted);
    }
}
</style>
