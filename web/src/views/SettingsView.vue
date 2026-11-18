<script setup lang="ts">
import TopBar from '@/components/TopBar.vue'
import InterfaceSettings from '@/components/settings/InterfaceSettings.vue'
import PasskeySettings from '@/components/settings/PasskeySettings.vue'
import SubscriptionSettings from '@/components/settings/subscriptions/SubscriptionSettings.vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
</script>

<template>
    <TopBar>
        <template #start
            ><span class="truncate font-medium">{{ t('settings.title') }}</span></template
        >
    </TopBar>
    <div class="min-h-0 flex-1 overflow-auto">
        <div class="motion-fade mx-auto flex max-w-[640px] flex-col gap-10 px-6 py-8 max-md:px-4">
            <InterfaceSettings />
            <SubscriptionSettings />
            <PasskeySettings />
        </div>
    </div>
</template>

<i18n lang="yaml">
ru:
    lead: Вход по отпечатку пальца, Face ID или аппаратному ключу вместо пароля.
    empty: Ключей пока нет.
    add: Добавить ключ доступа
en:
    lead: Sign in with a fingerprint, Face ID or a hardware key instead of a password.
    empty: No passkeys yet.
    add: Add a passkey
</i18n>
