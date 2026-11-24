<script setup lang="ts">
import '@xterm/xterm/css/xterm.css'

import { useTemplateRef } from 'vue'
import { Button } from '@ui-sdk/vue/button'
import { useI18n } from 'vue-i18n'

import { useTerminal } from '@drift/vue/terminal/useTerminal'
import { terminalLook } from '@/components/terminal/terminalLook'

const props = defineProps<{ sessionId: string }>()
const { t } = useI18n()
const host = useTemplateRef<HTMLElement>('host')
const { status, exitCode, reconnect } = useTerminal(host, () => props.sessionId, terminalLook)
</script>

<template>
    <div class="terminal-pane">
        <div ref="host" class="terminal-host" />
        <div v-if="status !== 'open'" class="terminal-status text-muted" role="status">
            <template v-if="status === 'exited'">{{ t('exited', { code: exitCode }) }}</template>
            <template v-else-if="status === 'offline'">
                {{ t('offline') }}
                <Button @click="reconnect()">{{ t('reconnect') }}</Button>
            </template>
            <template v-else>{{ t('connecting') }}</template>
        </div>
    </div>
</template>

<style scoped>
.terminal-pane {
    display: flex;
    min-height: 0;
    flex: 1;
    flex-direction: column;
    background: var(--ui-background);
}

.terminal-host {
    min-height: 0;
    flex: 1;
    padding: 8px 12px;
}

.terminal-status {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 8px 12px;
    border-top: 1px solid var(--ui-border);
}
</style>

<i18n lang="yaml">
ru:
    connecting: Подключение к терминалу…
    exited: 'Шелл завершился с кодом {code}. Нажмите любую клавишу, чтобы запустить новый.'
    offline: Нет связи с терминалом.
    reconnect: Переподключить
en:
    connecting: Connecting to the terminal…
    exited: 'The shell exited with code {code}. Press any key to start a new one.'
    offline: The terminal is unreachable.
    reconnect: Reconnect
</i18n>
