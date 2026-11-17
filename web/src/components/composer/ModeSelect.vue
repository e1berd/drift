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
</script>

<template>
    <Select v-model:value="value" :items="items">
        <SelectTrigger :aria-label="t('composer.permissions')"><SelectValue /></SelectTrigger>
        <SelectContent side="top" align="end">
            <SelectItem
                v-for="mode in PERMISSION_MODES"
                :key="mode.id"
                :value="mode.id"
                :title="t(mode.hint)"
            >
                <SelectItemText>{{ mode.label }}</SelectItemText>
                <SelectItemIndicator />
            </SelectItem>
        </SelectContent>
    </Select>
</template>
