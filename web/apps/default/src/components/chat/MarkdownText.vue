<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

import { copyCodeOnClick } from '@drift/sdk/markdownCopy'
import { renderMarkdown } from '@drift/sdk/markdown'

const props = defineProps<{ source: string }>()
const { t } = useI18n()

const html = computed(() => renderMarkdown(props.source, { copy: t('copy') }))
</script>

<template>
    <div class="markdown" @click="copyCodeOnClick" v-html="html" />
</template>

<i18n lang="yaml">
ru:
    copy: Копировать
en:
    copy: Copy
</i18n>
