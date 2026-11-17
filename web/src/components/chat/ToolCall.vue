<script setup lang="ts">
import { computed, ref } from 'vue'
import { PhArrowsIn, PhArrowsOut, PhCaretDown, PhCaretRight, PhCheck } from '@phosphor-icons/vue'

import RunSpinner from '@/components/RunSpinner.vue'
import type { MessageBody } from '@/gen'
import { useI18n } from '@/shared/i18n'

const VISIBLE_LINES = 12

const props = defineProps<{ body: MessageBody }>()
const { t } = useI18n()

const lines = computed(() => props.body.output ?? [])
const open = ref(props.body.status === 'running' || lines.value.length <= VISIBLE_LINES)
const full = ref(false)
const status = computed(() => props.body.status ?? 'done')

function lineClass(line: string) {
    if (line.startsWith('+') && !line.startsWith('+++')) return 'add'
    if (line.startsWith('-') && !line.startsWith('---')) return 'del'
    if (line.startsWith('@@')) return 'hunk'
    return ''
}

function toggleFull() {
    full.value = !full.value
    if (full.value) open.value = true
}
</script>

<template>
    <div
        class="tool"
        :data-open="open || undefined"
        :data-full="full || undefined"
        :data-status="status"
    >
        <div class="head">
            <button type="button" class="toggle" :aria-expanded="open" @click="open = !open">
                <PhCaretRight :size="14" class="chev" />
                <span class="name">{{ body.name }}</span>
                <span class="arg">{{ body.input }}</span>
                <span class="meta">
                    <template v-if="status === 'running'"
                        ><RunSpinner :size="13" fast />{{ t('chat.running') }}</template
                    >
                    <span
                        v-else-if="status === 'stopped' || status === 'error'"
                        class="text-danger"
                        >{{ t('chat.stopped') }}</span
                    >
                    <template v-else
                        >{{ body.meta }}<PhCheck :size="13" class="text-ok" weight="bold"
                    /></template>
                </span>
            </button>
            <button
                type="button"
                class="icon-btn small mt-1"
                :aria-pressed="full"
                :aria-label="full ? t('chat.collapse') : t('chat.expand')"
                :title="full ? t('chat.collapse') : t('chat.expand')"
                @click="toggleFull"
            >
                <Transition name="swap" mode="out-in">
                    <PhArrowsIn v-if="full" key="in" :size="14" />
                    <PhArrowsOut v-else key="out" :size="14" />
                </Transition>
            </button>
        </div>
        <div class="grid-reveal" :data-open="(open && lines.length > 0) || undefined">
            <div>
                <div class="body">
                    <div v-for="(line, i) in lines" :key="i" class="line" :class="lineClass(line)">
                        {{ line || ' ' }}
                    </div>
                    <button
                        v-if="!full && lines.length > VISIBLE_LINES"
                        type="button"
                        class="more"
                        @click="toggleFull"
                    >
                        <PhCaretDown :size="12" />{{ t('chat.showAll', { count: lines.length }) }}
                    </button>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.tool {
    margin-left: 36px;
    overflow: hidden;
    border: 1px solid var(--ui-border);
    border-radius: 10px;
    background: var(--surface);

    &[data-status='running'] {
        border-color: var(--border-strong);
    }
}

.head {
    display: flex;
    align-items: flex-start;
    gap: 2px;
    padding-right: 4px;

    &:hover {
        background: var(--surface-2);
    }
}

.toggle {
    display: flex;
    min-width: 0;
    flex: 1;
    align-items: flex-start;
    gap: 10px;
    padding: 10px 4px 10px 12px;
    border: 0;
    background: none;
    color: var(--ui-muted-foreground);
    font: 12.5px/16px var(--ui-font-mono);
    text-align: left;
    cursor: pointer;
}

.chev {
    margin-top: 1px;
    color: var(--faint);
    transition: transform var(--dur-3) var(--ease-out);
}

.tool[data-open] .chev {
    transform: rotate(90deg);
}

.name {
    flex: none;
    color: var(--ui-foreground);
    font-weight: 600;
}

.arg {
    min-width: 0;
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.tool[data-full] .arg {
    color: var(--ui-foreground);
    white-space: pre-wrap;
    word-break: break-all;
}

.meta {
    display: flex;
    flex: none;
    align-items: center;
    gap: 6px;
    color: var(--faint);
    font-size: 11.5px;
}

.body {
    position: relative;
    max-height: 248px;
    overflow: auto;
    padding: 8px 0;
    border-top: 1px solid var(--ui-border);
    background: var(--surface-2);
    font: 12px/1.65 var(--ui-font-mono);
    transition: max-height var(--dur-4) var(--ease-out);
}

.tool[data-full] .body {
    max-height: 1800px;
}

.line {
    padding: 0 14px;
    color: var(--ui-muted-foreground);
    white-space: pre;
    tab-size: 4;

    &.add {
        background: var(--add-bg);
        color: var(--add-fg);
    }

    &.del {
        background: var(--del-bg);
        color: var(--del-fg);
    }

    &.hunk {
        color: var(--faint);
    }
}

.tool[data-full] .line {
    white-space: pre-wrap;
    word-break: break-all;
}

.more {
    position: sticky;
    bottom: -8px;
    display: flex;
    width: 100%;
    height: 34px;
    align-items: flex-end;
    justify-content: center;
    gap: 6px;
    margin-top: -8px;
    padding-bottom: 8px;
    border: 0;
    background: linear-gradient(180deg, transparent, var(--surface-2) 55%);
    color: var(--accent);
    font: 500 12px var(--ui-font-sans);
    cursor: pointer;
}

.swap-enter-active,
.swap-leave-active {
    transition:
        opacity var(--dur-2) var(--ease-out),
        transform var(--dur-3) var(--ease-out);
}

.swap-enter-from,
.swap-leave-to {
    opacity: 0;
    transform: scale(0.6);
}

@media (max-width: 760px) {
    .tool {
        margin-left: 0;
    }
}
</style>
