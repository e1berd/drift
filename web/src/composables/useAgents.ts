import { computed } from 'vue'

import type { AgentId } from '@/shared/agents'
import { AGENTS } from '@/shared/agents'
import { storedRef } from '@/shared/storage'

const stored = storedRef<AgentId[]>('drift.agents', ['claude', 'codex'])
const connected = computed({
    get: () => stored.value.filter((id) => Object.hasOwn(AGENTS, id)),
    set: (list: AgentId[]) => (stored.value = list),
})

/** Agents marked as connected in settings; only these are offered for a new chat. */
export function useAgents() {
    const all = Object.values(AGENTS)
    const isConnected = (id: AgentId) => connected.value.includes(id)
    function toggle(id: AgentId, on: boolean) {
        connected.value = on
            ? [...new Set([...connected.value, id])]
            : connected.value.filter((agentId) => agentId !== id)
    }
    return { all, connected: computed(() => connected.value), isConnected, toggle }
}
