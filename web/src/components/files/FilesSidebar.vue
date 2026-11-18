<script setup lang="ts">
import { ref } from 'vue'
import { PhFilePlus, PhFolderPlus, PhFolderSimple } from '@phosphor-icons/vue'
import { Button } from '@ui-sdk/vue/button'
import { useI18n } from 'vue-i18n'

import type { Project } from '@/gen'

import FileTree from '@/components/files/FileTree.vue'
import NewEntryForm from '@/components/files/NewEntryForm.vue'

defineProps<{ project: Project }>()
const { t } = useI18n()
const creating = ref<'file' | 'dir' | null>(null)

function toggle(kind: 'file' | 'dir') {
    creating.value = creating.value === kind ? null : kind
}
</script>

<template>
    <aside
        class="flex min-h-0 flex-col border-r border-line max-md:max-h-[260px] max-md:border-r-0 max-md:border-b"
    >
        <div class="files-head">
            <PhFolderSimple class="shrink-0 text-fg" />
            <b class="truncate font-medium text-fg" :title="project.path">{{ project.name }}</b>
            <div class="ml-auto flex shrink-0 gap-0.5">
                <Button
                    :aria-label="t('newFile')"
                    :title="t('newFile')"
                    :aria-pressed="creating === 'file'"
                    @click="toggle('file')"
                >
                    <PhFilePlus />
                </Button>
                <Button
                    :aria-label="t('newDir')"
                    :title="t('newDir')"
                    :aria-pressed="creating === 'dir'"
                    @click="toggle('dir')"
                >
                    <PhFolderPlus />
                </Button>
            </div>
        </div>
        <NewEntryForm
            v-if="creating"
            :key="creating"
            :project-id="project.id"
            :dir="creating === 'dir'"
            @done="creating = null"
        />
        <FileTree />
    </aside>
</template>

<style scoped>
.files-head {
    display: flex;
    height: 40px;
    flex: none;
    align-items: center;
    gap: 8px;
    padding: 0 6px 0 14px;
    border-bottom: 1px solid var(--ui-border);
    font-size: 12.5px;
}
</style>

<i18n lang="yaml">
ru:
    newFile: Новый файл
    newDir: Новая папка
en:
    newFile: New file
    newDir: New folder
</i18n>
