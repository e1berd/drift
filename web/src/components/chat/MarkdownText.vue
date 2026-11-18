<script setup lang="ts">
import { computed } from 'vue'

import { renderMarkdown } from '@/shared/markdown'

const props = defineProps<{ source: string }>()

const html = computed(() => renderMarkdown(props.source))
</script>

<template>
    <div class="markdown" v-html="html" />
</template>
