import { storedRef } from '@/shared/storage'

const open = storedRef<boolean>('drift.infobar.open', true)

/** The session info bar beside the chat; its open state survives reloads. */
export function useInfobar() {
    return {
        open,
        toggle: () => {
            open.value = !open.value
        },
    }
}
