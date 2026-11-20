<script setup lang="ts">
import { PhDotsThree, PhPencilSimple, PhPlus, PhTrash } from '@phosphor-icons/vue'
import { Menu, MenuContent, MenuItem, MenuItemIcon, MenuTrigger } from '@ui-sdk/vue/menu'
import { useI18n } from 'vue-i18n'

const emit = defineEmits<{
    newChat: []
    rename: []
    remove: []
}>()
const { t } = useI18n()

const handlers: Record<string, () => void> = {
    newChat: () => emit('newChat'),
    rename: () => emit('rename'),
    remove: () => emit('remove'),
}

function choose(value: string) {
    handlers[value]?.()
}
</script>

<template>
    <Menu @action="choose">
        <MenuTrigger
            class="row-action"
            :aria-label="t('project.actions')"
            v-tooltip="t('project.actions')"
        >
            <PhDotsThree :size="16" weight="bold" />
        </MenuTrigger>
        <MenuContent align="end">
            <MenuItem value="newChat">
                <MenuItemIcon><PhPlus /></MenuItemIcon>
                {{ t('project.newChat') }}
            </MenuItem>
            <MenuItem value="rename">
                <MenuItemIcon><PhPencilSimple /></MenuItemIcon>
                {{ t('project.rename') }}
            </MenuItem>
            <MenuItem value="remove">
                <MenuItemIcon><PhTrash /></MenuItemIcon>
                {{ t('project.delete') }}
            </MenuItem>
        </MenuContent>
    </Menu>
</template>
