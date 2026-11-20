import type { AgentId } from '@/shared/agents'
import { AGENTS, type PermissionMode } from '@/shared/agents'
import { storedRef } from '@/shared/storage'

interface ComposerSettings {
    models: Partial<Record<AgentId, string>>
    sessionModels?: Record<string, string>
    mode: PermissionMode
}

const state = storedRef<ComposerSettings>('drift.composer', {
    models: {},
    sessionModels: {},
    mode: 'default',
})

const defaultModel = (agent: AgentId) =>
    state.value.models[agent] ?? AGENTS[agent].models.at(0)?.id ?? 'default'

/** Model per agent (or per session) and the permission mode last picked in any composer on this device. */
export function useComposerSettings() {
    return {
        modelFor: (agent: AgentId, sessionId?: string) =>
            (sessionId && state.value.sessionModels?.[sessionId]) || defaultModel(agent),
        setModel: (agent: AgentId, model: string, sessionId?: string) => {
            if (sessionId) (state.value.sessionModels ??= {})[sessionId] = model
            else state.value.models[agent] = model
        },
        mode: () => state.value.mode,
        setMode: (mode: PermissionMode) => {
            state.value.mode = mode
        },
    }
}
