<script setup lang="ts">
import { Dialog, DialogClose, DialogContent, DialogTitle } from '@ui-sdk/vue/dialog'
import { useI18n } from 'vue-i18n'

defineProps<{ url: string; name: string }>()
const emit = defineEmits<{ close: [] }>()
const { t } = useI18n()

function closeWhenHidden(open: boolean) {
    if (!open) emit('close')
}
</script>

<template>
    <Dialog :open="true" @open-change="closeWhenHidden">
        <DialogContent>
            <DialogTitle>{{ name }}</DialogTitle>
            <img :src="url" :alt="name" class="max-h-[70vh] max-w-full" />
            <DialogClose>{{ t('toast.close') }}</DialogClose>
        </DialogContent>
    </Dialog>
</template>
