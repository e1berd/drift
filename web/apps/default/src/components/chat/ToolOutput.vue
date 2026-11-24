<script setup lang="ts">
import { computed, ref } from 'vue'
import { Button } from '@ui-sdk/vue/button'
import { useI18n } from 'vue-i18n'

import { parseOutput } from '@drift/sdk/toolCalls'

const COLLAPSED_LINES = 16

const props = defineProps<{ lines: string[] }>()
const { t } = useI18n()

const full = ref(false)
const parsed = computed(() => parseOutput(props.lines))
const clipped = computed(() => !full.value && props.lines.length > COLLAPSED_LINES)
</script>

<template>
    <div class="tool-output" :data-clipped="clipped || undefined">
        <div class="tool-output-lines">
            <div
                v-for="(line, index) in parsed"
                :key="index"
                class="tool-output-line"
                :data-line="line.kind"
            >
                <span v-if="line.kind === 'numbered'" class="tool-output-number">{{
                    line.number
                }}</span>
                <span class="tool-output-text">{{ line.text || ' ' }}</span>
            </div>
        </div>
        <div v-if="clipped" class="tool-output-more">
            <Button @click="full = true">{{ t('chat.showAll', { count: lines.length }) }}</Button>
        </div>
    </div>
</template>

<style scoped>
.tool-output {
    position: relative;
    font-family: var(--ui-font-mono);
    font-size: 0.8125rem;
    line-height: 1.55;

    &[data-clipped] .tool-output-lines {
        max-height: calc(1.55em * 16 + 16px);
        overflow-y: hidden;
    }
}

.tool-output-lines {
    padding-block: calc(var(--ui-spacing) * 2);
    overflow-x: auto;
    scrollbar-width: thin;
}

.tool-output-line {
    display: flex;
    min-width: fit-content;
    padding-inline: calc(var(--ui-spacing) * 3);

    &[data-line='add'] {
        background: var(--add-bg);
    }

    &[data-line='del'] {
        background: var(--del-bg);
    }

    &[data-line='hunk'] {
        color: var(--ui-muted-foreground);
    }
}

.tool-output-number {
    flex: none;
    min-width: 3ch;
    margin-inline-end: calc(var(--ui-spacing) * 3);
    color: var(--ui-muted-foreground);
    text-align: end;
    user-select: none;
    opacity: 0.7;
}

.tool-output-text {
    white-space: pre;
    tab-size: 4;
}

.tool-output-more {
    position: absolute;
    inset-inline: 0;
    bottom: 0;
    display: flex;
    justify-content: center;
    padding: calc(var(--ui-spacing) * 6) 0 calc(var(--ui-spacing) * 2);
    background: linear-gradient(to bottom, transparent, var(--ui-muted) 70%);
}
</style>
