<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useTheme } from 'vuetify'

import { useLocale } from '@drift/vue/composables/useLocale'
import { colorSchemes, readColorScheme, saveColorScheme } from '@drift/sdk/colorScheme'
import type { ColorScheme } from '@drift/sdk/colorScheme'
import { locales } from '@drift/sdk/locale'
import type { Locale } from '@drift/sdk/locale'

const { t } = useI18n()
const theme = useTheme()
const { locale, setLocale } = useLocale()
const scheme = ref<ColorScheme>(readColorScheme())

const schemeChoice = computed({
    get: () => scheme.value,
    set: (next: ColorScheme) => {
        scheme.value = next
        saveColorScheme(next)
        theme.change(next)
    },
})
const languageChoice = computed({
    get: () => locale.value as Locale,
    set: (next: Locale) => setLocale(next),
})
</script>

<template>
    <v-card :title="t('settings.appearance')" variant="flat">
        <v-card-text class="d-flex flex-column ga-4">
            <div class="d-flex flex-wrap align-center justify-space-between ga-2">
                <span>{{ t('settings.theme') }}</span>
                <v-btn-toggle v-model="schemeChoice" mandatory divided variant="outlined">
                    <v-btn v-for="item in colorSchemes" :key="item" :value="item">{{
                        t(`theme.${item}`)
                    }}</v-btn>
                </v-btn-toggle>
            </div>
            <div class="d-flex flex-wrap align-center justify-space-between ga-2">
                <span>{{ t('settings.language') }}</span>
                <v-btn-toggle v-model="languageChoice" mandatory divided variant="outlined">
                    <v-btn v-for="item in locales" :key="item" :value="item">{{
                        t(`language.${item}`)
                    }}</v-btn>
                </v-btn-toggle>
            </div>
        </v-card-text>
    </v-card>
</template>
