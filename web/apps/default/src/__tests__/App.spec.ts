import { describe, it, expect, vi, afterEach, beforeEach } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'
import { QueryClient, VueQueryPlugin } from '@tanstack/vue-query'

import App from '../App.vue'
import { setupI18n } from '../i18n'
import { toastNotifier } from '../shared/notifier'
import { createNotifierPlugin } from '@drift/vue/notifier'

function requestURL(input: RequestInfo | URL): URL {
    if (input instanceof Request) return new URL(input.url)
    return new URL(input, 'http://localhost')
}

function requestedPaths(): string[] {
    return vi.mocked(fetch).mock.calls.map(([input]) => requestURL(input).pathname)
}

let delivery: Record<string, unknown>

const idleDelivery = {
    currentVersion: 'v1.0.0',
    releaseBuild: true,
    updateAvailable: false,
    repository: { url: 'https://github.com/e1berd/drift', stars: 1234 },
    installation: { state: 'idle', downloadedBytes: 0, totalBytes: 0 },
}

const whoami = {
    id: 's',
    identity: { id: 'i', traits: { email: 'root@example.com', name: 'Root' } },
}

const releasedDelivery = {
    ...idleDelivery,
    updateAvailable: true,
    latestRelease: {
        version: 'v1.1.0',
        name: 'v1.1.0',
        notes: '## Fixes\n- closed a hole',
        url: 'https://github.com/e1berd/drift/releases/tag/v1.1.0',
        publishTime: '2026-11-18T00:00:00Z',
        prerelease: false,
        installable: true,
    },
}

beforeEach(() => {
    delivery = idleDelivery
    localStorage.clear()
    Element.prototype.getAnimations ??= () => []
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
            if (pathname === '/api/delivery') return Response.json(delivery)
            return Response.json(whoami)
        }),
    )
})

afterEach(() => vi.unstubAllGlobals())

async function mountApp() {
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
        attachTo: document.body,
        global: {
            plugins: [
                router,
                setupI18n('en'),
                createNotifierPlugin(toastNotifier),
                [VueQueryPlugin, { queryClient: new QueryClient() }],
            ],
            directives: { autoAnimate: {} },
        },
    })
    await flushPromises()
    return wrapper
}

function buttonByText(text: string): HTMLElement {
    const button = [...document.querySelectorAll<HTMLElement>('button')].find(
        (element) => element.textContent?.trim() === text,
    )
    if (!button) throw new Error(`no button "${text}"`)
    return button
}

describe('App', () => {
    it('renders the shell for the signed-in identity', async () => {
        const wrapper = await mountApp()

        expect(wrapper.text()).toContain('Root')
        expect(wrapper.text()).toContain('home')
        expect(wrapper.text()).not.toContain('Offline')
        expect(requestedPaths()).toContain('/.ory/kratos/public/sessions/whoami')
        expect(requestedPaths()).toContain('/api/projects')
        expect(requestedPaths()).toContain('/api/drafts')
        expect(requestedPaths()).not.toContain('/api/sessions')
        expect(wrapper.text()).toContain('v1.0.0')
        wrapper.unmount()
    })

    it('shows the GitHub stars in the versions dialog', async () => {
        const wrapper = await mountApp()

        buttonByText('v1.0.0').click()
        await flushPromises()
        expect(document.body.textContent).toContain('1.2K')
        wrapper.unmount()
    })

    it('asks to confirm before skipping a new release', async () => {
        delivery = releasedDelivery
        const wrapper = await mountApp()

        expect(document.body.textContent).toContain('A new version of Drift is out')
        expect(document.body.textContent).toContain('v1.1.0')
        expect(document.body.textContent).toContain('closed a hole')

        buttonByText("Don't update").click()
        await flushPromises()
        expect(document.body.textContent).toContain('Skip the update?')

        buttonByText('Yes, skip it').click()
        await flushPromises()
        expect(JSON.parse(localStorage.getItem('drift.delivery.snooze') ?? '{}').version).toBe(
            'v1.1.0',
        )
        wrapper.unmount()
    })
})
