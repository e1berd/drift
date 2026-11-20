<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

import type { UsageWindow } from '@/gen'

import {
    InlineMeter,
    InlineMeterIndicator,
    InlineMeterTrack,
} from '@/components/composer/usage/inlineMeter'
import { formatResetDate, usageLevel } from '@/shared/usage'

const props = defineProps<{ short: string; name: string; window?: UsageWindow }>()
const { locale } = useI18n()

const percent = computed(() => Math.round(props.window?.usedPercent ?? 0))
const title = computed(() =>
    props.window?.resetTime
        ? `${props.name} · ${formatResetDate(props.window.resetTime, locale.value)}`
        : props.name,
)
</script>

<template>
    <InlineMeter
        class="usage-mini"
        :value="percent"
        :aria-label="name"
        :title="title"
        :data-level="window ? usageLevel(percent) : 'unknown'"
    >
        <span class="usage-mini-label" aria-hidden="true">{{ short }}</span>
        <InlineMeterTrack class="usage-mini-track"
            ><InlineMeterIndicator class="usage-mini-fill"
        /></InlineMeterTrack>
        <span class="usage-mini-value" aria-hidden="true">{{ window ? `${percent}%` : '—' }}</span>
    </InlineMeter>
</template>

<style scoped>
.usage-mini {
    display: flex;
    align-items: center;
    width: auto;
    min-width: 0;
    gap: calc(var(--ui-spacing) * 1.5);
    font-size: 0.6875rem;
    line-height: 1;
    font-variant-numeric: tabular-nums;

    & .usage-mini-track {
        position: relative;
        display: block;
        width: 2.25rem;
        height: 0.25rem;

        @container composer (width < 560px) {
            width: 1.5rem;
        }
        overflow: hidden;
        border-radius: 999px;
        background: var(--ui-muted);
    }

    & .usage-mini-fill {
        position: absolute;
        inset-block: 0;
        inset-inline-start: 0;
        background: var(--usage-normal);
    }

    &[data-level='high'] .usage-mini-fill {
        background: var(--usage-high);
    }

    &[data-level='critical'] {
        & .usage-mini-fill {
            background: var(--usage-critical);
        }

        & .usage-mini-value {
            color: var(--usage-critical);
        }
    }

    &[data-level='unknown'] .usage-mini-track {
        background: repeating-linear-gradient(
            90deg,
            var(--ui-muted) 0 0.25rem,
            transparent 0.25rem 0.375rem
        );
    }
}

.usage-mini-label {
    min-width: 1.25rem;

    @container composer (width < 560px) {
        display: none;
    }
}

.usage-mini-label,
.usage-mini-value {
    color: var(--ui-muted-foreground);
    font-weight: 500;
}

.usage-mini-value {
    min-width: 2.25ch;
    text-align: end;

    @container composer (width < 560px) {
        display: none;
    }
}
</style>
