<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { PhFile } from '@phosphor-icons/vue'
import { Progress, ProgressIndicator, ProgressTrack } from '@ui-sdk/vue/progress'

import type { AttachmentState } from '@/components/attachments/attachmentState'

const props = withDefaults(
    defineProps<{
        name: string
        state?: AttachmentState
        imageUrl?: string
        percent?: number
        detail?: string
    }>(),
    { state: 'ready', imageUrl: undefined, percent: 0, detail: '' },
)

const extension = computed(() => {
    const dot = props.name.lastIndexOf('.')
    return dot > 0 ? props.name.slice(dot + 1, dot + 5).toUpperCase() : ''
})
const showsProgress = computed(() => props.state === 'uploading' || props.state === 'paused')
const imageLoaded = ref(false)
watch(
    () => props.imageUrl,
    (current, previous) => {
        if (current !== previous) imageLoaded.value = false
    },
)
</script>

<template>
    <div
        class="tile"
        :data-kind="imageUrl ? 'image' : 'file'"
        :data-state="state"
        :data-progress="showsProgress || undefined"
    >
        <template v-if="imageUrl">
            <img
                :src="imageUrl"
                alt=""
                class="tile-image"
                :data-loaded="imageLoaded || undefined"
                decoding="async"
                @load="imageLoaded = true"
            />
            <span class="tile-lead tile-lead-image"><slot name="lead" /></span>
        </template>
        <template v-else>
            <span class="tile-lead tile-lead-file">
                <span class="tile-file-icon" aria-hidden="true">
                    <PhFile :size="18" />
                    <span v-if="extension" class="tile-extension">{{ extension }}</span>
                </span>
                <slot name="lead" />
            </span>
            <span class="tile-text">
                <span class="tile-name">{{ name }}</span>
                <span class="tile-detail">{{ detail }}</span>
            </span>
        </template>
        <Progress
            class="tile-progress"
            :data-shown="showsProgress || undefined"
            :value="percent"
            :max="100"
            :aria-label="name"
            :aria-hidden="!showsProgress"
        >
            <ProgressTrack><ProgressIndicator /></ProgressTrack>
        </Progress>
        <span class="tile-corner"><slot name="corner" /></span>
    </div>
</template>

<style scoped>
.tile {
    --tile-size: 56px;
    position: relative;
    display: flex;
    flex: none;
    align-items: center;
    gap: calc(var(--ui-spacing) * 2.5);
    height: var(--tile-size);
    overflow: hidden;
    border: 1px solid var(--ui-border);
    border-radius: var(--ui-radius);
    background: var(--ui-background);
    transition:
        border-color var(--dur-2),
        opacity var(--dur-3) var(--ease-out),
        scale var(--dur-3) var(--ease-out);

    @starting-style {
        opacity: 0;
        scale: 0.94;
    }

    &[data-kind='image'] {
        width: var(--tile-size);
        background: var(--ui-muted);
    }

    &[data-kind='file'] {
        width: 232px;
        max-width: 100%;
        padding-inline: calc((var(--tile-size) - 40px) / 2) calc(var(--ui-spacing) * 3);
    }

    &[data-state='failed'] {
        border-color: var(--ui-destructive);
    }

    &[data-state='ghost'] {
        border-style: dashed;
    }
}

.tile-image {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    object-fit: cover;
    opacity: 0;
    transition:
        opacity var(--dur-3) var(--ease-out),
        filter var(--dur-3) var(--ease-out);

    &[data-loaded] {
        opacity: 1;
    }

    [data-progress] > &,
    [data-state='failed'] > &,
    [data-state='ghost'] > & {
        filter: brightness(0.55) saturate(0.7);
    }
}

.tile-lead {
    position: relative;
    display: grid;
    flex: none;
    place-items: center;

    & > :deep([data-slot='button']) {
        position: absolute;
        inset: 0;
        width: 100%;
        height: 100%;
        min-height: 0;
        padding: 0;
        border-radius: inherit;
    }
}

.tile-lead-image {
    position: absolute;
    inset: 0;
    color: white;

    & > :deep([data-slot='button']) {
        color: inherit;
        background: none;
    }
}

.tile-lead-file {
    width: 40px;
    height: 40px;
    border-radius: calc(var(--ui-radius) * 0.75);
    background: var(--ui-muted);

    & > :deep([data-slot='button']) {
        background: var(--ui-muted);
        opacity: 0;
        transition: opacity var(--dur-1);
    }

    &:hover > :deep([data-slot='button']),
    & > :deep([data-slot='button']:focus-visible),
    [data-state='failed'] & > :deep([data-slot='button']),
    [data-state='ghost'] & > :deep([data-slot='button']),
    [data-state='paused'] & > :deep([data-slot='button']) {
        opacity: 1;
    }

    @media (hover: none) {
        & > :deep([data-slot='button']) {
            opacity: 1;
        }
    }
}

.tile-file-icon {
    display: grid;
    justify-items: center;
    color: var(--ui-muted-foreground);
}

.tile-extension {
    font-size: 0.5625rem;
    font-weight: 600;
    letter-spacing: 0.02em;
    line-height: 1;
}

.tile-text {
    display: flex;
    min-width: 0;
    flex: 1;
    flex-direction: column;
    gap: 2px;
    line-height: 1.25;
}

.tile-name {
    overflow: hidden;
    font-size: 0.8125rem;
    font-weight: 500;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.tile-detail {
    overflow: hidden;
    color: var(--ui-muted-foreground);
    font-size: 0.75rem;
    font-variant-numeric: tabular-nums;
    text-overflow: ellipsis;
    white-space: nowrap;

    [data-state='failed'] & {
        color: var(--ui-destructive);
    }
}

.tile-progress {
    position: absolute;
    inset-inline: 0;
    bottom: 0;
    display: block;
    min-width: 0;
    opacity: 0;
    transition: opacity var(--dur-3) var(--ease-out) 120ms;

    &[data-shown] {
        opacity: 1;
        transition-delay: 0ms;
    }

    & :deep([data-slot='progress-track']) {
        height: 3px;
        border-radius: 0;
        background: color-mix(in oklab, var(--ui-muted-foreground) 25%, transparent);
    }

    & :deep([data-slot='progress-indicator']) {
        transition: width 240ms linear;
    }

    [data-kind='image'] > & {
        inset-inline: 6px;
        bottom: 6px;

        & :deep([data-slot='progress-track']) {
            border-radius: 999px;
            background: color-mix(in oklab, white 30%, transparent);
        }

        & :deep([data-slot='progress-indicator']) {
            background: white;
        }
    }

    [data-state='paused'] > & :deep([data-slot='progress-indicator']) {
        background: var(--ui-muted-foreground);
    }
}

.tile-corner {
    position: absolute;
    top: 3px;
    right: 3px;
    opacity: 0;
    transition: opacity var(--dur-1);

    &:empty {
        display: none;
    }

    .tile:is(:hover, :focus-within) > & {
        opacity: 1;
    }

    @media (hover: none) {
        opacity: 1;
    }

    & > :deep([data-slot='button']) {
        width: 20px;
        height: 20px;
        min-height: 0;
        padding: 0;
        border-radius: 999px;
        background: color-mix(in oklab, var(--ui-foreground) 70%, transparent);
        color: var(--ui-background);

        & svg {
            width: 12px;
            height: 12px;
        }
    }
}
</style>
