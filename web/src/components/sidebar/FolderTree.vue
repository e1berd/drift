<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { Tree } from '@ui-sdk/vue/tree'
import { useI18n } from 'vue-i18n'

import type { TreeItemData } from '@ui-sdk/vue/tree'
import type { TreeEntry } from '@/gen'

import FileRows from '@/components/files/FileRows.vue'
import { listFolders, useListFolders } from '@/gen'
import { errorMessage } from '@/shared/errors'

const selected = defineModel<string | null>({ default: null })
const { t } = useI18n()

const children = reactive(new Map<string, TreeEntry[]>())
const expanded = ref<string[]>([])
const root = useListFolders({ query: { path: '.' } })

function toTreeItems(entries: TreeEntry[]): TreeItemData[] {
    return entries.map((entry) => ({
        value: entry.path,
        label: entry.name,
        children: toTreeItems(children.get(entry.path) ?? []),
    }))
}

const items = computed(() => toTreeItems(root.data.value?.entries ?? []))

async function loadChildren(path: string, { signal }: { signal: AbortSignal }) {
    const { data } = await listFolders({ query: { path }, signal })
    children.set(path, data.entries)
}
</script>

<template>
    <div class="folder-tree">
        <div v-if="root.isPending.value" class="flex flex-col gap-1">
            <div
                v-for="row in 6"
                :key="row"
                class="skeleton h-6"
                :style="{ opacity: 1 - row * 0.12 }"
            />
        </div>
        <p v-else-if="root.error.value" class="text-destructive">
            {{ errorMessage(root.error.value) }}
        </p>
        <p v-else-if="!items.length" class="text-muted">{{ t('empty') }}</p>
        <Tree
            v-else
            v-model="selected"
            v-model:expanded="expanded"
            :items="items"
            :load-children="loadChildren"
            :aria-label="t('label')"
        >
            <FileRows :items="items" />
        </Tree>
    </div>
</template>

<style scoped>
.folder-tree {
    height: min(60dvh, 420px);
    overflow: auto;
    padding: calc(var(--ui-spacing) * 1.5);
    border: 1px solid var(--ui-border);
    border-radius: var(--ui-radius);
}
</style>

<i18n lang="yaml">
ru:
    label: 'Папки рабочего каталога'
    empty: 'В рабочем каталоге нет папок'
en:
    label: 'Workspace folders'
    empty: 'The workspace has no folders'
</i18n>
