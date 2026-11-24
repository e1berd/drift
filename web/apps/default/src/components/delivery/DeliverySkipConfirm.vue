<script setup lang="ts">
import { PhShieldWarning } from '@phosphor-icons/vue'
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

export type SkipAnswer = 'skip' | 'update' | 'back'

const { t } = useI18n()
const confirm = createConfirm<{ version: string }>()
const SkipDialogContent = AlertDialogContent<{ version: string }>
let updateChosen = false

async function ask(version: string): Promise<SkipAnswer> {
    updateChosen = false
    const skipped = await confirm({ version })
    if (skipped) return 'skip'
    return updateChosen ? 'update' : 'back'
}

function chooseUpdate() {
    updateChosen = true
}

defineExpose({ ask })
</script>

<template>
    <AlertDialog :controller="confirm.controller">
        <AlertDialogBackdrop />
        <SkipDialogContent v-slot="{ state }">
            <div v-if="state.payload" class="flex flex-col gap-2">
                <AlertDialogTitle class="flex items-center gap-2">
                    <PhShieldWarning :size="20" />{{ t('title') }}
                </AlertDialogTitle>
                <AlertDialogDescription class="flex flex-col gap-2">
                    <span>{{ t('body', { version: state.payload.version }) }}</span>
                    <span>{{ t('later') }}</span>
                </AlertDialogDescription>
            </div>
            <div class="flex flex-wrap justify-end gap-2">
                <AlertDialogClose data-variant="destructive" value="confirm">{{
                    t('skip')
                }}</AlertDialogClose>
                <AlertDialogClose data-variant="solid" @click="chooseUpdate">{{
                    t('update')
                }}</AlertDialogClose>
            </div>
        </SkipDialogContent>
    </AlertDialog>
</template>

<i18n lang="yaml">
ru:
    title: Точно не обновлять?
    body: 'Обновления закрывают уязвимости и исправляют ошибки. Drift выполняет код агентов и даёт доступ к терминалу сервера, поэтому старая версия — это риск для безопасности всего сервера. Версия {version} уже вышла.'
    later: Напоминание появится снова через сутки.
    skip: Да, пропустить
    update: Обновить сейчас
en:
    title: Skip the update?
    body: 'Updates close vulnerabilities and fix bugs. Drift runs agent code and gives access to a server terminal, so an outdated version puts the whole server at risk. Version {version} is already out.'
    later: You will be reminded again in a day.
    skip: Yes, skip it
    update: Update now
</i18n>
