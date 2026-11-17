<script setup lang="ts">
import { computed } from 'vue'
import { PhDeviceMobile, PhPencilSimple } from '@phosphor-icons/vue'
import { RouterLink } from 'vue-router'

import { useDraftIndex } from '@/composables/useDrafts'
import { useShell } from '@/composables/useShell'
import { useI18n } from '@/shared/i18n'

const props = defineProps<{ draftKey: string; projectId: string }>()
const { t } = useI18n()
const drafts = useDraftIndex()
const { closeOnNarrow } = useShell()

const text = computed(() => drafts.get(props.draftKey)?.text ?? '')
const typing = computed(() => drafts.typing(props.draftKey))
const fromOther = computed(() => drafts.fromOther(props.draftKey))
</script>

<template>
    <RouterLink
        :to="{ name: 'new', query: { project: projectId } }"
        class="draft motion-enter"
        :title="text"
        @click="closeOnNarrow"
    >
        <span class="truncate italic">{{ text }}</span>
        <span class="tag" :class="{ typing }">
            <PhDeviceMobile v-if="fromOther" :size="11" />
            <PhPencilSimple v-else :size="11" />
            {{ typing ? t('session.typing') : t('session.draft') }}
            <span v-if="typing" class="typing-dots"><i /><i /><i /></span>
        </span>
    </RouterLink>
</template>

<style scoped>
.draft {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    align-items: center;
    gap: 8px;
    height: 34px;
    padding: 0 8px 0 10px;
    border: 1px dashed var(--border-strong);
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
}
</style>
