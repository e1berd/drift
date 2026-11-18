import { onScopeDispose, ref, watch } from 'vue'

import type { TerminalControl, TerminalEvent } from '@/gen'

import { reportError } from '@/shared/errors'
import { API_BASE } from '@/shared/http/client'

export type TerminalStatus = 'idle' | 'connecting' | 'open' | 'exited' | 'offline'

export interface TerminalSocketOptions {
    sessionId: () => string
    size: () => { cols: number; rows: number }
    onOutput: (data: Uint8Array) => void
    onReset: () => void
}

const RETRY_DELAYS = [500, 1000, 2000, 5000, 10_000]
const encoder = new TextEncoder()

function terminalUrl(sessionId: string, cols: number, rows: number): URL {
    const url = new URL(
        `${API_BASE}/sessions/${encodeURIComponent(sessionId)}/terminal`,
        location.href,
    )
    url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:'
    url.searchParams.set('cols', String(cols))
    url.searchParams.set('rows', String(rows))
    return url
}

function parseEvent(text: string): TerminalEvent | undefined {
    try {
        return JSON.parse(text) as TerminalEvent
    } catch (error) {
        reportError(error)
        return undefined
    }
}

/**
 * The WebSocket to a session's shell (GET /sessions/{id}/terminal). The shell lives on the
 * server, so every (re)connect starts from a reset screen and receives the recent output again.
 */
export function useTerminalSocket(options: TerminalSocketOptions) {
    const status = ref<TerminalStatus>('idle')
    const exitCode = ref<number>()
    let socket: WebSocket | undefined
    let attempt = 0
    let retryTimer: ReturnType<typeof setTimeout> | undefined

    function receive(data: unknown) {
        if (data instanceof ArrayBuffer) {
            options.onOutput(new Uint8Array(data))
            return
        }
        if (typeof data !== 'string') return
        const event = parseEvent(data)
        if (event?.type !== 'exit') return
        exitCode.value = event.exitCode
        status.value = 'exited'
    }

    function scheduleRetry(closed: WebSocket) {
        if (closed !== socket || status.value === 'exited') return
        const delay = RETRY_DELAYS[attempt]
        if (delay === undefined) {
            status.value = 'offline'
            return
        }
        attempt += 1
        status.value = 'connecting'
        retryTimer = setTimeout(open, delay)
    }

    function open() {
        clearTimeout(retryTimer)
        const { cols, rows } = options.size()
        const next = new WebSocket(terminalUrl(options.sessionId(), cols, rows))
        next.binaryType = 'arraybuffer'
        next.addEventListener('open', () => {
            attempt = 0
            exitCode.value = undefined
            status.value = 'open'
            options.onReset()
        })
        next.addEventListener('message', (event) => receive(event.data))
        next.addEventListener('close', () => scheduleRetry(next))
        socket = next
    }

    function disconnect() {
        clearTimeout(retryTimer)
        const current = socket
        socket = undefined
        current?.close()
    }

    /** Opens the socket, starting a new shell when the previous one has exited. */
    function connect() {
        disconnect()
        attempt = 0
        status.value = 'connecting'
        open()
    }

    function sendBytes(bytes: Uint8Array<ArrayBuffer>) {
        if (socket?.readyState === WebSocket.OPEN) socket.send(bytes)
    }

    /** Sends keyboard input typed into the terminal. */
    function send(text: string) {
        sendBytes(encoder.encode(text))
    }

    /** Sends raw bytes that xterm reports as a binary string, e.g. some mouse events. */
    function sendBinary(binary: string) {
        sendBytes(Uint8Array.from(binary, (char) => char.charCodeAt(0) & 0xff))
    }

    function resize(cols: number, rows: number) {
        if (socket?.readyState !== WebSocket.OPEN) return
        socket.send(JSON.stringify({ type: 'resize', cols, rows } satisfies TerminalControl))
    }

    watch(options.sessionId, () => {
        if (status.value !== 'idle') connect()
    })
    onScopeDispose(disconnect)

    return { status, exitCode, connect, send, sendBinary, resize }
}
