<script setup lang="ts">
import { computed } from 'vue'
import { PhTextAlignLeft, PhX } from '@phosphor-icons/vue'
import { Button } from '@ui-sdk/vue/button'
import { HoverCard, HoverCardContent, HoverCardTrigger } from '@ui-sdk/vue/hover-card'
import { useI18n } from 'vue-i18n'

import ScrollPane from '@/components/ScrollPane.vue'

import { countLines } from '@drift/vue/editor/pastedText'

const props = defineProps<{ text: string; removable: boolean }>()
const emit = defineEmits<{ remove: [] }>()
const { t } = useI18n()

const lines = computed(() => countLines(props.text))
const label = computed(() => t('label', { lines: lines.value }, lines.value))
</script>

<template>
    <span class="pasted">
        <HoverCard :delay="200" :close-delay="150">
            <HoverCardTrigger>
                <template #render="{ props: trigger }">
                    <span v-bind="trigger" class="pasted-chip" :aria-label="label">
                        <PhTextAlignLeft aria-hidden="true" />
                        <span>{{ label }}</span>
                    </span>
                </template>
            </HoverCardTrigger>
            <HoverCardContent side="top" align="start" class="pasted-card">
                <ScrollPane class="pasted-scroll" :label="t('preview')">
                    <pre class="pasted-text">{{ text }}</pre>
                </ScrollPane>
            </HoverCardContent>
        </HoverCard>
        <Button
            v-if="removable"
            class="pasted-remove"
            :aria-label="t('remove')"
            @mousedown.prevent
            @click="emit('remove')"
        >
            <PhX aria-hidden="true" />
        </Button>
    </span>
</template>

<style scoped>
.pasted {
    display: inline-flex;
    align-items: center;
    gap: 2px;
    margin-inline: 2px;
    padding: 1px 2px 1px 8px;
    border: 1px dashed var(--ui-input);
    border-radius: 999px;
    background: var(--ui-muted);
    color: var(--ui-muted-foreground);
    font-size: calc(var(--ui-font-size) - 0.0625rem);
    line-height: 1.6;
    user-select: none;

    :global([data-selected]) > & {
        outline: 2px solid var(--ui-ring);
        outline-offset: 1px;
    }
}

.pasted-chip {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    cursor: default;
}

.pasted-remove {
    width: auto;
    height: auto;
    min-height: 0;
    padding: 2px;
    border-radius: 999px;

    & svg {
        width: 12px;
        height: 12px;
    }
}

.pasted-card {
    width: min(40rem, 90vw);
    max-width: none;
    padding: 0;
}

.pasted-scroll {
    max-height: min(22rem, 50vh);
}

.pasted-text {
    margin: 0;
    padding: 8px 12px;
    font-family: var(--ui-font-mono, monospace);
    font-size: calc(var(--ui-font-size) - 0.125rem);
    white-space: pre-wrap;
    overflow-wrap: anywhere;
}
</style>

<i18n lang="yaml">
ru:
    label: 'Вставленный текст · {lines} стр.'
    preview: Вставленный текст
    remove: Убрать вставленный текст
en:
    label: 'Pasted text · {lines} line | Pasted text · {lines} lines'
    preview: Pasted text
    remove: Remove pasted text
</i18n>
