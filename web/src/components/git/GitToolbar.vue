<script setup lang="ts">
import { computed } from 'vue'
import {
    PhArrowClockwise,
    PhArrowDown,
    PhArrowUp,
    PhCloudArrowDown,
    PhCloudArrowUp,
} from '@phosphor-icons/vue'
import { Button } from '@ui-sdk/vue/button'
import { useI18n } from 'vue-i18n'

import GitBranchMenu from '@/components/git/GitBranchMenu.vue'
import { useSessionGit } from '@/composables/useSessionGit'
import { reportError } from '@/shared/errors'

const { t } = useI18n()
const git = useSessionGit()
const { status, pending } = git
const data = computed(() => status.data.value)
const pushLabel = computed(() =>
    data.value?.upstream ? t('push', { upstream: data.value.upstream }) : t('publish'),
)
</script>

<template>
    <div class="git-toolbar">
        <GitBranchMenu />
        <span
            v-if="data?.upstream && (data.ahead || data.behind)"
            class="status-tag shrink-0 tabular-nums"
            v-tooltip="t('aheadBehind', { ahead: data.ahead, behind: data.behind })"
        >
            <template v-if="data.behind"><PhArrowDown />{{ data.behind }}</template>
            <template v-if="data.ahead"><PhArrowUp />{{ data.ahead }}</template>
        </span>
        <div class="ml-auto flex shrink-0 gap-0.5">
            <Button
                :aria-label="t('pull')"
                v-tooltip="t('pull')"
                :disabled="pending.pull.value || !data?.upstream"
                @click="git.pull()"
            >
                <PhCloudArrowDown />
            </Button>
            <Button
                :aria-label="pushLabel"
                v-tooltip="pushLabel"
                :disabled="pending.push.value || data?.detached || !data?.head"
                @click="git.push()"
            >
                <PhCloudArrowUp />
            </Button>
            <Button
                :aria-label="t('refresh')"
                v-tooltip="t('refresh')"
                :disabled="status.isFetching.value"
                @click="git.refresh().catch(reportError)"
            >
                <PhArrowClockwise />
            </Button>
        </div>
    </div>
</template>

<style scoped>
.git-toolbar {
    display: flex;
    height: 40px;
    flex: none;
    align-items: center;
    gap: 8px;
    padding: 0 6px;
    border-bottom: 1px solid var(--ui-border);
    font-size: 12.5px;
}
</style>

<i18n lang="yaml">
ru:
    pull: Получить изменения (pull --ff-only)
    push: Отправить в {upstream}
    publish: Опубликовать ветку в origin
    refresh: Обновить
    aheadBehind: 'Впереди на {ahead}, позади на {behind}'
en:
    pull: Pull (fast-forward only)
    push: Push to {upstream}
    publish: Publish the branch to origin
    refresh: Refresh
    aheadBehind: '{ahead} ahead, {behind} behind'
</i18n>
