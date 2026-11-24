<script setup lang="ts">
import { computed, ref } from 'vue'
import { PhArrowSquareOut, PhCopy } from '@phosphor-icons/vue'
import { Button } from '@ui-sdk/vue/button'
import { Field, FieldDescription, FieldLabel } from '@ui-sdk/vue/field'
import { Input } from '@ui-sdk/vue/input'
import { toast } from '@ui-sdk/vue/toast'
import { useI18n } from 'vue-i18n'

import type { Subscription, SubscriptionLogin } from '@drift/vue/api'

import RunSpinner from '@/components/RunSpinner.vue'
import { useSubscriptionActions } from '@drift/vue/composables/useSubscriptionActions'
import { reportError } from '@drift/sdk/errors'

const props = defineProps<{ subscription: Subscription; login: SubscriptionLogin }>()
const { t, te } = useI18n()
const hasDeviceHint = computed(() => te(`deviceHint.${props.subscription.agent}`))
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
                <span>{{ t(`openPage.${subscription.agent}`) }}</span>
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
                    <Button :aria-label="t('copy')" v-tooltip="t('copy')" @click="copyUserCode">
                        <PhCopy />
                    </Button>
                </div>
            </li>
            <li v-if="login.needsCode">
                <form class="flex flex-col gap-2" @submit.prevent="submit">
                    <Field>
                        <FieldLabel>{{ t('pasteCode') }}</FieldLabel>
                        <FieldDescription v-if="subscription.agent === 'antigravity'">
                            {{ t('pasteAddress') }}
                        </FieldDescription>
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
        <FieldDescription v-if="!login.needsCode && hasDeviceHint">{{
            t(`deviceHint.${subscription.agent}`)
        }}</FieldDescription>
    </div>
</template>

<i18n lang="yaml">
ru:
    openPage:
        claude: Откройте страницу входа Anthropic и войдите в аккаунт с подпиской.
        codex: Откройте страницу OpenAI и войдите в аккаунт ChatGPT.
        antigravity: Откройте страницу Google и войдите в аккаунт с подпиской Google AI.
        kimi: Откройте страницу Kimi и войдите в аккаунт с подпиской.
        mimo: Откройте страницу Xiaomi и войдите в аккаунт с подпиской MiMo.
    open: Открыть страницу входа
    enterCode: Введите на странице этот код
    copy: Скопировать код
    copied: Код скопирован
    pasteCode: Вставьте код, который покажет страница
    pasteAddress: Если после входа браузер открыл пустую страницу localhost, скопируйте её адрес целиком и вставьте сюда.
    confirm: Подтвердить
    waiting: Ждём подтверждения входа…
    deviceHint:
        codex: Если страница не принимает код, включите вход по коду устройства для Codex в настройках безопасности ChatGPT.
en:
    openPage:
        claude: Open the Anthropic sign-in page and sign in to the account with the subscription.
        codex: Open the OpenAI page and sign in to your ChatGPT account.
        antigravity: Open the Google page and sign in to the account with the Google AI subscription.
        kimi: Open the Kimi page and sign in to the account with the subscription.
        mimo: Open the Xiaomi page and sign in to the account with the MiMo subscription.
    open: Open the sign-in page
    enterCode: Enter this code on the page
    copy: Copy the code
    copied: Code copied
    pasteCode: Paste the code the page shows
    pasteAddress: If the browser lands on an empty localhost page after sign-in, copy its whole address and paste it here.
    confirm: Confirm
    waiting: Waiting for the sign-in to finish…
    deviceHint:
        codex: If the page rejects the code, enable device code sign-in for Codex in ChatGPT security settings.
</i18n>
