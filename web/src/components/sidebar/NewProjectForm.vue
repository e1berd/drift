<script setup lang="ts">
import { onMounted, ref, useTemplateRef } from 'vue'
import { Button } from '@ui-sdk/vue/button'
import { Field, FieldDescription, FieldError, FieldLabel } from '@ui-sdk/vue/field'
import { Input } from '@ui-sdk/vue/input'
import { useI18n } from 'vue-i18n'

import { useCreateProject } from '@/gen'
import FolderPicker from '@/components/sidebar/FolderPicker.vue'
import { errorMessage } from '@/shared/errors'

const emit = defineEmits<{ done: [] }>()
const { t } = useI18n()

const name = ref('')
const path = ref('')
const pathTouched = ref(false)
const form = useTemplateRef<HTMLFormElement>('form')

const create = useCreateProject({ mutation: { onSuccess: () => emit('done') } })

function submit() {
    create.mutate({ body: { name: name.value, path: path.value } })
}

function suggestPath() {
    if (!pathTouched.value) path.value = name.value.trim().toLowerCase().replaceAll(/\s+/g, '-')
}

onMounted(() => form.value?.querySelector('input')?.focus())
</script>

<template>
    <form
        ref="form"
        class="flex flex-col gap-2 pb-2"
        @submit.prevent="submit"
        @keydown.esc="emit('done')"
    >
        <Field>
            <FieldLabel>{{ t('project.name') }}</FieldLabel>
            <Input v-model="name" required maxlength="120" @input="suggestPath" />
        </Field>
        <Field :invalid="!!create.error.value">
            <FieldLabel>{{ t('project.path') }}</FieldLabel>
            <div class="flex gap-2">
                <Input v-model="path" class="min-w-0 flex-1" required @input="pathTouched = true" />
                <FolderPicker v-model="path" @update:model-value="pathTouched = true" />
            </div>
            <FieldDescription>{{ t('project.pathHint') }}</FieldDescription>
            <FieldError v-if="create.error.value" match>{{
                errorMessage(create.error.value)
            }}</FieldError>
        </Field>
        <div class="flex justify-end gap-2">
            <Button type="button" @click="emit('done')">{{ t('project.cancel') }}</Button>
            <Button
                type="submit"
                data-variant="solid"
                :disabled="!name.trim() || !path.trim() || create.isPending.value"
            >
                {{ t('project.create') }}
            </Button>
        </div>
    </form>
</template>
