<script setup lang="ts">
import { computed, onBeforeUnmount } from 'vue'
import { PhFile, PhX } from '@phosphor-icons/vue'
import { Button } from '@ui-sdk/vue/button'
import { formatFileSize } from '@ui-sdk/vue/file-input'
import { useI18n } from 'vue-i18n'

const props = defineProps<{ file: File; removable?: boolean }>()
const emit = defineEmits<{ remove: []; open: [url: string] }>()
const { t } = useI18n()

const isImage = computed(() => props.file.type.startsWith('image/'))
const url = computed(() => (isImage.value ? URL.createObjectURL(props.file) : ''))

onBeforeUnmount(() => {
    if (url.value) URL.revokeObjectURL(url.value)
})
</script>

<template>
    <div class="flex items-center gap-1">
        <Button v-if="isImage" :aria-label="file.name" class="thumb" @click="emit('open', url)">
            <img :src="url" alt="" />
        </Button>
        <span v-else class="status-tag">
            <PhFile />
            <span class="max-w-48 truncate">{{ file.name }}</span>
            <span>{{ formatFileSize(file.size) }}</span>
        </span>
        <Button
            v-if="removable"
            :aria-label="t('composer.removeAttachment', { name: file.name })"
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
