/// <reference types="vite/client" />
/// <reference types="@intlify/unplugin-vue-i18n/messages" />

import type { Directive } from 'vue'
import type { AutoAnimateOptions } from '@formkit/auto-animate'

declare module 'vue' {
    interface GlobalDirectives {
        vAutoAnimate: Directive<HTMLElement, Partial<AutoAnimateOptions> | undefined>
    }
}
