<script setup lang="ts">
import { computed } from 'vue'
import { PhDesktop, PhMoon, PhSidebarSimple, PhSun } from '@phosphor-icons/vue'
import { usePreference } from '@ui-sdk/vue/preferences'
import {
    SegmentedControl,
    SegmentedControlIndicator,
    SegmentedControlItem,
} from '@ui-sdk/vue/segmented-control'

import { useI18n } from 'vue-i18n'

import type { ColorSchemeValue } from '@ui-sdk/core'
import type { Locale } from '@drift/sdk/locale'

import { useLocale } from '@drift/vue/composables/useLocale'
import { colorScheme, sidebarSide } from '@/shared/preferences'
import { locales } from '@drift/sdk/locale'

const { t } = useI18n()
const side = usePreference(sidebarSide)
const scheme = usePreference(colorScheme)
const { locale, setLocale } = useLocale()

const sideValue = computed({
    get: () => side.value.value,
    set: (value: string | null) => value && sidebarSide.set(value as 'left' | 'right'),
})
const schemeValue = computed({
    get: () => scheme.value.value,
    set: (value: string | null) => value && colorScheme.set(value as ColorSchemeValue),
})
const langValue = computed({
    get: () => locale.value,
    set: (value: string | null) => {
        if (locales.includes(value as Locale)) setLocale(value as Locale)
    },
})
</script>

<template>
    <section class="flex flex-col gap-4">
        <h2 class="m-0 text-base font-semibold">{{ t('settings.interface') }}</h2>
        <div class="flex flex-wrap items-center justify-between gap-3">
            <span id="pref-side">{{ t('settings.sidebarSide') }}</span>
            <SegmentedControl v-model:value="sideValue" aria-labelledby="pref-side">
                <SegmentedControlIndicator />
                <SegmentedControlItem value="left"
                    ><PhSidebarSimple />{{ t('settings.sidebarLeft') }}</SegmentedControlItem
                >
                <SegmentedControlItem value="right">
                    <PhSidebarSimple class="-scale-x-100" />{{ t('settings.sidebarRight') }}
                </SegmentedControlItem>
            </SegmentedControl>
        </div>
        <div class="flex flex-wrap items-center justify-between gap-3">
            <span id="pref-theme">{{ t('settings.theme') }}</span>
            <SegmentedControl v-model:value="schemeValue" aria-labelledby="pref-theme">
                <SegmentedControlIndicator />
                <SegmentedControlItem value="system"
                    ><PhDesktop />{{ t('theme.system') }}</SegmentedControlItem
                >
                <SegmentedControlItem value="light"
                    ><PhSun />{{ t('theme.light') }}</SegmentedControlItem
                >
                <SegmentedControlItem value="dark"
                    ><PhMoon />{{ t('theme.dark') }}</SegmentedControlItem
                >
            </SegmentedControl>
        </div>
        <div class="flex flex-wrap items-center justify-between gap-3">
            <span id="pref-lang">{{ t('settings.language') }}</span>
            <SegmentedControl v-model:value="langValue" aria-labelledby="pref-lang">
                <SegmentedControlIndicator />
                <SegmentedControlItem v-for="code in locales" :key="code" :value="code">
                    {{ t(`language.${code}`) }}
                </SegmentedControlItem>
            </SegmentedControl>
        </div>
    </section>
</template>
