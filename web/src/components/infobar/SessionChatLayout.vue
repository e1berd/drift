<script setup lang="ts">
import { computed, useTemplateRef } from 'vue'
import { Drawer, DrawerBackdrop, DrawerContent, DrawerTitle } from '@ui-sdk/vue/drawer'
import { useI18n } from 'vue-i18n'

import InfobarLinks from '@/components/infobar/InfobarLinks.vue'
import SessionInfobar from '@/components/infobar/SessionInfobar.vue'
import { summaryLinks } from '@/components/infobar/linkKeys'
import { useInfobar } from '@/composables/useInfobar'
import { useInfobarLinks } from '@/composables/useInfobarLinks'
import { useSessionSummary } from '@/composables/useSessionSummary'
import type { Message } from '@/gen'

const props = defineProps<{ sessionId: string; messages: Message[]; running: boolean }>()

const { t } = useI18n()
const { narrow, docked, drawerOpen, closeDrawer } = useInfobar()
const shown = computed(() => docked.value && !narrow.value)

const summary = useSessionSummary(
    () => props.sessionId,
    () => props.messages,
)
const layout = useTemplateRef<HTMLElement>('layout')
const links = computed(() => (shown.value ? summaryLinks(summary.data.value) : []))
const { routes, activeMessageId } = useInfobarLinks(layout, links)

function reveal(messageId: string) {
    closeDrawer()
    const card = layout.value?.querySelector(`[data-message-id="${CSS.escape(messageId)}"]`)
    const calm = matchMedia('(prefers-reduced-motion: reduce)').matches
    card?.scrollIntoView({ block: 'center', behavior: calm ? 'auto' : 'smooth' })
}
</script>

<template>
    <div ref="layout" class="chat-layout" :data-infobar-open="shown || undefined">
        <SessionInfobar
            v-if="!narrow"
            :inert="!shown"
            :summary="summary.data.value"
            :loading="summary.isPending.value"
            :failed="summary.isError.value"
            :running="running"
            :active-message-id="activeMessageId"
            @reveal="reveal"
        />
        <Drawer v-else v-model:open="drawerOpen" side="left">
            <DrawerBackdrop />
            <DrawerContent class="infobar-drawer">
                <DrawerTitle class="sr-only">{{ t('infobar.show') }}</DrawerTitle>
                <SessionInfobar
                    placement="drawer"
                    :summary="summary.data.value"
                    :loading="summary.isPending.value"
                    :failed="summary.isError.value"
                    :running="running"
                    :active-message-id="null"
                    @reveal="reveal"
                />
            </DrawerContent>
        </Drawer>
        <div class="chat-column"><slot /></div>
        <InfobarLinks v-if="shown" :routes="routes" :active-message-id="activeMessageId" />
    </div>
</template>

<style scoped>
.chat-layout {
    --infobar-width: 272px;

    position: relative;
    display: grid;
    min-height: 0;
    flex: 1;
    grid-template-columns: 0 minmax(0, 1fr);
    transition: grid-template-columns var(--dur-4) var(--ease-out);

    &[data-infobar-open] {
        grid-template-columns: var(--infobar-width) minmax(0, 1fr);
    }

    @media (max-width: 760px) {
        grid-template-columns: minmax(0, 1fr);
    }
}

:global(.infobar-drawer) {
    width: min(88vw, 340px);
    padding: 0;
}

.chat-column {
    display: flex;
    min-width: 0;
    min-height: 0;
    flex-direction: column;
}
</style>
