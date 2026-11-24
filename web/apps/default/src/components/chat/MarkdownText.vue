<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

import { reportError } from '@drift/sdk/errors'
import { renderMarkdown } from '@drift/sdk/markdown'

const COPIED_MS = 1500

const props = defineProps<{ source: string }>()
const { t } = useI18n()

const html = computed(() => renderMarkdown(props.source, { copy: t('copy') }))

function copyCode(event: MouseEvent) {
    const button = (event.target as Element | null)?.closest<HTMLElement>('[data-copy]')
    const code = button?.closest('.code-block')?.querySelector('pre')?.textContent
    if (!button || code === undefined || code === null) return
    navigator.clipboard
        .writeText(code)
        .then(() => {
            button.dataset.copied = ''
            setTimeout(() => delete button.dataset.copied, COPIED_MS)
        })
        .catch(reportError)
}
</script>

<template>
    <div class="markdown" @click="copyCode" v-html="html" />
</template>

<i18n lang="yaml">
ru:
    copy: Копировать
en:
    copy: Copy
</i18n>
