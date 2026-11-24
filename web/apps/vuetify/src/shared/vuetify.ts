import { createVuetify } from 'vuetify'
import { aliases, mdi } from 'vuetify/iconsets/mdi-svg'

import { readColorScheme } from '@drift/sdk/colorScheme'

export function createDriftVuetify() {
    return createVuetify({
        icons: { defaultSet: 'mdi', aliases, sets: { mdi } },
        theme: { defaultTheme: readColorScheme() },
        defaults: {
            VBtn: { variant: 'text' },
            VTextField: { variant: 'outlined', density: 'compact' },
            VSelect: { variant: 'outlined', density: 'compact', hideDetails: true },
        },
    })
}
