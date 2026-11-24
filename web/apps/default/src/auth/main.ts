import '@/styles/auth.css'

import htmx from 'htmx.org'
import 'hyperscript.org'
import '@ui-sdk/web/button'
import '@ui-sdk/web/field'
import '@ui-sdk/web/input'
import '@ui-sdk/web/password-input'
import '@ui-sdk/web/select'

import { passkeyAutofill, passkeyAvailable, passkeyLogin } from '@drift/sdk/webauthn'
import { bindPreferenceSwitches } from './preferences'

htmx.config.responseHandling = [
    { code: '204', swap: false },
    { code: '[23]..', swap: true },
    { code: '422', swap: true, error: false },
    { code: '[45]..', swap: false, error: true },
]

declare global {
    interface Window {
        htmx: typeof htmx
        drift: {
            passkeyLogin: typeof passkeyLogin
            passkeyAvailable: typeof passkeyAvailable
            passkeyAutofill: typeof passkeyAutofill
        }
    }
}

window.htmx = htmx
window.drift = { passkeyLogin, passkeyAvailable, passkeyAutofill }

bindPreferenceSwitches(document)

// The imports above upgraded every ui-* element; show the page the server kept transparent
// (data-upgrading in app/internal/auth/templates/layout.html) before the next paint.
requestAnimationFrame(() => document.documentElement.removeAttribute('data-upgrading'))
