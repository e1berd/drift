import { describe, expect, it } from 'vitest'

import { creationOptions, fromBase64Url, requestOptions, toBase64Url } from '../webauthn'

const bytes = (s: string) => new TextEncoder().encode(s)

describe('base64url', () => {
    it('round-trips binary data without padding', () => {
        const data = Uint8Array.from([251, 255, 0, 62, 63])
        const encoded = toBase64Url(data.buffer)
        expect(encoded).toBe('-_8APj8')
        expect([...fromBase64Url(encoded)]).toEqual([...data])
    })
})

describe('Kratos passkey options', () => {
    it('decodes settings-flow creation options', () => {
        const raw = JSON.stringify({
            publicKey: {
                rp: { name: 'Drift', id: 'localhost' },
                user: {
                    id: toBase64Url(bytes('user').buffer),
                    name: 'root@example.com',
                    displayName: 'root@example.com',
                },
                challenge: toBase64Url(bytes('challenge').buffer),
                pubKeyCredParams: [{ type: 'public-key', alg: -7 }],
            },
        })
        const options = creationOptions(raw)
        expect(new TextDecoder().decode(options.challenge as Uint8Array)).toBe('challenge')
        expect(new TextDecoder().decode(options.user.id as Uint8Array)).toBe('user')
        expect(options.user.name).toBe('root@example.com')
        expect(options.rp.id).toBe('localhost')
    })

    it('decodes registration-flow creation options with a display name', () => {
        const raw = JSON.stringify({
            displayNameFieldName: 'traits.email',
            credentialOptions: {
                publicKey: {
                    rp: { name: 'Drift' },
                    user: { id: toBase64Url(bytes('u').buffer) },
                    challenge: toBase64Url(bytes('c').buffer),
                    excludeCredentials: [
                        { type: 'public-key', id: toBase64Url(bytes('cred').buffer) },
                    ],
                },
            },
        })
        const options = creationOptions(raw, 'a@b.c')
        expect(options.user.displayName).toBe('a@b.c')
        expect(new TextDecoder().decode(options.excludeCredentials?.at(0)?.id as Uint8Array)).toBe(
            'cred',
        )
    })

    it('decodes login challenges', () => {
        const raw = JSON.stringify({
            publicKey: { challenge: toBase64Url(bytes('c').buffer), rpId: 'localhost' },
        })
        const options = requestOptions(raw)
        expect(new TextDecoder().decode(options.challenge as Uint8Array)).toBe('c')
        expect(options.allowCredentials).toBeUndefined()
    })
})
