<script setup lang="ts">
import { computed, ref } from 'vue'
import { PhCheck } from '@phosphor-icons/vue'
import { Button } from '@ui-sdk/vue/button'
import { Textarea } from '@ui-sdk/vue/textarea'
import { useI18n } from 'vue-i18n'

import { useSessionGit } from '@/composables/useSessionGit'
import { reportError } from '@drift/sdk/errors'

const { t } = useI18n()
const git = useSessionGit()
const message = ref('')
const stagedCount = computed(() => git.status.data.value?.staged.length ?? 0)
const canCommit = computed(
    () => stagedCount.value > 0 && !!message.value.trim() && !git.pending.commit.value,
)

function commit() {
    if (!canCommit.value) return
    git.commit(message.value)
        .then(() => {
            message.value = ''
        })
        .catch(reportError)
}
</script>

<template>
    <form class="flex flex-none flex-col gap-2 p-2" @submit.prevent="commit">
        <Textarea
            v-model="message"
            rows="2"
            class="commit-message"
            :aria-label="t('message')"
            :placeholder="t('placeholder')"
            @keydown.enter.ctrl.prevent="commit"
            @keydown.enter.meta.prevent="commit"
        />
        <Button type="submit" data-variant="solid" :disabled="!canCommit" v-tooltip="t('hint')">
            <PhCheck />{{ t('commit', { count: stagedCount }) }}
        </Button>
    </form>
</template>

<style scoped>
.commit-message {
    max-height: 160px;
    resize: vertical;
}
</style>

<i18n lang="yaml">
ru:
    message: Сообщение коммита
    placeholder: Сообщение коммита
    hint: Закоммитить подготовленные изменения (Ctrl+Enter)
    commit: 'Закоммитить ({count})'
en:
    message: Commit message
    placeholder: Commit message
    hint: Commit the staged changes (Ctrl+Enter)
    commit: 'Commit ({count})'
</i18n>
