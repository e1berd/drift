<script setup lang="ts">
import { computed, onBeforeUnmount } from 'vue'
import { PhX } from '@phosphor-icons/vue'

import { useI18n } from '@/shared/i18n'

const props = defineProps<{ file: File; removable?: boolean }>()
const emit = defineEmits<{ remove: []; open: [url: string] }>()
const { t } = useI18n()

const isImage = computed(() => props.file.type.startsWith('image/'))
const url = computed(() => (isImage.value ? URL.createObjectURL(props.file) : ''))
const ext = computed(() => props.file.name.split('.').pop()?.slice(0, 4) ?? '')
const size = computed(() => {
    const b = props.file.size
    if (b < 1024) return `${b} B`
    if (b < 1024 * 1024) return `${(b / 1024).toFixed(1)} KB`
    return `${(b / 1024 / 1024).toFixed(1)} MB`
})

onBeforeUnmount(() => {
    if (url.value) URL.revokeObjectURL(url.value)
})
</script>

<template>
    <div class="att">
        <button
            v-if="isImage"
            type="button"
            class="thumb"
            :aria-label="file.name"
            @click="emit('open', url)"
        >
            <img :src="url" alt="" />
        </button>
        <span v-else class="fchip">
            <span class="fic">{{ ext }}</span>
            <span class="min-w-0 leading-tight">
                <b class="block truncate text-[12.5px] font-medium">{{ file.name }}</b>
                <span class="text-[11.5px] text-faint">{{ size }}</span>
            </span>
        </span>
        <button
            v-if="removable"
            type="button"
            class="att-x"
            :aria-label="t('composer.removeAttachment', { name: file.name })"
            @click="emit('remove')"
        >
            <PhX :size="12" weight="bold" />
        </button>
    </div>
</template>

<style scoped>
.att {
    position: relative;
}

.thumb {
    display: block;
    width: 76px;
    height: 58px;
    overflow: hidden;
    padding: 0;
    border: 1px solid var(--border-strong);
    border-radius: 9px;
    background: var(--surface-2);
    cursor: zoom-in;

    & img {
        width: 100%;
        height: 100%;
        object-fit: cover;
    }
}

.fchip {
    display: flex;
    align-items: center;
    gap: 10px;
    max-width: 230px;
    height: 58px;
    padding: 0 32px 0 10px;
    border: 1px solid var(--border-strong);
    border-radius: 9px;
    background: var(--surface-2);
}

.fic {
    display: grid;
    place-items: center;
    width: 34px;
    height: 34px;
    flex: none;
    border-radius: 8px;
    background: var(--accent-soft);
    color: var(--accent);
    font: 600 9.5px var(--ui-font-mono);
    text-transform: uppercase;
}

.att-x {
    position: absolute;
    top: 4px;
    right: 4px;
    display: grid;
    place-items: center;
    width: 22px;
    height: 22px;
    border: 0;
    border-radius: 50%;
    background: rgb(10 11 13 / 0.72);
    color: #fff;
    opacity: 0;
    cursor: pointer;
    transition: opacity var(--dur-1);
}

.fchip + .att-x {
    top: 50%;
    margin-top: -11px;
    background: var(--surface-3);
    color: var(--ui-muted-foreground);
}

.att:is(:hover, :focus-within) .att-x {
    opacity: 1;
}

@media (hover: none) {
    .att-x {
        opacity: 1;
    }
}
</style>
