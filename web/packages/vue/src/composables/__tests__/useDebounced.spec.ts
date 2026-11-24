import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { effectScope, nextTick, ref } from 'vue'

import { useDebounced } from '../useDebounced'

describe('useDebounced', () => {
    beforeEach(() => vi.useFakeTimers())
    afterEach(() => vi.useRealTimers())

    it('follows the source only after it stops changing', async () => {
        const source = ref('')
        const scope = effectScope()
        const settled = scope.run(() => useDebounced(source, 250))!

        source.value = 'pl'
        await nextTick()
        vi.advanceTimersByTime(200)
        source.value = 'plan'
        await nextTick()
        vi.advanceTimersByTime(200)
        expect(settled.value).toBe('')

        vi.advanceTimersByTime(50)
        expect(settled.value).toBe('plan')
        scope.stop()
    })

    it('clears at once when the source is emptied', async () => {
        const source = ref('plan')
        const scope = effectScope()
        const settled = scope.run(() => useDebounced(source, 250))!

        source.value = 'plans'
        await nextTick()
        source.value = ''
        await nextTick()
        expect(settled.value).toBe('')

        vi.advanceTimersByTime(300)
        expect(settled.value).toBe('')
        scope.stop()
    })
})
