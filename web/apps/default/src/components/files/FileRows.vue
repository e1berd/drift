<script setup lang="ts">
import { PhFile, PhFolderSimple } from '@phosphor-icons/vue'
import { TreeGroup, TreeItem, TreeItemIndicator, TreeItemLabel } from '@ui-sdk/vue/tree'

import type { TreeItemData } from '@ui-sdk/vue/tree'

defineProps<{ items: TreeItemData[] }>()
</script>

<template>
    <TreeItem
        v-for="item in items"
        :key="item.value"
        :value="item.value"
        :branch="!!item.children"
        :label="item.label"
    >
        <TreeItemLabel>
            <TreeItemIndicator v-if="item.children" />
            <PhFolderSimple v-if="item.children" />
            <PhFile v-else />
            <span class="truncate">{{ item.label }}</span>
        </TreeItemLabel>
        <TreeGroup v-if="item.children">
            <FileRows :items="item.children" />
        </TreeGroup>
    </TreeItem>
</template>
