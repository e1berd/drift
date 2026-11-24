import { toast } from '@ui-sdk/vue/toast'

import type { ToastDismissDetails } from '@ui-sdk/vue/toast'
import type { NoticeOptions, Notifier } from '@drift/vue/notifier'

function dismissHandler(options: NoticeOptions) {
    return (details: ToastDismissDetails) => {
        if (details.reason === 'action' && options.action?.keepsNoticeOpen) details.cancel()
        else if (details.reason !== 'imperative') options.onUserDismiss?.()
    }
}

function toastOptions(options: NoticeOptions = {}): Parameters<typeof toast.info>[1] {
    const { action, onUserDismiss, ...rest } = options
    return {
        ...rest,
        ...(action && { action: { label: action.label, onAction: action.onAction } }),
        ...((action?.keepsNoticeOpen || onUserDismiss) && { onDismiss: dismissHandler(options) }),
    }
}

export const toastNotifier: Notifier = {
    info: (title, options) => toast.info(title, toastOptions(options)),
    success: (title, options) => toast.success(title, toastOptions(options)),
    error: (title, options) => toast.error(title, toastOptions(options)),
    loading: (title, options) => toast.loading(title, toastOptions(options)),
    describe: (id, description) => toast.update(id, { description }),
    dismiss: (id) => toast.dismiss(id),
}
