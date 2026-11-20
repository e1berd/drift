<script setup lang="ts">
import { computed } from 'vue'
import {
    Select,
    SelectContent,
    SelectItem,
    SelectItemIcon,
    SelectItemIndicator,
    SelectItemText,
    SelectTrigger,
} from '@ui-sdk/vue/select'
import { useI18n } from 'vue-i18n'

import type { AgentId, PermissionMode } from '@/shared/agents'

import { useComposerSettings } from '@/composables/useComposerSettings'
import { AGENTS } from '@/shared/agents'

const props = defineProps<{ agent: AgentId }>()
const { t } = useI18n()
const settings = useComposerSettings()

const modes = computed(() => AGENTS[props.agent].modes)
const items = computed(() => modes.value.map((mode) => ({ value: mode.id, label: mode.label })))
const value = computed({
    get: () => settings.modeFor(props.agent),
    set: (mode: PermissionMode | null) => {
        if (mode) settings.setMode(props.agent, mode)
    },
})
const current = computed(
    () => modes.value.find((mode) => mode.id === value.value) ?? modes.value.at(0)!,
)
</script>

<template>
    <span
        v-if="modes.length === 1"
        class="composer-select composer-mode-fixed"
        tabindex="0"
        :data-mode="current.id"
        :aria-label="`${t('composer.permissions')}: ${current.label}. ${t(current.hint)}`"
        v-tooltip="t(current.hint)"
    >
        <span class="composer-select-icon"><component :is="current.icon" weight="bold" /></span>
        <span class="composer-select-label">{{ current.short }}</span>
    </span>
    <Select v-else v-model:value="value" :items="items">
        <SelectTrigger
            class="composer-select"
            :data-mode="current.id"
            :aria-label="`${t('composer.permissions')}: ${current.label}`"
            v-tooltip="current.label"
        >
            <span class="composer-select-icon"><component :is="current.icon" weight="bold" /></span>
            <span class="composer-select-label">{{ current.short }}</span>
        </SelectTrigger>
        <SelectContent side="top" align="end">
            <SelectItem
                v-for="mode in modes"
                :key="mode.id"
                :value="mode.id"
                :data-mode="mode.id"
                v-tooltip="t(mode.hint)"
            >
                <SelectItemIcon><component :is="mode.icon" weight="bold" /></SelectItemIcon>
                <SelectItemText>{{ mode.label }}</SelectItemText>
                <SelectItemIndicator />
            </SelectItem>
        </SelectContent>
    </Select>
</template>

<style scoped>
[data-mode] {
    color: var(--mode-default);

    &[data-mode='acceptEdits'] {
        color: var(--mode-acceptEdits);
    }

    &[data-mode='auto'] {
        color: var(--mode-auto);
    }

    &[data-mode='plan'] {
        color: var(--mode-plan);
    }

    &[data-mode='bypassPermissions'] {
        color: var(--mode-bypassPermissions);
    }
}

.composer-mode-fixed {
    display: inline-flex;
    align-items: center;
    height: calc(var(--ui-spacing) * 8);
    border-radius: var(--ui-radius);
    font-size: var(--ui-font-size);
    cursor: default;

    &:focus-visible {
        outline: 2px solid var(--ui-ring);
        outline-offset: 2px;
    }
}
</style>
