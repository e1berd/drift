import { effectScope } from 'vue'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { useTerminalSocket } from '../useTerminalSocket'

class FakeSocket extends EventTarget {
    static OPEN = 1
    static instances: FakeSocket[] = []
    readyState = 0
    binaryType = 'blob'
    sent: unknown[] = []

    constructor(public url: URL) {
        super()
        FakeSocket.instances.push(this)
    }

    send(data: unknown) {
        this.sent.push(data)
    }

    close() {
        this.readyState = 3
        this.dispatchEvent(new Event('close'))
    }

    accept() {
        this.readyState = FakeSocket.OPEN
        this.dispatchEvent(new Event('open'))
    }

    receive(data: unknown) {
        this.dispatchEvent(new MessageEvent('message', { data }))
    }
}

function setup() {
    const output: Uint8Array[] = []
    const onReset = vi.fn()
    const scope = effectScope()
    const socket = scope.run(() =>
        useTerminalSocket({
            sessionId: () => 's1',
            size: () => ({ cols: 100, rows: 30 }),
            onOutput: (data) => output.push(data),
            onReset,
        }),
    )!
    return { socket, output, onReset, scope }
}

describe('useTerminalSocket', () => {
    beforeEach(() => {
        FakeSocket.instances = []
        vi.stubGlobal('WebSocket', FakeSocket)
    })

    afterEach(() => {
        vi.unstubAllGlobals()
        vi.useRealTimers()
    })

    it('connects to the session terminal with the current size', () => {
        const { socket } = setup()
        socket.connect()
        const url = FakeSocket.instances.at(0)?.url
        expect(url?.protocol).toBe('ws:')
        expect(url?.pathname).toBe('/api/sessions/s1/terminal')
        expect(url?.searchParams.get('cols')).toBe('100')
        expect(url?.searchParams.get('rows')).toBe('30')
    })

    it('resets the screen on open, then streams output and input', () => {
        const { socket, output, onReset } = setup()
        socket.connect()
        const server = FakeSocket.instances.at(0)!
        server.accept()
        expect(onReset).toHaveBeenCalledOnce()
        expect(socket.status.value).toBe('open')

        server.receive(new Uint8Array([104, 105]).buffer)
        expect(output.map((chunk) => Array.from(chunk))).toEqual([[104, 105]])

        socket.send('ls\r')
        socket.resize(120, 40)
        expect(Array.from(server.sent.at(0) as Uint8Array)).toEqual([108, 115, 13])
        expect(JSON.parse(server.sent.at(1) as string)).toEqual({
            type: 'resize',
            cols: 120,
            rows: 40,
        })
    })

    it('reports the exit code and does not reconnect after the shell exits', () => {
        vi.useFakeTimers()
        const { socket } = setup()
        socket.connect()
        const server = FakeSocket.instances.at(0)!
        server.accept()
        server.receive(JSON.stringify({ type: 'exit', exitCode: 3 }))
        server.close()
        vi.runAllTimers()
        expect(socket.status.value).toBe('exited')
        expect(socket.exitCode.value).toBe(3)
        expect(FakeSocket.instances).toHaveLength(1)
    })

    it('reconnects after a dropped connection and gives up after the last retry', () => {
        vi.useFakeTimers()
        const { socket } = setup()
        socket.connect()
        for (let attempt = 0; attempt < 6; attempt += 1) {
            FakeSocket.instances.at(-1)!.close()
            vi.runAllTimers()
        }
        expect(FakeSocket.instances).toHaveLength(6)
        expect(socket.status.value).toBe('offline')
    })

    it('closes the socket when its scope is disposed', () => {
        const { socket, scope } = setup()
        socket.connect()
        scope.stop()
        expect(FakeSocket.instances.at(0)?.readyState).toBe(3)
    })
})
