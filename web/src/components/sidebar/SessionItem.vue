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
        class="nav-row no-underline"
        :class="{ 'motion-flash': freshSessions.has(session.id) }"
        :data-active="active || undefined"
        :aria-current="active ? 'page' : undefined"
        @click="closeOnNarrow"
    >
        <span class="truncate">{{ session.title }}</span>
        <span class="status-tag">
            <RunSpinner v-if="running" :size="14" />
            <span v-else-if="typing" class="status-tag">
                <PhDeviceMobile :size="11" />{{ t('session.typing')
                }}<span class="typing-dots"><i /><i /><i /></span>
            </span>
            <span v-else-if="hasDraft" class="status-tag" :title="draftTitle"
                ><PhPencilSimple :size="11" />{{ t('session.draft') }}</span
            >
            <time v-else :datetime="session.updateTime">{{
                relativeTime(session.updateTime, locale)
            }}</time>
        </span>
        <AgentLogo :agent="session.agent" />
    </RouterLink>
</template>
