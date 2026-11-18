<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { RouterLink } from 'vue-router'

import type { MessageBody } from '@/gen'

defineProps<{ body: MessageBody }>()
const { t } = useI18n()
</script>

<template>
    <div class="flex flex-col items-center gap-1 self-center text-center text-muted">
        <p class="m-0">
            <template v-if="body.code">{{ t(`chat.notes.${body.code}`) }}</template>
            <RouterLink
                v-if="body.code === 'noSubscription'"
                :to="{ name: 'settings' }"
                class="ms-1"
            >
                {{ t('chat.notes.openSettings') }}
            </RouterLink>
        </p>
        <p v-if="body.text" class="m-0 font-mono whitespace-pre-wrap">{{ body.text }}</p>
    </div>
</template>
