import { createI18n } from 'vue-i18n'
// Every src/locales/*.yml, compiled at build time by @intlify/unplugin-vue-i18n.
import messages from '@intlify/unplugin-vue-i18n/messages'

import { defaultLocale, detectLocale } from '@drift/sdk/locale'

export function setupI18n(locale = detectLocale()) {
    return createI18n({
        legacy: false,
        locale,
        fallbackLocale: defaultLocale,
        messages,
    })
}
