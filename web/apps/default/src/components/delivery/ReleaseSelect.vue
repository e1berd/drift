<script setup lang="ts">
import { computed } from 'vue'
import {
    Select,
    SelectContent,
    SelectItem,
    SelectItemIndicator,
    SelectItemText,
    SelectTrigger,
    SelectValue,
} from '@ui-sdk/vue/select'
import { useI18n } from 'vue-i18n'

import type { Release } from '@drift/vue/api'

const props = defineProps<{ releases: Release[]; current: string; loading: boolean }>()
const model = defineModel<string>({ required: true })
const { t, locale } = useI18n()

const items = computed(() =>
    props.releases.map((release) => ({ value: release.version, label: release.version })),
)
const value = computed({
    get: () => model.value || null,
    set: (version: string | null) => {
        if (version) model.value = version
    },
})

function describe(release: Release) {
    const date = new Intl.DateTimeFormat(locale.value, { dateStyle: 'medium' }).format(
        new Date(release.publishTime),
    )
    const marks = [
        release.version === props.current ? t('current') : '',
        release.prerelease ? t('prerelease') : '',
    ].filter(Boolean)
    return [date, ...marks].join(' · ')
}
</script>

<template>
    <Select v-model:value="value" :items="items" :disabled="loading || !releases.length">
        <SelectTrigger :aria-label="t('label')">
            <SelectValue :placeholder="loading ? t('loading') : t('empty')" />
        </SelectTrigger>
        <SelectContent>
            <SelectItem v-for="release in releases" :key="release.version" :value="release.version">
                <SelectItemText>
                    <code>{{ release.version }}</code>
                    <span class="text-muted"> — {{ describe(release) }}</span>
                </SelectItemText>
                <SelectItemIndicator />
            </SelectItem>
        </SelectContent>
    </Select>
</template>

<i18n lang="yaml">
ru:
    label: Версия для установки
    loading: Загружаем релизы…
    empty: На GitHub пока нет релизов
    current: текущая
    prerelease: пре-релиз
en:
    label: Version to install
    loading: Loading releases…
    empty: No releases on GitHub yet
    current: current
    prerelease: pre-release
</i18n>
