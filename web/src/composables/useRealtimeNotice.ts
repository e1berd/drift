import { onScopeDispose, watch } from 'vue'
import { toast } from '@ui-sdk/vue/toast'
import { useI18n } from 'vue-i18n'

import type { ToastDismissDetails } from '@ui-sdk/vue/toast'

import { useRealtimeConnection } from '@/composables/realtimeConnection'

const NOTICE_ID = 'realtime-connection'
const SHOW_AFTER_MS = 2000
const RESTORED_DURATION_MS = 3000

/** Keeps a toast on screen while the realtime connection is down and confirms when it comes back. */
export function useRealtimeNotice() {
    const { t } = useI18n()
    const { status, browserOnline, reconnectNow } = useRealtimeConnection()
    let showTimer: ReturnType<typeof setTimeout> | undefined
    let shown = false

    function lostDescription() {
        return browserOnline.value ? t('realtime.reconnecting') : t('realtime.noNetwork')
    }

    function showLost() {
        shown = true
        toast.dismiss(NOTICE_ID)
        toast.loading(t('realtime.lost'), {
            id: NOTICE_ID,
            description: lostDescription(),
            duration: Infinity,
            action: { label: t('realtime.reconnectNow'), onAction: reconnectNow },
            onDismiss: keepOpenAfterReconnectPress,
        })
    }

    function keepOpenAfterReconnectPress(details: ToastDismissDetails) {
        if (details.reason === 'action') details.cancel()
        else if (details.reason !== 'imperative') shown = false
    }

    function showRestored() {
        shown = false
        toast.dismiss(NOTICE_ID)
        toast.success(t('realtime.restored'), { id: NOTICE_ID, duration: RESTORED_DURATION_MS })
    }

    function onStatusChange(next: string) {
        clearTimeout(showTimer)
        if (next === 'lost' && !shown) showTimer = setTimeout(showLost, SHOW_AFTER_MS)
        else if (next === 'open' && shown) showRestored()
    }

    watch(status, onStatusChange)
    watch(browserOnline, () => {
        if (shown) toast.update(NOTICE_ID, { description: lostDescription() })
    })
    onScopeDispose(() => clearTimeout(showTimer))
}
