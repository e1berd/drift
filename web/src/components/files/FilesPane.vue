<script setup lang="ts">
import { watch } from 'vue'
import { useRoute } from 'vue-router'

import FileEditor from '@/components/files/FileEditor.vue'
import FilesSidebar from '@/components/files/FilesSidebar.vue'
import { provideSessionFiles } from '@/composables/useSessionFiles'
import { reportError } from '@/shared/errors'

const props = defineProps<{ sessionId: string; title: string; hint?: string }>()
const files = provideSessionFiles(() => props.sessionId)
const route = useRoute()

function openRequestedFile(path: unknown) {
    if (typeof path !== 'string' || !path) return
    const name = path.split('/').at(-1) ?? path
    files.reveal({ name, path, dir: false }).catch(reportError)
}

watch(() => route.query.open, openRequestedFile, { immediate: true })
</script>

<template>
    <div class="split">
        <FilesSidebar :session-id="sessionId" :title="title" :hint="hint" />
        <FileEditor :session-id="sessionId" />
    </div>
</template>

<style scoped>
.split {
    display: grid;
    min-height: 0;
    flex: 1;
    grid-template-columns: 280px minmax(0, 1fr);

    @media (max-width: 760px) {
        grid-template-columns: 1fr;
        grid-template-rows: auto minmax(0, 1fr);
    }
}
</style>
