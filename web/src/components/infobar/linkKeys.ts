import type { GitSummary } from '@/gen'
import type { InfobarLink } from '@/composables/useInfobarLinks'

export function activityLinks(summary: GitSummary | undefined): InfobarLink[] {
    return (summary?.activity ?? []).map((entry) => ({
        key: entry.key,
        messageId: entry.messageId,
    }))
}
