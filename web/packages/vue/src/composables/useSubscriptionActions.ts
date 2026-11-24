import { useQueryClient } from '@tanstack/vue-query'
import { useErrorNotice } from '../notifier'

import {
    listSubscriptionsQueryKey,
    useActivateSubscription,
    useCompleteSubscriptionLogin,
    useCreateSubscription,
    useDeleteSubscription,
    useStartSubscriptionLogin,
} from '../api'

/** Mutations on subscriptions that refresh the list and show failures. */
export function useSubscriptionActions() {
    const showError = useErrorNotice()
    const qc = useQueryClient()
    const mutation = {
        onSuccess: () => qc.invalidateQueries({ queryKey: listSubscriptionsQueryKey() }),
        onError: showError,
    }
    return {
        create: useCreateSubscription({ mutation }),
        activate: useActivateSubscription({ mutation }),
        restart: useStartSubscriptionLogin({ mutation }),
        complete: useCompleteSubscriptionLogin({ mutation }),
        remove: useDeleteSubscription({ mutation }),
    }
}
