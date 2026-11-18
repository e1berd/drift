<script setup lang="ts">
import { onMounted, ref, useTemplateRef } from 'vue'
import { Field, FieldError } from '@ui-sdk/vue/field'
import { Input } from '@ui-sdk/vue/input'
import { useI18n } from 'vue-i18n'

import type { TreeEntry } from '@/gen'

import { folderOf, useSessionFiles } from '@/composables/useSessionFiles'
import { useCreateSessionTreeEntry } from '@/gen'
import { errorMessage, reportError } from '@/shared/errors'

const props = defineProps<{ sessionId: string; dir: boolean }>()
const emit = defineEmits<{ done: [] }>()
const { t } = useI18n()
const { selected, reveal } = useSessionFiles()

const startFolder = folderOf(selected.value)
const path = ref(startFolder ? `${startFolder}/` : '')
const form = useTemplateRef<HTMLFormElement>('form')

const create = useCreateSessionTreeEntry({
    mutation: {
        onSuccess: (entry: TreeEntry) => {
            emit('done')
            reveal(entry).catch(reportError)
        },
    },
})

function submit() {
    const value = path.value.trim().replace(/\/+$/, '')
    if (value)
        create.mutate({
            path: { sessionId: props.sessionId },
            body: { path: value, dir: props.dir },
        })
}

onMounted(() => form.value?.querySelector('input')?.focus())
</script>

<template>
    <form ref="form" class="px-2 pt-2" @submit.prevent="submit" @keydown.esc="emit('done')">
        <Field :invalid="!!create.error.value">
            <Input
                v-model="path"
                class="font-mono"
                required
                :placeholder="dir ? t('placeholderDir') : t('placeholderFile')"
                :aria-label="dir ? t('newDir') : t('newFile')"
                :disabled="create.isPending.value"
                @blur="!path.trim() && emit('done')"
            />
            <FieldError v-if="create.error.value" match>{{
                errorMessage(create.error.value)
            }}</FieldError>
        </Field>
    </form>
</template>

<i18n lang="yaml">
ru:
    newFile: Новый файл
    newDir: Новая папка
    placeholderFile: src/main.ts
    placeholderDir: src/components
en:
    newFile: New file
    newDir: New folder
    placeholderFile: src/main.ts
    placeholderDir: src/components
</i18n>
