<script setup lang="ts">
import { computed, ref } from 'vue'
import { Button } from '@ui-sdk/vue/button'
import { Field, FieldDescription, FieldLabel } from '@ui-sdk/vue/field'
import { Input } from '@ui-sdk/vue/input'
import {
    SegmentedControl,
    SegmentedControlIndicator,
    SegmentedControlItem,
} from '@ui-sdk/vue/segmented-control'
import { useI18n } from 'vue-i18n'

import type { AgentId } from '@/shared/agents'

import AgentLogo from '@/components/AgentLogo.vue'
import { useSubscriptionActions } from '@/composables/useSubscriptionActions'
import { AGENTS } from '@/shared/agents'

const emit = defineEmits<{ done: [] }>()
const { t } = useI18n()
const { create } = useSubscriptionActions()

const SIGN_IN_AGENTS: AgentId[] = ['claude', 'codex', 'antigravity']

const agent = ref<AgentId>('claude')
const method = ref<'login' | 'token'>('login')
const token = ref('')

const agentValue = computed({
    get: () => agent.value,
    set: (value: string | null) => {
        if (value && SIGN_IN_AGENTS.includes(value as AgentId)) agent.value = value as AgentId
    },
})
const methodValue = computed({
    get: () => method.value,
    set: (value: string | null) => value && (method.value = value as 'login' | 'token'),
})
const usesToken = computed(() => agent.value === 'claude' && method.value === 'token')

function submit() {
    create.mutate(
        { body: { agent: agent.value, token: usesToken.value ? token.value.trim() : undefined } },
        { onSuccess: () => emit('done') },
    )
}
</script>

<template>
    <form
        class="flex flex-col gap-4 rounded-(--ui-radius) border border-line p-3"
        @submit.prevent="submit"
        @keydown.esc="emit('done')"
    >
        <div class="flex flex-col gap-2">
            <span id="add-agent">{{ t('agent') }}</span>
            <SegmentedControl v-model:value="agentValue" aria-labelledby="add-agent">
                <SegmentedControlIndicator />
                <SegmentedControlItem
                    v-for="item in Object.values(AGENTS)"
                    :key="item.id"
                    :value="item.id"
                    :disabled="!SIGN_IN_AGENTS.includes(item.id)"
                    :title="SIGN_IN_AGENTS.includes(item.id) ? undefined : t('soon')"
                >
                    <AgentLogo :agent="item.id" decorative />{{ item.name }}
                </SegmentedControlItem>
            </SegmentedControl>
        </div>

        <div v-if="agent === 'claude'" class="flex flex-col gap-2">
            <span id="add-method">{{ t('method') }}</span>
            <SegmentedControl v-model:value="methodValue" aria-labelledby="add-method">
                <SegmentedControlIndicator />
                <SegmentedControlItem value="login">{{ t('browser') }}</SegmentedControlItem>
                <SegmentedControlItem value="token">{{ t('token') }}</SegmentedControlItem>
            </SegmentedControl>
        </div>

        <Field v-if="usesToken">
            <FieldLabel>{{ t('tokenLabel') }}</FieldLabel>
            <Input
                v-model="token"
                type="password"
                class="font-mono"
                autocomplete="off"
                placeholder="sk-ant-oat01-…"
                required
            />
            <FieldDescription>{{ t('tokenHint') }}</FieldDescription>
        </Field>
        <p v-else class="m-0 text-muted">
            {{ t(`browser.${agent}`) }}
        </p>

        <div class="flex justify-end gap-2">
            <Button type="button" @click="emit('done')">{{ t('cancel') }}</Button>
            <Button
                type="submit"
                data-variant="solid"
                :disabled="create.isPending.value || (usesToken && !token.trim())"
            >
                {{ usesToken ? t('connect') : t('signIn') }}
            </Button>
        </div>
    </form>
</template>

<i18n lang="yaml">
ru:
    agent: Агент
    method: Способ входа
    browser: Через браузер
    token: Токен
    soon: Появится позже
    tokenLabel: Токен из claude setup-token
    tokenHint: Выполните claude setup-token на своём компьютере и вставьте токен. Лимиты с таким входом не видны.
    browser:
        claude: Drift запустит Claude Code на сервере и даст ссылку на вход. Вы войдёте в браузере и вставите код со страницы.
        codex: Drift запустит Codex на сервере и покажет ссылку и код. Введите код на странице OpenAI, вход завершится сам.
        antigravity: Drift запустит Antigravity CLI на сервере и даст ссылку на вход Google. Вы войдёте в аккаунт с подпиской Google AI и вставите код со страницы.
    cancel: Отмена
    connect: Подключить
    signIn: Начать вход
en:
    agent: Agent
    method: Sign-in method
    browser: Browser
    token: Token
    soon: Coming later
    tokenLabel: Token from claude setup-token
    tokenHint: Run claude setup-token on your computer and paste the token. Limits are not visible with this sign-in.
    browser:
        claude: Drift runs Claude Code on the server and gives you a sign-in link. Sign in in the browser and paste the code from the page.
        codex: Drift runs Codex on the server and shows a link and a code. Enter the code on the OpenAI page; the sign-in finishes by itself.
        antigravity: Drift runs the Antigravity CLI on the server and gives you a Google sign-in link. Sign in to the account with the Google AI subscription and paste the code from the page.
    cancel: Cancel
    connect: Connect
    signIn: Start sign-in
</i18n>
