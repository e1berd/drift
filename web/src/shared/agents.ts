import claudeLogo from '@lobehub/icons-static-svg/icons/claudecode-color.svg?url'
import codexLogo from '@lobehub/icons-static-svg/icons/codex-color.svg?url'
import antigravityLogo from '@lobehub/icons-static-svg/icons/antigravity-color.svg?url'

import type { Component } from 'vue'
import {
    PhClipboardText,
    PhHandPalm,
    PhLightning,
    PhPencilSimpleLine,
    PhWarningOctagon,
} from '@phosphor-icons/vue'

import type { AgentIdKey } from '@/gen'

export type AgentId = AgentIdKey

export interface AgentModel {
    id: string
    /** Absent for the model the agent's CLI picks from its own config. */
    label?: string
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
        models: [{ id: 'default', hint: 'models.fromConfig' }],
    },
    antigravity: {
        id: 'antigravity',
        name: 'Antigravity',
        cli: 'agy',
        logo: antigravityLogo,
        models: [{ id: 'default', hint: 'models.fromConfig' }],
    },
}

export type PermissionMode = 'default' | 'acceptEdits' | 'auto' | 'plan' | 'bypassPermissions'

export interface PermissionModeMeta {
    id: PermissionMode
    label: string
    short: string
    hint: string
    icon: Component
}

/** Permission modes named as in Claude Code; colours live in `styles/tokens.css` as `--mode-<id>`. */
export const PERMISSION_MODES: PermissionModeMeta[] = [
    {
        id: 'default',
        label: 'Ask permissions',
        short: 'Ask',
        hint: 'modes.default',
        icon: PhHandPalm,
    },
    {
        id: 'acceptEdits',
        label: 'Accept edits',
        short: 'Edits',
        hint: 'modes.acceptEdits',
        icon: PhPencilSimpleLine,
    },
    { id: 'auto', label: 'Auto mode', short: 'Auto', hint: 'modes.auto', icon: PhLightning },
    { id: 'plan', label: 'Plan mode', short: 'Plan', hint: 'modes.plan', icon: PhClipboardText },
    {
        id: 'bypassPermissions',
        label: 'Bypass permissions',
        short: 'Bypass',
        hint: 'modes.bypass',
        icon: PhWarningOctagon,
    },
]
