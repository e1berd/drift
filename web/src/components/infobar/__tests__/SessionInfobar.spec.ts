import { defineComponent, h } from 'vue'
import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'

import SessionInfobar from '@/components/infobar/SessionInfobar.vue'
import { provideActivityActions } from '@/composables/useActivityActions'
import type { ActivityActions } from '@/composables/useActivityActions'
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
    activity: [
        {
            key: 'branch:feat/status',
            kind: 'branch',
            messageId: 'm1',
            name: 'feat/status',
            removed: false,
            current: true,
            commits: 2,
        },
        {
            key: 'worktree:drift-tree',
            kind: 'worktree',
            messageId: 'm2',
            name: 'drift-tree',
            removed: false,
            branch: 'wt',
            changes: 3,
        },
        {
            key: 'file:docs/plan.md',
            kind: 'file',
            messageId: 'm3',
            name: 'docs/plan.md',
            removed: false,
        },
        {
            key: 'file:old.txt',
            kind: 'file',
            messageId: 'm4',
            name: 'old.txt',
            removed: true,
            removedBy: 'user',
            removeTime: new Date().toISOString(),
        },
    ],
}

function render(props: Record<string, unknown> = {}) {
    const actions: ActivityActions = {
        reveal: vi.fn(),
        open: vi.fn(),
        remove: vi.fn(),
        canOpen: (entry) => entry.kind === 'file' && !entry.removed,
    }
    const Host = defineComponent({
        setup() {
            provideActivityActions(actions)
            return () =>
                h(SessionInfobar, {
                    summary,
                    loading: false,
                    failed: false,
                    running: false,
                    activeMessageId: null,
                    ...props,
                })
        },
    })
    const wrapper = mount(Host, {
        global: { plugins: [setupI18n('en')], directives: { autoAnimate: {} } },
    })
    return { wrapper, actions }
}

describe('SessionInfobar', () => {
    it('lists changes and the activity log in order as link targets', () => {
        const { wrapper } = render({ activeMessageId: 'm2' })
        expect(wrapper.text()).toContain('+12')
        const rows = wrapper.findAll('[data-link-target]')
        expect(rows.map((row) => row.attributes('data-link-target'))).toEqual([
            'branch:feat/status',
            'worktree:drift-tree',
            'file:docs/plan.md',
            'file:old.txt',
        ])
        expect(rows.at(1)?.attributes('data-active')).toBeDefined()
        expect(rows.at(0)?.text()).toContain('current · 2 commits · not pushed')
        expect(rows.at(2)?.text()).toContain('plan.md')
    })

    it('keeps removed entries, marked and without actions', () => {
        const { wrapper } = render()
        const removed = wrapper.find('[data-link-target="file:old.txt"]')
        expect(removed.attributes('data-removed')).toBeDefined()
        expect(removed.text()).toContain('removed by you')
        expect(removed.findAll('button')).toHaveLength(1)
    })

    it('offers removal for files and worktrees but not for branches', async () => {
        const { wrapper, actions } = render()
        expect(
            wrapper.find('[data-link-target="branch:feat/status"]').findAll('button'),
        ).toHaveLength(1)
        await wrapper.find('[aria-label="Remove drift-tree"]').trigger('click')
        await wrapper.find('[aria-label="Open docs/plan.md"]').trigger('click')
        await wrapper.find('[data-link-target="branch:feat/status"] button').trigger('click')
        expect(actions.remove).toHaveBeenCalledWith(summary.activity.at(1))
        expect(actions.open).toHaveBeenCalledWith(summary.activity.at(2))
        expect(actions.reveal).toHaveBeenCalledWith('m1')
    })

    it('still lists created files outside a repository', () => {
        const { wrapper } = render({ summary: { ...summary, repository: false } })
        expect(wrapper.text()).toContain('not a git repository')
        expect(wrapper.findAll('[data-link-target]')).toHaveLength(4)
    })
})
