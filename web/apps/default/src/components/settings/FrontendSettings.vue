<script setup lang="ts">
import { computed } from 'vue'
import {
    Select,
    SelectContent,
    SelectItem,
    SelectItemIndicator,
    SelectItemText,
    SelectTrigger,
} from '@ui-sdk/vue/select'
import { useI18n } from 'vue-i18n'

import { useFrontendChoice } from '@drift/vue/composables/useFrontendChoice'
import { useSession } from '@drift/vue/composables/useSession'

const { t } = useI18n()
const { isOwner } = useSession()
const choice = useFrontendChoice()

const items = computed(() =>
    choice.frontends.value.map((frontend) => ({
        value: frontend.name,
        label: frontend.displayName,
    })),
)

function displayName(name: string): string {
    return choice.frontends.value.find((frontend) => frontend.name === name)?.displayName ?? name
}

const mine = computed({
    get: () => choice.current.value,
    set: (name: string | null) => name && choice.switchTo(name),
})
const everyone = computed({
    get: () => choice.instanceDefault.value,
    set: (name: string | null) => name && choice.setInstanceDefault(name),
})
</script>

<template>
    <section class="flex flex-col gap-4">
        <h2 class="m-0 text-base font-semibold">{{ t('title') }}</h2>
        <p class="m-0 text-muted">{{ t('lead') }}</p>
        <div class="flex flex-wrap items-center justify-between gap-3">
            <span id="frontend-mine">{{ t('mine') }}</span>
            <Select v-model:value="mine" :items="items">
                <SelectTrigger aria-labelledby="frontend-mine" class="min-w-48">
                    {{ displayName(mine) }}
                </SelectTrigger>
                <SelectContent>
                    <SelectItem
                        v-for="frontend in choice.frontends.value"
                        :key="frontend.name"
                        :value="frontend.name"
                        v-tooltip="frontend.description"
                    >
                        <SelectItemText>{{ frontend.displayName }}</SelectItemText>
                        <SelectItemIndicator />
                    </SelectItem>
                </SelectContent>
            </Select>
        </div>
        <div v-if="isOwner" class="flex flex-wrap items-center justify-between gap-3">
            <span id="frontend-everyone">{{ t('everyone') }}</span>
            <Select v-model:value="everyone" :items="items" :disabled="choice.savingDefault.value">
                <SelectTrigger aria-labelledby="frontend-everyone" class="min-w-48">
                    {{ displayName(everyone) }}
                </SelectTrigger>
                <SelectContent>
                    <SelectItem
                        v-for="frontend in choice.frontends.value"
                        :key="frontend.name"
                        :value="frontend.name"
                    >
                        <SelectItemText>{{ frontend.displayName }}</SelectItemText>
                        <SelectItemIndicator />
                    </SelectItem>
                </SelectContent>
            </Select>
        </div>
    </section>
</template>

<i18n lang="yaml">
ru:
    title: Вид Drift
    lead: Drift может выглядеть по-разному. Выбор хранится в этом браузере, страница перезагрузится.
    mine: Интерфейс в этом браузере
    everyone: По умолчанию для всех
en:
    title: Drift look
    lead: Drift comes in several interfaces. The choice stays in this browser; the page reloads.
    mine: Interface in this browser
    everyone: Default for everyone
</i18n>
