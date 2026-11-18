<script setup lang="ts">
import type { Project } from '@/gen'

import FileEditor from '@/components/files/FileEditor.vue'
import FilesSidebar from '@/components/files/FilesSidebar.vue'
import { provideProjectFiles } from '@/composables/useProjectFiles'

const props = defineProps<{ project: Project }>()
provideProjectFiles(() => props.project.id)
</script>

<template>
    <div class="split">
        <FilesSidebar :project="project" />
        <FileEditor :project-id="project.id" />
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
