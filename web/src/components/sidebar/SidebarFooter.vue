<script setup lang="ts">
import { computed } from 'vue'
import { PhGearSix, PhMoon, PhSignOut, PhSun } from '@phosphor-icons/vue'
import { Avatar, AvatarFallback } from '@ui-sdk/vue/avatar'
import { Button } from '@ui-sdk/vue/button'
import { usePreference } from '@ui-sdk/vue/preferences'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'

import DeliveryFooter from '@/components/delivery/DeliveryFooter.vue'
import { useEventStream } from '@/composables/useEventStream'
import { useLocale } from '@/composables/useLocale'
import { useSession } from '@/composables/useSession'
import { useShell } from '@/composables/useShell'
import { colorScheme } from '@/shared/preferences'

const { t } = useI18n()
const router = useRouter()
const { locale, setLocale } = useLocale()
const { session } = useSession()
const { connected } = useEventStream()
const { closeOnNarrow } = useShell()
const scheme = usePreference(colorScheme)

const name = computed(
    () => session.value?.identity.traits.name || session.value?.identity.traits.email || '',
)
const initial = computed(() => (name.value.trim().at(0) ?? '?').toUpperCase())
const dark = computed(() => scheme.value.resolved !== 'light')
const themeLabel = computed(() => (dark.value ? t('theme.toLight') : t('theme.toDark')))

function toggleTheme() {
    colorScheme.set(dark.value ? 'light' : 'dark')
}

function toggleLocale() {
    setLocale(locale.value === 'en' ? 'ru' : 'en')
}

function openSettings() {
    closeOnNarrow()
    void router.push('/settings')
}

function signOut() {
    location.assign(new URL('/logout', location.origin))
}
</script>

<template>
    <div class="flex-none border-t border-line">
        <div class="flex items-center gap-2 p-2.5">
            <Avatar><AvatarFallback>{{ initial }}</AvatarFallback></Avatar>
            <span class="min-w-0 flex-1 leading-tight">
                <span class="block truncate font-medium">{{ name }}</span>
                <span class="block truncate text-muted">
                    {{ connected ? t('nav.online') : t('nav.offline') }}
                </span>
            </span>
            <Button
                :aria-label="t('nav.settings')"
                :title="t('nav.settings')"
                @click="openSettings"
            >
                <PhGearSix />
            </Button>
            <Button :aria-label="themeLabel" :title="themeLabel" @click="toggleTheme">
                <PhMoon v-if="dark" />
                <PhSun v-else />
            </Button>
            <Button :title="t('lang.switch')" @click="toggleLocale">{{
                locale.toUpperCase()
            }}</Button>
            <Button
                :aria-label="t('actions.sign_out')"
                :title="t('actions.sign_out')"
                @click="signOut"
            >
                <PhSignOut />
            </Button>
        </div>
        <DeliveryFooter />
    </div>
</template>
