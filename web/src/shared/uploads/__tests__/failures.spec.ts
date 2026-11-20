import { describe, expect, it } from 'vitest'

import { isRefusedForGood, uploadFailureStatus } from '../failures'

const failure = (status: number | null) =>
    Object.assign(new Error('upload failed'), {
        originalResponse: status === null ? null : { getStatus: () => status },
    })

describe('upload failures', () => {
    it('reads the status the tus server answered with', () => {
        expect(uploadFailureStatus(failure(413))).toBe(413)
        expect(uploadFailureStatus(failure(null))).toBeNull()
        expect(uploadFailureStatus(new Error('network'))).toBeNull()
    })

    it('gives up only on files the server refuses', () => {
        expect(isRefusedForGood(failure(413))).toBe(true)
        expect(isRefusedForGood(failure(400))).toBe(true)
        expect(isRefusedForGood(failure(507))).toBe(false)
        expect(isRefusedForGood(failure(500))).toBe(false)
        expect(isRefusedForGood(new Error('network'))).toBe(false)
    })
})
