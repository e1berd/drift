<script setup lang="ts">
import { computed } from 'vue'
import { PhArrowClockwise } from '@phosphor-icons/vue'
import { useMutation, useQueryClient } from '@tanstack/vue-query'
import { Button } from '@ui-sdk/vue/button'
import { Meter, MeterIndicator, MeterLabel, MeterTrack, MeterValue } from '@ui-sdk/vue/meter'
import { toast } from '@ui-sdk/vue/toast'
import { useI18n } from 'vue-i18n'

import { getSubscriptionUsage, getSubscriptionUsageQueryKey, useGetSubscriptionUsage } from '@/gen'
import { errorMessage } from '@/shared/errors'
import {
    formatReset,
    formatResetDate,
    secondsUntil,
    usageLevel,
    usageWindowKind,
} from '@/shared/usage'

const props = defineProps<{ subscriptionId: string }>()
const { t, locale } = useI18n()
const queryClient = useQueryClient()
const path = computed(() => ({ subscriptionId: props.subscriptionId }))

const usage = useGetSubscriptionUsage(
    { path },
    { query: { refetchInterval: 60_000, staleTime: 30_000 } },
)

const refresh = useMutation({
    mutationFn: async () =>
        (
            await getSubscriptionUsage({
                path: path.value,
                query: { refresh: true },
                throwOnError: true,
            })
        ).data,
    onSuccess: (data) => {
        queryClient.setQueryData(getSubscriptionUsageQueryKey({ path: path.value }), data)
    },
    onError: (error) => toast.error(errorMessage(error)),
})

const windows = computed(() =>
    (usage.data.value?.windows ?? []).map((window) => {
        const left = secondsUntil(window.resetTime)
        return {
            ...window,
            label:
                window.label ??
                t(`window.${usageWindowKind(window)}`, { minutes: window.windowMinutes ?? 0 }),
            reset: left === null ? '' : t('resets', { when: formatReset(left, locale.value) }),
            resetDate: window.resetTime ? formatResetDate(window.resetTime, locale.value) : '',
        }
    }),
)
</script>

<template>
    <div class="usage">
        <div class="usage-content">
            <div v-if="usage.isPending.value" class="skeleton h-8" />
            <p v-else-if="usage.isError.value" class="m-0 text-muted">
                {{ errorMessage(usage.error.value) }}
            </p>
            <p v-else-if="usage.data.value?.unavailable" class="m-0 text-muted">
                {{ t(`unavailable.${usage.data.value.unavailable}`) }}
            </p>
            <div v-else-if="windows.length" class="usage-grid">
                <Meter
                    v-for="window in windows"
                    :key="window.id"
                    :value="window.usedPercent"
                    :max="100"
                    :data-level="usageLevel(window.usedPercent)"
                >
                    <MeterLabel>{{ window.label }}</MeterLabel>
                    <MeterValue />
                    <MeterTrack><MeterIndicator /></MeterTrack>
                    <span v-if="window.reset" class="usage-reset text-muted">
                        {{ window.reset
                        }}<template v-if="window.resetDate"> · {{ window.resetDate }}</template>
                    </span>
                </Meter>
            </div>
            <p v-else class="m-0 text-muted">{{ t('none') }}</p>
        </div>
        <Button
            :aria-label="t('refresh')"
            v-tooltip="t('refresh')"
            :disabled="refresh.isPending.value || usage.isFetching.value"
            @click="refresh.mutate()"
        >
            <PhArrowClockwise :class="{ 'animate-spin': refresh.isPending.value }" />
        </Button>
    </div>
</template>

<style scoped>
.usage {
    display: flex;
    align-items: flex-start;
    gap: calc(var(--ui-spacing) * 2);
}

.usage-content {
    min-width: 0;
    flex: 1;
}

.usage-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(12rem, 1fr));
    gap: calc(var(--ui-spacing) * 3);
}

[data-level='normal'] [data-slot='meter-indicator'] {
    background: var(--usage-normal);
}

[data-level='high'] [data-slot='meter-indicator'] {
    background: var(--usage-high);
}

[data-level='critical'] [data-slot='meter-indicator'] {
    background: var(--usage-critical);
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
    refresh: Обновить лимиты
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
    refresh: Refresh limits
    none: The agent reported no limits.
    unavailable:
        signed_out: The subscription is not connected.
        expired: The sign-in is stale; limits refresh after the agent runs next time.
        token_login: A setup-token cannot read limits. Sign in through the browser to see them.
        unsupported: This agent does not report limits.
        error: Could not read the limits.
</i18n>
