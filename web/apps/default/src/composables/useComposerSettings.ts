import type { AgentId } from '@/shared/agents'
import { AGENTS, type PermissionMode } from '@/shared/agents'
import { storedRef } from '@drift/vue/storedRef'

interface ComposerSettings {
    models: Partial<Record<AgentId, string>>
    sessionModels?: Record<string, string>
    mode: PermissionMode
    agentModes?: Partial<Record<AgentId, PermissionMode>>
}

const state = storedRef<ComposerSettings>('drift.composer', {
    models: {},
    sessionModels: {},
    mode: 'default',
    agentModes: {},
})

const defaultModel = (agent: AgentId) =>
    state.value.models[agent] ?? AGENTS[agent].models.at(0)?.id ?? 'default'

const supportsMode = (agent: AgentId, mode: PermissionMode | undefined) =>
    AGENTS[agent].modes.some((meta) => meta.id === mode)

function modeFor(agent: AgentId): PermissionMode {
    const picked = state.value.agentModes?.[agent]
    if (picked && supportsMode(agent, picked)) return picked
    if (supportsMode(agent, state.value.mode)) return state.value.mode
    return AGENTS[agent].modes.at(0)?.id ?? 'default'
}

/** Model per agent (or per session) and permission mode per agent last picked in any composer on this device. */
export function useComposerSettings() {
    return {
        modelFor: (agent: AgentId, sessionId?: string) =>
            (sessionId && state.value.sessionModels?.[sessionId]) || defaultModel(agent),
        setModel: (agent: AgentId, model: string, sessionId?: string) => {
            if (sessionId) (state.value.sessionModels ??= {})[sessionId] = model
            else state.value.models[agent] = model
        },
        modeFor,
        setMode: (agent: AgentId, mode: PermissionMode) => {
            if (supportsMode(agent, mode)) (state.value.agentModes ??= {})[agent] = mode
        },
    }
}
