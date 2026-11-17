<script setup lang="ts">
import { ref } from 'vue'
import { PhCaretRight } from '@phosphor-icons/vue'

import type { MessageBody } from '@/gen'
import { useI18n } from '@/shared/i18n'

const props = defineProps<{ body: MessageBody }>()
const { t } = useI18n()
const open = ref(props.body.status === 'running')
</script>

<template>
    <div class="ml-9">
        <button type="button" class="head" :aria-expanded="open" @click="open = !open">
            <PhCaretRight :size="14" class="chev" :data-open="open || undefined" />
            <span v-if="body.status === 'running'" class="motion-shine">{{
                t('chat.thinking')
            }}</span>
            <span v-else>{{ t('chat.thought') }} {{ body.meta }}</span>
        </button>
        <div class="grid-reveal" :data-open="open || undefined">
            <div>
                <p
                    class="m-0 py-1 pr-3 pl-[26px] text-[13px] leading-relaxed whitespace-pre-wrap text-muted italic"
                >
                    {{ body.text }}
                </p>
            </div>
        </div>
    </div>
</template>

<style scoped>
.head {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    height: 28px;
    padding: 0 8px 0 4px;
    border: 0;
    border-radius: 7px;
    background: none;
    color: var(--faint);
    font-size: 12.5px;
    cursor: pointer;
    transition:
        color var(--dur-1),
        background var(--dur-1);

    &:hover {
        background: var(--surface-2);
        color: var(--ui-foreground);
    }
}

.chev {
    transition: transform var(--dur-3) var(--ease-out);

    &[data-open] {
        transform: rotate(90deg);
    }
}
</style>
