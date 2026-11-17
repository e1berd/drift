<script setup lang="ts">
import { PhCaretRight, PhFile, PhFolderSimple, PhFolderSimpleDashed } from '@phosphor-icons/vue'
import { TreeGroup, TreeItem, TreeItemLabel, type TreeItemData } from '@ui-sdk/vue/tree'

defineProps<{ items: TreeItemData[] }>()
</script>

<template>
    <TreeItem
        v-for="n in items"
        :key="n.value"
        :value="n.value"
        :branch="!!n.children"
        :label="n.label"
        class="tree-item"
    >
        <TreeItemLabel v-slot="{ state }" class="tree-label">
            <PhCaretRight v-if="n.children" :size="12" class="chev" />
            <span v-else class="w-3" />
            <PhFolderSimpleDashed
                v-if="n.children && state.loading"
                :size="15"
                class="text-accent"
            />
            <PhFolderSimple v-else-if="n.children" :size="15" class="text-accent" />
            <PhFile v-else :size="15" class="text-faint" />
            <span class="truncate">{{ n.label }}</span>
        </TreeItemLabel>
        <TreeGroup v-if="n.children" class="tree-group">
            <FileRows :items="n.children" />
        </TreeGroup>
    </TreeItem>
</template>

<style scoped>
.tree-item {
    outline: none;
}

.tree-label {
    display: flex;
    height: 28px;
    align-items: center;
    gap: 6px;
    padding-inline: calc(6px + (var(--tree-level, 1) - 1) * 14px) 8px;
    border-radius: 6px;
    color: var(--ui-muted-foreground);
    font-size: 13px;
    cursor: pointer;
    user-select: none;
    transition:
        background var(--dur-1),
        color var(--dur-1);

    &:hover {
        background: var(--surface-2);
        color: var(--ui-foreground);
    }

    &[data-selected] {
        background: var(--accent-soft);
        color: var(--ui-foreground);
    }

    &[data-failed] {
        color: var(--ui-destructive);
    }
}

.tree-item:focus-visible > .tree-label {
    outline: 2px solid var(--accent);
    outline-offset: -2px;
}

.chev {
    flex: none;
    color: var(--faint);
    transition: transform var(--dur-3) var(--ease-out);
}

.tree-label[data-expanded] .chev {
    transform: rotate(90deg);
}

.tree-group {
    height: var(--tree-group-height, auto);
    overflow: hidden;
    transition: height var(--dur-3) var(--ease-out);
}
</style>
