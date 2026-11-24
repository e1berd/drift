<script setup lang="ts">
import { PhArrowUp, PhStop } from '@phosphor-icons/vue'
import { Button } from '@ui-sdk/vue/button'
import { useI18n } from 'vue-i18n'

const props = defineProps<{
    running: boolean
    stopping: boolean
    canSend: boolean
    sendTitle: string
}>()
const emit = defineEmits<{ send: []; stop: [] }>()
const { t } = useI18n()
</script>

<template>
    <Button
        v-if="props.running"
        class="composer-stop"
        :data-variant="canSend ? undefined : 'destructive'"
        :disabled="stopping"
        :aria-label="t('composer.stop')"
        v-tooltip="t('composer.stop')"
        @click="emit('stop')"
    >
        <PhStop weight="fill" />
    </Button>
    <Button
        v-if="!props.running || canSend"
        class="composer-send"
        data-variant="solid"
        :disabled="!canSend"
        :aria-label="sendTitle"
        v-tooltip="sendTitle"
        @click="emit('send')"
    >
        <PhArrowUp weight="bold" />
    </Button>
</template>
