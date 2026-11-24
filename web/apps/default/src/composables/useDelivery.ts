import { computed, ref, watch } from 'vue'
import { useQueryClient } from '@tanstack/vue-query'
import { toast } from '@ui-sdk/vue/toast'

import type { Delivery } from '@drift/vue/api'

import {
    getDeliveryQueryKey,
    useCheckDelivery,
    useGetDelivery,
    useInstallRelease,
} from '@drift/vue/api'
import { errorMessage } from '@drift/sdk/errors'
import { storedRef } from '@drift/vue/storedRef'

interface Snooze {
    version: string
    until: number
}

const SNOOZE_MS = 24 * 60 * 60 * 1000
const RESTART_POLL_MS = 2000
const ACTIVE_STATES = new Set(['downloading', 'restarting'])

const snooze = storedRef<Snooze>('drift.delivery.snooze', { version: '', until: 0 })
const awaitingVersion = ref('')

function isSnoozed(version: string | undefined): boolean {
    return !!version && snooze.value.version === version && snooze.value.until > Date.now()
}

function reloadOnceRelaunched(delivery: Delivery | undefined, onFailed: (version: string) => void) {
    if (!delivery) return
    const { installation, currentVersion } = delivery
    if (installation.state === 'restarting' && installation.targetVersion)
        awaitingVersion.value = installation.targetVersion
    if (!awaitingVersion.value) return
    if (currentVersion === awaitingVersion.value) {
        location.reload()
        return
    }
    if (installation.state === 'idle') {
        awaitingVersion.value = ''
        onFailed(currentVersion)
    }
}

/** Update state of the instance: the running version, the latest release and an update in progress. */
export function useDelivery() {
    const qc = useQueryClient()
    const query = useGetDelivery({
        query: {
            staleTime: 5 * 60_000,
            refetchInterval: () => (awaitingVersion.value ? RESTART_POLL_MS : false),
            retry: () => !!awaitingVersion.value,
        },
    })
    const delivery = computed(() => query.data.value)
    const latest = computed(() => delivery.value?.latestRelease)
    const installing = computed(
        () =>
            !!awaitingVersion.value ||
            ACTIVE_STATES.has(delivery.value?.installation.state ?? 'idle'),
    )
    const offerUpdate = computed(
        () =>
            !!delivery.value?.updateAvailable &&
            !delivery.value.installBlocker &&
            !isSnoozed(latest.value?.version),
    )

    const store = (data: Delivery) => qc.setQueryData(getDeliveryQueryKey(), data)
    const onError = (error: unknown) => toast.error(errorMessage(error))
    const installMutation = useInstallRelease({ mutation: { onSuccess: store, onError } })
    const checkMutation = useCheckDelivery({ mutation: { onSuccess: store, onError } })

    function install(version: string) {
        return installMutation.mutateAsync({ body: { version } })
    }

    function skipLatest() {
        if (latest.value)
            snooze.value = { version: latest.value.version, until: Date.now() + SNOOZE_MS }
    }

    function remindNow() {
        snooze.value = { version: '', until: 0 }
    }

    return {
        delivery,
        latest,
        installing,
        offerUpdate,
        installPending: installMutation.isPending,
        install,
        check: () => checkMutation.mutate(undefined),
        checking: checkMutation.isPending,
        skipLatest,
        remindNow,
    }
}

/** Reloads the tab once the relaunched server reports the version being installed. */
export function watchRelaunch(onFailed: (version: string) => void) {
    const { delivery } = useDelivery()
    watch(delivery, (value) => reloadOnceRelaunched(value, onFailed), { immediate: true })
}
