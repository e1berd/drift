<script setup lang="ts">
import { computed, ref } from 'vue'
import { PhArrowClockwise, PhCheckCircle, PhTrash } from '@phosphor-icons/vue'
import { Button } from '@ui-sdk/vue/button'
import { useI18n } from 'vue-i18n'

import type { Subscription } from '@/gen'

import AgentLogo from '@/components/AgentLogo.vue'
import LoginStep from '@/components/settings/subscriptions/LoginStep.vue'
import UsageMeters from '@/components/settings/subscriptions/UsageMeters.vue'
import { useSubscriptionActions } from '@/composables/useSubscriptionActions'
import { AGENTS } from '@/shared/agents'

const props = defineProps<{ subscription: Subscription }>()
const { t } = useI18n()
const { activate, restart, remove } = useSubscriptionActions()
const confirming = ref(false)

const path = computed(() => ({ subscriptionId: props.subscription.id }))
const agentName = computed(() => AGENTS[props.subscription.agent].name)
const details = computed(() => {
    const { account, plan, method, state } = props.subscription
    if (!account && state !== 'active') return [t(`state.${state}`)]
    return [account ? agentName.value : '', plan, method === 'token' ? t('token') : ''].filter(
        Boolean,
    )
})
</script>

<template>
    <article class="subscription" :data-current="subscription.current || undefined">
        <header class="flex items-center gap-3">
            <AgentLogo :agent="subscription.agent" :size="20" decorative />
            <div class="flex min-w-0 flex-1 flex-col">
                <span class="truncate font-medium">{{ subscription.account || agentName }}</span>
                <span v-if="details.length" class="truncate text-muted">{{
                    details.join(' · ')
                }}</span>
            </div>
            <span v-if="subscription.current" class="status-tag text-fg">
                <PhCheckCircle weight="fill" />{{ t('current') }}
            </span>
            <Button
                v-else-if="subscription.state === 'active'"
                :disabled="activate.isPending.value"
                @click="activate.mutate({ path })"
            >
                {{ t('makeCurrent') }}
            </Button>
            <Button
                v-if="!confirming"
                :aria-label="t('remove')"
                :title="t('remove')"
                @click="confirming = true"
            >
                <PhTrash />
            </Button>
        </header>

        <div v-if="confirming" class="motion-fade flex flex-wrap items-center justify-end gap-2">
            <span class="me-auto">{{
                subscription.state === 'pending' ? t('confirmCancel') : t('confirmRemove')
            }}</span>
            <Button @click="confirming = false">{{ t('keep') }}</Button>
            <Button
                data-variant="destructive"
                :disabled="remove.isPending.value"
                @click="remove.mutate({ path })"
            >
                {{ t('remove') }}
            </Button>
        </div>

        <LoginStep
            v-if="subscription.state === 'pending' && subscription.login"
            :subscription="subscription"
            :login="subscription.login"
        />
        <div v-else-if="subscription.state === 'failed'" class="flex flex-col gap-2">
            <p class="m-0 text-danger">{{ subscription.error || t('failed') }}</p>
            <div v-if="subscription.method === 'login'">
                <Button :disabled="restart.isPending.value" @click="restart.mutate({ path })">
                    <PhArrowClockwise />{{ t('retry') }}
                </Button>
            </div>
        </div>
        <UsageMeters
            v-else-if="subscription.state === 'active'"
            :subscription-id="subscription.id"
        />
    </article>
</template>

<style scoped>
.subscription {
    display: flex;
    flex-direction: column;
    gap: calc(var(--ui-spacing) * 3);
    padding: calc(var(--ui-spacing) * 3);
    border: 1px solid var(--ui-border);
    border-radius: var(--ui-radius);

    &[data-current] {
        border-color: var(--ui-primary);
        background: color-mix(in oklch, var(--ui-primary) 4%, transparent);
    }
}
</style>

<i18n lang="yaml">
ru:
    current: Текущая
    makeCurrent: Сделать текущей
    token: вход по токену
    remove: Удалить
    keep: Оставить
    confirmRemove: Выйти из аккаунта и удалить подписку?
    confirmCancel: Отменить вход?
    failed: Вход не удался.
    retry: Войти заново
    state:
        pending: Ожидает входа
        active: Подключена
        failed: Не подключена
en:
    current: Current
    makeCurrent: Make current
    token: token sign-in
    remove: Remove
    keep: Keep
    confirmRemove: Sign out and remove the subscription?
    confirmCancel: Cancel the sign-in?
    failed: Sign-in failed.
    retry: Sign in again
    state:
        pending: Waiting for sign-in
        active: Connected
        failed: Not connected
</i18n>
