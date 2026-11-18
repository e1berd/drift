import { computed } from 'vue'

import type { AgentId } from '@/shared/agents'

import { useListSubscriptions } from '@/gen'
import { AGENTS } from '@/shared/agents'

/** Agents with a signed-in subscription; only these are offered for a new chat. */
export function useAgents() {
    const query = useListSubscriptions()
    const subscriptions = computed(() => query.data.value?.subscriptions ?? [])
    const connected = computed(() => [
        ...new Set(
            subscriptions.value
                .filter((subscription) => subscription.state === 'active')
                .map((subscription) => subscription.agent),
        ),
    ])
    const isConnected = (id: AgentId) => connected.value.includes(id)
    const available = computed(() => Object.values(AGENTS).filter((agent) => isConnected(agent.id)))
    const ready = computed(() => query.isSuccess.value)
    return {
        all: Object.values(AGENTS),
        available,
        subscriptions,
        connected,
        isConnected,
        ready,
        query,
    }
}
