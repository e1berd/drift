import { computed, ref, toValue, watch } from 'vue'

import type { MaybeRefOrGetter } from 'vue'

import { useCreateSession, useCreateStandaloneSession } from '../api'
import type { CreateSessionResponse, Session } from '../api'
import { useAgents } from '../composables/useAgents'
import { useComposerSettings } from '../composables/useComposerSettings'
import { useErrorNotice } from '../notifier'
import type { AgentId } from '@drift/sdk/agents'
import type { SentPrompt } from '@drift/sdk/attachments'

export interface AgentChoice {
    agent: AgentId
    model: string
}

/**
 * Starting a chat: the agent to talk to (kept on a connected one), and sending the first prompt,
 * which creates a session in `projectId` or a standalone one. `onCreated` opens it.
 */
export function useNewChat(
    projectId: MaybeRefOrGetter<string>,
    onCreated: (session: Session) => unknown,
) {
    const agents = useAgents()
    const settings = useComposerSettings()
    const agent = ref<AgentId>(agents.connected.value.at(0) ?? 'claude')

    watch(agents.connected, (connected) => {
        const first = connected.at(0)
        if (!connected.includes(agent.value) && first) agent.value = first
    })

    const created = {
        mutation: {
            onSuccess: ({ session }: CreateSessionResponse) => {
                settings.setModel(session.agent, settings.modelFor(session.agent), session.id)
                return onCreated(session)
            },
            onError: useErrorNotice(),
        },
    }
    const createInProject = useCreateSession(created)
    const createStandalone = useCreateStandaloneSession(created)
    const creating = computed(
        () => createInProject.isPending.value || createStandalone.isPending.value,
    )

    function selectAgent(next: AgentId) {
        if (agents.isConnected(next)) agent.value = next
    }

    function switchAgent(choice: AgentChoice) {
        settings.setModel(choice.agent, choice.model)
        agent.value = choice.agent
    }

    async function send({ text, attachments }: SentPrompt) {
        if (!agents.isConnected(agent.value))
            throw new Error(`agent ${agent.value} is not connected`)
        const attachmentIds = attachments.map((attachment) => attachment.id)
        if (!text && !attachmentIds.length) return
        const body = {
            agent: agent.value,
            text,
            title: text ? undefined : attachments.at(0)?.name.slice(0, 200),
            attachmentIds,
            model: settings.modelFor(agent.value),
            mode: settings.modeFor(agent.value),
        }
        const project = toValue(projectId)
        if (project) await createInProject.mutateAsync({ path: { projectId: project }, body })
        else await createStandalone.mutateAsync({ body })
    }

    return { agent, agents, creating, selectAgent, switchAgent, send }
}
