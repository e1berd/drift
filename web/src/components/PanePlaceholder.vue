<script setup lang="ts">
import type { Component } from 'vue'

import { useI18n } from 'vue-i18n'

defineProps<{ icon: Component }>()
const { t } = useI18n()
</script>

<template>
    <div class="grid flex-1 place-items-center p-6">
        <div
            class="flex max-w-[320px] flex-col items-center gap-3 text-center text-[13px] text-faint"
        >
            <span class="grid h-11 w-11 place-items-center rounded-xl bg-surface-2 text-muted">
                <component :is="icon" :size="22" />
            </span>
            {{ t('tabs.soon') }}
        </div>
    </div>
</template>
