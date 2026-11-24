<script setup lang="ts">
import { computed, useTemplateRef } from 'vue'
import { Drawer, DrawerBackdrop, DrawerContent, DrawerTitle } from '@ui-sdk/vue/drawer'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'

import InfobarLinks from '@/components/infobar/InfobarLinks.vue'
import RemoveEntryDialog from '@/components/infobar/RemoveEntryDialog.vue'
import SessionInfobar from '@/components/infobar/SessionInfobar.vue'
import { activityLinks } from '@/components/infobar/linkKeys'
import { provideActivityActions } from '@drift/vue/composables/useActivityActions'
import { useActivityRemoval } from '@/composables/useActivityRemoval'
import { useInfobar } from '@drift/vue/composables/useInfobar'
import { useInfobarLinks } from '@drift/vue/composables/useInfobarLinks'
import { useSessionMessages } from '@drift/vue/composables/useMessageWindow'
import { useSessionSummary } from '@drift/vue/composables/useSessionSummary'
import type { ActivityEntry, Message } from '@drift/vue/api'

const props = defineProps<{ sessionId: string; messages: readonly Message[]; running: boolean }>()

const { t } = useI18n()
const { narrow, docked, drawerOpen, closeDrawer } = useInfobar()
const shown = computed(() => docked.value && !narrow.value)

const summary = useSessionSummary(
    () => props.sessionId,
    () => props.messages,
)
const layout = useTemplateRef<HTMLElement>('layout')
const links = computed(() => (shown.value ? activityLinks(summary.data.value) : []))
const { routes, activeMessageId } = useInfobarLinks(layout, links)

const sessionMessages = useSessionMessages()

function reveal(messageId: string) {
    closeDrawer()
    if (!sessionMessages.reveal(messageId)) return
    const card = layout.value?.querySelector(`[data-message-id="${CSS.escape(messageId)}"]`)
    const calm = matchMedia('(prefers-reduced-motion: reduce)').matches
    card?.scrollIntoView({ block: 'center', behavior: calm ? 'auto' : 'smooth' })
}

const router = useRouter()

function canOpen(entry: ActivityEntry): boolean {
    const outsideFolder = /^(\/|~|\.\.\/)/.test(entry.name)
    return entry.kind === 'file' && !entry.removed && !outsideFolder
}

function open(entry: ActivityEntry) {
    closeDrawer()
    void router.push({
        name: 'session',
        params: { sessionId: props.sessionId, tab: 'files' },
        query: { open: entry.name },
    })
}

const removeDialog = useTemplateRef<InstanceType<typeof RemoveEntryDialog>>('removeDialog')
const { remove } = useActivityRemoval(
    () => props.sessionId,
    () => removeDialog.value?.ask,
)

provideActivityActions({ reveal, open, remove, canOpen })
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
                />
            </DrawerContent>
        </Drawer>
        <div class="chat-column"><slot /></div>
        <InfobarLinks v-if="shown" :routes="routes" :active-message-id="activeMessageId" />
        <RemoveEntryDialog ref="removeDialog" />
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
