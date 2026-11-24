import { computed, readonly, ref } from 'vue'

import { kratos, type KratosSession } from '@drift/sdk/kratos'

const session = ref<KratosSession | null>(null)
const loading = ref(false)
const isOwner = computed(() => session.value?.identity.metadata_public?.role === 'superuser')

/** Current Kratos session, shared across the app. */
export function useSession() {
    async function refresh() {
        loading.value = true
        try {
            const { status, body } = await kratos<KratosSession>('/sessions/whoami')
            if (status === 401) {
                window.location.assign('/login')
                return
            }
            session.value = status === 200 ? body : null
        } finally {
            loading.value = false
        }
    }

    return { session: readonly(session), loading: readonly(loading), isOwner, refresh }
}
