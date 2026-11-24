import { flushPromises, mount } from '@vue/test-utils'
import { QueryClient, VueQueryPlugin } from '@tanstack/vue-query'
import { afterEach, describe, expect, it, vi } from 'vitest'

import UsageMeters from '@/components/settings/subscriptions/UsageMeters.vue'
import { setupI18n } from '@/i18n'

const subscriptionId = '2f1c2a4e-9d0b-4c8e-8a7e-1b2c3d4e5f60'

function requestURL(input: RequestInfo | URL): URL {
    if (input instanceof Request) return new URL(input.url)
    return new URL(input, 'http://localhost')
}

afterEach(() => vi.unstubAllGlobals())

describe('UsageMeters', () => {
    it('refreshes limits after an initial read error', async () => {
        const requests: URL[] = []
        vi.stubGlobal(
            'fetch',
            vi.fn(async (input: RequestInfo | URL) => {
                const url = requestURL(input)
                requests.push(url)
                if (url.searchParams.get('refresh') === 'true') {
                    return Response.json({
                        windows: [
                            {
                                id: 'five_hour',
                                usedPercent: 42,
                                windowMinutes: 300,
                                resetTime: '2026-11-19T12:00:00Z',
                            },
                        ],
                        readTime: '2026-11-19T09:00:00Z',
                    })
                }
                return Response.json({
                    windows: [],
                    unavailable: 'error',
                    readTime: '2026-11-19T09:00:00Z',
                })
            }),
        )

        const wrapper = mount(UsageMeters, {
            props: { subscriptionId },
            global: {
                plugins: [setupI18n('en'), [VueQueryPlugin, { queryClient: new QueryClient() }]],
            },
        })
        await flushPromises()

        expect(wrapper.text()).toContain('Could not read the limits')
        expect(requests.at(0)?.searchParams.has('refresh')).toBe(false)

        await wrapper.find('button[aria-label="Refresh limits"]').trigger('click')
        await flushPromises()

        expect(requests.at(-1)?.searchParams.get('refresh')).toBe('true')
        expect(wrapper.text()).toContain('Session (5 hours)')
        expect(wrapper.text()).toContain('42%')
        expect(wrapper.text()).not.toContain('Could not read the limits')
        wrapper.unmount()
    })
})
