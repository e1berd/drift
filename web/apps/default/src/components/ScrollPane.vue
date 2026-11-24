<script setup lang="ts">
import {
    ScrollArea,
    ScrollAreaContent,
    ScrollAreaCorner,
    ScrollAreaScrollbar,
    ScrollAreaThumb,
    ScrollAreaViewport,
} from '@ui-sdk/vue/scroll-area'

const { axis = 'vertical' } = defineProps<{
    axis?: 'vertical' | 'horizontal' | 'both'
    contentClass?: string
    label?: string
}>()
</script>

<template>
    <ScrollArea class="scroll-pane">
        <ScrollAreaViewport :role="label ? 'region' : undefined" :aria-label="label">
            <ScrollAreaContent :class="contentClass"><slot /></ScrollAreaContent>
        </ScrollAreaViewport>
        <ScrollAreaScrollbar v-if="axis !== 'horizontal'" orientation="vertical">
            <ScrollAreaThumb />
        </ScrollAreaScrollbar>
        <ScrollAreaScrollbar v-if="axis !== 'vertical'" orientation="horizontal">
            <ScrollAreaThumb />
        </ScrollAreaScrollbar>
        <ScrollAreaCorner v-if="axis === 'both'" />
    </ScrollArea>
</template>
