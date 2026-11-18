<script setup lang="ts">
import { ref } from 'vue'
import { Button } from '@ui-sdk/vue/button'

import ImagePreview from '@/components/composer/ImagePreview.vue'
import type { Attachment } from '@/gen'
import { attachmentURL } from '@/shared/attachments'

defineProps<{ attachments: Attachment[] }>()

const preview = ref<{ url: string; name: string } | null>(null)
</script>

<template>
    <div class="flex flex-wrap justify-end gap-2">
        <Button
            v-for="attachment in attachments"
            :key="attachment.id"
            :aria-label="attachment.name"
            :title="attachment.name"
            class="thumb"
            @click="preview = { url: attachmentURL(attachment.id), name: attachment.name }"
        >
            <img :src="attachmentURL(attachment.id)" alt="" loading="lazy" decoding="async" />
        </Button>
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
</style>
