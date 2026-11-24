<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

import { useFrontendChoice } from '@drift/vue/composables/useFrontendChoice'
import { useSession } from '@drift/vue/composables/useSession'

const { t } = useI18n()
const { isOwner } = useSession()
const choice = useFrontendChoice()

const items = computed(() =>
    choice.frontends.value.map((frontend) => ({
        value: frontend.name,
        title: frontend.displayName,
        subtitle: frontend.description,
    })),
)
const mine = computed({
    get: () => choice.current.value,
    set: (name: string) => choice.switchTo(name),
})
const everyone = computed({
    get: () => choice.instanceDefault.value,
    set: (name: string) => choice.setInstanceDefault(name),
})
</script>

<template>
    <v-card :title="t('frontends.title')" :subtitle="t('frontends.lead')" variant="flat">
        <v-card-text class="d-flex flex-column ga-4">
            <v-select v-model="mine" :items="items" :label="t('frontends.mine')">
                <template #item="{ props: item, item: option }">
                    <v-list-item v-bind="item" :subtitle="option.subtitle" />
                </template>
            </v-select>
            <v-select
                v-if="isOwner"
                v-model="everyone"
                :items="items"
                :label="t('frontends.everyone')"
                :loading="choice.savingDefault.value"
            />
        </v-card-text>
    </v-card>
</template>
