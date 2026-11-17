import { useI18n } from 'vue-i18n'

import type { Locale } from '@/shared/preferences'

import { saveLocale } from '@/shared/preferences'

/** Current interface language; switching it also saves the cookie the server renders auth pages with. */
export function useLocale() {
    const { locale } = useI18n({ useScope: 'global' })

    function setLocale(next: Locale) {
        saveLocale(next)
        locale.value = next
        document.documentElement.lang = next
    }

    return { locale, setLocale }
}
