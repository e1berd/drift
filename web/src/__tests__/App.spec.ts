import { describe, it, expect, vi, afterEach, beforeEach } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'
import { QueryClient, VueQueryPlugin } from '@tanstack/vue-query'

import App from '../App.vue'
import { setupI18n } from '../i18n'

function requestURL(input: RequestInfo | URL): URL {
    if (input instanceof Request) return new URL(input.url)
    return new URL(input, 'http://localhost')
}

function requestedPaths(): string[] {
    return vi.mocked(fetch).mock.calls.map(([input]) => requestURL(input).pathname)
}

const whoami = {
    id: 's',
    identity: { id: 'i', traits: { email: 'root@example.com', name: 'Root' } },
}

beforeEach(() => {
    vi.stubGlobal(
        'ResizeObserver',
        class {
            observe() {}
            unobserve() {}
            disconnect() {}
        },
    )
    vi.stubGlobal(
        'fetch',
        vi.fn(async (input: RequestInfo | URL) => {
            const { pathname } = requestURL(input)
            if (pathname === '/api/projects') return Response.json({ projects: [] })
            if (pathname === '/api/drafts') return Response.json({ drafts: [] })
            if (pathname === '/api/sessions') return Response.json({ sessions: [] })
            return Response.json(whoami)
        }),
    )
})

afterEach(() => vi.unstubAllGlobals())

describe('App', () => {
    it('renders the shell for the signed-in identity', async () => {
        const router = createRouter({
            history: createMemoryHistory(),
            routes: [
                { path: '/', name: 'new', component: { template: '<p>home</p>' } },
                { path: '/settings', name: 'settings', component: { template: '<p>settings</p>' } },
            ],
        })
        router.push('/')
        await router.isReady()

        const wrapper = mount(App, {
            global: {
                plugins: [
                    router,
                    setupI18n('en'),
                    [VueQueryPlugin, { queryClient: new QueryClient() }],
                ],
                directives: { autoAnimate: {} },
            },
        })
        await flushPromises()

        expect(wrapper.text()).toContain('Root')
        expect(wrapper.text()).toContain('home')
        expect(wrapper.text()).toContain('Online')
        expect(requestedPaths()).toContain('/.ory/kratos/public/sessions/whoami')
        expect(requestedPaths()).toContain('/api/projects')
        expect(requestedPaths()).toContain('/api/drafts')
        expect(requestedPaths()).toContain('/api/sessions')
    })
})
