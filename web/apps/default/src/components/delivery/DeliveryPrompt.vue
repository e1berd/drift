<script setup lang="ts">
import { computed, useTemplateRef } from 'vue'
import { Dialog, DialogContent, DialogDescription, DialogTitle } from '@ui-sdk/vue/dialog'
import { toast } from '@ui-sdk/vue/toast'
import { useI18n } from 'vue-i18n'

import DeliveryOffer from '@/components/delivery/DeliveryOffer.vue'
import DeliveryProgress from '@/components/delivery/DeliveryProgress.vue'
import DeliverySkipConfirm from '@/components/delivery/DeliverySkipConfirm.vue'
import { useDelivery, watchRelaunch } from '@/composables/useDelivery'
import { reportError } from '@drift/sdk/errors'

const { t } = useI18n()
const { delivery, latest, installing, offerUpdate, installPending, install, skipLatest } =
    useDelivery()
const skipDialog = useTemplateRef<InstanceType<typeof DeliverySkipConfirm>>('skipDialog')

watchRelaunch((version) => toast.error(t('relaunchFailed', { version })))

const open = computed(() => installing.value || (offerUpdate.value && !!latest.value))

function update() {
    if (latest.value) install(latest.value.version).catch(reportError)
}

async function decline() {
    if (!latest.value || installing.value) return
    const answer = await skipDialog.value?.ask(latest.value.version)
    if (answer === 'skip') skipLatest()
    if (answer === 'update') update()
}
</script>

<template>
    <Dialog :open="open" :dismissible="false" @dismiss-attempt="decline">
        <DialogContent class="delivery-prompt">
            <div v-if="installing && delivery" class="flex flex-col gap-3">
                <DialogTitle>{{ t('updating') }}</DialogTitle>
                <DialogDescription class="sr-only">{{ t('updating') }}</DialogDescription>
                <DeliveryProgress :installation="delivery.installation" />
            </div>
            <DeliveryOffer
                v-else-if="delivery && latest"
                :delivery="delivery"
                :release="latest"
                :pending="installPending"
                @update="update"
                @decline="decline"
            />
        </DialogContent>
    </Dialog>
    <DeliverySkipConfirm ref="skipDialog" />
</template>

<style scoped>
.delivery-prompt {
    width: min(92vw, 40rem);
    max-height: 90dvh;
    display: flex;
    flex-direction: column;
}
</style>

<i18n lang="yaml">
ru:
    updating: Обновляем Drift
    relaunchFailed: 'Drift перезапустился, но работает версия {version} — обновление не применилось.'
en:
    updating: Updating Drift
    relaunchFailed: 'Drift restarted but is running {version}; the update did not apply.'
</i18n>
