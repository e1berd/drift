import type { GitSessionBranch, GitSessionWorktree, GitSummary } from '@/gen'
import type { InfobarLink } from '@/composables/useInfobarLinks'

export function branchLinkKey(branch: GitSessionBranch): string {
    return `branch:${branch.name}`
}

export function worktreeLinkKey(tree: GitSessionWorktree): string {
    return `worktree:${tree.path}`
}

export function summaryLinks(summary: GitSummary | undefined): InfobarLink[] {
    if (!summary) return []
    return [
        ...summary.branches.map((branch) => ({
            key: branchLinkKey(branch),
            messageId: branch.messageId,
        })),
        ...summary.worktrees.map((tree) => ({
            key: worktreeLinkKey(tree),
            messageId: tree.messageId,
        })),
    ]
}
