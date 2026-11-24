<script setup lang="ts">
import {
    AlertDialog,
    AlertDialogBackdrop,
    AlertDialogClose,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogTitle,
    createConfirm,
} from '@ui-sdk/vue/alert-dialog'
import { useI18n } from 'vue-i18n'

import type { GitFileChange } from '@drift/vue/api'

const { t } = useI18n()
const confirm = createConfirm<GitFileChange>()
const DiscardDialogContent = AlertDialogContent<GitFileChange>

defineExpose({ ask: (change: GitFileChange) => confirm(change) })
</script>

<template>
    <AlertDialog :controller="confirm.controller">
        <AlertDialogBackdrop />
        <DiscardDialogContent v-slot="{ state }">
            <div v-if="state.payload" class="flex flex-col gap-2">
                <AlertDialogTitle>{{ t('title') }}</AlertDialogTitle>
                <AlertDialogDescription>
                    <code class="break-all">{{ state.payload.path }}</code>
                    <br />
                    {{ state.payload.status === 'untracked' ? t('untracked') : t('tracked') }}
                </AlertDialogDescription>
            </div>
            <div class="flex justify-end gap-2">
                <AlertDialogClose>{{ t('cancel') }}</AlertDialogClose>
                <AlertDialogClose data-variant="destructive" value="confirm">{{
                    t('confirm')
                }}</AlertDialogClose>
            </div>
        </DiscardDialogContent>
    </AlertDialog>
</template>

<i18n lang="yaml">
ru:
    title: Отменить изменения?
    tracked: Неподготовленные изменения файла пропадут, вернуть их будет нельзя.
    untracked: Файл не отслеживается Git и будет удалён с диска.
    cancel: Оставить
    confirm: Отменить изменения
en:
    title: Discard changes?
    tracked: The unstaged changes of this file will be lost for good.
    untracked: The file is not tracked by Git and will be deleted from disk.
    cancel: Keep
    confirm: Discard changes
</i18n>
