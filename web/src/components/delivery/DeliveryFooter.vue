<script setup lang="ts">
import { computed, ref } from 'vue'
import { PhArrowCircleUp, PhGithubLogo, PhStar } from '@phosphor-icons/vue'
import { Button } from '@ui-sdk/vue/button'
import { useI18n } from 'vue-i18n'

import DeliveryVersionsDialog from '@/components/delivery/DeliveryVersionsDialog.vue'
import { useDelivery } from '@/composables/useDelivery'

const { t, locale } = useI18n()
const { delivery, latest, remindNow } = useDelivery()
const versionsOpen = ref(false)

const stars = computed(() =>
    new Intl.NumberFormat(locale.value, { notation: 'compact' }).format(
        delivery.value?.repository.stars ?? 0,
    ),
)

function showUpdate() {
    if (delivery.value?.installBlocker) versionsOpen.value = true
    else remindNow()
}
</script>

<template>
    <div v-if="delivery" class="delivery-footer">
        <Button
            class="delivery-version"
            :title="t('versions')"
            :aria-label="`${t('versions')}: ${delivery.currentVersion}`"
            @click="versionsOpen = true"
        >
            drift <code>{{ delivery.currentVersion }}</code>
        </Button>
        <Button
            v-if="delivery.updateAvailable && latest"
            data-variant="solid"
            class="delivery-update"
            :title="t('update', { version: latest.version })"
            @click="showUpdate"
        >
            <PhArrowCircleUp />{{ latest.version }}
        </Button>
        <a
            :href="delivery.repository.url"
            target="_blank"
            rel="noopener noreferrer"
            class="delivery-repo"
            :title="t('repo', { stars: delivery.repository.stars })"
        >
            <PhGithubLogo /><PhStar weight="fill" />{{ stars }}
        </a>
        <DeliveryVersionsDialog v-model:open="versionsOpen" />
    </div>
</template>

<style scoped>
.delivery-footer {
    display: flex;
    align-items: center;
    gap: 0.375rem;
    padding: 0 0.625rem 0.5rem;
    font-size: 0.8125rem;
    color: var(--ui-muted-foreground);
}

.delivery-version {
    min-width: 0;

    & code {
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }
}

.delivery-repo {
    display: inline-flex;
    align-items: center;
    gap: 0.25rem;
    margin-inline-start: auto;
    color: inherit;
    text-decoration: none;

    &:hover {
        color: var(--ui-foreground);
    }
}
</style>

<i18n lang="yaml">
ru:
    versions: Версия Drift и выбор другой версии
    update: 'Доступна версия {version}'
    repo: 'Drift на GitHub, звёзд: {stars}'
en:
    versions: Drift version and choosing another one
    update: 'Version {version} is available'
    repo: 'Drift on GitHub, {stars} stars'
</i18n>
