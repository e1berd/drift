interface TusResponse {
    getStatus(): number
}

/** HTTP status the tus server answered a failed upload with, or null for network and client errors. */
export function uploadFailureStatus(error: unknown): number | null {
    if (typeof error !== 'object' || error === null || !('originalResponse' in error)) return null
    const response = (error as { originalResponse?: TusResponse | null }).originalResponse
    return response ? response.getStatus() : null
}

const REFUSED_FOR_GOOD = [400, 411, 413]

/** Whether retrying cannot help: the server refused the file itself, not the connection. */
export function isRefusedForGood(error: unknown): boolean {
    const status = uploadFailureStatus(error)
    return status !== null && REFUSED_FOR_GOOD.includes(status)
}
