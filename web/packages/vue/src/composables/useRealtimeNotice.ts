import { onScopeDispose, watch } from 'vue'
import { useI18n } from 'vue-i18n'

import { useNotifier } from '../notifier'
import { useRealtimeConnection } from './realtimeConnection'

const NOTICE_ID = 'realtime-connection'
const SHOW_AFTER_MS = 2000
const RESTORED_DURATION_MS = 3000

/** Keeps a notice on screen while the realtime connection is down and confirms when it comes back. */
export function useRealtimeNotice() {
    const { t } = useI18n()
    const notifier = useNotifier()
    const { status, browserOnline, reconnectNow } = useRealtimeConnection()
    let showTimer: ReturnType<typeof setTimeout> | undefined
    let shown = false

    function lostDescription() {
        return browserOnline.value
            ? t('drift.realtime.reconnecting')
            : t('drift.realtime.noNetwork')
    }

    function showLost() {
        shown = true
        notifier.dismiss(NOTICE_ID)
        notifier.loading(t('drift.realtime.lost'), {
            id: NOTICE_ID,
            description: lostDescription(),
            duration: Infinity,
            action: {
                label: t('drift.realtime.reconnectNow'),
                onAction: reconnectNow,
                keepsNoticeOpen: true,
            },
            onUserDismiss: () => {
                shown = false
            },
        })
    }

    function showRestored() {
        shown = false
        notifier.dismiss(NOTICE_ID)
        notifier.success(t('drift.realtime.restored'), {
            id: NOTICE_ID,
            duration: RESTORED_DURATION_MS,
        })
    }

    function onStatusChange(next: string) {
        clearTimeout(showTimer)
        if (next === 'lost' && !shown) showTimer = setTimeout(showLost, SHOW_AFTER_MS)
        else if (next === 'open' && shown) showRestored()
    }

    watch(status, onStatusChange)
    watch(browserOnline, () => {
        if (shown) notifier.describe(NOTICE_ID, lostDescription())
    })
    onScopeDispose(() => clearTimeout(showTimer))
}
