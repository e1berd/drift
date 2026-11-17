<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Button } from '@ui-sdk/vue/button'

import {
    KratosRedirect,
    kratos,
    kratosURL,
    nodeValue,
    type KratosFlow,
    type KratosMessage,
} from '@/shared/kratos'
import { attestationJSON, creationOptions, passkeysSupported } from '@/shared/webauthn'
import { reportError } from '@/shared/errors'

const { t } = useI18n()
const route = useRoute()

const flow = ref<KratosFlow | null>(null)
const messages = ref<KratosMessage[]>([])

const passkeys = computed(() =>
    (flow.value?.ui.nodes ?? [])
        .filter((n) => n.attributes.name === 'passkey_remove')
        .map((n) => ({
            id: String(n.attributes.value),
            label: n.meta.label?.text ?? String(n.attributes.value),
        })),
)

const canAdd = computed(
    () =>
        passkeysSupported() && !!flow.value && nodeValue(flow.value, 'passkey_create_data') !== '',
)

function settingsFlowURL(id: string): URL {
    const url = kratosURL('/self-service/settings/flows')
    url.searchParams.set('id', id)
    return url
}

function collectMessages(f: KratosFlow) {
    messages.value = [...(f.ui.messages ?? []), ...f.ui.nodes.flatMap((n) => n.messages)]
}

async function load() {
    const id = typeof route.query.flow === 'string' ? route.query.flow : ''
    const { status, body } = id
        ? await kratos<KratosFlow>(settingsFlowURL(id))
        : await kratos<KratosFlow>('/self-service/settings/browser')
    if (status === 401) {
        window.location.assign('/login?return_to=/settings')
        return
    }
    if (status !== 200 && id) {
        await reload()
        return
    }
    if (status !== 200) throw new Error(`settings flow: ${status}`)
    flow.value = body
    collectMessages(body)
}

async function reload() {
    window.history.replaceState(null, '', '/settings')
    await load()
}

async function submit(payload: Record<string, string>) {
    if (!flow.value) return
    try {
        const { body } = await kratos<KratosFlow>(flow.value.ui.action, {
            method: 'POST',
            body: JSON.stringify({
                method: 'passkey',
                csrf_token: nodeValue(flow.value, 'csrf_token'),
                ...payload,
            }),
        })
        if (body?.ui) {
            flow.value = body
            collectMessages(body)
        }
    } catch (error) {
        if (error instanceof KratosRedirect) {
            window.location.assign(error.location)
            return
        }
        throw error
    }
}

async function addPasskey() {
    if (!flow.value) return
    const options = creationOptions(nodeValue(flow.value, 'passkey_create_data'))
    try {
        const credential = await navigator.credentials.create({ publicKey: options })
        if (credential instanceof PublicKeyCredential) {
            await submit({ passkey_settings_register: attestationJSON(credential) })
        }
    } catch (error) {
        if (error instanceof DOMException && error.name === 'NotAllowedError') return
        reportError(error)
    }
}

function removePasskey(id: string) {
    return submit({ passkey_remove: id })
}

onMounted(load)
</script>

<template>
    <section class="flex flex-col gap-4">
        <header class="flex flex-col gap-1">
            <h2 class="m-0 text-base font-semibold">{{ t('passkeys.title') }}</h2>
            <p class="text-sm text-(--ui-muted-foreground)">
                {{ t('lead') }}
            </p>
        </header>

        <ul v-if="messages.length" class="flex flex-col gap-2" role="status">
            <li
                v-for="m in messages"
                :key="m.id"
                class="rounded-(--ui-radius) border px-3 py-2 text-sm"
                :class="
                    m.type === 'error'
                        ? 'border-(--ui-destructive) text-(--ui-destructive)'
                        : 'border-(--ui-border)'
                "
            >
                {{ m.text }}
            </li>
        </ul>

        <ul
            v-if="passkeys.length"
            class="flex flex-col divide-y divide-(--ui-border) rounded-(--ui-radius) border border-(--ui-border)"
        >
            <li
                v-for="p in passkeys"
                :key="p.id"
                class="flex items-center justify-between gap-4 px-3 py-2 text-sm"
            >
                <span>{{ p.label }}</span>
                <Button class="btn danger h-7" @click="removePasskey(p.id)">{{
                    t('actions.delete')
                }}</Button>
            </li>
        </ul>
        <p v-else-if="flow" class="text-sm text-(--ui-muted-foreground)">{{ t('empty') }}</p>

        <div>
            <Button v-if="canAdd" class="btn primary" @click="addPasskey">{{ t('add') }}</Button>
            <p v-else-if="!passkeysSupported()" class="text-sm text-(--ui-muted-foreground)">
                {{ t('passkeys.unsupported') }}
            </p>
        </div>
    </section>
</template>

<i18n lang="yaml">
ru:
    lead: Вход по отпечатку пальца, Face ID или аппаратному ключу вместо пароля.
    empty: Ключей пока нет.
    add: Добавить ключ доступа
en:
    lead: Sign in with a fingerprint, Face ID or a hardware key instead of a password.
    empty: No passkeys yet.
    add: Add a passkey
</i18n>
