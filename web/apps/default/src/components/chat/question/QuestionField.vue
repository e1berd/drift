<script setup lang="ts">
import { computed, useId } from 'vue'
import { PhCheck } from '@phosphor-icons/vue'
import { Checkbox, CheckboxGroup, CheckboxIndicator } from '@ui-sdk/vue/checkbox'
import { Input } from '@ui-sdk/vue/input'
import { Radio, RadioGroup, RadioIndicator } from '@ui-sdk/vue/radio'
import { useI18n } from 'vue-i18n'

import type { Answer, Question } from '@drift/vue/api'

const props = defineProps<{ question: Question; disabled: boolean }>()
const answer = defineModel<Answer>({ required: true })
const { t } = useI18n()
const titleId = useId()

const single = computed({
    get: () => answer.value.selected?.at(0) ?? null,
    set: (value: string | null) => {
        answer.value = { ...answer.value, selected: value ? [value] : [] }
    },
})

const several = computed({
    get: () => answer.value.selected ?? [],
    set: (value: string[]) => {
        answer.value = { ...answer.value, selected: value }
    },
})

const custom = computed({
    get: () => answer.value.custom ?? '',
    set: (value: string) => {
        answer.value = { ...answer.value, custom: value }
    },
})

const options = computed(() => props.question.options ?? [])
</script>

<template>
    <fieldset class="question" :disabled="disabled">
        <legend :id="titleId" class="question-title">
            <span v-if="question.header" class="status-tag">{{ question.header }}</span>
            {{ question.question }}
            <span v-if="question.multiSelect" class="text-muted">{{ t('several') }}</span>
        </legend>
        <CheckboxGroup
            v-if="question.multiSelect && options.length"
            v-model="several"
            :aria-labelledby="titleId"
            :read-only="disabled"
            class="question-options"
        >
            <label v-for="option in options" :key="option.label" data-slot="label">
                <Checkbox :value="option.label">
                    <CheckboxIndicator><PhCheck aria-hidden="true" /></CheckboxIndicator>
                </Checkbox>
                <span class="question-option">
                    {{ option.label }}
                    <span v-if="option.description" class="text-muted">{{
                        option.description
                    }}</span>
                </span>
            </label>
        </CheckboxGroup>
        <RadioGroup
            v-else-if="options.length"
            v-model="single"
            :aria-labelledby="titleId"
            :read-only="disabled"
            class="question-options"
        >
            <label v-for="option in options" :key="option.label" data-slot="label">
                <Radio :value="option.label"><RadioIndicator /></Radio>
                <span class="question-option">
                    {{ option.label }}
                    <span v-if="option.description" class="text-muted">{{
                        option.description
                    }}</span>
                </span>
            </label>
        </RadioGroup>
        <Input
            v-if="!disabled || custom"
            v-model="custom"
            :placeholder="t('custom')"
            :aria-label="t('custom')"
            :readonly="disabled"
            maxlength="2000"
        />
    </fieldset>
</template>

<style scoped>
.question {
    display: flex;
    flex-direction: column;
    gap: calc(var(--ui-spacing) * 2);
    min-width: 0;
    margin: 0;
    padding: 0;
    border: 0;
}

.question-title {
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    gap: calc(var(--ui-spacing) * 2);
    padding: 0;
    margin-block-end: calc(var(--ui-spacing) * 2);
}

.question-options {
    display: flex;
    flex-direction: column;
    gap: calc(var(--ui-spacing) * 2);
}

.question-option {
    display: flex;
    flex-direction: column;
}
</style>

<i18n lang="yaml">
ru:
    several: можно выбрать несколько
    custom: Свой вариант
en:
    several: choose any
    custom: Your own answer
</i18n>
