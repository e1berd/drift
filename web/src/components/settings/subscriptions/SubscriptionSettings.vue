<script setup lang="ts">
import { ref } from 'vue'
import { PhPlus } from '@phosphor-icons/vue'
import { Button } from '@ui-sdk/vue/button'
import { useI18n } from 'vue-i18n'

import AddSubscription from '@/components/settings/subscriptions/AddSubscription.vue'
import SubscriptionItem from '@/components/settings/subscriptions/SubscriptionItem.vue'
import { useAgents } from '@/composables/useAgents'

const { t } = useI18n()
const { subscriptions, query } = useAgents()
const adding = ref(false)
</script>

<template>
    <section id="subscriptions" class="flex flex-col gap-4">
        <header class="flex items-start gap-3">
            <div class="flex flex-1 flex-col gap-1">
                <h2 class="m-0 text-base font-semibold">{{ t('settings.agents') }}</h2>
                <p class="m-0 text-muted">{{ t('settings.agentsLead') }}</p>
            </div>
            <Button v-if="!adding" @click="adding = true"><PhPlus />{{ t('add') }}</Button>
        </header>

        <AddSubscription v-if="adding" @done="adding = false" />

        <div v-if="query.isPending.value" class="flex flex-col gap-2">
            <div v-for="row in 2" :key="row" class="skeleton h-16" />
        </div>
        <p v-else-if="!subscriptions.length && !adding" class="m-0 text-muted">{{ t('empty') }}</p>
        <div v-auto-animate class="flex flex-col gap-3">
            <SubscriptionItem
                v-for="subscription in subscriptions"
                :key="subscription.id"
                :subscription="subscription"
            />
        </div>
    </section>
</template>

<i18n lang="yaml">
ru:
    add: Добавить подписку
    empty: Подписок пока нет. Добавьте подписку Claude или ChatGPT, чтобы начать чат с агентом.
en:
    add: Add a subscription
    empty: No subscriptions yet. Add a Claude or ChatGPT subscription to start chatting with an agent.
</i18n>
