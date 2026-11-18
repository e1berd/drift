<script setup lang="ts">
import { computed } from 'vue'
import { PhHourglassMedium } from '@phosphor-icons/vue'
import { Button } from '@ui-sdk/vue/button'
import { useI18n } from 'vue-i18n'

import { useLimitContinuation } from '@/composables/useLimitContinuation'
import type { Message } from '@/gen'
import { formatReset, formatResetClock, secondsUntil, usageWindowKind } from '@/shared/usage'

const props = defineProps<{ message: Message }>()
const { t, locale } = useI18n()
const { busy, continueNow, continueAfterReset, cancel } = useLimitContinuation(() => props.message)

const status = computed(() => props.message.body.status)
const open = computed(() => status.value === 'reached' || status.value === 'scheduled')
const window = computed(() =>
    props.message.body.window || props.message.body.windowMinutes
        ? t(
              `window.${usageWindowKind({ id: props.message.body.window ?? '', windowMinutes: props.message.body.windowMinutes })}`,
              {
                  minutes: props.message.body.windowMinutes ?? 0,
              },
          )
        : t('window.unknown'),
)
const reset = computed(() => {
    const iso = props.message.body.resetTime
    const left = secondsUntil(iso)
    if (!iso || left === null) return null
    return {
        clock: formatResetClock(iso, locale.value),
        relative: formatReset(left, locale.value),
        passed: left === 0,
    }
})
</script>

<template>
    <div class="chat-card" role="status" :data-closed="!open || undefined">
        <div class="chat-card-head">
            <PhHourglassMedium aria-hidden="true" />
            {{ t('title', { window }) }}
        </div>
        <p class="m-0">
            <template v-if="!reset">{{ t('resetUnknown') }}</template>
            <template v-else-if="reset.passed">{{ t('resetPassed') }}</template>
            <template v-else>{{
                t('resetAt', { clock: reset.clock, relative: reset.relative })
            }}</template>
        </p>
        <p v-if="message.body.text" class="m-0 text-muted">{{ message.body.text }}</p>
        <p v-if="status === 'scheduled' && reset" class="m-0">
            {{ t('scheduled', { clock: reset.clock }) }}
        </p>
        <div v-if="open" class="chat-card-actions">
            <Button
                v-if="status === 'reached' && reset && !reset.passed"
                data-variant="solid"
                :disabled="busy"
                @click="continueAfterReset"
            >
                {{ t('continueAfterReset') }}
            </Button>
            <Button :disabled="busy" @click="continueNow">{{ t('continueNow') }}</Button>
            <Button v-if="status === 'scheduled'" :disabled="busy" @click="cancel">
                {{ t('cancel') }}
            </Button>
        </div>
        <p v-else-if="status === 'continued'" class="m-0 text-muted">{{ t('continued') }}</p>
    </div>
</template>

<i18n lang="yaml">
ru:
    title: '{window} исчерпан'
    window:
        session: 5-часовой лимит
        week: Недельный лимит
        weekOpus: Недельный лимит Opus
        weekSonnet: Недельный лимит Sonnet
        other: 'Лимит на {minutes} мин'
        unknown: Лимит подписки
    resetAt: 'Сбросится в {clock} ({relative}).'
    resetPassed: Лимит уже должен был сброситься — можно продолжать.
    resetUnknown: Агент не сообщил, когда лимит сбросится.
    scheduled: 'Агент продолжит сам в {clock}. Сообщения, отправленные до этого, подождут и уйдут вместе с продолжением.'
    continueAfterReset: Продолжить, когда лимит сбросится
    continueNow: Продолжить сейчас
    cancel: Не продолжать
    continued: Работа продолжена.
en:
    title: '{window} reached'
    window:
        session: 5-hour limit
        week: Weekly limit
        weekOpus: Weekly Opus limit
        weekSonnet: Weekly Sonnet limit
        other: '{minutes}-minute limit'
        unknown: Subscription limit
    resetAt: 'Resets at {clock} ({relative}).'
    resetPassed: The limit should have reset by now — you can continue.
    resetUnknown: The agent did not say when the limit resets.
    scheduled: 'The agent continues on its own at {clock}. Messages sent before then wait and go along with it.'
    continueAfterReset: Continue when the limit resets
    continueNow: Continue now
    cancel: Don't continue
    continued: Work continued.
</i18n>
