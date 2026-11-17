<script setup lang="ts">
withDefaults(defineProps<{ size?: number; fast?: boolean; label?: string }>(), {
    size: 16,
    fast: false,
    label: '',
})
</script>

<template>
    <svg
        :width="size"
        :height="size"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2.2"
        stroke-linecap="round"
        :class="[fast ? 'motion-spin-fast' : 'motion-spin', 'shrink-0 text-accent']"
        :role="label ? 'img' : undefined"
        :aria-label="label || undefined"
        :aria-hidden="label ? undefined : 'true'"
    >
        <circle cx="12" cy="12" r="9" stroke-opacity=".2" />
        <path d="M12 3a9 9 0 0 1 9 9" />
    </svg>
</template>
