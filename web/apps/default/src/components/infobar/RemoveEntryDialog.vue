<script setup lang="ts">
import { ref } from 'vue'
import { PhCheck } from '@phosphor-icons/vue'
import {
    AlertDialog,
    AlertDialogBackdrop,
    AlertDialogClose,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogTitle,
    createConfirm,
} from '@ui-sdk/vue/alert-dialog'
import { Checkbox, CheckboxIndicator } from '@ui-sdk/vue/checkbox'
import { useI18n } from 'vue-i18n'

import type { ActivityEntry } from '@drift/vue/api'
import type { RemoveChoice } from '@drift/vue/composables/useActivityRemoval'

const { t } = useI18n()
const confirm = createConfirm<ActivityEntry>()
const RemoveDialogContent = AlertDialogContent<ActivityEntry>
const deleteBranch = ref(false)

async function ask(entry: ActivityEntry): Promise<RemoveChoice | null> {
    deleteBranch.value = false
    const confirmed = await confirm(entry)
    return confirmed ? { deleteBranch: deleteBranch.value } : null
}

defineExpose({ ask })
</script>

<template>
    <AlertDialog :controller="confirm.controller">
        <AlertDialogBackdrop />
        <RemoveDialogContent v-slot="{ state }">
            <div v-if="state.payload" class="flex flex-col gap-3">
                <AlertDialogTitle>{{ t(`title.${state.payload.kind}`) }}</AlertDialogTitle>
                <AlertDialogDescription class="flex flex-col gap-2">
                    <code class="break-all">{{ state.payload.name }}</code>
                    <span v-if="state.payload.changes" class="text-danger">
                        {{ t('changes', { count: state.payload.changes }) }}
                    </span>
                    <span>{{ t('agent') }}</span>
                </AlertDialogDescription>
                <label
                    v-if="state.payload.kind === 'worktree' && state.payload.branch"
                    class="flex items-center gap-2"
                >
                    <Checkbox v-model:checked="deleteBranch">
                        <CheckboxIndicator><PhCheck aria-hidden="true" /></CheckboxIndicator>
                    </Checkbox>
                    <span> {{ t('deleteBranch') }} <code>{{ state.payload.branch }}</code> </span>
                </label>
            </div>
            <div class="flex justify-end gap-2">
                <AlertDialogClose>{{ t('cancel') }}</AlertDialogClose>
                <AlertDialogClose data-variant="destructive" value="confirm">
                    {{ t('confirm') }}
                </AlertDialogClose>
            </div>
        </RemoveDialogContent>
    </AlertDialog>
</template>

<i18n lang="yaml">
ru:
    title:
        file: Удалить файл?
        worktree: Удалить worktree?
    changes: 'Незакоммиченных изменений: {count}. Они пропадут без возможности восстановления.'
    agent: Запись останется в активности как удалённая, агент узнает об этом из вашего следующего сообщения.
    deleteBranch: Удалить и ветку
    cancel: Оставить
    confirm: Удалить
en:
    title:
        file: Delete the file?
        worktree: Remove the worktree?
    changes: '{count} uncommitted changes will be lost for good.'
    agent: It stays in the activity log as removed; the agent learns about it with your next message.
    deleteBranch: Also delete the branch
    cancel: Keep
    confirm: Remove
</i18n>
