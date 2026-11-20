<script setup lang="ts">
import { computed } from 'vue'
import {
    Select,
    SelectContent,
    SelectGroup,
    SelectGroupLabel,
    SelectItem,
    SelectItemIndicator,
    SelectItemText,
    SelectSeparator,
    SelectTrigger,
} from '@ui-sdk/vue/select'
import { useI18n } from 'vue-i18n'

import type { AgentId, AgentModel } from '@/shared/agents'

import AgentLogo from '@/components/AgentLogo.vue'
import { useAgents } from '@/composables/useAgents'
import { useComposerSettings } from '@/composables/useComposerSettings'
import { AGENTS } from '@/shared/agents'

const props = defineProps<{ agent: AgentId }>()
const emit = defineEmits<{ switchAgent: [choice: { agent: AgentId; model: string }] }>()
const { t } = useI18n()
const settings = useComposerSettings()
const agents = useAgents()

const SEPARATOR = '/'
const optionValue = (agent: AgentId, model: string) => `${agent}${SEPARATOR}${model}`
const itemLabel = (model: AgentModel) => model.label ?? t('models.default')

const families = computed(() =>
    agents.all.filter((family) => family.id === props.agent || agents.isConnected(family.id)),
)
const items = computed(() =>
    families.value.flatMap((family) =>
        family.models.map((model) => ({
            value: optionValue(family.id, model.id),
            label: `${family.name} ${itemLabel(model)}`,
        })),
    ),
)
const currentModel = computed(() => {
    const id = settings.modelFor(props.agent)
    return AGENTS[props.agent].models.find((model) => model.id === id)
})
const triggerLabel = computed(() => currentModel.value?.label ?? AGENTS[props.agent].name)

const value = computed({
    get: () => optionValue(props.agent, settings.modelFor(props.agent)),
    set: (next: string | null) => {
        if (!next) return
        const [agent, model] = next.split(SEPARATOR) as [AgentId, string]
        if (agent === props.agent) settings.setModel(agent, model)
        else emit('switchAgent', { agent, model })
    },
})
</script>

<template>
    <Select v-model:value="value" :items="items">
        <SelectTrigger
            class="composer-select"
            :aria-label="`${t('composer.model')}: ${AGENTS[agent].name} ${triggerLabel}`"
            v-tooltip="`${AGENTS[agent].name} · ${triggerLabel}`"
        >
            <span class="composer-select-icon"><AgentLogo :agent="agent" decorative /></span>
            <span class="composer-select-label">{{ triggerLabel }}</span>
        </SelectTrigger>
        <SelectContent side="top" align="end">
            <template v-for="(family, index) in families" :key="family.id">
                <SelectSeparator v-if="index > 0" />
                <SelectGroup>
                    <SelectGroupLabel class="flex items-center gap-2">
                        <AgentLogo :agent="family.id" decorative />
                        {{ family.name }}
                        <span v-if="!agents.isConnected(family.id)" class="text-muted">
                            · {{ t('empty.notConnected') }}
                        </span>
                    </SelectGroupLabel>
                    <SelectItem
                        v-for="model in family.models"
                        :key="model.id"
                        :value="optionValue(family.id, model.id)"
                        :disabled="!agents.isConnected(family.id)"
                        v-tooltip="t(model.hint)"
                    >
                        <SelectItemText>{{ itemLabel(model) }}</SelectItemText>
                        <SelectItemIndicator />
                    </SelectItem>
                </SelectGroup>
            </template>
        </SelectContent>
    </Select>
</template>
