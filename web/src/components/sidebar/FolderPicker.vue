<script setup lang="ts">
import { ref } from 'vue'
import { PhFolderOpen } from '@phosphor-icons/vue'
import { Button } from '@ui-sdk/vue/button'
import {
    Dialog,
    DialogClose,
    DialogContent,
    DialogDescription,
    DialogTitle,
} from '@ui-sdk/vue/dialog'
import { useI18n } from 'vue-i18n'

import FolderTree from '@/components/sidebar/FolderTree.vue'

const path = defineModel<string>({ required: true })
const { t } = useI18n()

const open = ref(false)
const selected = ref<string | null>(null)

function browse() {
    selected.value = path.value.trim() || null
    open.value = true
}

function choose() {
    if (!selected.value) return
    path.value = selected.value
    open.value = false
}
</script>

<template>
    <Button type="button" :aria-label="t('browse')" :title="t('browse')" @click="browse">
        <PhFolderOpen />
    </Button>
    <Dialog v-model:open="open">
        <DialogContent class="flex flex-col gap-3">
            <DialogTitle>{{ t('title') }}</DialogTitle>
            <DialogDescription>{{ t('hint') }}</DialogDescription>
            <FolderTree v-if="open" v-model="selected" />
            <div class="flex items-center justify-end gap-2">
                <span class="mr-auto truncate text-muted">{{ selected ?? '' }}</span>
                <DialogClose>{{ t('cancel') }}</DialogClose>
                <Button type="button" data-variant="solid" :disabled="!selected" @click="choose">
                    {{ t('choose') }}
                </Button>
            </div>
        </DialogContent>
    </Dialog>
</template>

<i18n lang="yaml">
ru:
    browse: 'Выбрать папку на сервере'
    title: 'Папка проекта'
    hint: 'Пути считаются от рабочего каталога сервера. Отметьте папку и нажмите «Выбрать».'
    cancel: 'Отмена'
    choose: 'Выбрать'
en:
    browse: 'Pick a folder on the server'
    title: 'Project folder'
    hint: 'Paths are relative to the server workspace. Select a folder and press Choose.'
    cancel: 'Cancel'
    choose: 'Choose'
</i18n>
