<script setup lang="ts">
import { computed } from 'vue'
import { PhCaretDown, PhCheck, PhSparkle } from '@phosphor-icons/vue'
import {
    Select,
    SelectContent,
    SelectItem,
    SelectItemText,
    SelectTrigger,
    SelectValue,
} from '@ui-sdk/vue/select'

import { useComposerSettings } from '@/composables/useComposerSettings'
import type { AgentId } from '@/shared/agents'
import { AGENTS } from '@/shared/agents'
import { useI18n } from 'vue-i18n'

const props = defineProps<{ agent: AgentId }>()
const { t } = useI18n()
const settings = useComposerSettings()

const models = computed(() => AGENTS[props.agent].models)
const items = computed(() => models.value.map((m) => ({ value: m.id, label: m.label })))
const value = computed({
    get: () => settings.modelFor(props.agent),
    set: (v: string | null) => {
        if (v) settings.setModel(props.agent, v)
    },
})
</script>

<template>
    <Select v-model:value="value" :items="items">
        <SelectTrigger class="chip" :aria-label="t('composer.model')">
            <PhSparkle :size="15" />
            <SelectValue />
            <PhCaretDown :size="13" />
        </SelectTrigger>
        <SelectContent class="popup w-[280px]" side="top" align="end">
            <div class="popup-label">{{ t('composer.model') }} · {{ AGENTS[agent].name }}</div>
            <SelectItem v-for="m in models" :key="m.id" :value="m.id" class="option">
                <PhCheck :size="15" class="check" />
                <span>
                    <b><SelectItemText>{{ m.label }}</SelectItemText></b>
                    <small>{{ t(m.hint) }}</small>
                </span>
                <span />
            </SelectItem>
        </SelectContent>
    </Select>
</template>
