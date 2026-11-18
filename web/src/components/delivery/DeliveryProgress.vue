<script setup lang="ts">
import { computed } from 'vue'
import {
    Progress,
    ProgressIndicator,
    ProgressLabel,
    ProgressTrack,
    ProgressValue,
} from '@ui-sdk/vue/progress'
import { useI18n } from 'vue-i18n'

import type { DeliveryInstallation } from '@/gen'

const props = defineProps<{ installation: DeliveryInstallation }>()
const { t } = useI18n()

const restarting = computed(() => props.installation.state === 'restarting')
const percent = computed(() => {
    const { downloadedBytes, totalBytes } = props.installation
    if (restarting.value) return null
    return totalBytes > 0 ? Math.min(100, (downloadedBytes / totalBytes) * 100) : null
})
</script>

<template>
    <div class="flex flex-col gap-3">
        <Progress :value="percent" :max="100">
            <ProgressLabel>{{
                restarting
                    ? t('restarting', { version: installation.targetVersion })
                    : t('downloading', { version: installation.targetVersion })
            }}</ProgressLabel>
            <ProgressValue v-if="percent !== null" />
            <ProgressTrack><ProgressIndicator /></ProgressTrack>
        </Progress>
        <p class="m-0 text-muted">{{ t('hint') }}</p>
    </div>
</template>

<i18n lang="yaml">
ru:
    downloading: 'Загружаем {version} с GitHub…'
    restarting: 'Устанавливаем {version} и перезапускаем Drift…'
    hint: Страница перезагрузится сама, как только новая версия запустится. Работающие агенты будут прерваны.
en:
    downloading: 'Downloading {version} from GitHub…'
    restarting: 'Installing {version} and restarting Drift…'
    hint: The page reloads by itself once the new version is up. Running agents will be interrupted.
</i18n>
