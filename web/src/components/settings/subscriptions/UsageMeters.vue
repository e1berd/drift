<script setup lang="ts">
import { computed } from 'vue'
import { Meter, MeterIndicator, MeterLabel, MeterTrack, MeterValue } from '@ui-sdk/vue/meter'
import { useI18n } from 'vue-i18n'

import { useGetSubscriptionUsage } from '@/gen'
import { formatReset, secondsUntil, usageWindowKind } from '@/shared/usage'

const props = defineProps<{ subscriptionId: string }>()
const { t, locale } = useI18n()

const usage = useGetSubscriptionUsage(
    { path: computed(() => ({ subscriptionId: props.subscriptionId })) },
    { query: { refetchInterval: 60_000, staleTime: 30_000 } },
)

const windows = computed(() =>
    (usage.data.value?.windows ?? []).map((window) => {
        const left = secondsUntil(window.resetTime)
        return {
            ...window,
            label: t(`window.${usageWindowKind(window)}`, { minutes: window.windowMinutes ?? 0 }),
            reset: left === null ? '' : t('resets', { when: formatReset(left, locale.value) }),
        }
    }),
)
</script>

<template>
    <div v-if="usage.isPending.value" class="skeleton h-8" />
    <p v-else-if="usage.data.value?.unavailable" class="m-0 text-muted">
        {{ t(`unavailable.${usage.data.value.unavailable}`) }}
    </p>
    <div v-else-if="windows.length" class="usage-grid">
        <Meter v-for="window in windows" :key="window.id" :value="window.usedPercent" :max="100">
            <MeterLabel>{{ window.label }}</MeterLabel>
            <MeterValue />
            <MeterTrack><MeterIndicator /></MeterTrack>
            <span v-if="window.reset" class="usage-reset text-muted">{{ window.reset }}</span>
        </Meter>
    </div>
    <p v-else class="m-0 text-muted">{{ t('none') }}</p>
</template>

<style scoped>
.usage-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(12rem, 1fr));
    gap: calc(var(--ui-spacing) * 3);
}

.usage-reset {
    grid-column: 1 / -1;
    font-size: calc(var(--ui-font-size) - 0.125rem);
}
</style>

<i18n lang="yaml">
ru:
    window:
        session: Сессия (5 часов)
        week: Неделя
        weekOpus: Неделя · Opus
        weekSonnet: Неделя · Sonnet
        other: 'Окно {minutes} мин'
    resets: 'Сброс {when}'
    none: Агент не сообщил лимиты.
    unavailable:
        signed_out: Подписка не подключена.
        expired: Вход устарел, лимиты обновятся после следующего запуска агента.
        token_login: Токен из setup-token не даёт читать лимиты. Войдите через браузер, чтобы их видеть.
        unsupported: Этот агент не сообщает лимиты.
        error: Не удалось получить лимиты.
en:
    window:
        session: Session (5 hours)
        week: Week
        weekOpus: Week · Opus
        weekSonnet: Week · Sonnet
        other: '{minutes} min window'
    resets: 'Resets {when}'
    none: The agent reported no limits.
    unavailable:
        signed_out: The subscription is not connected.
        expired: The sign-in is stale; limits refresh after the agent runs next time.
        token_login: A setup-token cannot read limits. Sign in through the browser to see them.
        unsupported: This agent does not report limits.
        error: Could not read the limits.
</i18n>
