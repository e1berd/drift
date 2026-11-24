<script setup lang="ts">
import { useI18n } from 'vue-i18n'

import type { Message } from '@drift/vue/api'
import { useQuestionAnswers } from '@drift/vue/composables/useQuestionAnswers'
import type { AgentId } from '@drift/sdk/agents'

const props = defineProps<{ message: Message; agent: AgentId }>()
const { t } = useI18n()
const { answers, pending, complete, sending, submit } = useQuestionAnswers(
    () => props.message,
    () => props.agent,
)
</script>

<template>
    <v-card variant="tonal" color="primary" class="pa-1">
        <v-card-text
            v-for="(question, index) in message.body.questions"
            :key="index"
            class="d-flex flex-column ga-2"
        >
            <div class="font-weight-medium">{{ question.question }}</div>
            <v-chip-group
                v-if="answers[index]"
                v-model="answers[index].selected"
                :multiple="question.multiSelect"
                :disabled="!pending"
                column
            >
                <v-chip
                    v-for="option in question.options"
                    :key="option.label"
                    :value="option.label"
                    :title="option.description"
                    filter
                    >{{ option.label }}</v-chip
                >
            </v-chip-group>
            <v-text-field
                v-if="answers[index] && pending"
                v-model="answers[index].custom"
                :label="t('question.custom')"
                hide-details
            />
        </v-card-text>
        <v-card-actions v-if="pending">
            <v-spacer />
            <v-btn
                variant="flat"
                color="primary"
                :disabled="!complete"
                :loading="sending"
                @click="submit"
                >{{ t('question.answer') }}</v-btn
            >
        </v-card-actions>
    </v-card>
</template>
