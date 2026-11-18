<script setup lang="ts">
import { computed, useTemplateRef } from 'vue'
import { useI18n } from 'vue-i18n'

import ScrollPane from '@/components/ScrollPane.vue'
import type { GitFileChange } from '@/gen'

import GitDiscardDialog from '@/components/git/GitDiscardDialog.vue'
import GitFileList from '@/components/git/GitFileList.vue'
import { useSessionGit } from '@/composables/useSessionGit'
import { reportError } from '@/shared/errors'

const { t } = useI18n()
const git = useSessionGit()
const status = computed(() => git.status.data.value)
const clean = computed(() => !status.value?.staged.length && !status.value?.unstaged.length)
const discardDialog = useTemplateRef<InstanceType<typeof GitDiscardDialog>>('discardDialog')

function discard(change: GitFileChange) {
    discardDialog.value
        ?.ask(change)
        .then((confirmed) => {
            if (confirmed) git.discard([change.path])
        })
        .catch(reportError)
}
</script>

<template>
    <ScrollPane class="min-h-0 flex-1" content-class="flex flex-col gap-2 pb-2">
        <p v-if="clean" class="m-0 px-4 py-6 text-center text-muted">{{ t('clean') }}</p>
        <GitFileList :title="t('staged')" :changes="status?.staged ?? []" staged />
        <GitFileList
            :title="t('changes')"
            :changes="status?.unstaged ?? []"
            :staged="false"
            @discard="discard"
        />
        <p v-if="status?.truncated" class="m-0 px-4 text-muted">{{ t('truncated') }}</p>
        <GitDiscardDialog ref="discardDialog" />
    </ScrollPane>
</template>

<i18n lang="yaml">
ru:
    staged: Подготовлено
    changes: Изменения
    clean: Нет изменений — рабочая копия совпадает с последним коммитом.
    truncated: Показаны первые 2000 файлов.
en:
    staged: Staged
    changes: Changes
    clean: No changes; the working tree matches the last commit.
    truncated: Showing the first 2000 files.
</i18n>
