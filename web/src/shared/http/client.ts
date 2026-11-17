import { deviceId } from '@/shared/device'
import { ApiError } from '@/shared/http/ApiError'

export { ApiError }

export interface Auth {
    type: string
    scheme?: string
    name?: string
    in?: string
}

export interface RequestConfig {
    method: 'GET' | 'PUT' | 'PATCH' | 'POST' | 'DELETE' | 'OPTIONS' | 'HEAD'
    url: string
    baseURL?: string
    headers?: object
    path?: object
    query?: object
    cookies?: object
    body?: unknown
    signal?: AbortSignal
    keepalive?: boolean
    parseAs?: 'json' | 'text' | 'blob'
    contentType?: string | { request?: string; response?: string }
    throwOnError?: boolean
    security?: Array<Auth | Array<Auth>>
    styles?: object
    client?: typeof client
}

type DataShape = {
    body?: unknown
    cookies?: unknown
    headers?: unknown
    path?: unknown
    query?: unknown
}

export type Options<TData extends DataShape, ThrowOnError extends boolean = true> = Omit<
    RequestConfig,
    keyof DataShape | 'url' | 'method'
> &
    TData & { throwOnError?: ThrowOnError }

type SuccessStatus<TResponses> = Extract<keyof TResponses, `2${string}`>

type SuccessData<TResponses> = {
    [Status in SuccessStatus<TResponses>]: Status extends '204' ? undefined : TResponses[Status]
}[SuccessStatus<TResponses>]

type Success<TResponses> = { data: SuccessData<TResponses>; error: undefined; response: Response }
type Failure = { data: undefined; error: ApiError; response: Response }

export type RequestResult<
    TResponses,
    ThrowOnError extends boolean = true,
> = ThrowOnError extends true ? Success<TResponses> : Success<TResponses> | Failure

export type ResponseErrorConfig<TError = unknown> = ApiError & { body?: TError }

export const API_BASE = '/api'

function expandPath(template: string, values: Record<string, unknown>): string {
    return template.replaceAll(/\{(\w+)\}/g, (_match: string, name: string) =>
        encodeURIComponent(String(values[name])),
    )
}

function buildURL(config: RequestConfig): URL {
    const url = new URL(
        `${config.baseURL ?? API_BASE}${expandPath(config.url, { ...config.path })}`,
        globalThis.location?.origin ?? 'http://localhost',
    )
    for (const [name, value] of Object.entries(config.query ?? {})) {
        for (const item of Array.isArray(value) ? value : [value]) {
            if (item !== undefined && item !== null && item !== '')
                url.searchParams.append(name, String(item))
        }
    }
    return url
}

function buildRequest(config: RequestConfig): Request {
    const headers = new Headers(config.headers as HeadersInit | undefined)
    headers.set('Accept', 'application/json')
    headers.set('X-Drift-Device', deviceId())
    const hasBody = config.body !== undefined
    if (hasBody && !headers.has('Content-Type')) headers.set('Content-Type', 'application/json')
    return new Request(buildURL(config), {
        method: config.method,
        headers,
        body: hasBody ? JSON.stringify(config.body) : null,
        signal: config.signal,
        keepalive: config.keepalive,
        credentials: 'same-origin',
    })
}

async function readBody(response: Response, parseAs: RequestConfig['parseAs']): Promise<unknown> {
    if (response.status === 204 || response.headers.get('Content-Length') === '0') return undefined
    if (parseAs === 'blob') return response.blob()
    if (parseAs === 'text') return response.text()
    const type = response.headers.get('Content-Type') ?? ''
    return type.includes('json') ? response.json() : response.text()
}

function redirectToLogin() {
    const login = new URL('/login', location.origin)
    login.searchParams.set('return_to', location.pathname + location.search)
    location.assign(login)
}

async function send(config: RequestConfig) {
    const response = await fetch(buildRequest(config))
    if (response.status === 401) redirectToLogin()
    const body = await readBody(response, config.parseAs)
    if (response.ok) return { data: body, error: undefined, response }
    const message =
        typeof body === 'object' && body !== null && 'message' in body
            ? String(body.message)
            : response.statusText
    const error = new ApiError(response.status, message, body)
    if (config.throwOnError ?? true) throw error
    return { data: undefined, error, response }
}

/** Transport for every Kubb-generated operation. */
export const client = <TResponses, ThrowOnError extends boolean = true>(config: RequestConfig) =>
    send(config) as Promise<RequestResult<TResponses, ThrowOnError>>
