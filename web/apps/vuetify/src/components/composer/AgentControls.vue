<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

import { MODE_ICONS } from '@/shared/modeIcons'
import { useComposerSettings } from '@drift/vue/composables/useComposerSettings'
import { AGENTS } from '@drift/sdk/agents'
import type { AgentId, PermissionMode } from '@drift/sdk/agents'

const props = defineProps<{ agent: AgentId; sessionId?: string }>()
const { t } = useI18n()
const settings = useComposerSettings()

const models = computed(() =>
    AGENTS[props.agent].models.map((model) => ({
        value: model.id,
        title: model.label ?? t('drift.models.default'),
        subtitle: t(model.hint),
    })),
)
const modes = computed(() =>
    AGENTS[props.agent].modes.map((mode) => ({
        value: mode.id,
        title: mode.label,
        subtitle: t(mode.hint),
        icon: MODE_ICONS[mode.glyph],
    })),
)
const model = computed({
    get: () => settings.modelFor(props.agent, props.sessionId),
    set: (next: string) => settings.setModel(props.agent, next, props.sessionId),
})
const mode = computed({
    get: () => settings.modeFor(props.agent),
    set: (next: PermissionMode) => settings.setMode(props.agent, next),
})
const modeIcon = computed(() => modes.value.find((item) => item.value === mode.value)?.icon)
</script>

<template>
    <v-select
        v-model="model"
        :items="models"
        :aria-label="t('composer.model')"
        variant="plain"
        class="agent-control"
    >
        <template #item="{ props: item, item: option }">
            <v-list-item v-bind="item" :subtitle="option.subtitle" />
        </template>
    </v-select>
    <v-select
        v-model="mode"
        :items="modes"
        :disabled="modes.length === 1"
        :prepend-inner-icon="modeIcon"
        :aria-label="t('composer.mode')"
        variant="plain"
        class="agent-control"
    >
        <template #item="{ props: item, item: option }">
            <v-list-item v-bind="item" :prepend-icon="option.icon" :subtitle="option.subtitle" />
        </template>
    </v-select>
</template>

<style scoped>
.agent-control {
    flex: 0 1 auto;
    min-inline-size: 120px;
}
</style>
