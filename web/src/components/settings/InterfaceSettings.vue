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
import type { Locale } from '@/shared/preferences'

import { useLocale } from '@/composables/useLocale'
import { colorScheme, locales, sidebarSide } from '@/shared/preferences'

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
        <div class="pref-rows">
            <div class="pref-row">
                <span id="pref-side">{{ t('settings.sidebarSide') }}</span>
                <SegmentedControl
                    v-model:value="sideValue"
                    class="pref-seg"
                    aria-labelledby="pref-side"
                >
                    <SegmentedControlIndicator class="pref-seg-ind" />
                    <SegmentedControlItem value="left" class="pref-seg-item">
                        <PhSidebarSimple :size="15" />{{ t('settings.sidebarLeft') }}
                    </SegmentedControlItem>
                    <SegmentedControlItem value="right" class="pref-seg-item">
                        <PhSidebarSimple :size="15" class="-scale-x-100" />{{
                            t('settings.sidebarRight')
                        }}
                    </SegmentedControlItem>
                </SegmentedControl>
            </div>
            <div class="pref-row">
                <span id="pref-theme">{{ t('settings.theme') }}</span>
                <SegmentedControl
                    v-model:value="schemeValue"
                    class="pref-seg"
                    aria-labelledby="pref-theme"
                >
                    <SegmentedControlIndicator class="pref-seg-ind" />
                    <SegmentedControlItem value="system" class="pref-seg-item"
                        ><PhDesktop :size="15" />{{ t('theme.system') }}</SegmentedControlItem
                    >
                    <SegmentedControlItem value="light" class="pref-seg-item"
                        ><PhSun :size="15" />{{ t('theme.light') }}</SegmentedControlItem
                    >
                    <SegmentedControlItem value="dark" class="pref-seg-item"
                        ><PhMoon :size="15" />{{ t('theme.dark') }}</SegmentedControlItem
                    >
                </SegmentedControl>
            </div>
            <div class="pref-row">
                <span id="pref-lang">{{ t('settings.language') }}</span>
                <SegmentedControl
                    v-model:value="langValue"
                    class="pref-seg"
                    aria-labelledby="pref-lang"
                >
                    <SegmentedControlIndicator class="pref-seg-ind" />
                    <SegmentedControlItem value="ru" class="pref-seg-item">{{
                        t('language.ru')
                    }}</SegmentedControlItem>
                    <SegmentedControlItem value="en" class="pref-seg-item">{{
                        t('language.en')
                    }}</SegmentedControlItem>
                </SegmentedControl>
            </div>
        </div>
    </section>
</template>

<style>
.pref-rows {
    display: flex;
    flex-direction: column;
    overflow: hidden;
    border: 1px solid var(--ui-border);
    border-radius: 12px;
    background: var(--surface);
}

.pref-row {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 12px 14px;
    font-size: 13.5px;

    & + .pref-row {
        border-top: 1px solid var(--ui-border);
    }
}

.pref-seg {
    position: relative;
    display: inline-flex;
    gap: 2px;
    padding: 3px;
    border: 1px solid var(--ui-border);
    border-radius: 9px;
    background: var(--surface-2);
}

.pref-seg-ind {
    position: absolute;
    top: var(--segmented-control-indicator-top);
    left: var(--segmented-control-indicator-left);
    width: var(--segmented-control-indicator-width);
    height: var(--segmented-control-indicator-height);
    border-radius: 6px;
    background: var(--surface);
    box-shadow: inset 0 0 0 1px var(--ui-border);
    transition:
        left var(--dur-3) var(--ease-out),
        width var(--dur-3) var(--ease-out);
}

.pref-seg-item {
    position: relative;
    display: inline-flex;
    height: 28px;
    align-items: center;
    gap: 6px;
    padding: 0 10px;
    border: 0;
    border-radius: 6px;
    background: none;
    color: var(--ui-muted-foreground);
    font-size: 12.5px;
    font-weight: 500;
    cursor: pointer;
    transition: color var(--dur-2);

    &[data-checked] {
        color: var(--ui-foreground);
    }
}
</style>
