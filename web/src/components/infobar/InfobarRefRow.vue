<script setup lang="ts">
defineProps<{
    linkKey: string
    messageId: string
    name: string
    meta: string
    active: boolean
}>()
const emit = defineEmits<{ reveal: [messageId: string] }>()
</script>

<template>
    <li>
        <button
            type="button"
            class="ref-row"
            :data-link-target="linkKey"
            :data-link-source="messageId"
            :data-active="active || undefined"
            @click="emit('reveal', messageId)"
        >
            <span class="truncate font-medium" :title="name">{{ name }}</span>
            <span class="truncate text-muted">{{ meta }}</span>
        </button>
    </li>
</template>

<style scoped>
.ref-row {
    display: grid;
    width: 100%;
    gap: 2px;
    padding: calc(var(--ui-spacing) * 1.5) calc(var(--ui-spacing) * 2);
    border: 1px solid transparent;
    border-radius: var(--ui-radius);
    background: none;
    color: inherit;
    font: inherit;
    text-align: start;
    cursor: pointer;
    transition:
        background-color var(--dur-2) var(--ease-out),
        border-color var(--dur-2) var(--ease-out);

    &:hover,
    &[data-active] {
        border-color: color-mix(in oklab, var(--ui-primary) 45%, transparent);
        background: color-mix(in oklab, var(--ui-primary) 10%, transparent);
    }

    &:focus-visible {
        outline: 2px solid var(--ui-ring);
        outline-offset: 1px;
    }

    @media (pointer: coarse) {
        min-height: 44px;
        align-content: center;
    }
}
</style>
