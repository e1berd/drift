import type { GitSummary } from '@drift/vue/api'
import type { InfobarLink } from '@drift/vue/composables/useInfobarLinks'

export function activityLinks(summary: GitSummary | undefined): InfobarLink[] {
    return (summary?.activity ?? []).map((entry) => ({
        key: entry.key,
        messageId: entry.messageId,
    }))
}
