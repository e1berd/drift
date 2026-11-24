<script setup lang="ts">
import { computed } from 'vue'
import { PhWarning } from '@phosphor-icons/vue'
import { useI18n } from 'vue-i18n'

import type { MessageBody } from '@drift/vue/api'
import { formatReset, secondsUntil, usageWindowKind } from '@drift/sdk/usage'

const props = defineProps<{ body: MessageBody }>()
const { t, locale } = useI18n()

const window = computed(() =>
    t(
        `window.${usageWindowKind({ id: props.body.window ?? '', windowMinutes: props.body.windowMinutes })}`,
        {
            minutes: props.body.windowMinutes ?? 0,
        },
    ),
)
const reset = computed(() => {
    const left = secondsUntil(props.body.resetTime)
    return left ? formatReset(left, locale.value) : ''
})
</script>

<template>
    <p class="usage-warning m-0" role="status">
        <PhWarning aria-hidden="true" />
        <span>
            {{ t('left', { percent: body.percentLeft ?? 0, window }) }}
            <span v-if="reset" class="text-muted">· {{ t('resets', { when: reset }) }}</span>
        </span>
    </p>
</template>

<style scoped>
.usage-warning {
    display: flex;
    align-items: center;
    align-self: center;
    gap: calc(var(--ui-spacing) * 2);
}
</style>

<i18n lang="yaml">
ru:
    left: 'Осталось {percent}% от {window}'
    resets: 'сброс {when}'
    window:
        session: 5-часового лимита
        week: недельного лимита
        weekOpus: недельного лимита Opus
        weekSonnet: недельного лимита Sonnet
        other: 'лимита на {minutes} мин'
en:
    left: '{percent}% of the {window} left'
    resets: 'resets {when}'
    window:
        session: 5-hour limit
        week: weekly limit
        weekOpus: weekly Opus limit
        weekSonnet: weekly Sonnet limit
        other: '{minutes}-minute limit'
</i18n>
