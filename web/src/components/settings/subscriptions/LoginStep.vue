<script setup lang="ts">
import { ref } from 'vue'
import { PhArrowSquareOut, PhCopy } from '@phosphor-icons/vue'
import { Button } from '@ui-sdk/vue/button'
import { Field, FieldDescription, FieldLabel } from '@ui-sdk/vue/field'
import { Input } from '@ui-sdk/vue/input'
import { toast } from '@ui-sdk/vue/toast'
import { useI18n } from 'vue-i18n'

import type { Subscription, SubscriptionLogin } from '@/gen'

import RunSpinner from '@/components/RunSpinner.vue'
import { useSubscriptionActions } from '@/composables/useSubscriptionActions'
import { reportError } from '@/shared/errors'

const props = defineProps<{ subscription: Subscription; login: SubscriptionLogin }>()
const { t } = useI18n()
const { complete } = useSubscriptionActions()
const code = ref('')

function submit() {
    const value = code.value.trim()
    if (value)
        complete.mutate({ path: { subscriptionId: props.subscription.id }, body: { code: value } })
}

function openSignIn() {
    window.open(props.login.url, '_blank', 'noopener,noreferrer')
}

function copyUserCode() {
    if (!props.login.userCode) return
    navigator.clipboard
        .writeText(props.login.userCode)
        .then(() => toast.success(t('copied')))
        .catch(reportError)
}
</script>

<template>
    <div class="flex flex-col gap-3">
        <ol class="m-0 flex flex-col gap-3 ps-5">
            <li>
                <span>{{ login.needsCode ? t('openClaude') : t('openCodex') }}</span>
                <div class="mt-2">
                    <Button data-variant="solid" @click="openSignIn">
                        <PhArrowSquareOut />{{ t('open') }}
                    </Button>
                </div>
            </li>
            <li v-if="login.userCode">
                <span>{{ t('enterCode') }}</span>
                <div class="mt-2 flex items-center gap-2">
                    <code class="text-lg font-semibold tracking-widest">{{ login.userCode }}</code>
                    <Button :aria-label="t('copy')" :title="t('copy')" @click="copyUserCode">
                        <PhCopy />
                    </Button>
                </div>
            </li>
            <li v-if="login.needsCode">
                <form class="flex flex-col gap-2" @submit.prevent="submit">
                    <Field>
                        <FieldLabel>{{ t('pasteCode') }}</FieldLabel>
                        <div class="flex gap-2">
                            <Input
                                v-model="code"
                                class="min-w-0 flex-1 font-mono"
                                autocomplete="off"
                                spellcheck="false"
                                required
                            />
                            <Button
                                type="submit"
                                data-variant="solid"
                                :disabled="!code.trim() || complete.isPending.value"
                            >
                                {{ t('confirm') }}
                            </Button>
                        </div>
                    </Field>
                </form>
            </li>
        </ol>
        <p
            v-if="!login.needsCode || complete.isPending.value"
            class="m-0 flex items-center gap-2 text-muted"
        >
            <RunSpinner :size="14" />{{ t('waiting') }}
        </p>
        <FieldDescription v-if="!login.needsCode">{{ t('deviceHint') }}</FieldDescription>
    </div>
</template>

<i18n lang="yaml">
ru:
    openClaude: Откройте страницу входа Anthropic и войдите в аккаунт с подпиской.
    openCodex: Откройте страницу OpenAI и войдите в аккаунт ChatGPT.
    open: Открыть страницу входа
    enterCode: Введите на странице этот код
    copy: Скопировать код
    copied: Код скопирован
    pasteCode: Вставьте код, который покажет страница
    confirm: Подтвердить
    waiting: Ждём подтверждения входа…
    deviceHint: Если страница не принимает код, включите вход по коду устройства для Codex в настройках безопасности ChatGPT.
en:
    openClaude: Open the Anthropic sign-in page and sign in to the account with the subscription.
    openCodex: Open the OpenAI page and sign in to your ChatGPT account.
    open: Open the sign-in page
    enterCode: Enter this code on the page
    copy: Copy the code
    copied: Code copied
    pasteCode: Paste the code the page shows
    confirm: Confirm
    waiting: Waiting for the sign-in to finish…
    deviceHint: If the page rejects the code, enable device code sign-in for Codex in ChatGPT security settings.
</i18n>
