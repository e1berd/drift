<script setup lang="ts">
import { ref } from 'vue'
import { PhArrowCircleUp } from '@phosphor-icons/vue'
import { useI18n } from 'vue-i18n'

import DeliveryVersionsDialog from '@/components/delivery/DeliveryVersionsDialog.vue'
import { useDelivery } from '@/composables/useDelivery'

const { t } = useI18n()
const { delivery, latest, remindNow } = useDelivery()
const versionsOpen = ref(false)

function showUpdate() {
    if (delivery.value?.installBlocker) versionsOpen.value = true
    else remindNow()
}
</script>

<template>
    <span v-if="delivery" class="delivery-version">
        <button
            type="button"
            class="version-label"
            v-tooltip="t('versions')"
            :aria-label="`${t('versions')}: ${delivery.currentVersion}`"
            @click="versionsOpen = true"
        >
            {{ delivery.currentVersion }}
        </button>
        <button
            v-if="delivery.updateAvailable && latest"
            type="button"
            class="version-update"
            v-tooltip="t('update', { version: latest.version })"
            :aria-label="t('update', { version: latest.version })"
            @click="showUpdate"
        >
            <PhArrowCircleUp weight="fill" />
        </button>
        <DeliveryVersionsDialog v-model:open="versionsOpen" />
    </span>
</template>

<style scoped>
.delivery-version {
    display: inline-flex;
    align-items: center;
    gap: 0.25rem;
    min-width: 0;
}

.version-label,
.version-update {
    padding: 0;
    border: 0;
    background: none;
    font: inherit;
    cursor: pointer;
}

.version-label {
    flex: none;
    color: var(--ui-muted-foreground);
    font-family: var(--ui-font-mono);
    font-size: 0.75rem;
    white-space: nowrap;

    &:hover {
        color: var(--ui-foreground);
    }
}

.version-update {
    display: inline-flex;
    flex: none;
    color: var(--ui-primary);
    font-size: 1rem;

    &:hover {
        opacity: 0.8;
    }
}
</style>

<i18n lang="yaml">
ru:
    versions: Версия Drift и выбор другой версии
    update: 'Доступна версия {version}'
en:
    versions: Drift version and choosing another one
    update: 'Version {version} is available'
</i18n>
