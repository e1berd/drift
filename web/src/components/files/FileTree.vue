<script setup lang="ts">
import { computed } from 'vue'
import { Tree } from '@ui-sdk/vue/tree'

import type { TreeItemData } from '@ui-sdk/vue/tree'
import ScrollPane from '@/components/ScrollPane.vue'
import type { TreeEntry } from '@/gen'

import FileRows from '@/components/files/FileRows.vue'
import { useSessionFiles } from '@/composables/useSessionFiles'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const { root, children, expanded, selected, loadChildren } = useSessionFiles()

function toTreeItems(entries: TreeEntry[]): TreeItemData[] {
    return entries.map((entry) => ({
        value: entry.path,
        label: entry.name,
        children: entry.dir ? toTreeItems(children.get(entry.path) ?? []) : undefined,
    }))
}

function collectDirectories(entries: TreeEntry[], into: Set<string>): Set<string> {
    for (const entry of entries) {
        if (!entry.dir) continue
        into.add(entry.path)
        collectDirectories(children.get(entry.path) ?? [], into)
    }
    return into
}

const items = computed(() => toTreeItems(root.data.value?.entries ?? []))
const directories = computed(() => collectDirectories(root.data.value?.entries ?? [], new Set()))

function open(path: string) {
    if (!directories.value.has(path)) selected.value = path
}
</script>

<template>
    <ScrollPane class="min-h-0 flex-1" content-class="p-1.5">
        <div v-if="root.isPending.value" class="flex flex-col gap-1">
            <div
                v-for="row in 6"
                :key="row"
                class="skeleton h-6"
                :style="{ opacity: 1 - row * 0.12 }"
            />
        </div>
        <Tree
            v-else
            v-model:expanded="expanded"
            :value="selected"
            :items="items"
            :load-children="loadChildren"
            :expand-on-press="true"
            :aria-label="t('files.label')"
            @action="open"
        >
            <FileRows :items="items" />
        </Tree>
    </ScrollPane>
</template>
