import { computed, toValue } from 'vue'
import type { MaybeRefOrGetter } from 'vue'
import { toast } from '@ui-sdk/vue/toast'

import {
    useCancelContinueAfterLimit,
    useContinueAfterLimit,
    useScheduleContinueAfterLimit,
} from '@/gen'
import type { Message } from '@/gen'
import { errorMessage } from '@/shared/errors'

/** Continuing a session whose subscription ran out: now, after the reset, or not at all. */
export function useLimitContinuation(message: MaybeRefOrGetter<Message>) {
    const options = { mutation: { onError: (error: unknown) => toast.error(errorMessage(error)) } }
    const now = useContinueAfterLimit(options)
    const schedule = useScheduleContinueAfterLimit(options)
    const cancel = useCancelContinueAfterLimit(options)

    const path = () => ({ sessionId: toValue(message).sessionId, messageId: toValue(message).id })
    const busy = computed(
        () => now.isPending.value || schedule.isPending.value || cancel.isPending.value,
    )

    return {
        busy,
        continueNow: () => now.mutate({ path: path() }),
        continueAfterReset: () => schedule.mutate({ path: path() }),
        cancel: () => cancel.mutate({ path: path() }),
    }
}
