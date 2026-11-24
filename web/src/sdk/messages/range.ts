import type { MessageRange, WindowMessage } from './types'

const MESSAGE_ID = /^msg_[0-7][0-9a-hjkmnp-tv-z]{25}$/

export function isMessageId(value: unknown): value is string {
    return typeof value === 'string' && MESSAGE_ID.test(value)
}

export function rangeFromQuery(query: Record<string, unknown>): MessageRange {
    return {
        start: isMessageId(query.start) ? query.start : undefined,
        end: isMessageId(query.end) ? query.end : undefined,
    }
}

export function sameRange(left: MessageRange, right: MessageRange): boolean {
    return (left.start ?? '') === (right.start ?? '') && (left.end ?? '') === (right.end ?? '')
}

export function isPrompt(message: WindowMessage): boolean {
    return (
        message.role === 'user' &&
        (!!message.body.text?.trim() || !!message.body.attachments?.length)
    )
}
