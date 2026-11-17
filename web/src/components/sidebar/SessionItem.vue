<script setup lang="ts">
import { computed } from 'vue'
import { PhDeviceMobile, PhPencilSimple } from '@phosphor-icons/vue'
import { RouterLink, useRoute } from 'vue-router'

import AgentLogo from '@/components/AgentLogo.vue'
import RunSpinner from '@/components/RunSpinner.vue'
import { draftKey, useDraftIndex } from '@/composables/useDrafts'
import { useEventStream } from '@/composables/useEventStream'
import { useShell } from '@/composables/useShell'
import type { Session } from '@/gen'
import { useI18n } from 'vue-i18n'
import { relativeTime } from '@/shared/time'

const props = defineProps<{ session: Session }>()
const { t, locale } = useI18n()
const route = useRoute()
const drafts = useDraftIndex()
const { freshSessions } = useEventStream()
const { closeOnNarrow } = useShell()

const key = computed(() => draftKey(props.session.id))
const active = computed(() => route.params.sessionId === props.session.id)
const running = computed(() => props.session.status === 'running')
const typing = computed(() => !active.value && drafts.typing(key.value))
const hasDraft = computed(() => !active.value && !typing.value && drafts.has(key.value))
const draftTitle = computed(() => {
    const d = drafts.get(key.value)
    if (!d) return ''
    return drafts.fromOther(key.value) ? `${t('session.fromOther')}: ${d.text}` : d.text
})
</script>

<template>
    <RouterLink
        :to="{ name: 'session', params: { sessionId: session.id } }"
        class="sess"
        :class="{ 'motion-flash': freshSessions.has(session.id) }"
        :data-active="active || undefined"
        :aria-current="active ? 'page' : undefined"
        @click="closeOnNarrow"
    >
        <span class="truncate">{{ session.title }}</span>
        <span class="flex items-center gap-1 text-[11.5px] text-faint">
            <RunSpinner v-if="running" :size="14" />
            <span v-else-if="typing" class="tag typing">
                <PhDeviceMobile :size="11" />{{ t('session.typing')
                }}<span class="typing-dots"><i /><i /><i /></span>
            </span>
            <span v-else-if="hasDraft" class="tag" :title="draftTitle"
                ><PhPencilSimple :size="11" />{{ t('session.draft') }}</span
            >
            <time v-else :datetime="session.updateTime">{{
                relativeTime(session.updateTime, locale)
            }}</time>
        </span>
        <AgentLogo :agent="session.agent" />
    </RouterLink>
</template>

<style scoped>
.sess {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto 16px;
    align-items: center;
    gap: 8px;
    height: 32px;
    padding: 0 8px 0 10px;
    border-radius: 7px;
    color: var(--ui-muted-foreground);
    font-size: 13px;
    transition:
        background var(--dur-1),
        color var(--dur-1);

    &:hover {
        background: var(--surface-2);
        color: var(--ui-foreground);
    }

    &[data-active] {
        background: var(--accent-soft);
        color: var(--ui-foreground);
    }
}
</style>
