<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import AgentLogo from '@/components/AgentLogo.vue'
import type { SessionActivity } from '@/gen'
import type { AgentId } from '@/shared/agents'

const props = defineProps<{
    agent: AgentId
    activity?: SessionActivity
    runningTool?: string
    since: string
}>()
const { t } = useI18n()

const now = ref(Date.now())
const timer = setInterval(() => (now.value = Date.now()), 1000)
onBeforeUnmount(() => clearInterval(timer))

const seconds = computed(() =>
    Math.max(0, Math.round((now.value - new Date(props.since).getTime()) / 1000)),
)
const elapsed = computed(() =>
    seconds.value < 60
        ? t('seconds', { count: seconds.value })
        : t('minutes', { minutes: Math.floor(seconds.value / 60), seconds: seconds.value % 60 }),
)
const partial = computed(() =>
    props.activity?.phase === 'writing' && props.activity.text ? props.activity.text : '',
)
const label = computed(() => {
    if (props.runningTool) return t('running', { tool: props.runningTool })
    switch (props.activity?.phase) {
        case 'tool':
            return t('preparing', { tool: props.activity.tool ?? '' })
        case 'writing':
            return t('writing')
        case 'thinking':
            return props.activity.tokens
                ? t('thinkingTokens', { tokens: props.activity.tokens })
                : t('thinking')
        default:
            return t('working')
    }
})
</script>

<template>
    <div class="grid grid-cols-[20px_minmax(0,1fr)] gap-4" role="status" aria-live="polite">
        <AgentLogo :agent="agent" :size="20" decorative />
        <div class="flex min-w-0 flex-col gap-2">
            <p v-if="partial" class="m-0 whitespace-pre-wrap">
                {{ partial }}<span class="caret" aria-hidden="true" />
            </p>
            <p class="m-0 flex gap-2 text-muted">
                <span class="motion-shine">{{ label }}</span>
                <span class="tabular-nums">{{ elapsed }}</span>
            </p>
        </div>
    </div>
</template>

<style scoped>
.caret {
    display: inline-block;
    width: 0.5em;
    height: 1em;
    margin-inline-start: 2px;
    vertical-align: text-bottom;
    background: var(--ui-foreground);
    animation: caret-blink 1s steps(1) infinite;
}

@keyframes caret-blink {
    50% {
        opacity: 0;
    }
}

@media (prefers-reduced-motion: reduce) {
    .caret {
        animation: none;
    }
}
</style>

<i18n lang="yaml">
ru:
    working: Работает…
    thinking: Думает…
    thinkingTokens: Думает… ~{tokens} ток.
    writing: Пишет ответ…
    preparing: Готовит {tool}…
    running: Выполняет {tool}…
    seconds: '{count} с'
    minutes: '{minutes} мин {seconds} с'
en:
    working: Working…
    thinking: Thinking…
    thinkingTokens: Thinking… ~{tokens} tokens
    writing: Writing…
    preparing: Preparing {tool}…
    running: Running {tool}…
    seconds: '{count}s'
    minutes: '{minutes}m {seconds}s'
</i18n>
