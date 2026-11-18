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

import type { PermissionMode } from '@/shared/agents'

import { useComposerSettings } from '@/composables/useComposerSettings'
import { PERMISSION_MODES } from '@/shared/agents'

const { t } = useI18n()
const settings = useComposerSettings()

const items = PERMISSION_MODES.map((mode) => ({ value: mode.id, label: mode.label }))
const value = computed({
    get: () => settings.mode(),
    set: (mode: PermissionMode | null) => {
        if (mode) settings.setMode(mode)
    },
})
const current = computed(
    () => PERMISSION_MODES.find((mode) => mode.id === value.value) ?? PERMISSION_MODES.at(0)!,
)
</script>

<template>
    <Select v-model:value="value" :items="items">
        <SelectTrigger
            class="composer-select"
            :data-mode="current.id"
            :aria-label="`${t('composer.permissions')}: ${current.label}`"
            :title="current.label"
        >
            <span class="composer-select-icon"><component :is="current.icon" weight="bold" /></span>
            <span class="composer-select-label">{{ current.short }}</span>
        </SelectTrigger>
        <SelectContent side="top" align="end">
            <SelectItem
                v-for="mode in PERMISSION_MODES"
                :key="mode.id"
                :value="mode.id"
                :data-mode="mode.id"
                :title="t(mode.hint)"
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
</style>
