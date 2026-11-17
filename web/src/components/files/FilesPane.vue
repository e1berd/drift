<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { PhFolderSimple } from '@phosphor-icons/vue'

import type { Project } from '@/gen'

import CodeView from '@/components/files/CodeView.vue'
import FileTree from '@/components/files/FileTree.vue'
import { useReadFile } from '@/gen'
import { useI18n } from 'vue-i18n'

const props = defineProps<{ project: Project }>()
const { t } = useI18n()
const selected = ref<string | null>(null)

watch(
    () => props.project.id,
    () => (selected.value = null),
)

const file = useReadFile(
    {
        path: computed(() => ({ projectId: props.project.id })),
        query: computed(() => ({ path: selected.value ?? '' })),
    },
    { query: { enabled: () => !!selected.value } },
)

const lineCount = computed(() => file.data.value?.content.split('\n').length ?? 0)
</script>

<template>
    <div class="split">
        <aside
            class="flex min-h-0 flex-col border-r border-line max-md:max-h-[220px] max-md:border-r-0 max-md:border-b"
        >
            <div class="col-h">
                <PhFolderSimple :size="15" class="text-accent" />
                <b class="font-medium text-fg">{{ project.name }}</b>
                <span class="truncate font-mono text-[11.5px]">{{ project.path }}</span>
            </div>
            <FileTree v-model="selected" :project-id="project.id" />
        </aside>
        <section class="flex min-h-0 min-w-0 flex-col">
            <div class="col-h">
                <b class="truncate font-mono font-medium text-fg">{{
                    selected ?? t('files.pick')
                }}</b>
                <span v-if="file.data.value && !file.data.value.binary" class="ml-auto shrink-0">
                    {{ t('files.lines', { count: lineCount }) }} · {{ t('files.readOnly') }}
                </span>
            </div>
            <div v-if="!selected" class="grid flex-1 place-items-center text-[13px] text-faint">
                {{ t('files.pick') }}
            </div>
            <div v-else-if="file.isPending.value" class="flex flex-col gap-1.5 p-4">
                <div
                    v-for="i in 8"
                    :key="i"
                    class="skeleton h-3.5"
                    :style="{ width: `${40 + ((i * 37) % 50)}%` }"
                />
            </div>
            <div
                v-else-if="file.data.value?.binary"
                class="grid flex-1 place-items-center text-[13px] text-faint"
            >
                {{ t('files.binary') }}
            </div>
            <template v-else-if="file.data.value">
                <p v-if="file.data.value.truncated" class="m-0 px-4 pt-2 text-[12px] text-faint">
                    {{ t('files.truncated') }}
                </p>
                <CodeView :path="file.data.value.path" :content="file.data.value.content" />
            </template>
        </section>
    </div>
</template>

<style scoped>
.split {
    display: grid;
    min-height: 0;
    flex: 1;
    grid-template-columns: 280px minmax(0, 1fr);
}

.col-h {
    display: flex;
    height: 40px;
    flex: none;
    align-items: center;
    gap: 8px;
    padding: 0 14px;
    border-bottom: 1px solid var(--ui-border);
    color: var(--ui-muted-foreground);
    font-size: 12.5px;
}

@media (max-width: 760px) {
    .split {
        grid-template-columns: 1fr;
        grid-template-rows: auto minmax(0, 1fr);
    }
}
</style>
