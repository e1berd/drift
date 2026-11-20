<script setup lang="ts">
import { computed, onScopeDispose, ref, watch } from 'vue'
import { PhArrowUp } from '@phosphor-icons/vue'
import { useI18n } from 'vue-i18n'

import type { Message } from '@/gen'

const props = defineProps<{ message?: Message; viewport: HTMLElement | null }>()
const { t } = useI18n()

const outOfView = ref(false)
const summary = computed(() => {
    const text = props.message?.body.text?.trim()
    if (text) return text
    return t('attachments', { count: props.message?.body.attachments?.length ?? 0 })
})

function messageElement(): HTMLElement | null {
    if (!props.viewport || !props.message) return null
    return props.viewport.querySelector<HTMLElement>(
        `[data-message-id="${CSS.escape(props.message.id)}"]`,
    )
}

let observer: IntersectionObserver | null = null

function observeMessage() {
    observer?.disconnect()
    outOfView.value = false
    const target = messageElement()
    if (!target || !props.viewport || typeof IntersectionObserver === 'undefined') return
    observer = new IntersectionObserver(
        ([entry]) => (outOfView.value = !!entry && !entry.isIntersecting),
        { root: props.viewport },
    )
    observer.observe(target)
}

watch(() => [props.message?.id, props.viewport] as const, observeMessage, {
    immediate: true,
    flush: 'post',
})
onScopeDispose(() => observer?.disconnect())

function jump() {
    const target = messageElement()
    if (!target) return
    target.scrollIntoView({ block: 'start', behavior: 'smooth' })
    target.classList.remove('motion-flash')
    requestAnimationFrame(() => target.classList.add('motion-flash'))
}
</script>

<template>
    <button
        type="button"
        class="last-prompt"
        :data-visible="(message && outOfView) || undefined"
        :inert="!(message && outOfView)"
        @click="jump"
    >
        <span class="last-prompt-text">{{ summary }}</span>
        <span class="last-prompt-jump" aria-hidden="true"><PhArrowUp /></span>
        <span class="sr-only">{{ t('jump') }}</span>
    </button>
</template>

<style scoped>
.last-prompt {
    position: absolute;
    z-index: 2;
    inset-inline: 0;
    top: 0;
    display: flex;
    align-items: center;
    gap: calc(var(--ui-spacing) * 3);
    width: min(780px, calc(100% - 2rem));
    margin-inline: auto;
    margin-top: calc(var(--ui-spacing) * 2);
    padding: calc(var(--ui-spacing) * 2) calc(var(--ui-spacing) * 2) calc(var(--ui-spacing) * 2)
        calc(var(--ui-spacing) * 3);
    border: 1px solid var(--ui-border);
    border-radius: var(--ui-radius);
    background: color-mix(in oklab, var(--ui-background) 88%, transparent);
    backdrop-filter: blur(12px);
    box-shadow: var(--ui-shadow);
    color: var(--ui-foreground);
    font: inherit;
    text-align: start;
    cursor: pointer;
    opacity: 0;
    translate: 0 -6px;
    pointer-events: none;
    transition:
        opacity var(--dur-2) var(--ease-out),
        translate var(--dur-2) var(--ease-out);

    &[data-visible] {
        opacity: 1;
        translate: 0 0;
        pointer-events: auto;
    }

    &:hover .last-prompt-jump {
        background: var(--ui-muted);
        color: var(--ui-foreground);
    }

    &:focus-visible {
        outline: 4px solid var(--ui-ring);
    }
}

.last-prompt-text {
    display: -webkit-box;
    min-width: 0;
    flex: 1;
    overflow: hidden;
    color: var(--ui-muted-foreground);
    line-height: 1.4;
    white-space: pre-wrap;
    overflow-wrap: anywhere;
    -webkit-box-orient: vertical;
    -webkit-line-clamp: 2;
    line-clamp: 2;
}

.last-prompt-jump {
    display: grid;
    flex: none;
    place-items: center;
    width: 28px;
    height: 28px;
    border-radius: var(--ui-radius);
    color: var(--ui-muted-foreground);
    transition: background var(--dur-1);
}
</style>

<i18n lang="yaml">
ru:
    jump: Перейти к моему последнему сообщению
    attachments: 'Вложения: {count}'
en:
    jump: Jump to my last message
    attachments: 'Attachments: {count}'
</i18n>
