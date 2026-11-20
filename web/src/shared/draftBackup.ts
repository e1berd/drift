import { readStored, writeStored } from '@/shared/storage'

const STORAGE_KEY = 'drift.unsyncedDrafts'

/** Draft text the server has not confirmed yet; an empty text is a deletion still to be sent. */
export interface UnsyncedDraft {
    text: string
    updateTime: string
}

type UnsyncedDrafts = Record<string, UnsyncedDraft>

export function readUnsyncedDrafts(): UnsyncedDrafts {
    return readStored<UnsyncedDrafts>(STORAGE_KEY, {})
}

export function rememberUnsyncedDraft(key: string, text: string): void {
    writeStored(STORAGE_KEY, {
        ...readUnsyncedDrafts(),
        [key]: { text, updateTime: new Date().toISOString() },
    })
}

export function forgetSyncedDraft(key: string, syncedText: string): void {
    const unsynced = readUnsyncedDrafts()
    if (unsynced[key]?.text !== syncedText) return
    delete unsynced[key]
    writeStored(STORAGE_KEY, unsynced)
}
