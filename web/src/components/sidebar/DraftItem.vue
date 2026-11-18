<script setup lang="ts">
import { computed } from 'vue'
import { PhDeviceMobile, PhPencilSimple } from '@phosphor-icons/vue'
import { RouterLink } from 'vue-router'

import { useDraftIndex } from '@/composables/useDrafts'
import { useShell } from '@/composables/useShell'
import { useI18n } from 'vue-i18n'

const props = defineProps<{ draftKey: string; projectId?: string }>()
const { t } = useI18n()
const drafts = useDraftIndex()
const { closeOnNarrow } = useShell()

const text = computed(() => drafts.get(props.draftKey)?.text ?? '')
const typing = computed(() => drafts.typing(props.draftKey))
const fromOther = computed(() => drafts.fromOther(props.draftKey))
</script>

<template>
    <RouterLink
        :to="projectId ? { name: 'new', query: { project: projectId } } : { name: 'new' }"
        class="nav-row motion-enter no-underline"
        data-draft
        :title="text"
        @click="closeOnNarrow"
    >
        <span class="truncate italic">{{ text }}</span>
        <span class="status-tag">
            <PhDeviceMobile v-if="fromOther" :size="11" />
            <PhPencilSimple v-else :size="11" />
            {{ typing ? t('session.typing') : t('session.draft') }}
            <span v-if="typing" class="typing-dots"><i /><i /><i /></span>
        </span>
    </RouterLink>
</template>
