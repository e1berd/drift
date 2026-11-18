<script setup lang="ts">
import { computed, ref } from 'vue'
import { PhCaretDown, PhCheck, PhGitBranch, PhPlus } from '@phosphor-icons/vue'
import {
    Menu,
    MenuContent,
    MenuItem,
    MenuItemIcon,
    MenuSeparator,
    MenuTrigger,
} from '@ui-sdk/vue/menu'
import { useI18n } from 'vue-i18n'

import GitNewBranchDialog from '@/components/git/GitNewBranchDialog.vue'
import { useSessionGit } from '@/composables/useSessionGit'
import { useListGitBranches } from '@/gen'
import { reportError } from '@/shared/errors'

const CREATE = 'create'
const BRANCH_PREFIX = 'branch:'

const { t } = useI18n()
const git = useSessionGit()
const open = ref(false)
const creating = ref(false)

const branches = useListGitBranches(
    { path: () => ({ sessionId: git.sessionId.value }) },
    { query: { enabled: open, staleTime: 0 } },
)
const label = computed(() => {
    const status = git.status.data.value
    if (status?.branch) return status.branch
    return status?.head ? t('detached', { head: status.head.slice(0, 7) }) : 'HEAD'
})

function choose(value: string) {
    if (value === CREATE) {
        creating.value = true
        return
    }
    const branch = value.slice(BRANCH_PREFIX.length)
    if (branch !== git.status.data.value?.branch) git.checkout(branch).catch(reportError)
}
</script>

<template>
    <Menu v-model:open="open" @action="choose">
        <MenuTrigger class="branch-trigger" :disabled="git.busy.value" :title="t('switch')">
            <PhGitBranch class="shrink-0" />
            <span class="truncate font-mono">{{ label }}</span>
            <PhCaretDown class="shrink-0" />
        </MenuTrigger>
        <MenuContent class="max-h-[60dvh] overflow-auto">
            <MenuItem v-if="branches.isPending.value" value="" disabled>{{
                t('loading')
            }}</MenuItem>
            <MenuItem
                v-for="branch in branches.data.value?.branches ?? []"
                :key="branch.name"
                :value="BRANCH_PREFIX + branch.name"
            >
                <MenuItemIcon>
                    <PhCheck v-if="branch.current" />
                </MenuItemIcon>
                <span class="truncate font-mono">{{ branch.name }}</span>
            </MenuItem>
            <MenuSeparator />
            <MenuItem :value="CREATE">
                <MenuItemIcon><PhPlus /></MenuItemIcon>
                {{ t('newBranch') }}
            </MenuItem>
        </MenuContent>
    </Menu>
    <GitNewBranchDialog v-model:open="creating" />
</template>

<style scoped>
.branch-trigger {
    min-width: 0;
    max-width: 60%;
    gap: calc(var(--ui-spacing) * 1.5);
}
</style>

<i18n lang="yaml">
ru:
    switch: Сменить ветку
    detached: 'HEAD отсоединён ({head})'
    loading: Загрузка…
    newBranch: Новая ветка…
en:
    switch: Switch branch
    detached: 'Detached HEAD ({head})'
    loading: Loading…
    newBranch: New branch…
</i18n>
