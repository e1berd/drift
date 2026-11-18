<script setup lang="ts">
import { computed, watch } from 'vue'
import {
    Popover,
    PopoverContent,
    PopoverDescription,
    PopoverTitle,
    PopoverTrigger,
} from '@ui-sdk/vue/popover'
import { useI18n } from 'vue-i18n'
import { RouterLink } from 'vue-router'

import type { AgentId } from '@/shared/agents'

import UsageMiniMeter from '@/components/composer/usage/UsageMiniMeter.vue'
import UsageMeters from '@/components/settings/subscriptions/UsageMeters.vue'
import { useAgentUsage } from '@/composables/useAgentUsage'
import { reportError } from '@/shared/errors'
import { AGENTS } from '@/shared/agents'

const props = withDefaults(defineProps<{ agent: AgentId; running?: boolean }>(), {
    running: false,
})
const { t } = useI18n()
const usage = useAgentUsage(() => props.agent)

const summary = computed(() => {
    const percentOf = (used?: number) =>
        used === undefined ? t('unknown') : `${Math.round(used)}%`
    return t('summary', {
        session: percentOf(usage.headline.value.session?.usedPercent),
        week: percentOf(usage.headline.value.week?.usedPercent),
    })
})

watch(
    () => props.running,
    (running, wasRunning) => {
        if (wasRunning && !running) usage.refresh().catch(reportError)
    },
)
</script>

<template>
    <Popover v-if="usage.subscription.value">
        <PopoverTrigger
            class="usage-trigger"
            :aria-label="`${t('title')}: ${summary}`"
            :title="summary"
        >
            <UsageMiniMeter
                :short="t('short.session')"
                :name="t('name.session')"
                :window="usage.headline.value.session"
            />
            <UsageMiniMeter
                :short="t('short.week')"
                :name="t('name.week')"
                :window="usage.headline.value.week"
            />
        </PopoverTrigger>
        <PopoverContent side="top" align="end" class="usage-popover">
            <PopoverTitle>{{ t('heading', { agent: AGENTS[agent].name }) }}</PopoverTitle>
            <PopoverDescription v-if="usage.subscription.value.account">
                {{ usage.subscription.value.account }}
            </PopoverDescription>
            <UsageMeters :subscription-id="usage.subscription.value.id" />
            <RouterLink to="/settings" class="usage-settings text-muted">{{
                t('settings')
            }}</RouterLink>
        </PopoverContent>
    </Popover>
</template>

<style scoped>
.usage-trigger {
    flex-direction: column;
    align-items: stretch;
    justify-content: center;
    gap: 3px;
    padding-inline: calc(var(--ui-spacing) * 2);
}

.usage-popover {
    width: min(20rem, calc(100vw - 2rem));
}

.usage-settings {
    align-self: flex-start;
    font-size: calc(var(--ui-font-size) - 0.125rem);
    text-decoration: underline;
    text-underline-offset: 4px;
}
</style>

<i18n lang="yaml">
ru:
    title: Лимиты
    heading: 'Лимиты {agent}'
    summary: '5 часов: {session}, неделя: {week}'
    unknown: нет данных
    settings: Подписки в настройках
    short:
        session: 5ч
        week: 7д
    name:
        session: Лимит на 5 часов
        week: Недельный лимит
en:
    title: Limits
    heading: '{agent} limits'
    summary: '5 hours: {session}, week: {week}'
    unknown: no data
    settings: Subscriptions in settings
    short:
        session: 5h
        week: 7d
    name:
        session: 5-hour limit
        week: Weekly limit
</i18n>
