import type { GitSummary } from '../api'
import type { InfobarLink } from './useInfobarLinks'

export function activityLinks(summary: GitSummary | undefined): InfobarLink[] {
    return (summary?.activity ?? []).map((entry) => ({
        key: entry.key,
        messageId: entry.messageId,
    }))
}
