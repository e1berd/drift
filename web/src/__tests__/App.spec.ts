import { describe, it, expect, vi, afterEach } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'

import App from '../App.vue'

afterEach(() => vi.unstubAllGlobals())

describe('App', () => {
  it('shows the signed-in identity from Kratos', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () =>
        Response.json({ id: 's', identity: { id: 'i', traits: { email: 'root@example.com' } } }),
      ),
    )
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [{ path: '/', component: { template: '<p>home</p>' } }],
    })
    router.push('/')
    await router.isReady()

    const wrapper = mount(App, { global: { plugins: [router] } })
    await flushPromises()

    expect(wrapper.text()).toContain('root@example.com')
    expect(fetch).toHaveBeenCalledWith('/.ory/kratos/public/sessions/whoami', expect.anything())
  })
})
