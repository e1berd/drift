import { expect, it } from 'vitest'

import { setupI18n } from '../i18n'
it('serves messages of @drift/vue next to the app ones in both languages', () => {
    const { t } = setupI18n('en').global
    expect(t('drift.realtime.lost')).toBe('Realtime connection lost')
    expect(t('uploads.pause')).toBe('Pause')
    expect(t('drift.uploads.tooMany', { max: 3 })).toContain('3')
    const ru = setupI18n('ru').global
    expect(ru.t('drift.realtime.restored')).not.toBe('drift.realtime.restored')
    expect(ru.t('uploads.pause')).not.toBe('uploads.pause')
})
