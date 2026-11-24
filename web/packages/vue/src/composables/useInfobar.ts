import { computed, ref } from 'vue'

import { useMediaQuery } from './useMediaQuery'
import { storedRef } from '../storedRef'

const docked = storedRef<boolean>('drift.infobar.open', true)
const drawerOpen = ref(false)

/** The session info bar: docked beside the chat on wide screens (kept across reloads), a drawer on narrow ones. */
export function useInfobar() {
    const narrow = useMediaQuery('(max-width: 760px)')
    const open = computed(() => (narrow.value ? drawerOpen.value : docked.value))
    return {
        narrow,
        open,
        docked,
        drawerOpen,
        toggle: () => {
            if (narrow.value) drawerOpen.value = !drawerOpen.value
            else docked.value = !docked.value
        },
        closeDrawer: () => {
            drawerOpen.value = false
        },
    }
}
