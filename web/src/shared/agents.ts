import claudeLogo from '@lobehub/icons-static-svg/icons/claudecode-color.svg?url'
import codexLogo from '@lobehub/icons-static-svg/icons/codex-color.svg?url'
import antigravityLogo from '@lobehub/icons-static-svg/icons/antigravity-color.svg?url'

import type { AgentIdKey } from '@/gen'

export type AgentId = AgentIdKey

export interface AgentModel {
    id: string
    label: string
    hint: string
}

export interface AgentMeta {
    id: AgentId
    name: string
    cli: string
    logo: string
    models: AgentModel[]
}

/** Agents Drift can drive. Models come from each CLI's own settings once agents run. */
export const AGENTS: Record<AgentId, AgentMeta> = {
    claude: {
        id: 'claude',
        name: 'Claude Code',
        cli: 'claude',
        logo: claudeLogo,
        models: [
            { id: 'opus', label: 'Opus', hint: 'models.opus' },
            { id: 'sonnet', label: 'Sonnet', hint: 'models.sonnet' },
            { id: 'haiku', label: 'Haiku', hint: 'models.haiku' },
        ],
    },
    codex: {
        id: 'codex',
        name: 'Codex',
        cli: 'codex',
        logo: codexLogo,
        models: [{ id: 'default', label: 'Codex', hint: 'models.fromConfig' }],
    },
    antigravity: {
        id: 'antigravity',
        name: 'Antigravity',
        cli: 'agy',
        logo: antigravityLogo,
        models: [{ id: 'default', label: 'Antigravity', hint: 'models.fromConfig' }],
    },
}

export type PermissionMode = 'default' | 'acceptEdits' | 'auto' | 'plan' | 'bypassPermissions'

/** Permission modes named as in Claude Code. */
export const PERMISSION_MODES: { id: PermissionMode; label: string; hint: string }[] = [
    { id: 'default', label: 'Ask permissions', hint: 'modes.default' },
    { id: 'acceptEdits', label: 'Accept edits', hint: 'modes.acceptEdits' },
    { id: 'auto', label: 'Auto mode', hint: 'modes.auto' },
    { id: 'plan', label: 'Plan mode', hint: 'modes.plan' },
    { id: 'bypassPermissions', label: 'Bypass permissions', hint: 'modes.bypass' },
]
