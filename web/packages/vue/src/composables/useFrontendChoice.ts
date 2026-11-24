import { computed } from 'vue'
import { useQueryClient } from '@tanstack/vue-query'

import { getSettingsQueryKey, useGetSettings, useListFrontends, useUpdateSettings } from '../api'
import { useErrorNotice } from '../notifier'
import { DEFAULT_FRONTEND, pickFrontend, pickedFrontend } from '@drift/sdk/frontendChoice'

/**
 * Which Drift interface this browser shows, and, for the instance owner, which one everybody
 * gets by default. Switching reloads the page, since the server serves the chosen interface.
 */
export function useFrontendChoice() {
    const queryClient = useQueryClient()
    const list = useListFrontends()
    const settings = useGetSettings()
    const update = useUpdateSettings({
        mutation: {
            onSuccess: (saved) => queryClient.setQueryData(getSettingsQueryKey(), saved),
            onError: useErrorNotice(),
        },
    })

    const frontends = computed(() => list.data.value?.frontends ?? [])
    const instanceDefault = computed(() => settings.data.value?.defaultFrontend ?? DEFAULT_FRONTEND)
    const current = computed(() => {
        const picked = pickedFrontend()
        const known = frontends.value.some((frontend) => frontend.name === picked)
        return picked && known ? picked : instanceDefault.value
    })

    function switchTo(name: string) {
        if (name === current.value) return
        pickFrontend(name)
        window.location.reload()
    }

    function setInstanceDefault(name: string) {
        if (name !== instanceDefault.value) update.mutate({ body: { defaultFrontend: name } })
    }

    return {
        frontends,
        current,
        instanceDefault,
        switchTo,
        setInstanceDefault,
        savingDefault: update.isPending,
    }
}
