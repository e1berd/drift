<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { PhArrowsClockwise, PhGithubLogo, PhStar } from '@phosphor-icons/vue'
import { Button } from '@ui-sdk/vue/button'
import {
    Dialog,
    DialogClose,
    DialogContent,
    DialogDescription,
    DialogTitle,
} from '@ui-sdk/vue/dialog'
import { useI18n } from 'vue-i18n'

import ReleaseNotes from '@/components/delivery/ReleaseNotes.vue'
import ReleaseSelect from '@/components/delivery/ReleaseSelect.vue'
import { useListReleases } from '@drift/vue/api'
import { useDelivery } from '@/composables/useDelivery'
import { reportError } from '@drift/sdk/errors'

const open = defineModel<boolean>('open', { required: true })
const { t, locale } = useI18n()
const { delivery, latest, install, installPending, installing, check, checking } = useDelivery()
const releasesQuery = useListReleases({ query: { enabled: open } })

const releases = computed(() => releasesQuery.data.value?.releases ?? [])
const stars = computed(() =>
    new Intl.NumberFormat(locale.value, { notation: 'compact' }).format(
        delivery.value?.repository.stars ?? 0,
    ),
)
const current = computed(() => delivery.value?.currentVersion ?? '')
const chosenVersion = ref('')
const chosen = computed(() =>
    releases.value.find((release) => release.version === chosenVersion.value),
)

const isDowngrade = computed(() => {
    const currentIndex = releases.value.findIndex((release) => release.version === current.value)
    const chosenIndex = releases.value.findIndex(
        (release) => release.version === chosenVersion.value,
    )
    return currentIndex >= 0 && chosenIndex > currentIndex
})
const blocker = computed(() => delivery.value?.installBlocker)
const canInstall = computed(
    () =>
        !!chosen.value?.installable &&
        chosen.value.version !== current.value &&
        !blocker.value &&
        !installing.value,
)

watch([open, releases], () => {
    if (!open.value || chosen.value) return
    chosenVersion.value = latest.value?.version ?? releases.value.at(0)?.version ?? ''
})

function installChosen() {
    if (!chosen.value) return
    install(chosen.value.version)
        .then(() => (open.value = false))
        .catch(reportError)
}
</script>

<template>
    <Dialog v-model:open="open">
        <DialogContent class="delivery-versions">
            <div class="flex items-start gap-3">
                <div class="flex min-w-0 flex-1 flex-col gap-1">
                    <DialogTitle>{{ t('title') }}</DialogTitle>
                    <DialogDescription>
                        {{ t('current', { version: current }) }}
                        <template v-if="delivery && !delivery.releaseBuild"
                            >· {{ t('local') }}</template
                        >
                    </DialogDescription>
                </div>
                <a
                    v-if="delivery"
                    :href="delivery.repository.url"
                    target="_blank"
                    rel="noopener noreferrer"
                    class="delivery-repo"
                    v-tooltip="t('repo', { stars: delivery.repository.stars })"
                >
                    <PhGithubLogo /><PhStar weight="fill" />{{ stars }}
                </a>
            </div>

            <div class="flex items-center gap-2">
                <ReleaseSelect
                    v-model="chosenVersion"
                    class="min-w-0 flex-1"
                    :releases="releases"
                    :current="current"
                    :loading="releasesQuery.isPending.value"
                />
                <Button
                    v-tooltip="t('check')"
                    :aria-label="t('check')"
                    :disabled="checking"
                    @click="check"
                >
                    <PhArrowsClockwise />
                </Button>
            </div>

            <ReleaseNotes v-if="chosen" :release="chosen" />

            <p v-if="blocker" class="m-0 text-muted">{{ t(`blocker.${blocker}`) }}</p>
            <p v-else-if="chosen && !chosen.installable" class="m-0 text-muted">
                {{ t('noBuild') }}
            </p>
            <p v-else-if="isDowngrade" class="m-0 text-danger">{{ t('downgrade') }}</p>
            <p
                v-if="delivery?.installation.state === 'failed'"
                class="m-0 text-danger"
                role="alert"
            >
                {{ t('failed', { error: delivery.installation.error }) }}
            </p>

            <div class="flex justify-end gap-2">
                <DialogClose>{{ t('close') }}</DialogClose>
                <Button
                    data-variant="solid"
                    :disabled="!canInstall || installPending"
                    @click="installChosen"
                >
                    {{ chosen ? t('install', { version: chosen.version }) : t('installNone') }}
                </Button>
            </div>
        </DialogContent>
    </Dialog>
</template>

<style scoped>
.delivery-versions {
    width: min(92vw, 40rem);
    max-height: 90dvh;
    display: flex;
    flex-direction: column;
    gap: 1rem;
}

.delivery-repo {
    display: inline-flex;
    flex: none;
    align-items: center;
    gap: 0.25rem;
    font-size: 0.8125rem;
    color: var(--ui-muted-foreground);
    text-decoration: none;

    &:hover {
        color: var(--ui-foreground);
    }
}
</style>

<i18n lang="yaml">
ru:
    title: Версия Drift
    current: 'Сейчас работает {version}'
    local: локальная сборка
    check: Проверить обновления на GitHub
    noBuild: У этого релиза нет сборки для сервера Drift.
    downgrade: Это откат на более старую версию. Миграции базы данных не откатываются, поэтому старая версия может работать с новой схемой некорректно.
    failed: 'Не удалось обновиться: {error}'
    close: Закрыть
    install: 'Установить {version}'
    installNone: Установить
    repo: 'Drift на GitHub, звёзд: {stars}'
    blocker:
        development_build: Это dev-сборка (go run / air) — устанавливать релизы в неё нельзя.
        read_only: У процесса Drift нет прав на запись в папку бинарника, поэтому обновиться сам он не может.
en:
    title: Drift version
    current: 'Running {version}'
    local: local build
    check: Check GitHub for updates
    noBuild: This release has no build for the Drift server.
    downgrade: This rolls back to an older version. Database migrations are not reverted, so the older version may not work correctly with the newer schema.
    failed: 'The update failed: {error}'
    close: Close
    install: 'Install {version}'
    installNone: Install
    repo: 'Drift on GitHub, {stars} stars'
    blocker:
        development_build: This is a dev build (go run / air); releases cannot be installed into it.
        read_only: The Drift process cannot write to the folder of its binary, so it cannot update itself.
</i18n>
