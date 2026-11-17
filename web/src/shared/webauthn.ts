/**
 * WebAuthn glue for Ory Kratos passkey flows, replacing Kratos' own webauthn.js.
 * Kratos sends credential options as JSON with base64url-encoded binary fields and expects
 * the credential back in the same encoding.
 */

type Json = Record<string, unknown>

interface DescriptorJSON {
  id: string
  type: PublicKeyCredentialType
  transports?: AuthenticatorTransport[]
}

/** Decodes an unpadded base64url string. */
export function fromBase64Url(value: string): Uint8Array<ArrayBuffer> {
  const base64 = value.replace(/-/g, '+').replace(/_/g, '/')
  return Uint8Array.from(atob(base64), (c) => c.charCodeAt(0))
}

/** Encodes bytes as unpadded base64url. */
export function toBase64Url(value: ArrayBuffer | null): string {
  if (!value) return ''
  let binary = ''
  for (const byte of new Uint8Array(value)) binary += String.fromCharCode(byte)
  return btoa(binary).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '')
}

/** Whether the browser can use passkeys at all. */
export function passkeysSupported(): boolean {
  return typeof window !== 'undefined' && 'PublicKeyCredential' in window
}

function decodeDescriptors(list: unknown): PublicKeyCredentialDescriptor[] | undefined {
  if (!Array.isArray(list)) return undefined
  return (list as DescriptorJSON[]).map((d) => ({ ...d, id: fromBase64Url(d.id) }))
}

/** Converts Kratos `passkey_challenge` JSON into options for `navigator.credentials.get`. */
export function requestOptions(raw: string): PublicKeyCredentialRequestOptions {
  const { publicKey } = JSON.parse(raw) as { publicKey: Json }
  return {
    ...(publicKey as unknown as PublicKeyCredentialRequestOptions),
    challenge: fromBase64Url(String(publicKey.challenge)),
    allowCredentials: decodeDescriptors(publicKey.allowCredentials),
  }
}

/**
 * Converts Kratos `passkey_create_data` JSON into options for `navigator.credentials.create`.
 * Settings flows send `{ publicKey }`, registration flows wrap it in `{ credentialOptions }`.
 */
export function creationOptions(
  raw: string,
  displayName?: string,
): PublicKeyCredentialCreationOptions {
  const data = JSON.parse(raw) as { publicKey?: Json; credentialOptions?: { publicKey: Json } }
  const publicKey = data.credentialOptions?.publicKey ?? data.publicKey
  if (!publicKey) throw new Error('passkey_create_data has no publicKey')
  const user = publicKey.user as { id: string; name?: string; displayName?: string }
  const name = displayName || user.name || ''
  return {
    ...(publicKey as unknown as PublicKeyCredentialCreationOptions),
    challenge: fromBase64Url(String(publicKey.challenge)),
    user: {
      id: fromBase64Url(user.id),
      name,
      displayName: displayName || user.displayName || name,
    },
    excludeCredentials: decodeDescriptors(publicKey.excludeCredentials),
  }
}

/** Serialises an assertion the way Kratos expects in `passkey_login`. */
export function assertionJSON(credential: PublicKeyCredential): string {
  const response = credential.response as AuthenticatorAssertionResponse
  return JSON.stringify({
    id: credential.id,
    rawId: toBase64Url(credential.rawId),
    type: credential.type,
    response: {
      authenticatorData: toBase64Url(response.authenticatorData),
      clientDataJSON: toBase64Url(response.clientDataJSON),
      signature: toBase64Url(response.signature),
      userHandle: toBase64Url(response.userHandle),
    },
  })
}

/** Serialises an attestation the way Kratos expects in `passkey_settings_register`. */
export function attestationJSON(credential: PublicKeyCredential): string {
  const response = credential.response as AuthenticatorAttestationResponse
  return JSON.stringify({
    id: credential.id,
    rawId: toBase64Url(credential.rawId),
    type: credential.type,
    response: {
      attestationObject: toBase64Url(response.attestationObject),
      clientDataJSON: toBase64Url(response.clientDataJSON),
    },
  })
}

let conditional: AbortController | undefined

function field(form: HTMLFormElement, name: string): HTMLInputElement {
  const el = form.elements.namedItem(name)
  if (!(el instanceof HTMLInputElement)) throw new Error(`missing form field ${name}`)
  return el
}

function submitPasskey(form: HTMLFormElement, credential: PublicKeyCredential) {
  field(form, 'passkey_login').value = assertionJSON(credential)
  const method = document.createElement('input')
  method.type = 'hidden'
  method.name = 'method'
  method.value = 'passkey'
  form.append(method)
  form.submit()
}

/** Asks the authenticator for a passkey and submits the Kratos login form with it. */
export async function passkeyLogin(form: HTMLFormElement): Promise<void> {
  if (!passkeysSupported()) {
    alert('Этот браузер не поддерживает ключи доступа.')
    return
  }
  conditional?.abort()
  conditional = undefined
  try {
    const credential = await navigator.credentials.get({
      publicKey: requestOptions(field(form, 'passkey_challenge').value),
    })
    if (credential instanceof PublicKeyCredential) submitPasskey(form, credential)
  } catch (error) {
    if (!(error instanceof DOMException && error.name === 'NotAllowedError')) reportError(error)
    void passkeyAutofill(form)
  }
}

/** Offers saved passkeys in the email field's autofill (WebAuthn conditional mediation). */
export async function passkeyAutofill(form: HTMLFormElement): Promise<void> {
  if (!passkeysSupported() || !(await PublicKeyCredential.isConditionalMediationAvailable?.()))
    return
  conditional?.abort()
  conditional = new AbortController()
  try {
    const credential = await navigator.credentials.get({
      mediation: 'conditional',
      signal: conditional.signal,
      publicKey: requestOptions(field(form, 'passkey_challenge').value),
    })
    if (credential instanceof PublicKeyCredential) submitPasskey(form, credential)
  } catch (error) {
    if (
      !(
        error instanceof DOMException &&
        (error.name === 'AbortError' || error.name === 'NotAllowedError')
      )
    ) {
      reportError(error)
    }
  }
}
