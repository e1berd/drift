<script setup lang="ts">
import { useI18n } from 'vue-i18n'

import RunSpinner from '@/components/RunSpinner.vue'
import InfobarChanges from '@/components/infobar/InfobarChanges.vue'
import InfobarRefRow from '@/components/infobar/InfobarRefRow.vue'
import type { GitSessionBranch, GitSessionWorktree, GitSummary } from '@/gen'
import { branchLinkKey, worktreeLinkKey } from '@/components/infobar/linkKeys'

defineProps<{
    summary?: GitSummary
    loading: boolean
    failed: boolean
    running: boolean
    activeMessageId: string | null
}>()
const emit = defineEmits<{ reveal: [messageId: string] }>()
const { t } = useI18n()

function branchMeta(branch: GitSessionBranch): string {
    const parts = [t('commits', { count: branch.commits })]
    if (branch.current) parts.unshift(t('current'))
    parts.push(branch.upstream ? t('unpushed', { count: branch.unpushed }) : t('local'))
    return parts.join(' · ')
}

function worktreeMeta(tree: GitSessionWorktree): string {
    const changes = tree.changes ? t('changed', { count: tree.changes }) : t('clean')
    return [tree.branch ?? t('detached'), changes].join(' · ')
}
</script>

<template>
    <aside class="infobar" data-infobar :aria-label="t('label')">
        <div class="infobar-inner">
            <header class="infobar-header">
                <span class="font-medium">{{ t('title') }}</span>
                <span v-if="running" class="status-tag"><RunSpinner />{{ t('running') }}</span>
                <span v-else-if="summary?.base" class="status-tag" :title="t('baseHint')">
                    {{ t('base', { commit: summary.base.slice(0, 7) }) }}
                </span>
            </header>
            <div class="infobar-scroll" data-infobar-scroll>
                <div v-if="loading" class="flex flex-col gap-2">
                    <div v-for="i in 4" :key="i" class="skeleton h-3.5" />
                </div>
                <p v-else-if="failed" class="m-0 text-muted">{{ t('failed') }}</p>
                <p v-else-if="!summary?.repository" class="m-0 text-muted">
                    {{ t('notRepository') }}
                </p>
                <template v-else>
                    <InfobarChanges :changes="summary.changes" />
                    <section v-if="summary.branches.length" class="infobar-section motion-fade">
                        <h3 class="infobar-heading">
                            <span>{{ t('branches') }}</span>
                            <span class="tabular-nums">{{ summary.branches.length }}</span>
                        </h3>
                        <ul v-auto-animate class="infobar-list">
                            <InfobarRefRow
                                v-for="branch in summary.branches"
                                :key="branch.name"
                                :link-key="branchLinkKey(branch)"
                                :message-id="branch.messageId"
                                :name="branch.name"
                                :meta="branchMeta(branch)"
                                :active="activeMessageId === branch.messageId"
                                @reveal="emit('reveal', $event)"
                            />
                        </ul>
                    </section>
                    <section v-if="summary.worktrees.length" class="infobar-section motion-fade">
                        <h3 class="infobar-heading">
                            <span>{{ t('worktrees') }}</span>
                            <span class="tabular-nums">{{ summary.worktrees.length }}</span>
                        </h3>
                        <ul v-auto-animate class="infobar-list">
                            <InfobarRefRow
                                v-for="tree in summary.worktrees"
                                :key="tree.path"
                                :link-key="worktreeLinkKey(tree)"
                                :message-id="tree.messageId"
                                :name="tree.path"
                                :meta="worktreeMeta(tree)"
                                :active="activeMessageId === tree.messageId"
                                @reveal="emit('reveal', $event)"
                            />
                        </ul>
                    </section>
                </template>
            </div>
        </div>
    </aside>
</template>

<i18n lang="yaml">
ru:
    label: Статус сессии
    title: Статус сессии
    running: работает
    base: от {commit}
    baseHint: Коммит, с которого началась сессия; изменения считаются от него
    failed: Не удалось получить состояние git.
    notRepository: Папка сессии не является git-репозиторием.
    branches: Ветки сессии
    worktrees: Worktree сессии
    current: текущая
    commits: 'коммитов: {count}'
    unpushed: 'не запушено: {count}'
    local: не запушена
    changed: 'изменений: {count}'
    clean: без изменений
    detached: detached HEAD
en:
    label: Session status
    title: Session status
    running: running
    base: from {commit}
    baseHint: The commit the session started at; changes are counted from it
    failed: Could not read the git state.
    notRepository: The session folder is not a git repository.
    branches: Session branches
    worktrees: Session worktrees
    current: current
    commits: '{count} commits'
    unpushed: '{count} unpushed'
    local: not pushed
    changed: '{count} changes'
    clean: clean
    detached: detached HEAD
</i18n>

<style scoped>
.infobar {
    min-width: 0;
    overflow: hidden;
    border-inline-end: 1px solid var(--ui-border);
    transition: border-color var(--dur-4) var(--ease-out);

    &[inert] {
        border-color: transparent;
    }
}

.infobar-inner {
    display: flex;
    width: var(--infobar-width);
    height: 100%;
    flex-direction: column;
    animation: infobar-in var(--dur-3) var(--ease-out) both;
    transition:
        opacity var(--dur-3) var(--ease-out),
        translate var(--dur-3) var(--ease-out);

    .infobar[inert] & {
        opacity: 0;
        translate: -8px 0;
    }

    @media (prefers-reduced-motion: reduce) {
        animation-name: infobar-fade;
        translate: none;
    }
}

.infobar-header {
    display: flex;
    min-height: 44px;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    padding-inline: calc(var(--ui-spacing) * 4);
    border-bottom: 1px solid var(--ui-border);
}

.infobar-scroll {
    display: flex;
    min-height: 0;
    flex: 1;
    flex-direction: column;
    gap: calc(var(--ui-spacing) * 5);
    padding: calc(var(--ui-spacing) * 4);
    overflow-y: auto;
    scrollbar-width: thin;
}

@keyframes infobar-in {
    from {
        opacity: 0;
        translate: -8px 0;
    }
}

@keyframes infobar-fade {
    from {
        opacity: 0;
    }
}
</style>
