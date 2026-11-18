import { computed, toValue } from 'vue'
import type { MaybeRefOrGetter } from 'vue'

import type { AgentId } from '@/shared/agents'

import { useAgents } from '@/composables/useAgents'
import { useComposerSettings } from '@/composables/useComposerSettings'
import { useGetSubscriptionUsage } from '@/gen'
import { headlineWindows } from '@/shared/usage'

const USAGE_REFRESH_MS = 60_000

export function useAgentUsage(agent: MaybeRefOrGetter<AgentId>) {
    const agents = useAgents()
    const settings = useComposerSettings()

    const subscription = computed(() => {
        const active = agents.subscriptions.value.filter(
            (item) => item.agent === toValue(agent) && item.state === 'active',
        )
        return active.find((item) => item.current) ?? active.at(0)
    })

    const usage = useGetSubscriptionUsage(
        { path: () => ({ subscriptionId: subscription.value?.id ?? '' }) },
        {
            query: {
                enabled: () => !!subscription.value,
                refetchInterval: USAGE_REFRESH_MS,
                staleTime: USAGE_REFRESH_MS / 2,
            },
        },
    )

    const windows = computed(() => usage.data.value?.windows ?? [])
    const headline = computed(() =>
        headlineWindows(windows.value, settings.modelFor(toValue(agent))),
    )

    return {
        subscription,
        windows,
        headline,
        unavailable: computed(() => usage.data.value?.unavailable),
        loading: computed(() => !!subscription.value && usage.isPending.value),
        refresh: () => usage.refetch(),
    }
}
