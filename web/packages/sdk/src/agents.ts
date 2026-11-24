import claudeLogo from '@lobehub/icons-static-svg/icons/claudecode-color.svg?url'
import codexLogo from '@lobehub/icons-static-svg/icons/codex-color.svg?url'
import antigravityLogo from '@lobehub/icons-static-svg/icons/antigravity-color.svg?url'
import kimiLogo from '@lobehub/icons-static-svg/icons/kimi-color.svg?url'

import mimoLogo from './assets/mimo-color.svg?url'

import type { AgentIdKey } from './api'

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

/** Which picture a mode wears; each interface maps it to an icon of its own set. */
export type ModeGlyph = 'ask' | 'edits' | 'auto' | 'plan' | 'bypass' | 'build'

export interface PermissionModeMeta {
    id: PermissionMode
    label: string
    short: string
    hint: string
    glyph: ModeGlyph
}

const ASK_MODE: PermissionModeMeta = {
    id: 'default',
    label: 'Ask permissions',
    short: 'Ask',
    hint: 'drift.modes.default',
    glyph: 'ask',
}

const ACCEPT_EDITS_MODE: PermissionModeMeta = {
    id: 'acceptEdits',
    label: 'Accept edits',
    short: 'Edits',
    hint: 'drift.modes.acceptEdits',
    glyph: 'edits',
}

const AUTO_MODE: PermissionModeMeta = {
    id: 'auto',
    label: 'Auto mode',
    short: 'Auto',
    hint: 'drift.modes.auto',
    glyph: 'auto',
}

const PLAN_MODE: PermissionModeMeta = {
    id: 'plan',
    label: 'Plan mode',
    short: 'Plan',
    hint: 'drift.modes.plan',
    glyph: 'plan',
}

const BYPASS_MODE: PermissionModeMeta = {
    id: 'bypassPermissions',
    label: 'Bypass permissions',
    short: 'Bypass',
    hint: 'drift.modes.bypass',
    glyph: 'bypass',
}

/** Permission modes named as in Claude Code; colours live in `styles/tokens.css` as `--mode-<id>`. */
const PERMISSION_MODES = [ASK_MODE, ACCEPT_EDITS_MODE, AUTO_MODE, PLAN_MODE, BYPASS_MODE]

const KIMI_PROMPT_MODE: PermissionModeMeta = { ...AUTO_MODE, hint: 'drift.modes.kimiPrompt' }

const MIMO_MODES: PermissionModeMeta[] = [
    { id: 'auto', label: 'Build', short: 'Build', hint: 'drift.modes.mimoBuild', glyph: 'build' },
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
            { id: 'opus', label: 'Opus', hint: 'drift.models.opus' },
            { id: 'sonnet', label: 'Sonnet', hint: 'drift.models.sonnet' },
            { id: 'haiku', label: 'Haiku', hint: 'drift.models.haiku' },
        ],
    },
    codex: {
        id: 'codex',
        name: 'Codex',
        cli: 'codex',
        logo: codexLogo,
        modes: PERMISSION_MODES,
        models: [{ id: 'default', hint: 'drift.models.fromConfig' }],
    },
    antigravity: {
        id: 'antigravity',
        name: 'Antigravity',
        cli: 'agy',
        logo: antigravityLogo,
        modes: PERMISSION_MODES,
        models: [{ id: 'default', hint: 'drift.models.fromConfig' }],
    },
    kimi: {
        id: 'kimi',
        name: 'Kimi Code',
        cli: 'kimi',
        logo: kimiLogo,
        modes: [KIMI_PROMPT_MODE],
        models: [
            { id: 'default', hint: 'drift.models.default' },
            { id: 'kimi-code/k3.1-pro', label: 'K3.1 Pro', hint: 'drift.models.kimiPro' },
            { id: 'kimi-code/k3', label: 'K3', hint: 'drift.models.kimiK3' },
            { id: 'kimi-code/kimi-for-coding', label: 'K2.8', hint: 'drift.models.kimiCoding' },
        ],
    },
    mimo: {
        id: 'mimo',
        name: 'MiMo Code',
        cli: 'mimo',
        logo: mimoLogo,
        modes: MIMO_MODES,
        models: [{ id: 'default', hint: 'drift.models.fromConfig' }],
    },
}
