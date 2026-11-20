<script setup lang="ts">
import type { DrawnLink } from '@/composables/useInfobarLinks'

defineProps<{ routes: DrawnLink[]; activeMessageId: string | null }>()
</script>

<template>
    <svg class="infobar-links" :data-focused="activeMessageId ? '' : undefined" aria-hidden="true">
        <defs>
            <marker
                id="infobar-link-arrow"
                viewBox="0 0 10 10"
                refX="9"
                refY="5"
                markerWidth="6"
                markerHeight="6"
                orient="auto-start-reverse"
            >
                <path class="infobar-link-head" d="M0 0L10 5L0 10z" />
            </marker>
        </defs>
        <TransitionGroup name="infobar-link">
            <path
                v-for="route in routes"
                :key="route.id"
                class="infobar-link"
                :d="route.d"
                :data-dashed="route.dashed || undefined"
                :data-active="route.messageId === activeMessageId || undefined"
                marker-end="url(#infobar-link-arrow)"
            />
        </TransitionGroup>
    </svg>
</template>

<style scoped>
.infobar-links {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    overflow: hidden;
    pointer-events: none;
}

.infobar-link-head {
    fill: var(--ui-primary);
}

.infobar-link {
    fill: none;
    stroke: color-mix(in oklab, var(--ui-primary) 70%, transparent);
    stroke-width: 1.25;
    stroke-linejoin: miter;
    transition:
        stroke var(--dur-2) var(--ease-out),
        opacity var(--dur-2) var(--ease-out);

    &[data-dashed] {
        stroke-dasharray: 4 3;
    }

    &[data-active] {
        stroke: var(--ui-primary);
        stroke-width: 2;
    }
}

.infobar-links[data-focused] .infobar-link:not([data-active]) {
    opacity: 0.35;
}

.infobar-link-enter-active,
.infobar-link-leave-active {
    transition: opacity var(--dur-3) var(--ease-out);
}

.infobar-link-enter-from,
.infobar-link-leave-to {
    opacity: 0;
}
</style>
