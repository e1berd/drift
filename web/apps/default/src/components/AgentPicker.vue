<script setup lang="ts">
import { nextTick, useTemplateRef, watch } from 'vue'
import {
    SegmentedControl,
    SegmentedControlIndicator,
    SegmentedControlItem,
} from '@ui-sdk/vue/segmented-control'

import type { AgentId } from '@drift/sdk/agents'

import AgentLogo from '@/components/AgentLogo.vue'

const agent = defineModel<string | null>({ required: true })
defineProps<{ agents: { id: AgentId; name: string }[]; label: string }>()

const slider = useTemplateRef<HTMLElement>('slider')

async function revealSelectedAgent() {
    await nextTick()
    slider.value
        ?.querySelector('[data-slot="segmented-control-item"][data-checked]')
        ?.scrollIntoView({ inline: 'nearest', block: 'nearest', behavior: 'smooth' })
}

watch(agent, revealSelectedAgent)
</script>

<template>
    <div ref="slider" class="agent-slider">
        <SegmentedControl v-model:value="agent" :aria-label="label">
            <SegmentedControlIndicator />
            <SegmentedControlItem v-for="item in agents" :key="item.id" :value="item.id">
                <AgentLogo :agent="item.id" decorative />{{ item.name }}
            </SegmentedControlItem>
        </SegmentedControl>
    </div>
</template>

<style scoped>
@property --agent-fade-start {
    syntax: '<length>';
    inherits: false;
    initial-value: 0px;
}

@property --agent-fade-end {
    syntax: '<length>';
    inherits: false;
    initial-value: 0px;
}

@keyframes agent-slider-fade {
    from {
        --agent-fade-start: 0px;
        --agent-fade-end: 2.5rem;
    }
    to {
        --agent-fade-start: 2.5rem;
        --agent-fade-end: 0px;
    }
}

.agent-slider {
    max-width: 100%;
    overflow-x: auto;
    overscroll-behavior-x: contain;
    scroll-snap-type: x proximity;
    scroll-padding-inline: calc(var(--ui-spacing) * 4);
    scrollbar-width: none;
    mask-image: linear-gradient(
        to right,
        transparent,
        black var(--agent-fade-start),
        black calc(100% - var(--agent-fade-end)),
        transparent
    );
    animation: agent-slider-fade linear both;
    animation-timeline: scroll(self inline);

    &::-webkit-scrollbar {
        display: none;
    }

    & :deep([data-slot='segmented-control']) {
        width: max-content;
    }

    & :deep([data-slot='segmented-control-item']) {
        flex: none;
        gap: calc(var(--ui-spacing) * 2);
        scroll-snap-align: start;
    }
}
</style>
