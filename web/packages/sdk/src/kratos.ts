/** Kratos public API as exposed on the application origin by the Go server. */
export const KRATOS_PUBLIC = '/.ory/kratos/public'

export interface KratosMessage {
    id: number
    text: string
    type: 'error' | 'info' | 'success'
}

export interface KratosNode {
    type: string
    group: string
    attributes: { name: string; type: string; value?: unknown; node_type: string }
    messages: KratosMessage[]
    meta: { label?: KratosMessage }
}

export interface KratosFlow {
    id: string
    state?: string
    ui: { action: string; method: string; nodes: KratosNode[]; messages?: KratosMessage[] }
}

export interface KratosSession {
    id: string
    identity: {
        id: string
        traits: { email: string; name?: string }
        metadata_public?: { role?: string } | null
    }
}

/** A Kratos response that asks the browser to navigate (re-authentication, expired flow, …). */
export class KratosRedirect extends Error {
    constructor(readonly location: string) {
        super(`redirect to ${location}`)
    }
}

/** Resolves a Kratos public API path, or an absolute flow action URL, on the application origin. */
export function kratosURL(path: string | URL): URL {
    if (path instanceof URL) return path
    return new URL(path.startsWith('http') ? path : `${KRATOS_PUBLIC}${path}`, location.origin)
}

type KratosBody<T> = T & { redirect_browser_to?: string }

async function readJSON<T>(response: Response): Promise<KratosBody<T>> {
    const isJSON = response.headers.get('Content-Type')?.includes('json') ?? false
    return isJSON ? response.json() : ({} as KratosBody<T>)
}

function isRedirectStatus(status: number): boolean {
    return status === 403 || status === 410 || status === 422
}

/** Calls the Kratos public API with the session cookie and JSON in both directions. */
export async function kratos<T>(
    path: string | URL,
    init: RequestInit = {},
): Promise<{ status: number; body: T }> {
    const headers = new Headers(init.headers)
    headers.set('Accept', 'application/json')
    if (init.body && !headers.has('Content-Type')) headers.set('Content-Type', 'application/json')
    const response = await fetch(
        new Request(kratosURL(path), { ...init, headers, credentials: 'same-origin' }),
    )
    const body = await readJSON<T>(response)
    if (body.redirect_browser_to && isRedirectStatus(response.status)) {
        throw new KratosRedirect(body.redirect_browser_to)
    }
    return { status: response.status, body }
}

/** Returns the value of the named node in a flow. */
export function nodeValue(flow: KratosFlow, name: string): string {
    const value = flow.ui.nodes.find((n) => n.attributes.name === name)?.attributes.value
    return typeof value === 'string' ? value : ''
}
