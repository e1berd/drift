<script setup lang="ts">
import { ref } from 'vue'
import { PhGitBranch } from '@phosphor-icons/vue'
import { Button } from '@ui-sdk/vue/button'
import {
    SegmentedControl,
    SegmentedControlIndicator,
    SegmentedControlItem,
} from '@ui-sdk/vue/segmented-control'
import { useI18n } from 'vue-i18n'

import GitChanges from '@/components/git/GitChanges.vue'
import GitCommitBox from '@/components/git/GitCommitBox.vue'
import GitDiff from '@/components/git/GitDiff.vue'
import GitHistory from '@/components/git/GitHistory.vue'
import GitToolbar from '@/components/git/GitToolbar.vue'
import { provideSessionGit } from '@/composables/useSessionGit'
import { errorMessage } from '@/shared/errors'

const props = defineProps<{ sessionId: string }>()
const { t } = useI18n()
const git = provideSessionGit(() => props.sessionId)
const { status, repository, pending } = git
const view = ref<'changes' | 'history'>('changes')
</script>

<template>
    <div v-if="status.isPending.value" class="flex flex-col gap-1.5 p-4">
        <div v-for="i in 6" :key="i" class="skeleton h-3.5" :style="{ width: `${30 + i * 9}%` }" />
    </div>
    <div v-else-if="status.isError.value" class="git-empty text-muted">
        <p class="m-0">{{ errorMessage(status.error.value) }}</p>
        <Button @click="git.refresh()">{{ t('retry') }}</Button>
    </div>
    <div v-else-if="!repository" class="git-empty text-muted">
        <span class="grid h-11 w-11 place-items-center rounded-xl bg-surface">
            <PhGitBranch :size="22" />
        </span>
        <p class="m-0 max-w-[320px]">{{ t('notRepository') }}</p>
        <Button data-variant="solid" :disabled="pending.init.value" @click="git.init()">
            git init
        </Button>
    </div>
    <div v-else class="git-split">
        <aside class="git-side">
            <GitToolbar />
            <div class="px-2 pt-2">
                <SegmentedControl v-model:value="view" :aria-label="t('view')" class="w-full">
                    <SegmentedControlIndicator />
                    <SegmentedControlItem value="changes" class="flex-1">{{
                        t('changes')
                    }}</SegmentedControlItem>
                    <SegmentedControlItem value="history" class="flex-1">{{
                        t('history')
                    }}</SegmentedControlItem>
                </SegmentedControl>
            </div>
            <template v-if="view === 'changes'">
                <GitCommitBox />
                <GitChanges />
            </template>
            <GitHistory v-else />
        </aside>
        <GitDiff />
    </div>
</template>

<style scoped>
.git-split {
    display: grid;
    min-height: 0;
    flex: 1;
    grid-template-columns: 320px minmax(0, 1fr);

    @media (max-width: 760px) {
        grid-template-columns: 1fr;
        grid-template-rows: auto minmax(0, 1fr);
    }
}

.git-side {
    display: flex;
    min-height: 0;
    flex-direction: column;
    border-right: 1px solid var(--ui-border);

    @media (max-width: 760px) {
        max-height: 55dvh;
        border-right: 0;
        border-bottom: 1px solid var(--ui-border);
    }
}

.git-empty {
    display: flex;
    flex: 1;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: calc(var(--ui-spacing) * 3);
    padding: calc(var(--ui-spacing) * 6);
    text-align: center;
}
</style>

<i18n lang="yaml">
ru:
    notRepository: Папка сессии — не Git-репозиторий. Создайте его, чтобы видеть изменения и делать коммиты.
    retry: Повторить
    view: Раздел Git
    changes: Изменения
    history: История
en:
    notRepository: The session folder is not a Git repository. Create one to see changes and commit them.
    retry: Retry
    view: Git section
    changes: Changes
    history: History
</i18n>
