<script setup lang="ts">
import { computed, ref } from 'vue'
import { PhArrowsIn, PhArrowsOut, PhCheck } from '@phosphor-icons/vue'
import { Button } from '@ui-sdk/vue/button'
import { Collapsible, CollapsiblePanel, CollapsibleTrigger } from '@ui-sdk/vue/collapsible'
import { useI18n } from 'vue-i18n'

import type { MessageBody } from '@/gen'

import RunSpinner from '@/components/RunSpinner.vue'
import ScrollPane from '@/components/ScrollPane.vue'

const VISIBLE_LINES = 12

const props = defineProps<{ body: MessageBody }>()
const { t } = useI18n()

const lines = computed(() => props.body.output ?? [])
const status = computed(() => props.body.status ?? 'done')
const open = ref(status.value === 'running' || lines.value.length <= VISIBLE_LINES)
const full = ref(false)
const fullLabel = computed(() => (full.value ? t('chat.collapse') : t('chat.expand')))

function lineKind(line: string): string | undefined {
    if (line.startsWith('@@')) return 'hunk'
    if (line.startsWith('+') && !line.startsWith('+++')) return 'add'
    if (line.startsWith('-') && !line.startsWith('---')) return 'del'
    return undefined
}

function toggleFull() {
    full.value = !full.value
    if (full.value) open.value = true
}
</script>

<template>
    <Collapsible v-model:open="open" class="tool" :data-full="full || undefined">
        <div class="flex items-start gap-1">
            <CollapsibleTrigger class="min-w-0 flex-1">
                <span class="tool-head">
                    <b>{{ body.name }}</b>
                    <span class="tool-input">{{ body.input }}</span>
                    <span class="status-tag shrink-0">
                        <template v-if="status === 'running'"
                            ><RunSpinner :size="13" fast />{{ t('chat.running') }}</template
                        >
                        <span v-else-if="status === 'stopped'" class="text-danger">{{
                            t('chat.stopped')
                        }}</span>
                        <span v-else-if="status === 'error'" class="text-danger">{{
                            t('chat.failed')
                        }}</span>
                        <template v-else>{{ body.meta }}<PhCheck /></template>
                    </span>
                </span>
            </CollapsibleTrigger>
            <Button
                :aria-label="fullLabel"
                v-tooltip="fullLabel"
                :aria-pressed="full"
                @click="toggleFull"
            >
                <PhArrowsIn v-if="full" />
                <PhArrowsOut v-else />
            </Button>
        </div>
        <CollapsiblePanel>
            <ScrollPane axis="both" class="tool-output" content-class="py-2">
                <div
                    v-for="(line, index) in lines"
                    :key="index"
                    class="tool-line"
                    :data-line="lineKind(line)"
                >
                    {{ line || ' ' }}
                </div>
            </ScrollPane>
            <Button v-if="!full && lines.length > VISIBLE_LINES" @click="toggleFull">
                {{ t('chat.showAll', { count: lines.length }) }}
            </Button>
        </CollapsiblePanel>
    </Collapsible>
</template>

<style scoped>
.tool {
    margin-inline-start: 36px;

    @media (max-width: 760px) {
        margin-inline-start: 0;
    }
}

.tool-head {
    display: flex;
    min-width: 0;
    flex: 1;
    align-items: baseline;
    gap: 8px;
    font-family: var(--ui-font-mono);
}

.tool-input {
    min-width: 0;
    flex: 1;
    overflow: hidden;
    color: var(--ui-muted-foreground);
    text-overflow: ellipsis;
    white-space: nowrap;

    [data-full] & {
        white-space: pre-wrap;
        word-break: break-all;
    }
}

.tool-output {
    max-height: 248px;
    border-radius: var(--ui-radius);
    background: var(--ui-muted);
    font-family: var(--ui-font-mono);
    transition: max-height var(--dur-4) var(--ease-out);

    [data-full] & {
        max-height: 1800px;
    }

    & .tool-line {
        padding-inline: 12px;
        white-space: pre;
        tab-size: 4;

        [data-full] & {
            white-space: pre-wrap;
            word-break: break-all;
        }
    }

    & [data-line='add'] {
        background: var(--add-bg);
    }

    & [data-line='del'] {
        background: var(--del-bg);
    }

    & [data-line='hunk'] {
        color: var(--ui-muted-foreground);
    }
}
</style>
