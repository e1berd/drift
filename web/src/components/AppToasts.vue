<script setup lang="ts">
import { PhX } from '@phosphor-icons/vue'
import {
    Toast,
    ToastAction,
    ToastClose,
    ToastDescription,
    ToastRegion,
    ToastTitle,
} from '@ui-sdk/vue/toast'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
</script>

<template>
    <ToastRegion side="top" align="end">
        <Toast>
            <ToastTitle />
            <ToastDescription />
            <ToastAction />
            <ToastClose :aria-label="t('toast.close')"><PhX :size="14" /></ToastClose>
        </Toast>
    </ToastRegion>
</template>
