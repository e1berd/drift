<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useQueryClient } from '@tanstack/vue-query'
import { Button } from '@ui-sdk/vue/button'
import {
    NumberInput,
    NumberInputControl,
    NumberInputDecrement,
    NumberInputGroup,
    NumberInputIncrement,
} from '@ui-sdk/vue/number-input'
import { toast } from '@ui-sdk/vue/toast'
import { useI18n } from 'vue-i18n'

import { getSettingsQueryKey, useGetSettings, useUpdateSettings } from '@drift/vue/api'
import { useSession } from '@drift/vue/composables/useSession'
import { errorMessage } from '@drift/sdk/errors'

const GB = 1_000_000_000
const MIN_GB = 0.1
const MAX_GB = 1000

const { t } = useI18n()
const { session } = useSession()
const queryClient = useQueryClient()
const settings = useGetSettings()
const update = useUpdateSettings({
    mutation: {
        onSuccess: (saved) => {
            queryClient.setQueryData(getSettingsQueryKey(), saved)
            toast.success(t('saved'))
        },
        onError: (error) => toast.error(errorMessage(error)),
    },
})

const isOwner = computed(() => session.value?.identity.metadata_public?.role === 'superuser')
const savedGb = computed(() =>
    settings.data.value ? roundGb(settings.data.value.maxUploadBytes / GB) : null,
)
const limitGb = ref<number | null>(null)
watch(savedGb, (value) => (limitGb.value = value), { immediate: true })

const changed = computed(() => limitGb.value !== null && limitGb.value !== savedGb.value)

function roundGb(value: number): number {
    return Math.round(value * 10) / 10
}

function save() {
    if (limitGb.value === null) return
    update.mutate({ body: { maxUploadBytes: Math.round(limitGb.value * GB) } })
}
</script>

<template>
    <section v-if="isOwner" class="flex flex-col gap-4">
        <h2 class="m-0 text-base font-semibold">{{ t('title') }}</h2>
        <p class="m-0 text-muted">{{ t('lead') }}</p>
        <div class="flex flex-wrap items-center justify-between gap-3">
            <span id="upload-limit">{{ t('limit') }}</span>
            <div class="flex items-center gap-2">
                <NumberInput
                    v-model:value="limitGb"
                    aria-labelledby="upload-limit"
                    :min="MIN_GB"
                    :max="MAX_GB"
                    :step="0.1"
                    :disabled="!settings.data.value"
                >
                    <NumberInputGroup>
                        <NumberInputDecrement />
                        <NumberInputControl class="w-24" />
                        <NumberInputIncrement />
                    </NumberInputGroup>
                </NumberInput>
                <span class="text-muted">{{ t('unit') }}</span>
                <Button
                    data-variant="solid"
                    :disabled="!changed || update.isPending.value"
                    @click="save"
                    >{{ t('save') }}</Button
                >
            </div>
        </div>
    </section>
</template>

<i18n lang="yaml">
ru:
    title: Файлы
    lead: Самый большой файл, который можно прикрепить к сообщению. Загрузка докачивается после обрыва связи и перезагрузки страницы.
    limit: Размер файла не больше
    unit: ГБ
    save: Сохранить
    saved: Лимит сохранён
en:
    title: Files
    lead: The largest file a message can carry. Uploads resume after a dropped connection or a page reload.
    limit: Largest file
    unit: GB
    save: Save
    saved: Limit saved
</i18n>
