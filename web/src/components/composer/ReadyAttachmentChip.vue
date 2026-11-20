<script setup lang="ts">
import { computed } from 'vue'
import { PhFile, PhX } from '@phosphor-icons/vue'
import { Button } from '@ui-sdk/vue/button'
import { formatFileSize } from '@ui-sdk/vue/file-input'
import { useI18n } from 'vue-i18n'

import type { ReadyAttachment } from '@/shared/attachments'
import { attachmentURL, isInlineImage } from '@/shared/attachments'

const props = defineProps<{ attachment: ReadyAttachment }>()
const emit = defineEmits<{ remove: []; open: [url: string] }>()
const { t } = useI18n()

const url = computed(() => attachmentURL(props.attachment.id))
</script>

<template>
    <div class="flex items-center gap-1">
        <Button
            v-if="isInlineImage(attachment.type)"
            :aria-label="attachment.name"
            v-tooltip="attachment.name"
            class="thumb"
            @click="emit('open', url)"
        >
            <img :src="url" alt="" />
        </Button>
        <span v-else class="status-tag" v-tooltip="attachment.name">
            <PhFile />
            <span class="max-w-48 truncate">{{ attachment.name }}</span>
            <span>{{ formatFileSize(attachment.size) }}</span>
        </span>
        <Button
            :aria-label="t('composer.removeAttachment', { name: attachment.name })"
            @click="emit('remove')"
        >
            <PhX />
        </Button>
    </div>
</template>

<style scoped>
.thumb {
    width: auto;
    height: 48px;
    padding: 0;
    overflow: hidden;

    & img {
        height: 100%;
    }
}
</style>
