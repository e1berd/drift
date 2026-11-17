<script setup lang="ts">
import { Dialog, DialogClose, DialogContent, DialogTitle } from '@ui-sdk/vue/dialog'

import { useI18n } from 'vue-i18n'

defineProps<{ url: string; name: string }>()
const emit = defineEmits<{ close: [] }>()
const { t } = useI18n()
</script>

<template>
    <Dialog :open="true" @open-change="(open: boolean) => !open && emit('close')">
        <DialogContent class="lightbox">
            <img
                :src="url"
                :alt="name"
                class="max-h-[80vh] max-w-full rounded-xl shadow-(--shadow)"
            />
            <div class="flex items-center justify-between gap-4 text-[13px] text-[#ecedef]">
                <DialogTitle class="truncate font-medium">{{ name }}</DialogTitle>
                <DialogClose class="btn h-7">{{ t('toast.close') }}</DialogClose>
            </div>
        </DialogContent>
    </Dialog>
</template>

<style>
.lightbox {
    display: flex;
    max-width: min(900px, calc(100vw - 48px));
    flex-direction: column;
    gap: 10px;
    padding: 0;
    border: 0;
    background: transparent;
    transition:
        opacity var(--dur-3) var(--ease-out),
        scale var(--dur-4) var(--ease-out);

    &::backdrop {
        background: rgb(6 7 9 / 0.72);
    }

    &:is([data-starting-style], [data-ending-style]) {
        opacity: 0;
        scale: 0.97;
    }
}
</style>
