<script setup lang="ts">
import '@xterm/xterm/css/xterm.css'

import { onBeforeUnmount, onMounted, useTemplateRef } from 'vue'
import { FitAddon } from '@xterm/addon-fit'
import { WebLinksAddon } from '@xterm/addon-web-links'
import { Terminal } from '@xterm/xterm'
import { Button } from '@ui-sdk/vue/button'
import { useI18n } from 'vue-i18n'

import { useTerminalTheme } from '@/components/terminal/useTerminalTheme'
import { useTerminalSocket } from '@/composables/useTerminalSocket'

const props = defineProps<{ sessionId: string }>()
const { t } = useI18n()
const host = useTemplateRef<HTMLElement>('host')

const terminal = new Terminal({
    cursorBlink: true,
    fontSize: 13,
    scrollback: 5000,
    minimumContrastRatio: 4.5,
})
const fit = new FitAddon()
terminal.loadAddon(fit)
terminal.loadAddon(new WebLinksAddon())

const socket = useTerminalSocket({
    sessionId: () => props.sessionId,
    size: () => ({ cols: terminal.cols, rows: terminal.rows }),
    onOutput: (data) => terminal.write(data),
    onReset: () => terminal.reset(),
})
const { status, exitCode } = socket
const theme = useTerminalTheme(terminal, () => host.value)

terminal.onData((data) => {
    if (status.value === 'exited') socket.connect()
    else socket.send(data)
})
terminal.onBinary((data) => socket.sendBinary(data))
terminal.onResize(({ cols, rows }) => socket.resize(cols, rows))

function fitToHost() {
    if (host.value?.clientWidth && host.value.clientHeight) fit.fit()
}

const resizeObserver = new ResizeObserver(fitToHost)

onMounted(() => {
    if (!host.value) return
    theme.apply()
    terminal.open(host.value)
    fitToHost()
    resizeObserver.observe(host.value)
    socket.connect()
    terminal.focus()
})

onBeforeUnmount(() => {
    resizeObserver.disconnect()
    terminal.dispose()
})
</script>

<template>
    <div class="terminal-pane">
        <div ref="host" class="terminal-host" />
        <div v-if="status !== 'open'" class="terminal-status text-muted" role="status">
            <template v-if="status === 'exited'">{{ t('exited', { code: exitCode }) }}</template>
            <template v-else-if="status === 'offline'">
                {{ t('offline') }}
                <Button @click="socket.connect()">{{ t('reconnect') }}</Button>
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
