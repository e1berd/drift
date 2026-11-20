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
            class="project-menu-trigger"
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

<style scoped>
.project-menu-trigger {
    flex: none;
    width: 20px;
    height: 20px;
    padding: 0;
    border: 0;
    border-radius: calc(var(--ui-radius) / 2);
    background: none;
    color: var(--ui-muted-foreground);

    &:hover:not([data-disabled]),
    &[data-open] {
        background: none;
        color: var(--ui-foreground);
    }
}
</style>
