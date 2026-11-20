<script setup lang="ts">
import { nextTick, useTemplateRef, watch } from 'vue'
import { useI18n } from 'vue-i18n'

import InfobarActivityRow from '@/components/infobar/InfobarActivityRow.vue'
import type { ActivityEntry } from '@/gen'

const props = defineProps<{ activity: ActivityEntry[]; activeMessageId: string | null }>()
const { t } = useI18n()

const STICK_DISTANCE = 48
const scroller = useTemplateRef<HTMLElement>('scroller')

function nearEnd(element: HTMLElement): boolean {
    return element.scrollHeight - element.scrollTop - element.clientHeight < STICK_DISTANCE
}

watch(
    () => props.activity.length,
    async (count, previous) => {
        const element = scroller.value
        if (!element || count <= (previous ?? 0) || (previous && !nearEnd(element))) return
        await nextTick()
        element.scrollTo({ top: element.scrollHeight, behavior: previous ? 'smooth' : 'auto' })
    },
    { flush: 'pre' },
)
</script>

<template>
    <section class="activity">
        <h3 class="infobar-heading activity-heading">
            <span>{{ t('title') }}</span>
            <span class="tabular-nums">{{ activity.length || '' }}</span>
        </h3>
        <div ref="scroller" class="activity-scroll" data-infobar-scroll>
            <p v-if="!activity.length" class="m-0 text-muted">{{ t('empty') }}</p>
            <ul v-else v-auto-animate class="infobar-list">
                <InfobarActivityRow
                    v-for="entry in activity"
                    :key="entry.key"
                    :entry="entry"
                    :active="activeMessageId === entry.messageId"
                />
            </ul>
        </div>
    </section>
</template>

<style scoped>
.activity {
    display: flex;
    min-height: 0;
    flex: 1;
    flex-direction: column;
    gap: calc(var(--ui-spacing) * 2);
}

.activity-heading {
    padding-inline: calc(var(--ui-spacing) * 4);
}

.activity-scroll {
    min-height: 0;
    flex: 1;
    padding: 0 calc(var(--ui-spacing) * 4) calc(var(--ui-spacing) * 4);
    overflow-y: auto;
    overscroll-behavior: contain;
    scrollbar-width: thin;
}
</style>

<i18n lang="yaml">
ru:
    title: Активность
    empty: Агент пока не создал ни веток, ни worktree, ни файлов.
en:
    title: Activity
    empty: The agent has not created branches, worktrees or files yet.
</i18n>
