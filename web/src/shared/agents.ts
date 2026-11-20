import claudeLogo from '@lobehub/icons-static-svg/icons/claudecode-color.svg?url'
import codexLogo from '@lobehub/icons-static-svg/icons/codex-color.svg?url'
import antigravityLogo from '@lobehub/icons-static-svg/icons/antigravity-color.svg?url'
import kimiLogo from '@lobehub/icons-static-svg/icons/kimi-color.svg?url'

import mimoLogo from '@/assets/mimo-color.svg?url'

import type { Component } from 'vue'
import {
    PhClipboardText,
    PhHammer,
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
    modes: PermissionModeMeta[]
}

export type PermissionMode = 'default' | 'acceptEdits' | 'auto' | 'plan' | 'bypassPermissions'

export interface PermissionModeMeta {
    id: PermissionMode
    label: string
    short: string
    hint: string
    icon: Component
}

const ASK_MODE: PermissionModeMeta = {
    id: 'default',
    label: 'Ask permissions',
    short: 'Ask',
    hint: 'modes.default',
    icon: PhHandPalm,
}

const ACCEPT_EDITS_MODE: PermissionModeMeta = {
    id: 'acceptEdits',
    label: 'Accept edits',
    short: 'Edits',
    hint: 'modes.acceptEdits',
    icon: PhPencilSimpleLine,
}

const AUTO_MODE: PermissionModeMeta = {
    id: 'auto',
    label: 'Auto mode',
    short: 'Auto',
    hint: 'modes.auto',
    icon: PhLightning,
}

const PLAN_MODE: PermissionModeMeta = {
    id: 'plan',
    label: 'Plan mode',
    short: 'Plan',
    hint: 'modes.plan',
    icon: PhClipboardText,
}

const BYPASS_MODE: PermissionModeMeta = {
    id: 'bypassPermissions',
    label: 'Bypass permissions',
    short: 'Bypass',
    hint: 'modes.bypass',
    icon: PhWarningOctagon,
}

/** Permission modes named as in Claude Code; colours live in `styles/tokens.css` as `--mode-<id>`. */
const PERMISSION_MODES = [ASK_MODE, ACCEPT_EDITS_MODE, AUTO_MODE, PLAN_MODE, BYPASS_MODE]

const KIMI_PROMPT_MODE: PermissionModeMeta = { ...AUTO_MODE, hint: 'modes.kimiPrompt' }

const MIMO_MODES: PermissionModeMeta[] = [
    { id: 'auto', label: 'Build', short: 'Build', hint: 'modes.mimoBuild', icon: PhHammer },
    { ...PLAN_MODE, label: 'Plan' },
]

/** Agents Drift can drive. Models come from each CLI's own settings once agents run. */
export const AGENTS: Record<AgentId, AgentMeta> = {
    claude: {
        id: 'claude',
        name: 'Claude Code',
        cli: 'claude',
        logo: claudeLogo,
        modes: PERMISSION_MODES,
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
        modes: PERMISSION_MODES,
        models: [{ id: 'default', hint: 'models.fromConfig' }],
    },
    antigravity: {
        id: 'antigravity',
        name: 'Antigravity',
        cli: 'agy',
        logo: antigravityLogo,
        modes: PERMISSION_MODES,
        models: [{ id: 'default', hint: 'models.fromConfig' }],
    },
    kimi: {
        id: 'kimi',
        name: 'Kimi Code',
        cli: 'kimi',
        logo: kimiLogo,
        modes: [KIMI_PROMPT_MODE],
        models: [
            { id: 'default', hint: 'models.default' },
            { id: 'kimi-code/k3.1-pro', label: 'K3.1 Pro', hint: 'models.kimiPro' },
            { id: 'kimi-code/k3', label: 'K3', hint: 'models.kimiK3' },
            { id: 'kimi-code/kimi-for-coding', label: 'K2.8', hint: 'models.kimiCoding' },
        ],
    },
    mimo: {
        id: 'mimo',
        name: 'MiMo Code',
        cli: 'mimo',
        logo: mimoLogo,
        modes: MIMO_MODES,
        models: [{ id: 'default', hint: 'models.fromConfig' }],
    },
}
