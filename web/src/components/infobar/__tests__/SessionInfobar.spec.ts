import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import SessionInfobar from '@/components/infobar/SessionInfobar.vue'
import type { GitSummary } from '@/gen'
import { setupI18n } from '@/i18n'

const summary: GitSummary = {
    repository: true,
    base: '0123456789ab',
    changes: {
        additions: 12,
        deletions: 3,
        fileCount: 2,
        files: [{ path: 'app/main.go', additions: 12, deletions: 3 }],
    },
    branches: [{ name: 'feat/status', messageId: 'm1', current: true, commits: 2, unpushed: 0 }],
    worktrees: [{ path: 'drift-tree', branch: 'wt', messageId: 'm2', changes: 0 }],
}

function render(props: Partial<InstanceType<typeof SessionInfobar>['$props']> = {}) {
    return mount(SessionInfobar, {
        props: {
            summary,
            loading: false,
            failed: false,
            running: false,
            activeMessageId: null,
            ...props,
        },
        global: { plugins: [setupI18n('en')], directives: { autoAnimate: {} } },
    })
}

describe('SessionInfobar', () => {
    it('lists changes, session branches and worktrees as link targets', () => {
        const wrapper = render({ activeMessageId: 'm2' })
        expect(wrapper.text()).toContain('+12')
        expect(wrapper.text()).toContain('from 0123456')
        const rows = wrapper.findAll('[data-link-target]')
        expect(rows.map((row) => row.attributes('data-link-target'))).toEqual([
            'branch:feat/status',
            'worktree:drift-tree',
        ])
        expect(rows.at(1)?.attributes('data-active')).toBeDefined()
        expect(rows.at(0)?.text()).toContain('current · 2 commits · not pushed')
    })

    it('asks to reveal the tool call that created a row', async () => {
        const wrapper = render()
        await wrapper.find('[data-link-target="branch:feat/status"]').trigger('click')
        expect(wrapper.emitted('reveal')).toEqual([['m1']])
    })

    it('says when the folder is not a repository', () => {
        const wrapper = render({ summary: { ...summary, repository: false } })
        expect(wrapper.find('[data-link-target]').exists()).toBe(false)
        expect(wrapper.text()).toContain('not a git repository')
    })

    it('fills the drawer on narrow screens', () => {
        const wrapper = render({ placement: 'drawer' })
        expect(wrapper.attributes('data-placement')).toBe('drawer')
    })
})
