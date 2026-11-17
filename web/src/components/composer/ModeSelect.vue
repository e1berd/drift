<script setup lang="ts">
import { computed } from 'vue'
import { PhCaretDown, PhCheck, PhShieldCheck } from '@phosphor-icons/vue'
import {
    Select,
    SelectContent,
    SelectItem,
    SelectItemText,
    SelectTrigger,
    SelectValue,
} from '@ui-sdk/vue/select'

import { useComposerSettings } from '@/composables/useComposerSettings'
import { PERMISSION_MODES, type PermissionMode } from '@/shared/agents'
import { useI18n, type MessageKey } from '@/shared/i18n'

const { t } = useI18n()
const settings = useComposerSettings()

const items = PERMISSION_MODES.map((m) => ({ value: m.id, label: m.label }))
const value = computed({
    get: () => settings.mode(),
    set: (v: PermissionMode | null) => {
        if (v) settings.setMode(v)
    },
})
</script>

<template>
    <Select v-model:value="value" :items="items">
        <SelectTrigger
            class="chip mode-chip"
            :data-mode="value"
            :aria-label="t('composer.permissions')"
        >
            <PhShieldCheck :size="15" />
            <SelectValue />
            <PhCaretDown :size="13" />
        </SelectTrigger>
        <SelectContent class="popup w-[320px]" side="top" align="end">
            <div class="popup-label">{{ t('composer.permissions') }}</div>
            <SelectItem v-for="m in PERMISSION_MODES" :key="m.id" :value="m.id" class="option">
                <PhCheck :size="15" class="check" />
                <span>
                    <b><SelectItemText>{{ m.label }}</SelectItemText></b>
                    <small>{{ t(m.hint as MessageKey) }}</small>
                </span>
                <span v-if="m.id === 'acceptEdits' || m.id === 'plan'" class="kbd">⇧Tab</span>
            </SelectItem>
        </SelectContent>
    </Select>
</template>

<style>
.mode-chip[data-mode='acceptEdits'] {
    color: var(--accent);
}

.mode-chip[data-mode='auto'] {
    color: var(--auto);
}

.mode-chip[data-mode='plan'] {
    color: var(--plan);
}

.mode-chip[data-mode='bypassPermissions'] {
    color: var(--ui-destructive);
}
</style>
