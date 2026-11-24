import { useOpenProjects } from '@/composables/useOpenProjects'
import { storedRef } from '@/shared/storage'

export type SidebarTab = 'projects' | 'chats'

const tab = storedRef<SidebarTab>('drift.sidebar.tab', 'projects')

export function useSidebarTab() {
    const { setOpen } = useOpenProjects()

    function revealSessionPlace(projectId: string | undefined) {
        tab.value = projectId ? 'projects' : 'chats'
        if (projectId) setOpen(projectId, true)
    }

    return { tab, revealSessionPlace }
}
