import {
    PhClipboardText,
    PhHammer,
    PhHandPalm,
    PhLightning,
    PhPencilSimpleLine,
    PhWarningOctagon,
} from '@phosphor-icons/vue'

import type { Component } from 'vue'
import type { ModeGlyph } from '@drift/sdk/agents'

export const MODE_ICONS: Record<ModeGlyph, Component> = {
    ask: PhHandPalm,
    edits: PhPencilSimpleLine,
    auto: PhLightning,
    plan: PhClipboardText,
    bypass: PhWarningOctagon,
    build: PhHammer,
}
