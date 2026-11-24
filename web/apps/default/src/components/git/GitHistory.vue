<script setup lang="ts">
import { computed, useTemplateRef } from 'vue'
import { useI18n } from 'vue-i18n'

import ScrollPane from '@/components/ScrollPane.vue'
import { useInfiniteSentinel } from '@drift/vue/composables/useInfiniteSentinel'
import { useSessionGit } from '@/composables/useSessionGit'
import { useListGitCommitsInfinite } from '@drift/vue/api'
import { errorMessage } from '@drift/sdk/errors'
import { relativeTime } from '@drift/sdk/time'

const PAGE_SIZE = 30

const { t, locale } = useI18n()
const git = useSessionGit()
const history = useListGitCommitsInfinite(
    {
        path: () => ({ sessionId: git.sessionId.value }),
        query: { page_size: PAGE_SIZE },
    },
    { query: { staleTime: 0 } },
)
const commits = computed(() => history.data.value?.pages.flatMap((page) => page.commits) ?? [])
const sentinel = useTemplateRef<HTMLElement>('sentinel')

useInfiniteSentinel(sentinel, () => {
    if (history.hasNextPage.value && !history.isFetchingNextPage.value) void history.fetchNextPage()
})
</script>

<template>
    <ScrollPane class="min-h-0 flex-1" content-class="flex flex-col py-2">
        <p v-if="history.isError.value" class="m-0 px-4 text-danger">
            {{ errorMessage(history.error.value) }}
        </p>
        <p
            v-else-if="history.isSuccess.value && !commits.length"
            class="m-0 px-4 py-6 text-center text-muted"
        >
            {{ t('empty') }}
        </p>
        <ol class="m-0 list-none p-0">
            <li
                v-for="commit in commits"
                :key="commit.hash"
                class="commit-row"
                v-tooltip="`${commit.hash}\n${commit.authorName} <${commit.authorEmail}>`"
            >
                <span class="truncate text-fg">{{ commit.subject }}</span>
                <span class="commit-meta">
                    <code>{{ commit.shortHash }}</code>
                    <span class="truncate">{{ commit.authorName }}</span>
                    <time class="ml-auto shrink-0" :datetime="commit.authorTime">{{
                        relativeTime(commit.authorTime, locale)
                    }}</time>
                </span>
            </li>
        </ol>
        <div ref="sentinel" class="h-px flex-none" />
        <p v-if="history.isFetching.value" class="m-0 px-4 py-2 text-muted">{{ t('loading') }}</p>
    </ScrollPane>
</template>

<style scoped>
.commit-row {
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding: 6px 12px;
    border-bottom: 1px solid var(--ui-border);
}

.commit-meta {
    display: flex;
    min-width: 0;
    align-items: baseline;
    gap: 8px;
    color: var(--ui-muted-foreground);
    font-size: 12px;

    & > code {
        font-family: var(--ui-font-mono);
    }
}
</style>

<i18n lang="yaml">
ru:
    empty: Коммитов пока нет.
    loading: Загрузка…
en:
    empty: No commits yet.
    loading: Loading…
</i18n>
