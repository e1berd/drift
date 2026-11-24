import { onMounted } from 'vue'
import { useQueryClient } from '@tanstack/vue-query'

import { loadDrafts } from './useDrafts'
import { startEventStream } from './useEventStream'
import { useRealtimeNotice } from './useRealtimeNotice'
import { reportError } from '@drift/sdk/errors'

/** What every interface shell starts once mounted: realtime events, saved drafts, the link notice. */
export function useDriftStartup() {
    const queryClient = useQueryClient()
    useRealtimeNotice()
    onMounted(() => {
        startEventStream(queryClient)
        loadDrafts().catch(reportError)
    })
}
