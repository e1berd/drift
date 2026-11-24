import { useQueryClient } from '@tanstack/vue-query'
import { toast } from '@ui-sdk/vue/toast'

import {
    listSubscriptionsQueryKey,
    useActivateSubscription,
    useCompleteSubscriptionLogin,
    useCreateSubscription,
    useDeleteSubscription,
    useStartSubscriptionLogin,
} from '@drift/vue/api'
import { errorMessage } from '@drift/sdk/errors'

/** Mutations on subscriptions that refresh the list and toast failures. */
export function useSubscriptionActions() {
    const qc = useQueryClient()
    const mutation = {
        onSuccess: () => qc.invalidateQueries({ queryKey: listSubscriptionsQueryKey() }),
        onError: (error: unknown) => toast.error(errorMessage(error)),
    }
    return {
        create: useCreateSubscription({ mutation }),
        activate: useActivateSubscription({ mutation }),
        restart: useStartSubscriptionLogin({ mutation }),
        complete: useCompleteSubscriptionLogin({ mutation }),
        remove: useDeleteSubscription({ mutation }),
    }
}
