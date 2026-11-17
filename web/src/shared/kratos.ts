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

/** Calls the Kratos public API with the session cookie and JSON in both directions. */
export async function kratos<T>(
  path: string,
  init: RequestInit = {},
): Promise<{ status: number; body: T }> {
  const url = path.startsWith('http') ? path : `${KRATOS_PUBLIC}${path}`
  const res = await fetch(url, {
    ...init,
    credentials: 'same-origin',
    headers: { Accept: 'application/json', 'Content-Type': 'application/json', ...init.headers },
  })
  const body = (await res.json().catch(() => ({}))) as T & { redirect_browser_to?: string }
  if (
    body.redirect_browser_to &&
    (res.status === 403 || res.status === 410 || res.status === 422)
  ) {
    throw new KratosRedirect(body.redirect_browser_to)
  }
  return { status: res.status, body }
}

/** Returns the value of the named node in a flow. */
export function nodeValue(flow: KratosFlow, name: string): string {
  const value = flow.ui.nodes.find((n) => n.attributes.name === name)?.attributes.value
  return typeof value === 'string' ? value : ''
}
