import {
    mdiAlertOctagonOutline,
    mdiClipboardTextOutline,
    mdiFlashOutline,
    mdiHammerWrench,
    mdiHandBackRightOutline,
    mdiPencilOutline,
} from '@mdi/js'

import type { ModeGlyph } from '@drift/sdk/agents'

export const MODE_ICONS: Record<ModeGlyph, string> = {
    ask: mdiHandBackRightOutline,
    edits: mdiPencilOutline,
    auto: mdiFlashOutline,
    plan: mdiClipboardTextOutline,
    bypass: mdiAlertOctagonOutline,
    build: mdiHammerWrench,
}
