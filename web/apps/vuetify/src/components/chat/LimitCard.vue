<script setup lang="ts">
import { useI18n } from 'vue-i18n'

import type { Message } from '@drift/vue/api'
import { useLimitContinuation } from '@drift/vue/composables/useLimitContinuation'

const props = defineProps<{ message: Message }>()
const { t } = useI18n()
const { busy, continueNow } = useLimitContinuation(() => props.message)
</script>

<template>
    <v-alert type="warning" variant="tonal" :title="t('limit.title')">
        {{ message.body.text }}
        <template v-if="message.body.status === 'reached'" #append>
            <v-btn :loading="busy" @click="continueNow">{{ t('limit.continue') }}</v-btn>
        </template>
    </v-alert>
</template>
