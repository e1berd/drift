<script setup lang="ts">
import { computed, ref } from 'vue'
import {
    PhCaretRight,
    PhCheck,
    PhFilePlus,
    PhFileText,
    PhGlobe,
    PhListChecks,
    PhMagnifyingGlass,
    PhPencilSimple,
    PhRobot,
    PhTerminal,
    PhWarningCircle,
    PhWrench,
} from '@phosphor-icons/vue'
import { Collapsible, CollapsiblePanel, CollapsibleTrigger } from '@ui-sdk/vue/collapsible'
import { useI18n } from 'vue-i18n'

import type { Component } from 'vue'

import type { MessageBody } from '@drift/vue/api'

import RunSpinner from '@/components/RunSpinner.vue'
import ToolOutput from '@/components/chat/ToolOutput.vue'
import ToolSummary from '@/components/chat/ToolSummary.vue'
import type { ToolFamily } from '@drift/sdk/toolCalls'
import { shellCommand, toolFamily } from '@drift/sdk/toolCalls'

const ICONS: Record<ToolFamily, Component> = {
    shell: PhTerminal,
    edit: PhPencilSimple,
    write: PhFilePlus,
    read: PhFileText,
    search: PhMagnifyingGlass,
    web: PhGlobe,
    todo: PhListChecks,
    agent: PhRobot,
    other: PhWrench,
}

const props = defineProps<{ body: MessageBody }>()
const { t } = useI18n()

const name = computed(() => props.body.name ?? '')
const input = computed(() => props.body.input ?? '')
const family = computed(() => toolFamily(name.value))
const lines = computed(() => props.body.output ?? [])
const status = computed(() => props.body.status ?? 'done')
const failed = computed(() => status.value === 'error' || status.value === 'stopped')
const fullInput = computed(() =>
    family.value === 'shell' ? shellCommand(input.value) : input.value,
)
const showsFullInput = computed(() => fullInput.value.includes('\n') || fullInput.value.length > 80)
const expandable = computed(() => lines.value.length > 0 || showsFullInput.value)
const open = ref(status.value === 'error' && lines.value.length > 0)
</script>

<template>
    <Collapsible
        v-model:open="open"
        class="tool"
        :data-status="status"
        :data-expandable="expandable || undefined"
        :disabled="!expandable"
    >
        <CollapsibleTrigger class="tool-head">
            <component :is="ICONS[family]" class="tool-icon" />
            <span class="tool-name">{{ name }}</span>
            <ToolSummary :family="family" :input="input" />
            <span class="tool-status status-tag">
                <template v-if="status === 'running'"><RunSpinner :size="13" fast /></template>
                <span v-else-if="failed" class="text-danger tool-failed"
                    ><PhWarningCircle />{{
                        status === 'stopped' ? t('chat.stopped') : t('chat.failed')
                    }}</span
                >
                <template v-else
                    ><span v-if="body.meta">{{ body.meta }}</span><PhCheck class="tool-done"
                /></template>
            </span>
            <PhCaretRight v-if="expandable" class="tool-caret" />
        </CollapsibleTrigger>
        <CollapsiblePanel class="tool-panel">
            <pre
                v-if="showsFullInput"
                class="tool-input"
                :data-shell="family === 'shell' || undefined"
                >{{ fullInput }}</pre>
            <ToolOutput v-if="lines.length" :lines="lines" />
        </CollapsiblePanel>
    </Collapsible>
</template>

<style scoped>
.tool {
    margin-inline-start: 36px;
    overflow: hidden;
    border: 1px solid transparent;
    border-radius: var(--ui-radius);
    transition:
        border-color var(--dur-2),
        background var(--dur-2);

    &:has(> .tool-head[data-open]) {
        border-color: var(--ui-border);
        background: var(--ui-muted);
    }

    @media (max-width: 760px) {
        margin-inline-start: 0;
    }
}

.tool-head {
    display: flex;
    width: 100%;
    min-width: 0;
    align-items: center;
    gap: calc(var(--ui-spacing) * 2);
    min-height: 32px;
    padding: calc(var(--ui-spacing) * 1) calc(var(--ui-spacing) * 2);
    border: 0;
    border-radius: var(--ui-radius);
    background: none;
    color: var(--ui-foreground);
    font: inherit;
    text-align: start;

    [data-expandable] > & {
        cursor: pointer;

        &:hover {
            background: var(--ui-muted);
        }
    }

    &:focus-visible {
        outline: 2px solid var(--ui-ring-border);
        outline-offset: -2px;
    }
}

.tool-icon {
    flex: none;
    color: var(--ui-muted-foreground);
}

.tool-name {
    flex: none;
    color: var(--ui-muted-foreground);
    font-size: 0.8125rem;
    font-weight: 500;
}

.tool-status {
    flex: none;
}

.tool-failed {
    display: inline-flex;
    align-items: center;
    gap: 4px;
}

.tool-done {
    color: var(--ui-muted-foreground);
}

.tool-caret {
    flex: none;
    color: var(--ui-muted-foreground);
    transition: rotate var(--dur-2) var(--ease-out);

    .tool-head[data-open] > & {
        rotate: 90deg;
    }
}

.tool-panel {
    border-top: 1px solid var(--ui-border);
}

.tool-input {
    margin: 0;
    padding: calc(var(--ui-spacing) * 2) calc(var(--ui-spacing) * 3);
    border-bottom: 1px solid var(--ui-border);
    font-family: var(--ui-font-mono);
    font-size: 0.8125rem;
    line-height: 1.55;
    white-space: pre-wrap;
    overflow-wrap: anywhere;

    &[data-shell]::before {
        content: '$ ';
        color: var(--ui-muted-foreground);
    }

    &:last-child {
        border-bottom: 0;
    }
}
</style>
