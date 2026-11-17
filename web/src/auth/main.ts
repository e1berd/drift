import '@/styles/base.css'

import htmx from 'htmx.org'
import 'hyperscript.org'
import '@ui-sdk/web/button'
import '@ui-sdk/web/field'
import '@ui-sdk/web/input'
import '@ui-sdk/web/password-input'

import { passkeyAutofill, passkeyLogin } from '@/shared/webauthn'

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
      passkeyAutofill: typeof passkeyAutofill
    }
  }
}

window.htmx = htmx
window.drift = { passkeyLogin, passkeyAutofill }
