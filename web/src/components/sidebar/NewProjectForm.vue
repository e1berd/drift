<script setup lang="ts">
import { onMounted, ref, useTemplateRef } from 'vue'

import { useCreateProject } from '@/gen'
import { errorMessage } from '@/shared/errors'
import { useI18n } from 'vue-i18n'

const emit = defineEmits<{ done: [] }>()
const { t } = useI18n()

const name = ref('')
const path = ref('')
const pathTouched = ref(false)
const nameInput = useTemplateRef<HTMLInputElement>('nameInput')

const create = useCreateProject({ mutation: { onSuccess: () => emit('done') } })

function submit() {
    create.mutate({ body: { name: name.value, path: path.value } })
}

function onName() {
    if (!pathTouched.value) path.value = name.value.trim().toLowerCase().replaceAll(/\s+/g, '-')
}

onMounted(() => nameInput.value?.focus())
</script>

<template>
    <form
        class="flex flex-col gap-2 rounded-[10px] bg-surface-2 p-2.5"
        @submit.prevent="submit"
        @keydown.esc="emit('done')"
    >
        <label class="flex flex-col gap-1 text-[12px] text-muted">
            {{ t('project.name') }}
            <input
                ref="nameInput"
                v-model="name"
                class="field"
                required
                maxlength="120"
                @input="onName"
            />
        </label>
        <label class="flex flex-col gap-1 text-[12px] text-muted">
            {{ t('project.path') }}
            <input v-model="path" class="field font-mono" required @input="pathTouched = true" />
        </label>
        <p v-if="create.error.value" class="text-[12px] text-danger">
            {{ errorMessage(create.error.value) }}
        </p>
        <div class="flex justify-end gap-1.5">
            <button type="button" class="btn h-7" @click="emit('done')">
                {{ t('project.cancel') }}
            </button>
            <button
                type="submit"
                class="btn primary h-7"
                :disabled="!name.trim() || !path.trim() || create.isPending.value"
            >
                {{ t('project.create') }}
            </button>
        </div>
    </form>
</template>
