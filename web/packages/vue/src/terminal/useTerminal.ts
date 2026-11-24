import { onBeforeUnmount, onMounted } from 'vue'
import { FitAddon } from '@xterm/addon-fit'
import { WebLinksAddon } from '@xterm/addon-web-links'
import { Terminal } from '@xterm/xterm'

import type { Ref } from 'vue'

import { useTerminalSocket } from '../composables/useTerminalSocket'
import { useTerminalTheme } from './useTerminalTheme'
import type { TerminalLook } from './useTerminalTheme'

/** The session's shell in an xterm mounted into `host`, sized to it and wired to the server. */
export function useTerminal(
    host: Readonly<Ref<HTMLElement | null>>,
    sessionId: () => string,
    look: TerminalLook,
) {
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
        sessionId,
        size: () => ({ cols: terminal.cols, rows: terminal.rows }),
        onOutput: (data) => terminal.write(data),
        onReset: () => terminal.reset(),
    })
    const theme = useTerminalTheme(terminal, () => host.value, look)

    terminal.onData((data) => {
        if (socket.status.value === 'exited') socket.connect()
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

    return { status: socket.status, exitCode: socket.exitCode, reconnect: socket.connect }
}
