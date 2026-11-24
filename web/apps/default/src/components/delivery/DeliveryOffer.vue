<script setup lang="ts">
import { PhShieldCheck } from '@phosphor-icons/vue'
import { Button } from '@ui-sdk/vue/button'
import { DialogDescription, DialogTitle } from '@ui-sdk/vue/dialog'
import { useI18n } from 'vue-i18n'

import type { Delivery, Release } from '@drift/vue/api'

import ReleaseNotes from '@/components/delivery/ReleaseNotes.vue'
import VersionChange from '@/components/delivery/VersionChange.vue'

defineProps<{ delivery: Delivery; release: Release; pending: boolean }>()
const emit = defineEmits<{ update: []; decline: [] }>()
const { t } = useI18n()
</script>

<template>
    <div class="flex min-h-0 flex-col gap-4">
        <div class="flex flex-col gap-1">
            <DialogTitle class="flex items-center gap-2">
                <PhShieldCheck :size="22" />{{ t('title') }}
            </DialogTitle>
            <DialogDescription>{{ t('body') }}</DialogDescription>
        </div>
        <VersionChange :current="delivery.currentVersion" :next="release.version" />
        <ReleaseNotes :release="release" />
        <p v-if="delivery.installation.state === 'failed'" class="m-0 text-danger" role="alert">
            {{ t('failed', { error: delivery.installation.error }) }}
        </p>
        <div class="flex flex-wrap justify-end gap-2">
            <Button :disabled="pending" @click="emit('decline')">{{ t('decline') }}</Button>
            <Button data-variant="solid" :disabled="pending" @click="emit('update')">
                {{ delivery.installation.state === 'failed' ? t('retry') : t('update') }}
            </Button>
        </div>
    </div>
</template>

<i18n lang="yaml">
ru:
    title: Вышла новая версия Drift
    body: Обновление скачается из GitHub Releases, проверится по контрольной сумме и установится с перезапуском сервера — это займёт меньше минуты.
    failed: 'Не удалось обновиться: {error}'
    decline: Не обновлять
    update: Обновить
    retry: Попробовать снова
en:
    title: A new version of Drift is out
    body: The update is downloaded from GitHub Releases, verified by its checksum and installed with a server restart; it takes under a minute.
    failed: 'The update failed: {error}'
    decline: "Don't update"
    update: Update
    retry: Try again
</i18n>
