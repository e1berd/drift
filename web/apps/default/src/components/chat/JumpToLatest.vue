<script setup lang="ts">
import { PhArrowDown } from '@phosphor-icons/vue'
import { Button } from '@ui-sdk/vue/button'
import { useI18n } from 'vue-i18n'

defineProps<{ visible: boolean; unread: number }>()
const emit = defineEmits<{ jump: [] }>()
const { t } = useI18n()
</script>

<template>
    <div class="jump-to-latest" :data-visible="visible || undefined" :inert="!visible">
        <Button
            :data-unread="unread || undefined"
            :aria-label="unread ? t('newMessages', { count: unread }) : t('chat.jumpToEnd')"
            v-tooltip="t('chat.jumpToEnd')"
            @click="emit('jump')"
        >
            <span v-if="unread" class="unread-dot" aria-hidden="true" />
            <span v-if="unread">{{ t('newMessages', { count: unread }) }}</span>
            <PhArrowDown />
        </Button>
    </div>
</template>

<style scoped>
.jump-to-latest {
    position: absolute;
    z-index: 2;
    bottom: calc(var(--ui-spacing) * 3);
    left: 50%;
    opacity: 0;
    translate: -50% 8px;
    pointer-events: none;
    transition:
        opacity var(--dur-2) var(--ease-out),
        translate var(--dur-2) var(--ease-out);

    &[data-visible] {
        opacity: 1;
        translate: -50% 0;
        pointer-events: auto;
    }

    & > [data-slot='button'] {
        border: 1px solid var(--ui-border);
        border-radius: 999px;
        background: var(--ui-background);
        box-shadow: var(--ui-shadow);
    }
}

.unread-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--ui-primary);
    animation: drift-blink 1.6s infinite;
}
</style>

<i18n lang="yaml">
ru:
    newMessages: 'Новые: {count}'
en:
    newMessages: '{count} new'
</i18n>
