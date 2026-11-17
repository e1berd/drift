<script setup lang="ts">
withDefaults(defineProps<{ size?: number }>(), { size: 20 })
</script>

<template>
    <svg
        :width="size"
        :height="size"
        viewBox="0 0 24 24"
        fill="none"
        stroke="var(--accent)"
        stroke-width="2"
        stroke-linecap="round"
        aria-hidden="true"
    >
        <path d="M3 8c3-3 6 3 9 0s6 3 9 0" />
        <path d="M3 13c3-3 6 3 9 0s6 3 9 0" opacity=".65" />
        <path d="M3 18c3-3 6 3 9 0s6 3 9 0" opacity=".35" />
    </svg>
</template>
