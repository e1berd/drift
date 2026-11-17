import type { AgentId } from '@/shared/agents'
import { AGENTS, type PermissionMode } from '@/shared/agents'
import { storedRef } from '@/shared/storage'

interface ComposerSettings {
    models: Partial<Record<AgentId, string>>
    mode: PermissionMode
}

const state = storedRef<ComposerSettings>('drift.composer', { models: {}, mode: 'default' })

/** Model per agent and the permission mode last picked in any composer on this device. */
export function useComposerSettings() {
    return {
        modelFor: (agent: AgentId) =>
            state.value.models[agent] ?? AGENTS[agent].models.at(0)?.id ?? 'default',
        setModel: (agent: AgentId, model: string) => {
            state.value.models[agent] = model
        },
        mode: () => state.value.mode,
        setMode: (mode: PermissionMode) => {
            state.value.mode = mode
        },
    }
}
