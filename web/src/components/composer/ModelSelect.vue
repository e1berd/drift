<script setup lang="ts">
import { computed } from 'vue'
import {
    Select,
    SelectContent,
    SelectItem,
    SelectItemIndicator,
    SelectItemText,
    SelectTrigger,
    SelectValue,
} from '@ui-sdk/vue/select'
import { useI18n } from 'vue-i18n'

import type { AgentId } from '@/shared/agents'

import { useComposerSettings } from '@/composables/useComposerSettings'
import { AGENTS } from '@/shared/agents'

const props = defineProps<{ agent: AgentId }>()
const { t } = useI18n()
const settings = useComposerSettings()

const models = computed(() => AGENTS[props.agent].models)
const items = computed(() => models.value.map((model) => ({ value: model.id, label: model.label })))
const value = computed({
    get: () => settings.modelFor(props.agent),
    set: (model: string | null) => {
        if (model) settings.setModel(props.agent, model)
    },
})
</script>

<template>
    <Select v-model:value="value" :items="items">
        <SelectTrigger :aria-label="t('composer.model')"><SelectValue /></SelectTrigger>
        <SelectContent side="top" align="end">
            <SelectItem
                v-for="model in models"
                :key="model.id"
                :value="model.id"
                :title="t(model.hint)"
            >
                <SelectItemText>{{ model.label }}</SelectItemText>
                <SelectItemIndicator />
            </SelectItem>
        </SelectContent>
    </Select>
</template>
