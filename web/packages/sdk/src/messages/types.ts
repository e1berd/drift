export interface WindowMessageBody {
    text?: string
    attachments?: readonly unknown[]
}

export interface WindowMessage {
    id: string
    role: string
    createTime: string
    body: WindowMessageBody
}

export interface MessageRange {
    start?: string
    end?: string
}

export interface WindowPageRequest extends MessageRange {
    pageSize: number
}

export interface WindowPage<Message extends WindowMessage> {
    messages: Message[]
    hasOlder: boolean
    hasNewer: boolean
    lastPrompt?: Message
}
