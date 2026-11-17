<script setup lang="ts">
import { computed } from 'vue'
import { PhGearSix, PhMoon, PhSignOut, PhSun } from '@phosphor-icons/vue'
import { usePreference } from '@ui-sdk/vue/preferences'
import { RouterLink } from 'vue-router'

import { useEventStream } from '@/composables/useEventStream'
import { useSession } from '@/composables/useSession'
import { useShell } from '@/composables/useShell'
import { useI18n } from 'vue-i18n'
import { useLocale } from '@/composables/useLocale'
import { colorScheme } from '@/shared/preferences'

const { t } = useI18n()
const { locale, setLocale } = useLocale()
const { session } = useSession()
const { connected } = useEventStream()
const { closeOnNarrow } = useShell()
const scheme = usePreference(colorScheme)

const name = computed(
    () => session.value?.identity.traits.name || session.value?.identity.traits.email || '',
)
const email = computed(() => session.value?.identity.traits.email ?? '')
const initial = computed(() => (name.value.trim().at(0) ?? '?').toUpperCase())
const dark = computed(() => scheme.value.resolved !== 'light')

function toggleTheme() {
    colorScheme.set(dark.value ? 'light' : 'dark')
}

function toggleLang() {
    setLocale(locale.value === 'en' ? 'ru' : 'en')
}
</script>

<template>
    <div class="foot">
        <span class="avatar" :data-online="connected || undefined" aria-hidden="true">{{
            initial
        }}</span>
        <span class="min-w-0 flex-1 leading-tight">
            <b class="block truncate text-[13px] font-medium">{{ name }}</b>
            <span class="block truncate text-[11.5px] text-faint" :title="email">
                {{ connected ? t('nav.online') : t('nav.offline') }}
            </span>
        </span>
        <RouterLink
            to="/settings"
            class="icon-btn"
            :aria-label="t('nav.settings')"
            :title="t('nav.settings')"
            @click="closeOnNarrow"
        >
            <PhGearSix :size="18" />
        </RouterLink>
        <button
            class="icon-btn"
            :aria-label="dark ? t('theme.toLight') : t('theme.toDark')"
            :title="dark ? t('theme.toLight') : t('theme.toDark')"
            @click="toggleTheme"
        >
            <Transition name="swap" mode="out-in">
                <PhMoon v-if="dark" key="moon" :size="18" />
                <PhSun v-else key="sun" :size="18" />
            </Transition>
        </button>
        <button
            class="icon-btn font-mono text-[11px] font-semibold uppercase"
            :aria-label="t('lang.switch')"
            :title="t('lang.switch')"
            @click="toggleLang"
        >
            {{ locale }}
        </button>
        <a href="/logout" class="icon-btn" aria-label="Logout" title="Logout"
            ><PhSignOut :size="18"
        /></a>
    </div>
</template>

<style scoped>
.foot {
    display: flex;
    flex: none;
    align-items: center;
    gap: 8px;
    padding: 10px 8px 10px 12px;
    border-top: 1px solid var(--ui-border);
}

.avatar {
    position: relative;
    display: grid;
    place-items: center;
    width: 32px;
    height: 32px;
    flex: none;
    border-radius: 50%;
    background: var(--accent-soft);
    color: var(--accent);
    font-size: 13px;
    font-weight: 600;

    &::after {
        content: '';
        position: absolute;
        right: -1px;
        bottom: -1px;
        width: 9px;
        height: 9px;
        border: 2px solid var(--surface);
        border-radius: 50%;
        background: var(--faint);
        transition: background var(--dur-3);
    }

    &[data-online]::after {
        background: var(--ok);
    }
}

.swap-enter-active,
.swap-leave-active {
    transition:
        opacity var(--dur-2) var(--ease-out),
        transform var(--dur-3) var(--ease-out);
}

.swap-enter-from {
    opacity: 0;
    transform: rotate(-40deg) scale(0.6);
}

.swap-leave-to {
    opacity: 0;
    transform: rotate(40deg) scale(0.6);
}
</style>
