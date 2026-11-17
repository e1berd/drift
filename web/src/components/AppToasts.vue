<script setup lang="ts">
import { PhCheck, PhDeviceMobile, PhInfo, PhStop, PhWarning, PhX } from '@phosphor-icons/vue'
import {
    Toast,
    ToastAction,
    ToastClose,
    ToastDescription,
    ToastRegion,
    ToastTitle,
} from '@ui-sdk/vue/toast'

import { useI18n } from '@/shared/i18n'

const { t } = useI18n()

const icons = {
    success: PhCheck,
    info: PhInfo,
    warning: PhWarning,
    error: PhStop,
    device: PhDeviceMobile,
}

function iconOf(item: { type?: string; data?: unknown }) {
    const kind = (item.data as { icon?: keyof typeof icons } | undefined)?.icon ?? item.type
    return icons[kind as keyof typeof icons] ?? PhCheck
}
</script>

<template>
    <ToastRegion class="toast-region" side="top" align="end">
        <template #default="{ toast: item }">
            <Toast class="toast">
                <span class="toast-icon" :data-type="item.type ?? 'success'">
                    <component :is="iconOf(item)" :size="15" weight="bold" />
                </span>
                <span class="min-w-0 flex-1 leading-snug">
                    <ToastTitle class="block text-[13px] font-semibold" />
                    <ToastDescription class="block truncate text-[12.5px] text-muted" />
                </span>
                <ToastAction class="toast-action" />
                <ToastClose class="toast-close" :aria-label="t('toast.close')">
                    <PhX :size="14" />
                </ToastClose>
            </Toast>
        </template>
    </ToastRegion>
</template>

<style>
.toast-region {
    position: fixed;
    top: 64px;
    right: 20px;
    z-index: 90;
    width: min(340px, calc(100vw - 24px));
    height: var(--toast-front-height);
    transition: height var(--dur-3) var(--ease-out);

    &[data-expanded] {
        height: calc(var(--toast-stack-height) + (var(--toast-count) - 1) * 8px);
    }
}

.toast {
    --y: calc(var(--toast-index) * 8px);
    position: absolute;
    inset-inline: 0;
    top: 0;
    z-index: calc(100 - var(--toast-index));
    display: flex;
    align-items: flex-start;
    gap: 12px;
    height: var(--toast-height);
    overflow: hidden;
    padding: 12px 10px 12px 14px;
    border: 1px solid var(--border-strong);
    border-radius: 12px;
    background: var(--surface);
    box-shadow: var(--shadow);
    translate: var(--toast-swipe-x, 0px) calc(var(--y) + var(--toast-swipe-y, 0px));
    scale: calc(1 - var(--toast-index) * 0.04);
    transition-property: translate, scale, opacity, height;
    transition-duration: var(--dur-4);
    transition-timing-function: var(--ease-out);
    user-select: none;

    &:not([data-front], [data-expanded]) {
        height: var(--toast-front-height);
    }

    &[data-expanded] {
        --y: calc(var(--toast-offset) + var(--toast-index) * 8px);
        scale: 1;
    }

    &[data-starting-style] {
        transition: none;
    }

    &[data-starting-style][data-front] {
        --y: -12px;
        opacity: 0;
    }

    &:is([data-starting-style]:not([data-front]), [data-ending-style]) {
        opacity: 0;
    }

    &[data-ending-style] {
        transition-duration: var(--dur-2);
        transition-timing-function: var(--ease-in);
    }

    &[data-swiping] {
        transition: none;
    }
}

.toast-icon {
    display: grid;
    place-items: center;
    width: 26px;
    height: 26px;
    flex: none;
    border-radius: 50%;
    background: var(--add-bg);
    color: var(--ok);

    &[data-type='info'] {
        background: var(--accent-soft);
        color: var(--accent);
    }

    &[data-type='error'],
    &[data-type='warning'] {
        background: var(--del-bg);
        color: var(--ui-destructive);
    }
}

.toast-action:not(:empty) {
    height: 28px;
    padding: 0 10px;
    border: 1px solid var(--border-strong);
    border-radius: 7px;
    background: var(--surface-2);
    font-size: 12px;
    font-weight: 500;
}

.toast-close {
    display: grid;
    place-items: center;
    width: 26px;
    height: 26px;
    border: 0;
    border-radius: 7px;
    background: none;
    color: var(--faint);

    &:hover {
        background: var(--surface-2);
        color: var(--ui-foreground);
    }
}
</style>
