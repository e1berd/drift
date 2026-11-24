<script setup lang="ts">
import { computed } from 'vue'
import {
    PhCaretUpDown,
    PhGearSix,
    PhMoon,
    PhSignOut,
    PhSun,
    PhTranslate,
} from '@phosphor-icons/vue'
import { Avatar, AvatarFallback } from '@ui-sdk/vue/avatar'
import { Button } from '@ui-sdk/vue/button'
import {
    Menu,
    MenuContent,
    MenuItem,
    MenuItemIcon,
    MenuSeparator,
    MenuTrigger,
} from '@ui-sdk/vue/menu'
import { usePreference } from '@ui-sdk/vue/preferences'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'

import { useEventStream } from '@drift/vue/composables/useEventStream'
import { useLocale } from '@/composables/useLocale'
import { useSession } from '@drift/vue/composables/useSession'
import { useShell } from '@drift/vue/composables/useShell'
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
const otherLocale = computed(() => (locale.value === 'en' ? 'ru' : 'en'))

function openSettings() {
    closeOnNarrow()
    void router.push('/settings')
}

function choose(value: string) {
    if (value === 'theme') colorScheme.set(dark.value ? 'light' : 'dark')
    else if (value === 'locale') setLocale(otherLocale.value)
    else if (value === 'signOut') location.assign(new URL('/logout', location.origin))
}
</script>

<template>
    <div class="flex flex-none items-center gap-1 border-t border-line p-2">
        <Menu @action="choose">
            <MenuTrigger class="account-trigger" v-tooltip="name">
                <Avatar class="flex-none"><AvatarFallback>{{ initial }}</AvatarFallback></Avatar>
                <span class="min-w-0 flex-1 text-left leading-tight">
                    <span class="block truncate font-medium">{{ name }}</span>
                    <span v-if="!connected" role="status" class="offline">
                        {{ t('nav.offline') }}
                    </span>
                </span>
                <PhCaretUpDown class="flex-none text-muted" />
            </MenuTrigger>
            <MenuContent side="top" align="start">
                <MenuItem value="theme">
                    <MenuItemIcon><PhSun v-if="dark" /><PhMoon v-else /></MenuItemIcon>
                    {{ themeLabel }}
                </MenuItem>
                <MenuItem value="locale">
                    <MenuItemIcon><PhTranslate /></MenuItemIcon>
                    {{ t(`language.${otherLocale}`) }}
                </MenuItem>
                <MenuSeparator />
                <MenuItem value="signOut">
                    <MenuItemIcon><PhSignOut /></MenuItemIcon>
                    {{ t('actions.sign_out') }}
                </MenuItem>
            </MenuContent>
        </Menu>
        <Button
            class="flex-none"
            :aria-label="t('nav.settings')"
            v-tooltip="t('nav.settings')"
            @click="openSettings"
        >
            <PhGearSix />
        </Button>
    </div>
</template>

<style scoped>
.account-trigger {
    display: flex;
    flex: 1;
    align-items: center;
    gap: 0.5rem;
    min-width: 0;
    padding: 0.25rem;
    border: 0;
    border-radius: var(--ui-radius);
    background: none;
    color: inherit;
    font: inherit;
    cursor: pointer;

    &:hover,
    &[data-state='open'],
    &[aria-expanded='true'] {
        background: var(--ui-muted);
    }
}

.offline {
    display: block;
    overflow: hidden;
    color: var(--ui-destructive);
    font-size: 0.75rem;
    text-overflow: ellipsis;
    white-space: nowrap;
}
</style>
