<script setup lang="ts">
import { ref, watch } from 'vue'
import { Button } from '@ui-sdk/vue/button'
import {
    Dialog,
    DialogClose,
    DialogContent,
    DialogDescription,
    DialogTitle,
} from '@ui-sdk/vue/dialog'
import { Field, FieldLabel } from '@ui-sdk/vue/field'
import { Input } from '@ui-sdk/vue/input'
import { useI18n } from 'vue-i18n'

import { useSessionGit } from '@/composables/useSessionGit'
import { reportError } from '@/shared/errors'

const open = defineModel<boolean>('open', { required: true })
const { t } = useI18n()
const git = useSessionGit()
const name = ref('')
const saving = ref(false)

watch(open, (value) => {
    if (value) name.value = ''
})

async function create() {
    const branch = name.value.trim()
    if (!branch || saving.value) return
    saving.value = true
    try {
        await git.checkout(branch, true)
        open.value = false
    } catch (error) {
        reportError(error)
    } finally {
        saving.value = false
    }
}
</script>

<template>
    <Dialog v-model:open="open">
        <DialogContent class="flex flex-col gap-3">
            <DialogTitle>{{ t('title') }}</DialogTitle>
            <DialogDescription>{{
                t('hint', { from: git.status.data.value?.branch ?? 'HEAD' })
            }}</DialogDescription>
            <form class="flex flex-col gap-3" @submit.prevent="create">
                <Field>
                    <FieldLabel>{{ t('name') }}</FieldLabel>
                    <Input
                        v-model="name"
                        class="font-mono"
                        required
                        autofocus
                        placeholder="feature/login"
                        :disabled="saving"
                    />
                </Field>
                <div class="flex justify-end gap-2">
                    <DialogClose type="button">{{ t('cancel') }}</DialogClose>
                    <Button type="submit" data-variant="solid" :disabled="!name.trim() || saving">
                        {{ t('create') }}
                    </Button>
                </div>
            </form>
        </DialogContent>
    </Dialog>
</template>

<i18n lang="yaml">
ru:
    title: Новая ветка
    hint: 'Ветка создаётся от {from}, незакоммиченные изменения переходят в неё.'
    name: Название ветки
    cancel: Отмена
    create: Создать и перейти
en:
    title: New branch
    hint: 'The branch starts at {from}; uncommitted changes move with you.'
    name: Branch name
    cancel: Cancel
    create: Create and switch
</i18n>
