/** An HTTP error from the Drift API, carrying the status and the server's message. */
export class ApiError extends Error {
    constructor(
        readonly status: number,
        message: string,
        readonly body?: unknown,
    ) {
        super(message)
        this.name = 'ApiError'
    }
}
